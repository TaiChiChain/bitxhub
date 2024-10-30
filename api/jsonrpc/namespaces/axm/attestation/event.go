package attestation

import (
	"sync"
	"time"

	consensusTypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type EventSystem struct {
	api            api.CoreAPI
	attestationSub event.Subscription // Subscription for new attestation event
	attestationCh  chan events.AttestationEvent
	install        chan *subscription
	uninstall      chan *subscription
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(api api.CoreAPI) *EventSystem {
	m := &EventSystem{
		attestationCh: make(chan events.AttestationEvent, 10),
		api:           api,
		install:       make(chan *subscription),
		uninstall:     make(chan *subscription),
	}

	// Subscribe events
	m.attestationSub = m.api.Feed().SubscribeNewAttestationEvent(m.attestationCh)
	// m.pendingLogsSub = m.api.SubscribePendingLogsEvent(m.pendingLogsCh)

	// Make sure none of the subscriptions are empty
	if m.attestationSub == nil {
		log.Crit("Subscribe for event system failed")
	}

	go m.eventLoop()
	return m
}

type Type byte

const (
	UnknownSubscription Type = iota
	AttestationsSubscription
	LastIndexSubscription
)

type subscription struct {
	id          rpc.ID
	typ         Type
	created     time.Time
	attestation chan *consensusTypes.Attestation
	installed   chan struct{} // closed when the filter is installed
	err         chan error    // closed when the filter is uninstalled
}

// Subscription is created when the client registers itself for a particular event.
type Subscription struct {
	ID        rpc.ID
	f         *subscription
	es        *EventSystem
	unsubOnce sync.Once
}

// Err returns a channel that is closed when unsubscribed.
func (sub *Subscription) Err() <-chan error {
	return sub.f.err
}

// Unsubscribe uninstalls the subscription from the event broadcast loop.
func (sub *Subscription) Unsubscribe() {
	sub.unsubOnce.Do(func() {
	uninstallLoop:
		for {
			// write uninstall request and consume logs/txs. This prevents
			// the eventLoop broadcast method to deadlock when writing to the
			// filter event channel while the subscription loop is waiting for
			// this method to return (and thus not reading these events).
			select {
			case sub.es.uninstall <- sub.f:
				break uninstallLoop
			case <-sub.f.attestation:
			}
		}

		// wait for filter to be uninstalled in work loop before returning
		// this ensures that the manager won't use the event channel which
		// will probably be closed by the client asap after this method returns.
		<-sub.Err()
	})
}

// subscribe installs the subscription in the event broadcast loop.
func (es *EventSystem) subscribe(sub *subscription) *Subscription {
	es.install <- sub
	<-sub.installed
	return &Subscription{ID: sub.id, f: sub, es: es}
}

func (es *EventSystem) SubscribeNewAttestations(attestations chan *consensusTypes.Attestation) *Subscription {
	sub := &subscription{
		id:          rpc.NewID(),
		typ:         AttestationsSubscription,
		created:     time.Now(),
		attestation: attestations,
		installed:   make(chan struct{}),
		err:         make(chan error),
	}
	return es.subscribe(sub)
}

type filterIndex map[Type]map[rpc.ID]*subscription

func (es *EventSystem) handleAttestationEvent(filters filterIndex, ev events.AttestationEvent) {
	for _, f := range filters[AttestationsSubscription] {
		f.attestation <- ev.AttestationData
	}
}

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	// Ensure all subscriptions get cleaned up
	defer func() {
		es.attestationSub.Unsubscribe()
	}()

	index := make(filterIndex)
	for i := UnknownSubscription; i < LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*subscription)
	}

	for {
		select {
		case f := <-es.install:
			index[f.typ][f.id] = f
			close(f.installed)

		case ev := <-es.attestationCh:
			es.handleAttestationEvent(index, ev)
		case f := <-es.uninstall:
			delete(index[f.typ], f.id)
			close(f.err)

		// System stopped
		case <-es.attestationSub.Err():
		}
	}
}
