package eth

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

var (
	ErrNotSupportApiError = errors.New("unsupported interface")
)

func getStateLedgerAt(api api.CoreAPI, blockNrOrHash *rpctypes.BlockNumberOrHash) (ledger.StateLedger, error) {
	var blockHeader *types.BlockHeader
	var err error

	meta, err := api.Chain().Meta()
	if err != nil {
		return nil, err
	}

	if blockNrOrHash != nil {
		if blockNumber, ok := blockNrOrHash.Number(); ok {
			if blockNumber == rpctypes.PendingBlockNumber || blockNumber == rpctypes.LatestBlockNumber {
				blockHeader, err = api.Broker().GetBlockHeaderByNumber(meta.Height)
				if err != nil {
					return nil, err
				}
			} else {
				blockHeader, err = api.Broker().GetBlockHeaderByNumber(uint64(blockNumber))
				if err != nil {
					return nil, err
				}
			}
		} else if blockHash, ok := blockNrOrHash.Hash(); ok {
			blockHeader, err = api.Broker().GetBlockHeaderByHash(types.NewHash(blockHash.Bytes()))
			if err != nil {
				return nil, err
			}
		}
	} else {
		// default case: use the latest committed blockHeader
		blockHeader, err = api.Broker().GetBlockHeaderByNumber(meta.Height)
		if err != nil {
			return nil, err
		}
	}

	enableSnapshot := blockHeader.Number == meta.Height
	lg, err := api.Broker().GetViewStateLedger().NewViewWithoutCache(blockHeader, enableSnapshot)
	if err != nil {
		return nil, fmt.Errorf("GetViewStateLedger error: %v", err)
	}
	return lg, nil
}

// NewRPCTransaction returns a transaction that will serialize to the RPC representation
func NewRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *rpctypes.RPCTransaction {
	from := common.BytesToAddress(tx.GetFrom().Bytes())
	var to *common.Address
	if tx.GetTo() != nil {
		toAddr := common.BytesToAddress(tx.GetTo().Bytes())
		to = &toAddr
	}
	v, r, s := tx.GetRawSignature()
	result := &rpctypes.RPCTransaction{
		Type:     hexutil.Uint64(tx.GetType()),
		From:     from,
		Gas:      hexutil.Uint64(tx.GetGas()),
		GasPrice: (*hexutil.Big)(tx.GetGasPrice()),
		Hash:     tx.GetHash().ETHHash(),
		Input:    hexutil.Bytes(tx.GetPayload()),
		Nonce:    hexutil.Uint64(tx.GetNonce()),
		To:       to,
		Value:    (*hexutil.Big)(tx.GetValue()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}

	switch tx.GetType() {
	case types.AccessListTxType:
		al := tx.GetInner().GetAccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.GetChainID())
	case types.DynamicFeeTxType:
		al := tx.GetInner().GetAccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.GetChainID())
		result.GasFeeCap = (*hexutil.Big)(tx.GetInner().GetGasFeeCap())
		result.GasTipCap = (*hexutil.Big)(tx.GetInner().GetGasTipCap())
		result.GasPrice = result.GasFeeCap
	case types.IncentiveTxType:
		al := tx.GetInner().GetAccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.GetChainID())
		result.GasFeeCap = (*hexutil.Big)(tx.GetInner().GetGasFeeCap())
		result.GasTipCap = (*hexutil.Big)(tx.GetInner().GetGasTipCap())
		result.GasPrice = result.GasFeeCap
		result.IncentiveAddress = *tx.Inner.GetIncentiveAddress()
	}
	return result
}

// GetBlockCumulativeGas returns the cumulative gas used on a block up to a given transaction index (inclusive)
func getBlockCumulativeGas(api api.CoreAPI, block *types.Block, idx uint64) (uint64, error) {
	var gasUsed uint64
	txs := block.Transactions

	for i := 0; i <= int(idx) && i < len(txs); i++ {
		receipt, err := api.Broker().GetReceipt(txs[i].GetHash())
		if err != nil {
			return 0, err
		}

		gasUsed += receipt.GasUsed
	}

	return gasUsed, nil
}
