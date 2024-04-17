package finance

import (
	"math"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
)

type Gas struct {
	chainState *chainstate.ChainState
	logger     logrus.FieldLogger
}

func NewGas(chainState *chainstate.ChainState) *Gas {
	logger := loggers.Logger(loggers.Finance)
	return &Gas{chainState: chainState, logger: logger}
}

// CalNextGasPrice returns the current block gas price, based on the formula:
//
// G_c = G_p * (1 + (0.5 * (MaxBatchSize - TxsCount) / MaxBatchSize))
//
// if G_c <= minGasPrice, G_c = minGasPrice
//
// if G_c >= maxGasPrice, G_c = maxGasPrice
func (gas *Gas) CalNextGasPrice(parentGasPrice uint64, txs int) (uint64, error) {
	// TODO: use big.Int
	currentEpoch := gas.chainState.EpochInfo
	max := currentEpoch.FinanceParams.MaxGasPrice.ToBigInt().Uint64()
	min := currentEpoch.FinanceParams.MinGasPrice.ToBigInt().Uint64()
	if parentGasPrice < min || parentGasPrice > max {
		gas.logger.Errorf("gas price is out of range, parent gas price is %d, min is %d, max is %d", parentGasPrice, min, max)
		return 0, ErrGasOutOfRange
	}
	total := int(gas.chainState.EpochInfo.ConsensusParams.BlockMaxTxNum)
	if txs > total {
		return 0, ErrTxsOutOfRange
	}
	percentage := 2 * float64(txs-total/2) / float64(total) * float64(currentEpoch.FinanceParams.GasChangeRateValue) / math.Pow10(int(currentEpoch.FinanceParams.GasChangeRateDecimals))
	currentPrice := uint64(float64(parentGasPrice) * (1 + percentage))
	if currentPrice > max {
		gas.logger.Warningf("gas price is touching ceiling, current price is %d, max is %d", currentPrice, max)
		currentPrice = max
	}
	if currentPrice < min {
		gas.logger.Warningf("gas price is touching floor, current price is %d, min is %d", currentPrice, min)
		currentPrice = min
	}
	return currentPrice, nil
}
