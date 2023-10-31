package adaptor

import (
	"context"
	"crypto/ecdsa"
	"sync"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/block_sync"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	p2p "github.com/axiomesh/axiom-p2p"
)

var _ rbft.ExternalStack[types.Transaction, *types.Transaction] = (*RBFTAdaptor)(nil)
var _ rbft.Storage = (*RBFTAdaptor)(nil)
var _ rbft.Network = (*RBFTAdaptor)(nil)
var _ rbft.Crypto = (*RBFTAdaptor)(nil)
var _ rbft.ServiceOutbound[types.Transaction, *types.Transaction] = (*RBFTAdaptor)(nil)
var _ rbft.EpochService = (*RBFTAdaptor)(nil)

type RBFTAdaptor struct {
	epochStore        storage.Storage
	store             storage.Storage
	priv              *ecdsa.PrivateKey
	network           network.Network
	msgPipes          map[int32]p2p.Pipe
	globalMsgPipe     p2p.Pipe
	ReadyC            chan *Ready
	BlockC            chan *common.CommitEvent
	logger            logrus.FieldLogger
	getChainMetaFunc  func() *types.ChainMeta
	getBlockFunc      func(uint64) (*types.Block, error)
	StateUpdating     bool
	StateUpdateHeight uint64

	currentSyncHeight uint64
	Cancel            context.CancelFunc
	config            *common.Config
	EpochInfo         *rbft.EpochInfo

	sync           block_sync.Sync
	quitSync       chan struct{}
	broadcastNodes []string
	ctx            context.Context

	lock sync.Mutex
}

type Ready struct {
	Txs             []*types.Transaction
	LocalList       []bool
	Height          uint64
	Timestamp       int64
	ProposerAccount string
}

func NewRBFTAdaptor(config *common.Config) (*RBFTAdaptor, error) {
	store, err := storagemgr.Open(repo.GetStoragePath(config.RepoRoot, storagemgr.Consensus))
	if err != nil {
		return nil, err
	}

	epochStore, err := storagemgr.Open(repo.GetStoragePath(config.RepoRoot, storagemgr.Epoch))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	stack := &RBFTAdaptor{
		epochStore:       epochStore,
		store:            store,
		priv:             config.PrivKey,
		network:          config.Network,
		ReadyC:           make(chan *Ready, 1024),
		BlockC:           make(chan *common.CommitEvent, 1024),
		quitSync:         make(chan struct{}, 1),
		logger:           config.Logger,
		getChainMetaFunc: config.GetChainMetaFunc,
		getBlockFunc:     config.GetBlockFunc,
		config:           config,

		sync: config.BlockSync,

		ctx:    ctx,
		Cancel: cancel,
	}

	return stack, nil
}

func (a *RBFTAdaptor) UpdateEpoch() error {
	e, err := a.config.GetCurrentEpochInfoFromEpochMgrContractFunc()
	if err != nil {
		return err
	}
	a.EpochInfo = e
	a.broadcastNodes = lo.Map(lo.Flatten([][]*rbft.NodeInfo{a.EpochInfo.ValidatorSet, a.EpochInfo.CandidateSet}), func(item *rbft.NodeInfo, index int) string {
		return item.P2PNodeID
	})
	return nil
}

func (a *RBFTAdaptor) SetMsgPipes(msgPipes map[int32]p2p.Pipe, globalMsgPipe p2p.Pipe) {
	a.msgPipes = msgPipes
	a.globalMsgPipe = globalMsgPipe
}
