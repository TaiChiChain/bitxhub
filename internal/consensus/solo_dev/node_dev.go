package solo_dev

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const checkpoint = 10

type GetAccountNonceFunc func(address *types.Address) uint64

func init() {
	repo.Register(repo.ConsensusTypeSoloDev, false)
}

type NodeDev struct {
	config          *common.Config
	proposerAccount string
	persistDoneC    chan struct{}            // signal of tx had been persisted
	commitC         chan *common.CommitEvent // block channel
	lastExec        uint64                   // the index of the last-applied block
	mutex           sync.Mutex
	logger          logrus.FieldLogger // logger
	GetAccountNonce GetAccountNonceFunc
	txFeed          event.Feed
	mockBlockFeed   event.Feed
}

func NewNode(config *common.Config) (*NodeDev, error) {
	proposerAccount, err := repo.KeyToNodeID(config.PrivKey)
	if err != nil {
		return nil, err
	}

	return &NodeDev{
		config:          config,
		proposerAccount: proposerAccount,
		persistDoneC:    make(chan struct{}),
		commitC:         make(chan *common.CommitEvent),
		lastExec:        config.Applied,
		logger:          config.Logger,
		GetAccountNonce: config.GetAccountNonce,
	}, nil
}

func (n *NodeDev) Start() error {
	n.logger.Info("consensus dev started")
	return nil
}

func (n *NodeDev) Stop() {
	n.logger.Info("consensus dev stopped")
}

func (n *NodeDev) Prepare(tx *types.Transaction) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	block := &types.Block{
		BlockHeader: &types.BlockHeader{
			Epoch:           1,
			Number:          n.lastExec + 1,
			Timestamp:       time.Now().Unix(),
			ProposerAccount: n.proposerAccount,
		},
		Transactions: []*types.Transaction{tx},
	}
	n.commitC <- &common.CommitEvent{
		Block: block,
	}
	n.lastExec++
	// ensure this tx had been persist
	<-n.persistDoneC
	return nil
}

func (n *NodeDev) SubmitTxsFromRemote(_ [][]byte) error {
	return nil
}

func (n *NodeDev) Commit() chan *common.CommitEvent {
	return n.commitC
}

func (n *NodeDev) Step(_ []byte) error {
	return nil
}

func (n *NodeDev) Ready() error {
	return nil
}

func (n *NodeDev) ReportState(height uint64, blockHash *types.Hash, txPointerList []*events.TxPointer, _ *consensus.Checkpoint, _ bool) {
	if height%checkpoint == 0 {
		n.logger.WithFields(logrus.Fields{
			"height": height,
			"hash":   blockHash,
			"txs":    txPointerList,
		}).Info("Report checkpoint")
	}
	n.logger.Debugf("ReportState", height, blockHash, txPointerList)
	n.persistDoneC <- struct{}{}
}

func (n *NodeDev) Quorum(_ uint64) uint64 {
	return 1
}

func (n *NodeDev) GetPendingTxCountByAccount(account string) uint64 {
	nonce := n.GetAccountNonce(types.NewAddressByStr(account))
	return nonce
}

func (n *NodeDev) GetPendingTxByHash(_ *types.Hash) *types.Transaction {
	return nil
}

func (n *NodeDev) DelNode(_ uint64) error {
	return nil
}

func (n *NodeDev) SubscribeTxEvent(events chan<- []*types.Transaction) event.Subscription {
	return n.txFeed.Subscribe(events)
}

func (n *NodeDev) SubscribeMockBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return n.mockBlockFeed.Subscribe(ch)
}

func (n *NodeDev) GetTotalPendingTxCount() uint64 {
	return 0
}

func (n *NodeDev) GetLowWatermark() uint64 {
	return n.lastExec
}

func (n *NodeDev) GetAccountPoolMeta(account string, full bool) *common.AccountMeta {
	return nil
}

func (n *NodeDev) GetPoolMeta(full bool) *common.Meta {
	return nil
}
