package attestation

import (
	"context"
	"sync"
	"time"

	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

type AttestationAPI struct {
	ctx       context.Context
	cancel    context.CancelFunc
	rep       *repo.Repo
	api       api.CoreAPI
	events    *EventSystem
	filters   map[rpc.ID]*filter
	filtersMu sync.Mutex

	logger  logrus.FieldLogger
	timeout time.Duration
}

// filter is a helper struct that holds meta information over the filter type
// and associated subscription in the event system.
type filter struct {
	typ         Type
	deadline    *time.Timer // filter is inactiv when deadline triggers
	attestation *consensustypes.Attestation
	s           *Subscription // associated subscription in event system
}

func NewAttestationAPI(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *AttestationAPI {
	ctx, cancel := context.WithCancel(context.Background())
	return &AttestationAPI{
		ctx:     ctx,
		cancel:  cancel,
		rep:     rep,
		api:     api,
		events:  NewEventSystem(api),
		filters: make(map[rpc.ID]*filter),
		logger:  logger,
		timeout: 10 * time.Second,
	}
}

func (api *AttestationAPI) timeoutLoop() {
	var toUninstall []*Subscription
	ticker := time.NewTicker(api.timeout)
	defer ticker.Stop()
	for {
		<-ticker.C
		api.filtersMu.Lock()
		for id, f := range api.filters {
			select {
			case <-f.deadline.C:
				toUninstall = append(toUninstall, f.s)
				delete(api.filters, id)
			default:
				continue
			}
		}
		api.filtersMu.Unlock()

		// Unsubscribes are processed outside the lock to avoid the following scenario:
		// event loop attempts broadcasting events to still active filters while
		// Unsubscribe is waiting for it to process the uninstall request.
		for _, s := range toUninstall {
			s.Unsubscribe()
		}
		toUninstall = nil
	}
}

// NewAttestationsFilter creates a filter that fetches attestations that are imported into the chain.
func (api *AttestationAPI) NewAttestationsFilter() rpc.ID {
	var (
		at     = make(chan *consensustypes.Attestation)
		attSub = api.events.SubscribeNewAttestations(at)
	)

	api.filtersMu.Lock()
	api.filters[attSub.ID] = &filter{typ: AttestationsSubscription, deadline: time.NewTimer(api.timeout), attestation: &consensustypes.Attestation{}, s: attSub}
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case a := <-at:
				api.filtersMu.Lock()
				if f, found := api.filters[attSub.ID]; found {
					f.attestation = a
				}
				api.filtersMu.Unlock()
			case <-attSub.Err():
				api.filtersMu.Lock()
				delete(api.filters, attSub.ID)
				api.filtersMu.Unlock()
				return
			}
		}
	}()

	return attSub.ID
}

// NewAttestations send a notification each time a new attestation is appended to the chain.
func (api *AttestationAPI) NewAttestations(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()

	go func() {
		attestations := make(chan *consensustypes.Attestation)
		headersSub := api.events.SubscribeNewAttestations(attestations)

		for {
			select {
			case a := <-attestations:
				err := notifier.Notify(rpcSub.ID, a)
				if err != nil {
					api.logger.Warn("notifier notify error", err)
				}
			case <-rpcSub.Err():
				headersSub.Unsubscribe()
				return
			case <-notifier.Closed():
				headersSub.Unsubscribe()
				return
			}
		}
	}()

	return rpcSub, nil
}
