package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	consensuscommon "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethvm "github.com/axiomesh/eth-kit/evm"
)

const (
	srcMethod   = "did:axiom-ledger:addr1:."
	dstMethod   = "did:axiom-ledger:addr2:."
	from        = "0x3f9d18f7c3a6e5e4c0b877fe3e688ab08840b997"
	minGasPrice = 1000000000000
)

func TestNew(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		BlockHash: types.NewHashByStr(from),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()

	executor, err := New(r, mockLedger)
	assert.Nil(t, err)
	assert.NotNil(t, executor)

	assert.Equal(t, mockLedger, executor.ledger)
	assert.NotNil(t, executor.blockC)
	assert.Equal(t, chainMeta.Height, executor.currentHeight)
}

func TestGetEvm(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		BlockHash: types.NewHashByStr(from),
	}
	// mock block for ledger
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(mockBlock(1, nil), nil).Times(2)
	stateLedger.EXPECT().NewView(gomock.Any()).Return(stateLedger).AnyTimes()
	chainLedger.EXPECT().GetBlockHash(gomock.Any()).Return(&types.Hash{}).AnyTimes()

	executor, err := New(r, mockLedger)
	assert.Nil(t, err)
	assert.NotNil(t, executor)

	txCtx := ethvm.TxContext{}
	evm, err := executor.NewEvmWithViewLedger(txCtx, ethvm.Config{NoBaseFee: true})
	assert.NotNil(t, evm)
	assert.Nil(t, err)

	h := evm.Context.GetHash(0)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", h.String())

	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(nil, errors.New("get block error")).Times(1)
	evmErr, err := executor.NewEvmWithViewLedger(txCtx, ethvm.Config{NoBaseFee: true})
	assert.Nil(t, evmErr)
	assert.NotNil(t, err)
}

func TestSubscribeLogsEvent(t *testing.T) {
	executor := executor_start(t)
	ch := make(chan []*types.EvmLog, 10)
	subscription := executor.SubscribeLogsEvent(ch)
	assert.NotNil(t, subscription)
}

func TestGetLogsForReceipt(t *testing.T) {
	executor := executor_start(t)
	receipts := []*types.Receipt{{
		EvmLogs: []*types.EvmLog{{
			BlockNumber: 0,
		}},
	}}
	executor.getLogsForReceipt(receipts, &types.Hash{})
}

func TestGetChainConfig(t *testing.T) {
	executor := executor_start(t)
	config := executor.GetChainConfig()
	assert.NotNil(t, config)
}

func TestGetBlockHashFunc(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		BlockHash: types.NewHashByStr(from),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	chainLedger.EXPECT().GetBlockHash(gomock.Any()).Return(nil)

	executor, _ := New(r, mockLedger)

	getHash := getBlockHashFunc(executor.ledger.ChainLedger)
	getHash(10)
}

