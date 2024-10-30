package data_syncer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc"
	consensus_types "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/concurrency"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

type client struct {
	*rpc.Client
	eventName            string
	sendNewAttestationCh chan<- *consensus_types.Attestation
	recvCloseC           chan struct{}
	ctx                  context.Context
	cancel               context.CancelFunc
	logger               logrus.FieldLogger
}

func newClient(bindAddr string, closeCh chan struct{}, sendCh chan<- *consensus_types.Attestation) (*client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cli, err := rpc.DialContext(ctx, "ws://"+bindAddr)
	if err != nil {
		cancel()
		return nil, err
	}

	return &client{
		Client:               cli,
		eventName:            "newAttestations",
		recvCloseC:           closeCh,
		sendNewAttestationCh: sendCh,
		ctx:                  ctx,
		cancel:               cancel,
	}, nil

}

func (c *client) start(wg *sync.WaitGroup, ap *concurrency.AsyncPool) error {
	return ap.AsyncDo(func() {
		if err := retry.Retry(func(attempt uint) error {
			return c.startTask()
		}, strategy.Limit(common.MaxRetryCount), strategy.Wait(30*time.Second)); err != nil {
			panic(fmt.Errorf("retry subscribe bind node failed: %v", err))
		}
	}, wg)
}

func (c *client) startTask() error {
	resultChan := make(chan *consensus_types.Attestation)
	sub, err := c.Subscribe(c.ctx, jsonrpc.AxmNamespace, resultChan, c.eventName)
	if err != nil {
		return err
	}
	for {
		select {
		case <-c.recvCloseC:
			c.cancel()
			c.Client.Close()
			return nil
		case result := <-resultChan:
			channel.SafeSend(c.sendNewAttestationCh, result, c.ctx.Done())
		case err = <-sub.Err():
			c.cancel()
			c.Client.Close()
			sub.Unsubscribe()
			return fmt.Errorf("failed to subscribe %s: %w", c.eventName, err)
		}
	}
}
