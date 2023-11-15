package solo

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-ledger/pkg/txpool"

	rbft "github.com/axiomesh/axiom-bft"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNode_Start(t *testing.T) {
	repoRoot := t.TempDir()
	r, err := repo.Load(repoRoot)
	require.Nil(t, err)

	mockCtl := gomock.NewController(t)
	mockNetwork := mock_network.NewMockNetwork(mockCtl)

	config, err := common.GenerateConfig(
		common.WithConfig(r.RepoRoot, r.ConsensusConfig),
		common.WithGenesisEpochInfo(r.Config.Genesis.EpochInfo),
		common.WithLogger(log.NewWithModule("consensus")),
		common.WithApplied(1),
		common.WithNetwork(mockNetwork),
		common.WithApplied(1),
		common.WithGetAccountNonceFunc(func(address *types.Address) uint64 {
			return 0
		}),
		common.WithGetAccountBalanceFunc(func(address *types.Address) *big.Int {
			return big.NewInt(adminBalance)
		}),
		common.WithGetChainMetaFunc(func() *types.ChainMeta {
			return &types.ChainMeta{
				GasPrice: big.NewInt(0),
			}
		}),
		common.WithTxPool(mockTxPool(t)),
		common.WithGetCurrentEpochInfoFromEpochMgrContractFunc(func() (*rbft.EpochInfo, error) {
			return &rbft.EpochInfo{ConsensusParams: rbft.ConsensusParams{
				EnableTimedGenEmptyBlock: false,
			}}, nil
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
	require.Equal(t, uint64(1), solo.Quorum())

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
	require.Equal(t, uint64(2), commitEvent.Block.BlockHeader.Number)
	require.Equal(t, 1, len(commitEvent.Block.Transactions))
	blockHash := commitEvent.Block.Hash()

	txHashList := make([]*types.Hash, 0)
	txHashList = append(txHashList, tx.GetHash())
	solo.ReportState(commitEvent.Block.Height(), blockHash, txHashList, nil, false)
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
	ast.Equal(event1.Block.BlockHeader.Number, uint64(1))

	event2 := <-node.commitC
	ast.NotNil(event2)
	ast.Equal(len(event2.Block.Transactions), 0)
	ast.Equal(event2.Block.BlockHeader.Number, uint64(2))
}

func TestNode_ReportState(t *testing.T) {
	ast := assert.New(t)
	node, err := mockSoloNode(t, false)
	ast.Nil(err)

	err = node.Start()
	ast.Nil(err)
	defer node.Stop()
	node.batchDigestM[10] = "test"
	node.ReportState(10, types.NewHashByStr("0x123"), []*types.Hash{}, nil, false)
	time.Sleep(10 * time.Millisecond)
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
	t.Run("test pool full", func(t *testing.T) {
		ast.Equal(10, len(node.batchDigestM))
		tx11, err := types.GenerateTransactionWithSigner(uint64(11),
			types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(0), nil, signer)

		ast.Nil(err)
		err = node.Prepare(tx11)
		ast.NotNil(err)
		<-txSubscribeCh
		ast.Contains(err.Error(), ErrPoolFull)
		ast.Equal(10, len(node.batchDigestM), "the pool should be full, tx11 is not add in txpool successfully")
	})

	ast.NotNil(node.txpool.GetPendingTxByHash(txList[9].RbftGetTxHash()), "tx10 should be in txpool")
	// trigger the report state
	node.ReportState(10, types.NewHashByStr("0x123"), []*types.Hash{txList[9].GetHash()}, nil, false)
	time.Sleep(100 * time.Millisecond)
	ast.Equal(0, len(node.batchDigestM))
	ast.Nil(node.txpool.GetPendingTxByHash(txList[9].RbftGetTxHash()), "tx10 should be removed from txpool")
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
