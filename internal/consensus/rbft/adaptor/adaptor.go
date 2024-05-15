package adaptor

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/event"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	sync_comm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	p2p "github.com/axiomesh/axiom-p2p"
)

var _ rbft.ExternalStack[types.Transaction, *types.Transaction] = (*RBFTAdaptor)(nil)
var _ rbft.Storage = (*RBFTAdaptor)(nil)
var _ rbft.Network = (*RBFTAdaptor)(nil)
var _ rbft.Crypto = (*RBFTAdaptor)(nil)
var _ rbft.ServiceOutbound[types.Transaction, *types.Transaction] = (*RBFTAdaptor)(nil)
var _ rbft.EpochService = (*RBFTAdaptor)(nil)
var _ rbft.Ledger = (*RBFTAdaptor)(nil)

type RBFTAdaptor struct {
	epochStore            storage.Storage
	store                 rbft.Storage
	libp2pKey             libp2pcrypto.PrivKey
	p2pID2PubKeyCache     map[string]libp2pcrypto.PubKey
	p2pID2PubKeyCacheLock *sync.RWMutex
	network               network.Network
	msgPipes              map[int32]p2p.Pipe
	ReadyC                chan *Ready
	BlockC                chan *common.CommitEvent
	logger                logrus.FieldLogger
	getChainMetaFunc      func() *types.ChainMeta
	getBlockHeaderFunc    func(uint64) (*types.BlockHeader, error)
	StateUpdating         bool
	StateUpdateHeight     uint64

	currentSyncHeight uint64
	Cancel            context.CancelFunc
	config            *common.Config
	EpochInfo         *rbft.EpochInfo

	sync           sync_comm.Sync
	quitSync       chan struct{}
	broadcastNodes []string
	ctx            context.Context

	MockBlockFeed event.Feed
}

type Ready struct {
	Txs             []*types.Transaction
	LocalList       []bool
	BatchDigest     string
	Height          uint64
	Timestamp       int64
	ProposerAccount string
	ProposerNodeID  uint64
}

func NewRBFTAdaptor(config *common.Config) (*RBFTAdaptor, error) {
	var err error
	storePath := repo.GetStoragePath(config.RepoRoot, storagemgr.Consensus)
	var store rbft.Storage
	switch config.ConsensusStorageType {
	case repo.ConsensusStorageTypeMinifile:
		store, err = OpenMinifile(storePath)
	case repo.ConsensusStorageTypeRosedb:
		store, err = OpenRosedb(storePath)
	default:
		return nil, errors.Errorf("unsupported consensus storage type: %s", config.ConsensusStorageType)
	}
	if err != nil {
		return nil, err
	}

	libp2pKey, err := repo.Libp2pKeyFromECDSAKey(config.PrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ecdsa p2pKey: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stack := &RBFTAdaptor{
		epochStore:            config.EpochStore,
		store:                 store,
		libp2pKey:             libp2pKey,
		p2pID2PubKeyCache:     make(map[string]libp2pcrypto.PubKey),
		p2pID2PubKeyCacheLock: new(sync.RWMutex),
		network:               config.Network,
		ReadyC:                make(chan *Ready, 1024),
		BlockC:                make(chan *common.CommitEvent, 1024),
		quitSync:              make(chan struct{}, 1),
		logger:                config.Logger,
		getChainMetaFunc:      config.GetChainMetaFunc,
		getBlockHeaderFunc:    config.GetBlockHeaderFunc,
		config:                config,

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
	a.broadcastNodes = lo.Map(lo.Flatten([][]rbft.NodeInfo{a.EpochInfo.ValidatorSet, a.EpochInfo.CandidateSet, a.EpochInfo.DataSyncerSet}), func(item rbft.NodeInfo, index int) string {
		return item.P2PNodeID
	})
	return nil
}

func (a *RBFTAdaptor) SetMsgPipes(msgPipes map[int32]p2p.Pipe) {
	a.msgPipes = msgPipes
}
