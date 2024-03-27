package solo

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/axiomesh/axiom-ledger/internal/components/timer"
	"github.com/axiomesh/axiom-ledger/internal/consensus/rbft/testutil"
	txpoolImpl "github.com/axiomesh/axiom-ledger/internal/txpool"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck/mock_precheck"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNode_Start(t *testing.T) {
	repoRoot := t.TempDir()
	r, err := repo.Load(repo.DefaultKeyJsonPassword, repoRoot, true)
	require.Nil(t, err)

	mockCtl := gomock.NewController(t)
	mockNetwork := mock_network.NewMockNetwork(mockCtl)
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	config, err := common.GenerateConfig(
		common.WithPrivKey(s.Sk),
		common.WithConfig(r.RepoRoot, r.ConsensusConfig),
		common.WithGenesisEpochInfo(r.GenesisConfig.EpochInfo),
		common.WithLogger(log.NewWithModule("consensus")),
		common.WithApplied(1),
		common.WithNetwork(mockNetwork),
		common.WithApplied(1),
		common.WithGetAccountNonceFunc(func(address *types.Address) uint64 {
			return 0
		}),
		common.WithGetAccountBalanceFunc(func(address string) *big.Int {
			maxGasPrice := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e9))
			return new(big.Int).Mul(big.NewInt(math.MaxInt64), maxGasPrice)
		}),
		common.WithGetChainMetaFunc(func() *types.ChainMeta {
			return &types.ChainMeta{
				GasPrice: big.NewInt(0),
			}
		}),
		common.WithTxPool(mockTxPool(t)),
		common.WithGetCurrentEpochInfoFromEpochMgrContractFunc(func() (*rbft.EpochInfo, error) {
			return r.GenesisConfig.EpochInfo, nil
		}),
	)
	require.Nil(t, err)

	solo, err := NewNode(config)
	require.Nil(t, err)

	err = solo.Start()
	require.Nil(t, err)

	err = solo.SubmitTxsFromRemote(nil)
	require.Nil(t, err)

	var msg []byte
	require.Nil(t, solo.Step(msg))
	require.Equal(t, uint64(1), solo.Quorum(1))

	tx, err := types.GenerateEmptyTransactionAndSigner()
	require.Nil(t, err)

	for {
		time.Sleep(200 * time.Millisecond)
		err := solo.Ready()
		if err == nil {
			break
		}
	}

	txSubscribeCh := make(chan []*types.Transaction, 1)
	sub := solo.SubscribeTxEvent(txSubscribeCh)
	defer sub.Unsubscribe()

	txSubscribeMockCh := make(chan events.ExecutedEvent, 1)
	mockSub := solo.SubscribeMockBlockEvent(txSubscribeMockCh)
	defer mockSub.Unsubscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockAddTx(solo, ctx)

	err = solo.Prepare(tx)
	require.Nil(t, err)
	subTxs := <-txSubscribeCh
	require.EqualValues(t, 1, len(subTxs))
	require.EqualValues(t, tx, subTxs[0])

	solo.notifyGenerateBatch(txpool.GenBatchSizeEvent)

	commitEvent := <-solo.Commit()
	require.Equal(t, uint64(2), commitEvent.Block.Header.Number)
	require.Equal(t, 1, len(commitEvent.Block.Transactions))
	blockHash := commitEvent.Block.Hash()

	txPointerList := make([]*events.TxPointer, 0)
	txPointerList = append(txPointerList, &events.TxPointer{Hash: tx.GetHash(), Account: tx.RbftGetFrom(), Nonce: tx.RbftGetNonce()})
	solo.ReportState(commitEvent.Block.Height(), blockHash, txPointerList, nil, false)
	solo.Stop()
}

func TestTimedBlock(t *testing.T) {
	ast := assert.New(t)
	node, err := mockSoloNode(t, true)
	ast.Nil(err)
	defer node.Stop()

	err = node.Start()
	ast.Nil(err)
	defer node.Stop()

	event1 := <-node.commitC
	ast.NotNil(event1)
	ast.Equal(len(event1.Block.Transactions), 0)
	ast.Equal(event1.Block.Header.Number, uint64(1))

	event2 := <-node.commitC
	ast.NotNil(event2)
	ast.Equal(len(event2.Block.Transactions), 0)
	ast.Equal(event2.Block.Header.Number, uint64(2))
}

