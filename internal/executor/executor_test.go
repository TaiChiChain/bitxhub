package executor

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	consensuscommon "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	from        = "0x3f9d18f7c3a6e5e4c0b877fe3e688ab08840b997"
	minGasPrice = 1000000000000
)

func TestNew(t *testing.T) {
	r := repo.MockRepo(t)
	mockLedger, err := ledger.NewMemory(r)
	assert.Nil(t, err)

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    0,
		BlockHash: types.NewHashByStr(from),
	}
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.ChainMeta = chainMeta

	executor, err := New(r, mockLedger, chainState)
	assert.Nil(t, err)
	assert.NotNil(t, executor)

	assert.Equal(t, mockLedger, executor.ledger)
	assert.NotNil(t, executor.blockC)
	assert.Equal(t, chainMeta.Height, executor.currentHeight)
}

func TestGetEvm(t *testing.T) {
	r := repo.MockRepo(t)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    0,
		BlockHash: types.NewHashByStr(from),
	}
	// mock block for ledger
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	stateLedger.EXPECT().NewView(gomock.Any(), gomock.Any()).Return(stateLedger, nil).AnyTimes()
	blockHeader := &types.BlockHeader{
		Number:         0,
		StateRoot:      &types.Hash{},
		TxRoot:         &types.Hash{},
		ReceiptRoot:    &types.Hash{},
		ParentHash:     &types.Hash{},
		Timestamp:      0,
		Epoch:          0,
		Bloom:          &types.Bloom{},
		GasPrice:       0,
		GasUsed:        0,
		ProposerNodeID: 0,
	}
	chainLedger.EXPECT().GetBlockHeader(gomock.Any()).Return(blockHeader, nil).Times(3)
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.ChainMeta = chainMeta
	executor, err := New(r, mockLedger, chainState)
	assert.Nil(t, err)
	assert.NotNil(t, executor)

	txCtx := vm.TxContext{}
	evm, err := executor.NewEvmWithViewLedger(txCtx, vm.Config{NoBaseFee: true})
	assert.NotNil(t, evm)
	assert.Nil(t, err)

	h := evm.Context.GetHash(0)
	assert.Equal(t, blockHeader.Hash().String(), h.String())

	chainLedger.EXPECT().GetBlockHeader(gomock.Any()).Return(nil, errors.New("get block error")).Times(1)
	evmErr, err := executor.NewEvmWithViewLedger(txCtx, vm.Config{NoBaseFee: true})
	assert.Nil(t, evmErr)
	assert.NotNil(t, err)
}

func TestSubscribeLogsEvent(t *testing.T) {
	executor := executorStart(t)
	ch := make(chan []*types.EvmLog, 10)
	subscription := executor.SubscribeLogsEvent(ch)
	assert.NotNil(t, subscription)
}

func TestGetLogsForReceipt(t *testing.T) {
	executor := executorStart(t)
	receipts := []*types.Receipt{{
		EvmLogs: []*types.EvmLog{{
			BlockNumber: 0,
		}},
	}}
	executor.updateLogsBlockHash(receipts, &types.Hash{})
}

func TestGetChainConfig(t *testing.T) {
	executor := executorStart(t)
	config := executor.GetChainConfig()
	assert.NotNil(t, config)
}

func TestGetBlockHashFunc(t *testing.T) {
	r := repo.MockRepo(t)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    0,
		BlockHash: types.NewHashByStr(from),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	chainLedger.EXPECT().GetBlockHeader(gomock.Any()).Return(&types.BlockHeader{
		Number:         0,
		StateRoot:      &types.Hash{},
		TxRoot:         &types.Hash{},
		ReceiptRoot:    &types.Hash{},
		ParentHash:     &types.Hash{},
		Timestamp:      0,
		Epoch:          0,
		Bloom:          &types.Bloom{},
		GasPrice:       0,
		GasUsed:        0,
		ProposerNodeID: 0,
	}, nil).AnyTimes()
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.ChainMeta = chainMeta
	executor, _ := New(r, mockLedger, chainState)

	getHash := getBlockHashFunc(executor.ledger.ChainLedger)
	getHash(10)
}

