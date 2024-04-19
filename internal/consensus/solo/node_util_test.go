package solo

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/txpool/mock_txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck/mock_precheck"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	batchTimeout     = 50 * time.Millisecond
	noTxBatchTimeout = 100 * time.Millisecond
)

func mockSoloNode(t *testing.T, enableTimed bool) (*Node, error) {
	logger := log.NewWithModule("consensus")
	logger.Logger.SetLevel(logrus.DebugLevel)
	rep := repo.MockRepo(t)

	recvCh := make(chan consensusEvent, maxChanSize)
	mockCtl := gomock.NewController(t)
	mockNetwork := mock_network.NewMockNetwork(mockCtl)

	ctx, cancel := context.WithCancel(context.Background())

	mockPool := mock_txpool.NewMockMinimalTxPool[types.Transaction, *types.Transaction](500, mockCtl)
	mockPrecheck := mock_precheck.NewMockMinPreCheck(mockCtl, mockPool)

	soloNode := &Node{
		config: &common.Config{
			Repo:       rep,
			ChainState: chainstate.NewMockChainState(rep.GenesisConfig, nil),
		},
		lastExec:     uint64(0),
		commitC:      make(chan *common.CommitEvent, maxChanSize),
		blockCh:      make(chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction], maxChanSize),
		txpool:       mockPool,
		network:      mockNetwork,
		batchDigestM: make(map[uint64]string),
		recvCh:       recvCh,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		txPreCheck:   mockPrecheck,
		epcCnf: &epochConfig{
			epochPeriod:         rep.GenesisConfig.EpochInfo.EpochPeriod,
			startBlock:          rep.GenesisConfig.EpochInfo.StartBlock,
			checkpoint:          rep.GenesisConfig.EpochInfo.ConsensusParams.CheckpointPeriod,
			enableGenEmptyBlock: enableTimed,
		},
	}
	batchTimerMgr := timer.NewTimerManager(logger)
	err := batchTimerMgr.CreateTimer(common.Batch, batchTimeout, soloNode.handleTimeoutEvent)
	require.Nil(t, err)
	err = batchTimerMgr.CreateTimer(common.NoTxBatch, noTxBatchTimeout, soloNode.handleTimeoutEvent)
	require.Nil(t, err)
	soloNode.batchMgr = &batchTimerManager{Timer: batchTimerMgr}
	return soloNode, nil
}
