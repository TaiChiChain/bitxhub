package data_syncer

import (
	"context"
	"math"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	consensus_types "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/sirupsen/logrus"
)

type Node struct {
	client        *client
	bindNode      string
	ctx           context.Context
	cancel        context.CancelFunc
	logger        logrus.FieldLogger
	recvBlock     chan *consensus_types.AttestationAndBlock
	chainState    *chainstate.ChainState
	currentHeight uint64
	verifier      *verifier

	recvBlockCache *blockCache
}

func (n *Node) Start() error {
	cli, err := newClient(n.bindNode, n.ctx)
	if err != nil {
		return err
	}
	n.client = cli
	n.client.listenNewBlock(n.recvBlock)
	return nil
}

func (n *Node) ProcessEvent() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case b := <-n.recvBlock:

		}
	}

}

func (n *Node) handleNewBlock(b *types.Block) {
	if b.Height() < n.currentHeight {
		n.logger.Warningf("receive new block height %d is less than current height %d, just ignore it...",
			b.Height(), n.currentHeight)
		return
	}

	n.verifier.
	if b.Height() > n.currentHeight {
	}
}

func (n *Node) Stop() {
	n.cancel()
}

func (bc *blockCache) pushBlock(b *types.Block) {
	bc.blockM[b.Height()] = b
	bc.heightIndex.Push(b.Height())
}

func (n *Node) popBlock() []*types.Block {
	matched := make([]*types.Block, 0)
	for h := n.recvBlockCache.heightIndex.PeekItem(); h != math.MaxUint64 && h == n.currentHeight; n.recvBlockCache.heightIndex.Pop() {
		block, ok := n.recvBlockCache.blockM[h]
		if !ok {
			n.logger.Errorf("block %d not found in cache, but exists in heightIndex", h)
			return nil
		}
		if block.Header.Epoch != n.chainState.EpochInfo.Epoch {
			n.logger.Infof("block %d epoch %d not match current epoch %d, need wait for next epoch", h, block.Header.Epoch, n.chainState.EpochInfo.Epoch)
			return matched
		}
		matched = append(matched, n.recvBlockCache.blockM[h])
		delete(n.recvBlockCache.blockM, h)
	}
	return matched
}