func TestNode_ReportState(t *testing.T) {
	t.Run("test report state", func(t *testing.T) {
		ast := assert.New(t)
		node, err := mockSoloNode(t, false)
		ast.Nil(err)

		err = node.Start()
		ast.Nil(err)
		defer node.Stop()
		node.batchDigestM[10] = "test"
		node.ReportState(10, types.NewHashByStr("0x123"), []*events.TxPointer{}, nil, false)
		// ensure last event(report state) had been processed
		node.GetLowWatermark()
		ast.Equal(0, len(node.batchDigestM))

		txList, signer := prepareMultiTx(t, 10)
		ast.Equal(10, len(txList))

		txSubscribeCh := make(chan []*types.Transaction, 1)
		sub := node.SubscribeTxEvent(txSubscribeCh)
		defer sub.Unsubscribe()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mockAddTx(node, ctx)

		for _, tx := range txList {
			err = node.Prepare(tx)
			ast.Nil(err)
			<-txSubscribeCh
			// sleep to make sure the tx is generated to the batch
			time.Sleep(batchTimeout + 10*time.Millisecond)
		}

		// test pool full
		ast.Equal(10, len(node.batchDigestM))
		tx11, err := types.GenerateTransactionWithSigner(uint64(11),
			types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(0), nil, signer)

		ast.Nil(err)
		err = node.Prepare(tx11)
		ast.NotNil(err)
		<-txSubscribeCh
		ast.Contains(err.Error(), txpoolImpl.ErrTxPoolFull.Error())
		ast.Equal(10, len(node.batchDigestM), "the pool should be full, tx11 is not add in txpool successfully")

		ast.NotNil(node.txpool.GetPendingTxByHash(txList[9].RbftGetTxHash()), "tx10 should be in txpool")
		// trigger the report state
		pointer9 := &events.TxPointer{
			Hash:    txList[9].GetHash(),
			Account: txList[9].RbftGetFrom(),
			Nonce:   txList[9].RbftGetNonce(),
		}
		node.getCurrentEpochInfoFunc = func() (*rbft.EpochInfo, error) {
			return &rbft.EpochInfo{Epoch: 2, StartBlock: 10, EpochPeriod: 10, ConsensusParams: rbft.ConsensusParams{EnableTimedGenEmptyBlock: true}}, nil
		}
		node.epcCnf.epochPeriod = 10
		node.ReportState(10, types.NewHashByStr("0x123"), []*events.TxPointer{pointer9}, nil, false)
		node.GetLowWatermark()
		ast.Nil(node.txpool.GetPendingTxByHash(txList[9].RbftGetTxHash()), "tx10 should be removed from txpool")
		ast.Equal(0, len(node.batchDigestM))
		ast.True(node.epcCnf.enableGenEmptyBlock)
		ast.Equal(uint64(10), node.epcCnf.startBlock)
	})
}

func prepareMultiTx(t *testing.T, count int) ([]*types.Transaction, *types.Signer) {
	signer, err := types.GenerateSigner()
	require.Nil(t, err)
	txList := make([]*types.Transaction, 0)
	for i := 0; i < count; i++ {
		tx, err := types.GenerateTransactionWithSigner(uint64(i),
			types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(0), nil, signer)
		require.Nil(t, err)
		txList = append(txList, tx)
	}
	return txList, signer
}