func TestBlockExecutor_ExecuteBlock(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)
	r.Config.Genesis.EpochInfo.StartBlock = 2
	r.Config.Genesis.EpochInfo.EpochPeriod = 1

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	genesisBlock := &types.Block{
		BlockHeader: &types.BlockHeader{
			Number: 1,
		},
		BlockHash: types.NewHash([]byte(from)),
	}
	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    genesisBlock.Height(),
		GasPrice:  big.NewInt(minGasPrice),
		BlockHash: genesisBlock.BlockHash,
	}
	// block := &types.Block{
	//	BlockHeader: &types.BlockHeader{
	//		GasPrice: minGasPrice,
	//	},
	// }

	evs := make([]*types.Event, 0)
	m := make(map[string]uint64)
	m[from] = 3
	data, err := json.Marshal(m)
	assert.Nil(t, err)
	ev := &types.Event{
		TxHash:    types.NewHash([]byte(from)),
		Data:      data,
		EventType: types.EventOTHER,
	}
	ev2 := &types.Event{
		TxHash:    types.NewHash([]byte(from)),
		Data:      data,
		EventType: types.EventOTHER,
	}
	ev3 := &types.Event{
		TxHash:    types.NewHash([]byte(from)),
		Data:      data,
		EventType: types.EventOTHER,
	}
	stateRoot := &types.Hash{}

	evs = append(evs, ev, ev2, ev3)
	stateLedger.EXPECT().NewView(gomock.Any()).Return(stateLedger).AnyTimes()
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	stateLedger.EXPECT().Commit().Return(stateRoot, nil).AnyTimes()
	stateLedger.EXPECT().Clear().AnyTimes()
	stateLedger.EXPECT().GetEVMNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().GetBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().GetEVMBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().SetCode(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetCode(gomock.Any()).Return([]byte("10")).AnyTimes()
	stateLedger.EXPECT().GetLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	chainLedger.EXPECT().PersistExecutionResult(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().PrepareBlock(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Finalise().AnyTimes()
	stateLedger.EXPECT().Snapshot().Return(1).AnyTimes()
	stateLedger.EXPECT().RevertToSnapshot(1).AnyTimes()
	stateLedger.EXPECT().PrepareEVM(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Close().AnyTimes()
	chainLedger.EXPECT().Close().AnyTimes()
	stateLedger.EXPECT().GetEVMCode(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMRefund().AnyTimes()
	stateLedger.EXPECT().GetEVMCodeHash(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SubEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetEVMNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().ExistEVM(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().CreateEVMAccount(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(genesisBlock, nil).Times(2)
	stateLedger.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(addr *types.Address, key []byte, value []byte) {},
	).AnyTimes()
	stateLedger.EXPECT().GetState(gomock.Any(), gomock.Any()).DoAndReturn(
		func(addr *types.Address, key []byte) (bool, []byte) {
			return true, []byte("10")
		}).AnyTimes()

	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "executor"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(2, ld, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	err = base.InitEpochInfo(stateLedger, r.Config.Genesis.EpochInfo)
	assert.Nil(t, err)

	exec, err := New(r, mockLedger)
	assert.Nil(t, err)

	var txs []*types.Transaction
	emptyDataTx := mockTx(t)
	txs = append(txs, emptyDataTx)

	invalidTx := mockTx(t)
	invalidTx.Inner.(*types.LegacyTx).Nonce = 1000
	txs = append(txs, invalidTx)

	assert.Nil(t, exec.Start())

	ch := make(chan events.ExecutedEvent)
	remoteCh := make(chan events.ExecutedEvent)
	blockSub := exec.SubscribeBlockEvent(ch)
	remoteBlockSub := exec.SubscribeBlockEventForRemote(remoteCh)
	defer func() {
		blockSub.Unsubscribe()
		remoteBlockSub.Unsubscribe()
	}()

	// send blocks to executor
	commitEvent1 := mockCommitEvent(uint64(2), nil)

	commitEvent2 := mockCommitEvent(uint64(3), txs)
	exec.AsyncExecuteBlock(commitEvent1)
	exec.AsyncExecuteBlock(commitEvent2)

	blockRes1 := <-ch
	assert.EqualValues(t, 2, blockRes1.Block.BlockHeader.Number)
	assert.Equal(t, 0, len(blockRes1.Block.Transactions))
	assert.Equal(t, 0, len(blockRes1.TxHashList))

	remoteBlockRes1 := <-remoteCh
	assert.Equal(t, blockRes1, remoteBlockRes1)

	blockRes2 := <-ch
	assert.EqualValues(t, 3, blockRes2.Block.BlockHeader.Number)
	assert.Equal(t, 2, len(blockRes2.Block.Transactions))
	assert.Equal(t, 2, len(blockRes2.TxHashList))

	remoteBlockRes2 := <-remoteCh
	assert.Equal(t, blockRes2, remoteBlockRes2)

	t.Run("test rollback block", func(t *testing.T) {
		chainLedger.EXPECT().GetBlock(gomock.Any()).Return(genesisBlock, nil).Times(3)
		// send bigger block to executor
		oldHeight := exec.currentHeight
		biggerCommitEvent := mockCommitEvent(uint64(5), nil)
		exec.processExecuteEvent(biggerCommitEvent)
		assert.Equal(t, oldHeight, exec.currentHeight, "ignore illegal block")

		chainLedger.EXPECT().RollbackBlockChain(gomock.Any()).Return(nil).AnyTimes()
		stateLedger.EXPECT().RollbackState(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		oldBlock := blockRes1.Block
		// send rollback block to executor
		rollbackCommitEvent := mockCommitEvent(uint64(2), txs)
		exec.AsyncExecuteBlock(rollbackCommitEvent)

		blockRes := <-ch
		assert.EqualValues(t, 2, blockRes.Block.BlockHeader.Number)
		assert.Equal(t, len(txs), len(blockRes.Block.Transactions))
		assert.Equal(t, genesisBlock.BlockHash.String(), blockRes.Block.BlockHeader.ParentHash.String())
		assert.NotEqual(t, oldBlock.BlockHash.String(), blockRes.Block.BlockHash.String())

		remoteBlockRes := <-remoteCh
		assert.Equal(t, blockRes, remoteBlockRes)

		errorRollbackStateLedger := errors.New("rollback state ledger err")
		stateLedger.EXPECT().RollbackState(gomock.Any(), gomock.Any()).Return(errorRollbackStateLedger).Times(1)

		// handle panic error
		defer func() {
			if r := recover(); r != nil {
				assert.NotNil(t, r)
				assert.Contains(t, fmt.Sprintf("%v", r), errorRollbackStateLedger.Error())
			}
		}()

		// send rollback block to executor, but rollback error
		rollbackCommitEvent1 := mockCommitEvent(uint64(1), nil)
		exec.processExecuteEvent(rollbackCommitEvent1)
	})

	t.Run("test rollback block with error", func(t *testing.T) {
		chainLedger.EXPECT().RollbackBlockChain(gomock.Any()).Return(nil).AnyTimes()
		stateLedger.EXPECT().RollbackState(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		errorGetBlock := errors.New("get block error")
		chainLedger.EXPECT().GetBlock(gomock.Any()).Return(nil, errorGetBlock).Times(1)
		// handle panic error
		defer func() {
			if r := recover(); r != nil {
				assert.NotNil(t, r)
				assert.Contains(t, fmt.Sprintf("%v", r), errorGetBlock.Error())
			}
		}()
		// send rollback block to executor, but rollback error
		rollbackCommitEvent1 := mockCommitEvent(uint64(1), nil)
		exec.processExecuteEvent(rollbackCommitEvent1)
	})

	t.Run("test get block with error", func(t *testing.T) {
		errorGetBlock := errors.New("get block error")
		chainLedger.EXPECT().GetBlock(gomock.Any()).Return(nil, errorGetBlock).Times(1)
		// handle panic error
		defer func() {
			if r := recover(); r != nil {
				assert.NotNil(t, r)
				assert.Contains(t, fmt.Sprintf("%v", r), errorGetBlock.Error())
			}
		}()
		// send rollback block to executor, but rollback error
		exec.currentHeight = 0
		rollbackCommitEvent1 := mockCommitEvent(uint64(1), nil)
		exec.processExecuteEvent(rollbackCommitEvent1)
	})

	assert.Nil(t, exec.Stop())
}

func TestBlockExecutor_ApplyReadonlyTransactions(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)
	repoRoot := r.RepoRoot
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		GasPrice:  big.NewInt(minGasPrice),
		BlockHash: types.NewHashByStr(from),
	}

	id := fmt.Sprintf("%s-%s-%d", srcMethod, dstMethod, 1)

	hash := types.NewHash([]byte{1})
	val, err := json.Marshal(hash)
	assert.Nil(t, err)

	ld, err := leveldb.New(filepath.Join(repoRoot, "executor"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(2, ld, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateRoot := &types.Hash{}

	contractAddr := types.NewAddressByStr("0xdac17f958d2ee523a2206206994597c13d831ec7")
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	stateLedger.EXPECT().Commit().Return(stateRoot, nil).AnyTimes()
	stateLedger.EXPECT().Clear().AnyTimes()
	stateLedger.EXPECT().GetState(contractAddr, []byte(fmt.Sprintf("index-tx-%s", id))).Return(true, val).AnyTimes()
	chainLedger.EXPECT().PersistExecutionResult(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().SetNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Finalise().AnyTimes()
	stateLedger.EXPECT().Snapshot().Return(1).AnyTimes()
	stateLedger.EXPECT().RevertToSnapshot(1).AnyTimes()
	stateLedger.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	chainLedger.EXPECT().LoadChainMeta().Return(chainMeta, nil).AnyTimes()
	stateLedger.EXPECT().GetLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(mockBlock(10, nil), nil).AnyTimes()
	stateLedger.EXPECT().PrepareEVM(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().PrepareBlock(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetEVMNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().GetBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().GetEVMBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().SubEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().ExistEVM(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().CreateEVMAccount(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMCode(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMRefund().AnyTimes()
	stateLedger.EXPECT().GetEVMCodeHash(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	signer, err := types.GenerateSigner()
	assert.Nil(t, err)
	err = governance.InitCouncilMembers(stateLedger, []*repo.Admin{
		{
			Address: signer.Addr.String(),
			Weight:  1,
			Name:    "111",
		},
		{
			Address: "0x1220000000000000000000000000000000000000",
			Weight:  1,
			Name:    "222",
		},
		{
			Address: "0x1230000000000000000000000000000000000000",
			Weight:  1,
			Name:    "333",
		},
		{
			Address: "0x1240000000000000000000000000000000000000",
			Weight:  1,
			Name:    "444",
		},
	}, "1000000")
	assert.Nil(t, err)
	err = governance.InitNodeMembers(stateLedger, []*governance.NodeMember{
		{
			NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
		},
	})
	assert.Nil(t, err)

	exec, err := New(r, mockLedger)
	assert.Nil(t, err)

	rawTx := "0xf86c8085147d35700082520894f927bb571eaab8c9a361ab405c9e4891c5024380880de0b6b3a76400008025a00b8e3b66c1e7ae870802e3ef75f1ec741f19501774bd5083920ce181c2140b99a0040c122b7ebfb3d33813927246cbbad1c6bf210474f5d28053990abff0fd4f53"
	tx4 := &types.Transaction{}
	err = tx4.Unmarshal(hexutil.Decode(rawTx))
	assert.Nil(t, err)

	var txs3 []*types.Transaction
	tx5, _, err := types.GenerateTransactionAndSigner(uint64(0), types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil)
	assert.Nil(t, err)
	// test system contract
	data := generateNodeAddProposeData(t, NodeExtraArgs{
		Nodes: []*NodeMember{
			{
				NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})
	assert.Nil(t, err)
	tx6, err := types.GenerateTransactionWithSigner(uint64(1), types.NewAddressByStr(common.NodeManagerContractAddr), big.NewInt(0), data, signer)
	assert.Nil(t, err)

	txs3 = append(txs3, tx4, tx5, tx6)
	res := exec.ApplyReadonlyTransactions(txs3)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, types.ReceiptSUCCESS, res[0].Status)
	assert.Equal(t, types.ReceiptSUCCESS, res[1].Status)
	assert.Equal(t, types.ReceiptSUCCESS, res[2].Status)
}

func TestBlockExecutor_ApplyReadonlyTransactionsWithError(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)
	repoRoot := r.RepoRoot
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		GasPrice:  big.NewInt(minGasPrice),
		BlockHash: types.NewHashByStr(from),
	}

	id := fmt.Sprintf("%s-%s-%d", srcMethod, dstMethod, 1)

	hash := types.NewHash([]byte{1})
	val, err := json.Marshal(hash)
	assert.Nil(t, err)

	ld, err := leveldb.New(filepath.Join(repoRoot, "executor"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(1, ld, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	contractAddr := types.NewAddressByStr("0xdac17f958d2ee523a2206206994597c13d831ec7")
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	stateLedger.EXPECT().Commit().Return(nil, nil).AnyTimes()
	stateLedger.EXPECT().Clear().AnyTimes()
	stateLedger.EXPECT().GetState(contractAddr, []byte(fmt.Sprintf("index-tx-%s", id))).Return(true, val).AnyTimes()
	chainLedger.EXPECT().PersistExecutionResult(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().SetNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Finalise().AnyTimes()
	stateLedger.EXPECT().Snapshot().Return(1).AnyTimes()
	stateLedger.EXPECT().RevertToSnapshot(1).AnyTimes()
	stateLedger.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(nil, errors.New("block not found")).AnyTimes()
	stateLedger.EXPECT().PrepareEVM(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().PrepareBlock(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetEVMNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().GetBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().GetEVMBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().SubEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().ExistEVM(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().CreateEVMAccount(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddEVMBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMCode(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetEVMRefund().AnyTimes()
	stateLedger.EXPECT().GetEVMCodeHash(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	signer, err := types.GenerateSigner()
	assert.Nil(t, err)
	err = governance.InitCouncilMembers(stateLedger, []*repo.Admin{
		{
			Address: signer.Addr.String(),
			Weight:  1,
			Name:    "111",
		},
		{
			Address: "0x1220000000000000000000000000000000000000",
			Weight:  1,
			Name:    "222",
		},
		{
			Address: "0x1230000000000000000000000000000000000000",
			Weight:  1,
			Name:    "333",
		},
		{
			Address: "0x1240000000000000000000000000000000000000",
			Weight:  1,
			Name:    "444",
		},
	}, "1000000")
	assert.Nil(t, err)
	err = governance.InitNodeMembers(stateLedger, []*governance.NodeMember{
		{
			NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
		},
	})
	assert.Nil(t, err)

	exec, err := New(r, mockLedger)
	assert.Nil(t, err)

	rawTx := "0xf86c8085147d35700082520894f927bb571eaab8c9a361ab405c9e4891c5024380880de0b6b3a76400008025a00b8e3b66c1e7ae870802e3ef75f1ec741f19501774bd5083920ce181c2140b99a0040c122b7ebfb3d33813927246cbbad1c6bf210474f5d28053990abff0fd4f53"
	tx4 := &types.Transaction{}
	err = tx4.Unmarshal(hexutil.Decode(rawTx))
	assert.Nil(t, err)

	var txs3 []*types.Transaction
	tx5, _, err := types.GenerateTransactionAndSigner(uint64(0), types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil)
	assert.Nil(t, err)
	// test system contract
	data := generateNodeAddProposeData(t, NodeExtraArgs{
		Nodes: []*NodeMember{
			{
				NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})
	assert.Nil(t, err)
	tx6, err := types.GenerateTransactionWithSigner(uint64(1), types.NewAddressByStr(common.NodeManagerContractAddr), big.NewInt(0), data, signer)
	assert.Nil(t, err)

	txs3 = append(txs3, tx4, tx5, tx6)
	exec.ApplyReadonlyTransactions(txs3)
}

func generateNodeAddProposeData(t *testing.T, extraArgs NodeExtraArgs) []byte {
	// test system contract
	gabi, err := governance.GetABI()
	assert.Nil(t, err)

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)

	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(governance.ProposeMethod, uint8(governance.NodeUpgrade), title, desc, blockNumber, extra)
	assert.Nil(t, err)
	return data
}

// NodeExtraArgs is Node proposal extra arguments
type NodeExtraArgs struct {
	Nodes []*NodeMember
}

type NodeMember struct {
	NodeId string
}

func mockCommitEvent(blockNumber uint64, txs []*types.Transaction) *consensuscommon.CommitEvent {
	return &consensuscommon.CommitEvent{
		Block: mockBlock(blockNumber, txs),
	}
}

func mockBlock(blockNumber uint64, txs []*types.Transaction) *types.Block {
	header := &types.BlockHeader{
		Number:    blockNumber,
		Timestamp: time.Now().Unix(),
	}

	block := &types.Block{
		BlockHeader:  header,
		Transactions: txs,
	}
	block.BlockHash = block.Hash()

	return block
}

func mockTx(t *testing.T) *types.Transaction {
	tx, _, err := types.GenerateTransactionAndSigner(uint64(0), types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil)
	assert.Nil(t, err)
	return tx
}

func TestBlockExecutor_ExecuteBlock_Transfer(t *testing.T) {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {},
		"pebble":  {},
	}

	for kvType := range testcase {
		t.Run(kvType, func(t *testing.T) {
			err = storagemgr.Initialize(kvType, repo.KVStorageCacheSize)
			require.Nil(t, err)
			ldg, err := ledger.NewLedger(createMockRepo(t))
			require.Nil(t, err)

			signer, err := types.GenerateSigner()
			require.Nil(t, err)
			to := types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7")

			ldg.StateLedger.SetBalance(signer.Addr, new(big.Int).Mul(big.NewInt(5000000000000), big.NewInt(21000*10000)))
			rootHash, err := ldg.StateLedger.Commit()
			require.Nil(t, err)
			require.NotNil(t, rootHash)
			err = ldg.ChainLedger.PersistExecutionResult(mockBlock(1, nil), nil)
			require.Nil(t, err)

			// mock data for ledger
			chainMeta := &types.ChainMeta{
				Height:    1,
				GasPrice:  big.NewInt(minGasPrice),
				BlockHash: types.NewHash([]byte(from)),
			}
			ldg.ChainLedger.UpdateChainMeta(chainMeta)

			executor, err := New(r, ldg)
			require.Nil(t, err)
			err = executor.Start()
			require.Nil(t, err)

			ch := make(chan events.ExecutedEvent)
			sub := executor.SubscribeBlockEvent(ch)
			defer sub.Unsubscribe()

			tx1 := mockTransferTx(t, signer, to, 0, 1)
			tx2 := mockTransferTx(t, signer, to, 1, 1)
			tx3 := mockTransferTx(t, signer, to, 2, 1)
			commitEvent := mockCommitEvent(2, []*types.Transaction{tx1, tx2, tx3})
			executor.AsyncExecuteBlock(commitEvent)

			block := <-ch
			require.EqualValues(t, 2, block.Block.Height())
			require.EqualValues(t, 3, ldg.StateLedger.GetBalance(to).Uint64())
		})
	}
}

func mockTransferTx(t *testing.T, s *types.Signer, to *types.Address, nonce, amount int) *types.Transaction {
	tx, err := types.GenerateTransactionWithSigner(uint64(nonce), to, big.NewInt(int64(amount)), nil, s)
	assert.Nil(t, err)
	return tx
}

func createMockRepo(t *testing.T) *repo.Repo {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	return r
}

func executor_start(t *testing.T) *BlockExecutor {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		BlockHash: types.NewHashByStr(from),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()

	executor, _ := New(r, mockLedger)
	return executor
}
