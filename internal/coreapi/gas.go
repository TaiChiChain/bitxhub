package coreapi

import (
	"context"
	"math/big"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
)

type GasAPI CoreAPI

var _ api.ChainAPI = (*ChainAPI)(nil)

func (gas *GasAPI) GetGasPrice() (uint64, error) {
	gasPrice := gas.axiomLedger.ViewLedger.ChainLedger.GetChainMeta().GasPrice
	return gasPrice.Uint64(), nil
}

func (gas *GasAPI) GetCurrentGasPrice(blockHeight uint64) (uint64, error) {
	// since all block header records the gas price of the next block, so here we need to dec to get the current block's gas price, genesis block will return its own gas price
	if blockHeight != 1 {
		blockHeight--
	}
	blockHeader, err := gas.axiomLedger.ViewLedger.ChainLedger.GetBlockHeader(blockHeight)
	if err != nil {
		return 0, err
	}
	return uint64(blockHeader.GasPrice), nil
}

// TODO: support gas price oracle (refer to https://github.com/ethereum/go-ethereum/tree/master/eth/gasprice)

func (gas *GasAPI) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return nil, errors.Errorf("not implemented")
}

func (gas *GasAPI) FeeHistory(ctx context.Context, blockCount uint64, lastBlock uint64, rewardPercentiles []float64) (*big.Int, [][]*big.Int, []*big.Int, []float64, error) {
	return nil, nil, nil, nil, errors.Errorf("not implemented")
}
