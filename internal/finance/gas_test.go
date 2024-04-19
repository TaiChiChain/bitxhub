package finance

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestGetGasPrice(t *testing.T) {
	// Gas should be less than 5000
	GasPriceBySize(t, 100, 5000000000000, nil)
	// Gas should be larger than 5000
	GasPriceBySize(t, 400, 5000000000000, nil)
	// Gas should be equals to 5000
	GasPriceBySize(t, 250, 5000000000000, nil)
	// Gas touch the ceiling
	GasPriceBySize(t, 400, 10000000000000, nil)
	// Gas touch the floor
	GasPriceBySize(t, 100, 1000000000000, nil)
	// Txs too much error
	GasPriceBySize(t, 700, 5000000000000, ErrTxsOutOfRange)
	// parent gas out of range error
	GasPriceBySize(t, 100, 11000000000000, ErrGasOutOfRange)
	// parent gas out of range error
	GasPriceBySize(t, 100, 900000000000, ErrGasOutOfRange)
}

func GasPriceBySize(t *testing.T, size int, parentGasPrice uint64, expectErr error) uint64 {
	// mock block for ledger
	block := &types.Block{
		Header:       &types.BlockHeader{GasPrice: parentGasPrice},
		Transactions: []*types.Transaction{},
	}
	prepareTxs := func(size int) []*types.Transaction {
		var txs []*types.Transaction
		for i := 0; i < size; i++ {
			txs = append(txs, &types.Transaction{})
		}
		return txs
	}
	block.Transactions = prepareTxs(size)
	rep := repo.MockRepo(t)
	chainState := chainstate.NewMockChainState(rep.GenesisConfig, nil)
	gasPrice := NewGas(chainState)
	gas, err := gasPrice.CalNextGasPrice(parentGasPrice, size)
	if expectErr != nil {
		assert.EqualError(t, err, expectErr.Error())
		return 0
	}
	assert.Nil(t, err)
	return checkResult(t, block, rep, parentGasPrice, gas)
}

func checkResult(t *testing.T, block *types.Block, config *repo.Repo, parentGasPrice uint64, gas uint64) uint64 {
	epochInfo := config.GenesisConfig.EpochInfo
	percentage := 2 * (float64(len(block.Transactions)) - float64(epochInfo.ConsensusParams.BlockMaxTxNum)/2) / float64(epochInfo.ConsensusParams.BlockMaxTxNum)
	actualGas := uint64(float64(parentGasPrice) * (1 + percentage*float64(epochInfo.FinanceParams.GasChangeRateValue)/math.Pow10(int(epochInfo.FinanceParams.GasChangeRateDecimals))))
	if actualGas > epochInfo.FinanceParams.MaxGasPrice.ToBigInt().Uint64() {
		actualGas = epochInfo.FinanceParams.MaxGasPrice.ToBigInt().Uint64()
	}
	if actualGas < config.GenesisConfig.EpochInfo.FinanceParams.MinGasPrice.ToBigInt().Uint64() {
		actualGas = epochInfo.FinanceParams.MinGasPrice.ToBigInt().Uint64()
	}
	assert.Equal(t, actualGas, gas, "Gas price is not correct")
	return gas
}
