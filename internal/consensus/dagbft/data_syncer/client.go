package data_syncer

import (
	"context"

	consensus_types "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

type client struct {
	*rpc.Client
	eventName string
	ctx       context.Context
	logger    logrus.FieldLogger
}

func newClient(bindAddr string, ctx context.Context) (*client, error) {
	cli, err := rpc.DialContext(ctx, "ws://"+bindAddr)
	if err != nil {
		return nil, err
	}

	return &client{
		Client:    cli,
		eventName: "newBlockAndProofs",
		ctx:       ctx,
	}, nil

}

func (c *client) listenNewBlock(newBlockCh chan *consensus_types.AttestationAndBlock) {
	resultChan := make(chan *consensus_types.AttestationAndBlock)
	sub, err := c.EthSubscribe(c.ctx, resultChan, c.eventName)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.Client.Close()
				return
			case result := <-resultChan:
				newBlockCh <- result
			case <-sub.Err():
				c.Client.Close()
				return
			}
		}
	}()
}
