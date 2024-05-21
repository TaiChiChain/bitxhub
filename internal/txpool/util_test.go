package txpool

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	log2 "github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
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
	r := repo.MockRepo(t)
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.EpochInfo.FinanceParams.MinGasPrice = types.CoinNumberByMol(0)
	pool, err := newTxPoolImpl[T, Constraint](NewMockTxPoolConfig(t), chainState)
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

func mockTxPoolImplWithTyp[T any, Constraint types.TXConstraint[T]](t *testing.T, typ string) *txPoolImpl[T, Constraint] {
	ast := assert.New(t)
	poolConf := NewMockTxPoolConfig(t)
	poolConf.GenerateBatchType = typ
	r := repo.MockRepo(t)
	chainState := chainstate.NewMockChainState(r.GenesisConfig, nil)
	chainState.EpochInfo.FinanceParams.MinGasPrice = types.CoinNumberByMol(defaultGasPrice)

	pool, err := newTxPoolImpl[T, Constraint](poolConf, chainState)
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
		RepoRoot:              dir,
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
		RotateTxLocalsInterval: DefaultRotateTxLocalsInterval,
		EnableLocalsPersist:    true,
		PriceLimit:             0,
		PriceBump:              DefaultPriceBump,
		GenerateBatchType:      repo.GenerateBatchByTime,

		ChainInfo: &txpool.ChainInfo{
			Height:   1,
			GasPrice: big.NewInt(defaultGasPrice),
			EpochConf: &txpool.EpochConfig{
				EnableGenEmptyBatch: false,
				BatchSize:           DefaultTestBatchSize,
			},
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

func constructPoolTxByGas(s *types.Signer, nonce uint64, gasPrice *big.Int) *internalTransaction[types.Transaction, *types.Transaction] {
	inner := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      21000,
		Value:    big.NewInt(0),
		Data:     nil,
	}
	t := to.ETHAddress()
	inner.To = &t
	tx := &types.Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.Sign(s.Sk); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Millisecond)
	poolTx := &internalTransaction[types.Transaction, *types.Transaction]{
		rawTx:       tx,
		local:       true,
		lifeTime:    time.Now().Unix(),
		arrivedTime: time.Now().Unix(),
	}
	return poolTx
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

func constructPoolTxListByGas(s *types.Signer, count int, gasPrice *big.Int) []*internalTransaction[types.Transaction, *types.Transaction] {
	poolTxs := make([]*internalTransaction[types.Transaction, *types.Transaction], count)
	for i := 0; i < count; i++ {
		poolTxs[i] = constructPoolTxByGas(s, uint64(i), gasPrice)
	}
	return poolTxs
}

func getPrioritySize(pool *txPoolImpl[types.Transaction, *types.Transaction]) int {
	if pool.enablePricePriority {
		return int(pool.txStore.priorityByPrice.size())
	}
	return pool.txStore.priorityByTime.size()
}
