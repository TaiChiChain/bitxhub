package txpool

import (
	"math"
	"math/big"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	log2 "github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
)

// nolint
const (
	DefaultTestBatchSize = uint64(4)
	defaultGasPrice      = 1000
	defaultGasLimit      = 21000
)

// nolint
var to = types.NewAddressByStr("0xa2f28344131970356c4a112d1e634e51589aa57c")

// nolint
func mockTxPoolImpl[T any, Constraint types.TXConstraint[T]](t *testing.T) *txPoolImpl[T, Constraint] {
	ast := assert.New(t)
	pool, err := newTxPoolImpl[T, Constraint](NewMockTxPoolConfig(t))
	ast.Nil(err)
	conf := txpool.ConsensusConfig{
		SelfID: 1,
		NotifyGenerateBatchFn: func(typ int) {
			// do nothing
		},
	}
	pool.Init(conf)
	return pool
}

// NewMockTxPoolConfig returns the default test config
func NewMockTxPoolConfig(t *testing.T) Config {
	dir := t.TempDir()
	log := log2.NewWithModule("txpool")
	log.Logger.SetLevel(logrus.DebugLevel)
	poolConfig := Config{
		BatchSize:             DefaultTestBatchSize,
		PoolSize:              DefaultPoolSize,
		Logger:                log,
		ToleranceTime:         DefaultToleranceTime,
		CleanEmptyAccountTime: DefaultCleanEmptyAccountTime,
		GetAccountNonce: func(address string) uint64 {
			return 0
		},
		GetAccountBalance: func(address string) *big.Int {
			return big.NewInt(math.MaxInt64)
		},
		IsTimed:                false,
		RotateTxLocalsInterval: DefaultRotateTxLocalsInterval,
		EnableLocalsPersist:    true,
		TxRecordsFile:          path.Join(dir, "txpool/tx_records.pb"),

		ChainInfo: &txpool.ChainInfo{
			Height:   1,
			GasPrice: big.NewInt(defaultGasPrice),
		},
	}
	return poolConfig
}

// nolint
func constructTx(s *types.Signer, nonce uint64) *types.Transaction {
	tx, err := types.GenerateTransactionWithSigner(nonce, to, big.NewInt(0), nil, s)
	if err != nil {
		panic(err)
	}
	return tx
}

// nolint
func constructTxs(s *types.Signer, count int) []*types.Transaction {
	txs := make([]*types.Transaction, count)
	for i := 0; i < count; i++ {
		tx, err := types.GenerateTransactionWithSigner(uint64(i), to, big.NewInt(0), nil, s)
		if err != nil {
			panic(err)
		}
		txs[i] = tx
	}
	return txs
}