func TestNode_RemoveTxFromPool(t *testing.T) {
	ast := assert.New(t)
	node, err := mockSoloNode(t, false)
	ast.Nil(err)

	err = node.Start()
	ast.Nil(err)
	defer node.Stop()

	txList, _ := prepareMultiTx(t, 10)
	// remove the first tx
	txList = txList[1:]
	ast.Equal(9, len(txList))

	txSubscribeCh := make(chan []*types.Transaction, 1)
	sub := node.SubscribeTxEvent(txSubscribeCh)
	defer sub.Unsubscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockAddTx(node, ctx)

	for _, tx := range txList {
		err = node.Prepare(tx)
		<-txSubscribeCh
		ast.Nil(err)
	}
	// lack nonce 0, so the txs will not be generated to the batch
	ast.Equal(0, len(node.batchDigestM))
	ast.NotNil(node.txpool.GetPendingTxByHash(txList[8].RbftGetTxHash()), "tx9 should be in txpool")
	// sleep to make sure trigger the remove tx from pool
	time.Sleep(2*removeTxTimeout + 500*time.Millisecond)

	ast.Nil(node.txpool.GetPendingTxByHash(txList[8].RbftGetTxHash()), "tx9 should be removed from txpool")
}

func TestNode_GetLowWatermark(t *testing.T) {
	ast := assert.New(t)
	node, err := mockSoloNode(t, false)
	ast.Nil(err)

	err = node.Start()
	ast.Nil(err)
	defer node.Stop()

	ast.Equal(uint64(0), node.GetLowWatermark())

	tx, err := types.GenerateEmptyTransactionAndSigner()
	require.Nil(t, err)
	txSubscribeCh := make(chan []*types.Transaction, 1)
	sub := node.SubscribeTxEvent(txSubscribeCh)
	defer sub.Unsubscribe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockAddTx(node, ctx)

	err = node.Prepare(tx)
	<-txSubscribeCh
	ast.Nil(err)
	commitEvent := <-node.commitC
	ast.NotNil(commitEvent)
	ast.Equal(commitEvent.Block.Height(), node.GetLowWatermark())
}

