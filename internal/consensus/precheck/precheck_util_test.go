package precheck

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool/mock_txpool"
	"github.com/axiomesh/axiom-kit/types"
	common2 "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	basicGas = 21000
	to       = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
)

type mockDb struct {
	db map[string]*big.Int
}

func newMockPreCheckMgr(ledger *mockDb, t *testing.T) (*TxPreCheckMgr, *logrus.Entry, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.NewWithModule("precheck")

	getAccountBalance := func(address string) *big.Int {
		val, ok := ledger.db[address]
		if !ok {
			return big.NewInt(0)
		}
		return val
	}
	getChainmetaFn := func() *types.ChainMeta {
		return &types.ChainMeta{
			GasPrice: big.NewInt(0),
		}
	}

	ctrl := gomock.NewController(t)
	mockPool := mock_txpool.NewMockMinimalTxPool[types.Transaction, *types.Transaction](500, ctrl)

	cnf := &common2.Config{
		Logger: logger,
		GenesisEpochInfo: &types.EpochInfo{
			MiscParams: rbft.MiscParams{
				TxMaxSize: repo.DefaultTxMaxSize,
			},
		},
		GetChainMetaFunc:  getChainmetaFn,
		GetAccountBalance: getAccountBalance,
		TxPool:            mockPool,
	}

	return NewTxPreCheckMgr(ctx, cnf), logger, cancel
}

func (db *mockDb) setBalance(address string, balance *big.Int) {
	db.db[address] = balance
}

func (db *mockDb) getBalance(address string) *big.Int {
	val, ok := db.db[address]
	if !ok {
		return big.NewInt(0)
	}
	return val
}

func createLocalTxEvent(tx *types.Transaction) *common2.UncheckedTxEvent {
	return &common2.UncheckedTxEvent{
		EventType: common2.LocalTxEvent,
		Event: &common2.TxWithResp{
			Tx:      tx,
			CheckCh: make(chan *common2.TxResp),
			PoolCh:  make(chan *common2.TxResp),
		},
	}
}

func createRemoteTxEvent(txs []*types.Transaction) *common2.UncheckedTxEvent {
	return &common2.UncheckedTxEvent{
		EventType: common2.RemoteTxEvent,
		Event:     txs,
	}
}

func generateBatchTx(s *types.Signer, size, illegalIndex int) ([]*types.Transaction, error) {
	toAddr := common.HexToAddress(to)
	txs := make([]*types.Transaction, size)
	for i := 0; i < size; i++ {
		if i != illegalIndex {
			tx, err := generateLegacyTx(s, &toAddr, uint64(i), nil, uint64(basicGas), 1, big.NewInt(0))
			if err != nil {
				return nil, err
			}
			txs[i] = tx
		}
	}
	// illegal tx
	tx, err := generateLegacyTx(s, nil, uint64(illegalIndex), nil, uint64(basicGas+1), 1, big.NewInt(0))
	if err != nil {
		return nil, err
	}
	txs[illegalIndex] = tx

	return txs, nil
}

func generateLegacyTx(s *types.Signer, to *common.Address, nonce uint64, data []byte, gasLimit, gasPrice uint64, value *big.Int) (*types.Transaction, error) {
	inner := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(int64(gasPrice)),
		Gas:      gasLimit,
		To:       to,
		Data:     data,
		Value:    value,
	}
	tx := &types.Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.SignByTxType(s.Sk); err != nil {
		return nil, err
	}
	return tx, nil
}

func generateDynamicFeeTx(s *types.Signer, to *common.Address, data []byte,
	gasLimit uint64, value, gasFeeCap, gasTipCap *big.Int) (*types.Transaction, error) {
	inner := &types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     0,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        to,
		Data:      data,
		Value:     value,
	}
	tx := &types.Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.SignByTxType(s.Sk); err != nil {
		return nil, err
	}
	return tx, nil
}