func TestBlockExecutor_ExecuteBlock(t *testing.T) {
	r := repo.MockRepo(t)
	r.GenesisConfig.EpochInfo.StartBlock = 0
	r.GenesisConfig.EpochInfo.EpochPeriod = 1

	mockLedger, err := ledger.NewMemory(r)
	require.Nil(t, err)

	genesisBlock := &types.Block{
		Header: &types.BlockHeader{
			Number:   0,
			GasPrice: minGasPrice * 5,
		},
	}

	nvm := system.New()
	err = nvm.GenesisInit(r.GenesisConfig, mockLedger.StateLedger)
	assert.Nil(t, err)

	mockLedger.StateLedger.Finalise()

	stateRoot, err := mockLedger.StateLedger.Commit()
	assert.Nil(t, err)
	genesisBlock.Header.StateRoot = stateRoot
	mockLedger.PersistBlockData(&ledger.BlockData{Block: genesisBlock})

	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.ChainMeta = &types.ChainMeta{
		Height:    0,
		GasPrice:  new(big.Int),
		BlockHash: genesisBlock.Hash(),
	}
	exec, err := New(r, mockLedger, chainState)
	assert.Nil(t, err)

	var txs []*types.Transaction
	nonce := uint64(0)
	emptyDataTx := mockTx(t, nonce)
	nonce++
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
	commitEvent1 := mockCommitEvent(uint64(1), nil)

	commitEvent2 := mockCommitEvent(uint64(2), txs)
	exec.AsyncExecuteBlock(commitEvent1)
	exec.AsyncExecuteBlock(commitEvent2)

	blockRes1 := <-ch

	assert.EqualValues(t, 1, blockRes1.Block.Header.Number)
	assert.Equal(t, 0, len(blockRes1.Block.Transactions))
	assert.Equal(t, 0, len(blockRes1.TxPointerList))

	remoteBlockRes1 := <-remoteCh
	assert.Equal(t, blockRes1, remoteBlockRes1)

	blockRes2 := <-ch
	assert.EqualValues(t, 2, blockRes2.Block.Header.Number)
	assert.Equal(t, 2, len(blockRes2.Block.Transactions))
	assert.Equal(t, 2, len(blockRes2.TxPointerList))

	remoteBlockRes2 := <-remoteCh
	assert.Equal(t, blockRes2, remoteBlockRes2)

	t.Run("test rollback block", func(t *testing.T) {
		// send bigger block to executor
		oldHeight := exec.currentHeight
		biggerCommitEvent := mockCommitEvent(uint64(5), nil)
		exec.processExecuteEvent(biggerCommitEvent)
		assert.Equal(t, oldHeight, exec.currentHeight, "ignore illegal block")

		oldBlock := blockRes1.Block
		// send rollback block to executor
		rollbackCommitEvent := mockCommitEvent(uint64(1), txs)
		exec.AsyncExecuteBlock(rollbackCommitEvent)

		blockRes := <-ch
		assert.EqualValues(t, 1, blockRes.Block.Header.Number)
		assert.Equal(t, len(txs), len(blockRes.Block.Transactions))
		assert.Equal(t, genesisBlock.Hash().String(), blockRes.Block.Header.ParentHash.String())
		assert.NotEqual(t, oldBlock.Hash().String(), blockRes.Block.Hash().String())

		remoteBlockRes := <-remoteCh
		assert.Equal(t, blockRes, remoteBlockRes)

		// handle panic error
		defer func() {
			r := recover()
			assert.NotNil(t, r)
			assert.Contains(t, fmt.Sprintf("%v", r), "cannot rollback genesis block")
		}()

		// send rollback block to executor, but rollback error
		rollbackCommitEvent1 := mockCommitEvent(uint64(0), nil)
		exec.processExecuteEvent(rollbackCommitEvent1)
	})

	t.Run("test rollback block with error", func(t *testing.T) {
		// handle panic error
		defer func() {
			if r := recover(); r != nil {
				assert.NotNil(t, r)
				assert.Contains(t, fmt.Sprintf("%v", r), "cannot rollback genesis block")
			}
		}()
		// send rollback block to executor, but rollback error
		rollbackCommitEvent1 := mockCommitEvent(uint64(0), nil)
		exec.processExecuteEvent(rollbackCommitEvent1)
	})

	t.Run("test get block with error", func(t *testing.T) {
		errorGetBlock := errors.New("out-order")
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

// NodeExtraArgs is Node proposal extra arguments

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
		Header:       header,
		Transactions: txs,
	}
	return block
}

func generateTransactionAndSigner(nonce uint64, to *types.Address, value *big.Int, data []byte) (*types.Transaction, *types.Signer, error) {
	sk, err := ethcrypto.HexToECDSA("b6477143e17f889263044f6cf463dc37177ac4526c4c39a7a344198457024a2f")
	if err != nil {
		return nil, nil, err
	}
	a := ethcrypto.PubkeyToAddress(sk.PublicKey)
	s := &types.Signer{
		Sk:   sk,
		Addr: types.NewAddress(a.Bytes()),
	}
	tx, err := types.GenerateTransactionWithSigner(nonce, to, value, data, s)
	if err != nil {
		return nil, nil, err
	}

	return tx, s, nil
}

func mockTx(t *testing.T, nonce ...uint64) *types.Transaction {
	var localNonce uint64 = 0
	if nonce != nil {
		localNonce = nonce[0]
	}
	tx, _, err := generateTransactionAndSigner(localNonce, types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil)
	assert.Nil(t, err)
	return tx
}

func TestBlockExecutor_ExecuteBlock_Transfer(t *testing.T) {
	r := repo.MockRepo(t)

	ldg, err := ledger.NewMemory(r)
	require.Nil(t, err)

	signer, err := types.GenerateSigner()
	require.Nil(t, err)
	to := types.NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7")

	dummyRootHash := common2.Hash{}
	ldg.StateLedger.PrepareBlock(types.NewHash(dummyRootHash[:]), 1)
	ldg.StateLedger.SetBalance(signer.Addr, new(big.Int).Mul(big.NewInt(5000000000000), big.NewInt(21000*10000)))
	account := ldg.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	balance, _ := new(big.Int).SetString("100000000000000000000", 10)
	account.SetState([]byte("axcBalances-0xf16F8B02df2Dd7c4043C41F3f1EBB17f15358888"), balance.Bytes())
	ldg.StateLedger.Finalise()
	rootHash, err := ldg.StateLedger.Commit()
	require.Nil(t, err)
	require.NotNil(t, rootHash)
	block1 := mockBlock(0, nil)
	// refresh block
	block1.Header.StateRoot = rootHash
	err = ldg.ChainLedger.PersistExecutionResult(block1, nil)
	require.Nil(t, err)

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    0,
		GasPrice:  big.NewInt(minGasPrice),
		BlockHash: types.NewHash([]byte(from)),
	}
	ldg.ChainLedger.UpdateChainMeta(chainMeta)

	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	executor, err := New(r, ldg, chainState)
	require.Nil(t, err)
	err = executor.Start()
	require.Nil(t, err)

	ch := make(chan events.ExecutedEvent)
	sub := executor.SubscribeBlockEvent(ch)
	defer sub.Unsubscribe()

	tx1 := mockTransferTx(t, signer, to, 0, 1)
	tx2 := mockTransferTx(t, signer, to, 1, 1)
	tx3 := mockTransferTx(t, signer, to, 2, 1)
	commitEvent := mockCommitEvent(1, []*types.Transaction{tx1, tx2, tx3})
	executor.AsyncExecuteBlock(commitEvent)

	block := <-ch
	require.EqualValues(t, 1, block.Block.Height())
	require.EqualValues(t, 3, ldg.StateLedger.GetBalance(to).Uint64())
}

func mockTransferTx(t *testing.T, s *types.Signer, to *types.Address, nonce, amount int) *types.Transaction {
	tx, err := types.GenerateTransactionWithSigner(uint64(nonce), to, big.NewInt(int64(amount)), nil, s)
	assert.Nil(t, err)
	return tx
}

func executorStart(t *testing.T) *BlockExecutor {
	r := repo.MockRepo(t)

	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    0,
		BlockHash: types.NewHashByStr(from),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.ChainMeta = chainMeta
	executor, err := New(r, mockLedger, chainState)
	require.Nil(t, err)
	return executor
}