func TestNode_Prepare(t *testing.T) {
	t.Parallel()
	t.Run("test prepare tx success, generate batch timeout", func(t *testing.T) {
		ast := assert.New(t)
		node, err := mockSoloNode(t, false)
		ast.Nil(err)
		ctrl := gomock.NewController(t)
		wrongPrecheckMgr := mock_precheck.NewMockPreCheck(ctrl)
		wrongPrecheckMgr.EXPECT().Start().AnyTimes()
		wrongPrecheckMgr.EXPECT().PostUncheckedTxEvent(gomock.Any()).Do(func(ev *common.UncheckedTxEvent) {
			event := ev.Event.(*common.TxWithResp)
			event.CheckCh <- &common.TxResp{
				Status:   false,
				ErrorMsg: "check error",
			}
		}).Times(1)
		node.txPreCheck = wrongPrecheckMgr
		tx, err := types.GenerateEmptyTransactionAndSigner()
		require.Nil(t, err)

		err = node.Start()
		require.Nil(t, err)

		err = node.Prepare(tx)
		ast.NotNil(err)
		ast.Contains(err.Error(), "check error")

		wrongPrecheckMgr.EXPECT().PostUncheckedTxEvent(gomock.Any()).Do(func(ev *common.UncheckedTxEvent) {
			event := ev.Event.(*common.TxWithResp)
			event.CheckCh <- &common.TxResp{
				Status: true,
			}
			event.PoolCh <- &common.TxResp{
				Status:   false,
				ErrorMsg: "add pool error",
			}
		}).Times(1)

		err = node.Prepare(tx)
		ast.NotNil(err)
		ast.Contains(err.Error(), "add pool error")

		rightPrecheckMgr := mock_precheck.NewMockMinPreCheck(ctrl, validTxsCh)
		node.txPreCheck = rightPrecheckMgr

		txSubscribeCh := make(chan []*types.Transaction, 1)
		sub := node.SubscribeTxEvent(txSubscribeCh)
		defer sub.Unsubscribe()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mockAddTx(node, ctx)
		err = node.Prepare(tx)
		<-txSubscribeCh
		ast.Nil(err)
		block := <-node.Commit()
		ast.Equal(1, len(block.Block.Transactions))
	})
	t.Run("test prepare txs success, generate batch size", func(t *testing.T) {
		ast := assert.New(t)

		logger := log.NewWithModule("consensus")
		repoRoot := t.TempDir()
		r, err := repo.Load(repo.DefaultKeyJsonPassword, repoRoot, true)
		ast.Nil(err)
		cfg := r.ConsensusConfig

		recvCh := make(chan consensusEvent, maxChanSize)
		mockCtl := gomock.NewController(t)
		mockNetwork := mock_network.NewMockNetwork(mockCtl)
		mockPrecheck := mock_precheck.NewMockMinPreCheck(mockCtl, validTxsCh)

		batchSize := 4
		txpoolConf := txpoolImpl.Config{
			Logger:              &common.Logger{FieldLogger: log.NewWithModule("pool")},
			PoolSize:            poolSize,
			ToleranceRemoveTime: removeTxTimeout,
			GetAccountNonce: func(address string) uint64 {
				return 0
			},
			GetAccountBalance: func(address string) *big.Int {
				maxGasPrice := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e9))
				return new(big.Int).Mul(big.NewInt(math.MaxInt64), maxGasPrice)
			},
			ChainInfo: &txpool.ChainInfo{
				Height:   1,
				GasPrice: new(big.Int).Mul(big.NewInt(5000), big.NewInt(1e9)),
				EpochConf: &txpool.EpochConfig{
					EnableGenEmptyBatch: false,
					BatchSize:           uint64(batchSize),
				},
			},
		}

		txpoolInst, err := txpoolImpl.NewTxPool[types.Transaction, *types.Transaction](txpoolConf)
		ast.Nil(err)

		s, err := types.GenerateSigner()
		ast.Nil(err)

		batchTime := 100 * time.Second // sleep 100s to ensure not to generate batch timeout

		ctx, cancel := context.WithCancel(context.Background())
		soloNode := &Node{
			config: &common.Config{
				Config:  r.ConsensusConfig,
				PrivKey: s.Sk,
			},
			lastExec:         uint64(0),
			noTxBatchTimeout: batchTime,
			batchTimeout:     cfg.Solo.BatchTimeout.ToDuration(),
			commitC:          make(chan *common.CommitEvent, maxChanSize),
			blockCh:          make(chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction], maxChanSize),
			txpool:           txpoolInst,
			network:          mockNetwork,
			batchDigestM:     make(map[uint64]string),
			recvCh:           recvCh,
			logger:           logger,
			ctx:              ctx,
			cancel:           cancel,
			txPreCheck:       mockPrecheck,
			epcCnf: &epochConfig{
				epochPeriod:         r.GenesisConfig.EpochInfo.EpochPeriod,
				startBlock:          r.GenesisConfig.EpochInfo.StartBlock,
				checkpoint:          r.GenesisConfig.EpochInfo.ConsensusParams.CheckpointPeriod,
				enableGenEmptyBlock: txpoolConf.ChainInfo.EpochConf.EnableGenEmptyBatch,
			},
		}

		soloNode.txpool = txpoolInst
		batchTimerMgr := timer.NewTimerManager(logger)
		err = batchTimerMgr.CreateTimer(timer.Batch, batchTime, soloNode.handleTimeoutEvent)
		require.Nil(t, err)
		err = batchTimerMgr.CreateTimer(timer.NoTxBatch, batchTime, soloNode.handleTimeoutEvent)
		require.Nil(t, err)
		soloNode.batchMgr = &batchTimerManager{Timer: batchTimerMgr}

		txSubscribeCh := make(chan []*types.Transaction, 1)
		sub := soloNode.SubscribeTxEvent(txSubscribeCh)
		defer sub.Unsubscribe()

		err = soloNode.Start()
		require.Nil(t, err)
		mockAddTx(soloNode, ctx)

		txs := testutil.ConstructTxs(s, batchSize)

		lo.ForEach(txs, func(tx *types.Transaction, _ int) {
			err = soloNode.Prepare(tx)
			<-txSubscribeCh
			ast.Nil(err)
		})
		block := <-soloNode.Commit()
		ast.Equal(batchSize, len(block.Block.Transactions))

	})
}
