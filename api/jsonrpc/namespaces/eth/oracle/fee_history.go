package oracle

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"golang.org/x/exp/slices"

	"github.com/axiomesh/axiom-kit/hexutil"
	types2 "github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
)

var (
	errInvalidPercentile = errors.New("invalid reward percentile")
	errRequestBeyondHead = errors.New("request beyond head block")
)

const (
	// maxBlockFetchers is the max number of goroutines to spin up to pull blocks
	// for the fee history calculation (mostly relevant for LES).
	maxBlockFetchers = 4
)

// blockFees represents a single block for processing
type blockFees struct {
	// set by the caller
	blockNumber uint64

	header       *types2.BlockHeader
	transactions []*types2.Transaction // only set if reward percentiles are requested
	receipts     []*types2.Receipt

	// filled by processBlock
	results processedFees

	err error
}

type cacheKey struct {
	number      uint64
	percentiles string
}

// processedFees contains the results of a processed block.
type processedFees struct {
	reward               []*hexutil.Big
	baseFee, nextBaseFee *hexutil.Big
	gasUsedRatio         float64
}

// txGasAndReward is sorted in ascending order based on reward
type txGasAndReward struct {
	gasUsed uint64
	reward  *big.Int
}

// processBlock takes a blockFees structure with the blockNumber, the header and optionally
// the block field filled in, retrieves the block from the backend if not present yet and
// fills in the rest of the fields.
func (oracle *Oracle) processBlock(bf *blockFees, percentiles []float64) {
	bf.results.baseFee = new(hexutil.Big)
	bf.results.nextBaseFee = new(hexutil.Big)
	bf.results.gasUsedRatio = float64(bf.header.GasUsed) / float64(oracle.gasLimit)
	if len(percentiles) == 0 {
		// rewards were not requested, return null
		return
	}
	if bf.transactions == nil || (bf.receipts == nil && len(bf.transactions) != 0) {
		oracle.logger.Error("Block or receipts are missing while reward percentiles are requested")
		return
	}

	bf.results.reward = make([]*hexutil.Big, len(percentiles))
	if len(bf.transactions) == 0 {
		// return an all zero row if there are no transactions to gather data from
		for i := range bf.results.reward {
			bf.results.reward[i] = new(hexutil.Big)
		}
		return
	}

	sorter := make([]txGasAndReward, len(bf.transactions))
	for i, tx := range bf.transactions {
		reward := tx.GetGasTipCap()
		sorter[i] = txGasAndReward{gasUsed: bf.receipts[i].GasUsed, reward: reward}
	}
	slices.SortStableFunc(sorter, func(a, b txGasAndReward) int {
		return a.reward.Cmp(b.reward)
	})

	var txIndex int
	sumGasUsed := sorter[0].gasUsed

	for i, p := range percentiles {
		thresholdGasUsed := uint64(float64(bf.header.GasUsed) * p / 100)
		for sumGasUsed < thresholdGasUsed && txIndex < len(bf.transactions)-1 {
			txIndex++
			sumGasUsed += sorter[txIndex].gasUsed
		}
		bf.results.reward[i] = (*hexutil.Big)(sorter[txIndex].reward)
	}
}

// resolveBlockRange resolves the specified block range to absolute block numbers while also
// enforcing backend specific limitations. The pending block and corresponding receipts are
// also returned if requested and available.
// Note: an error is only returned if retrieving the head header has failed. If there are no
// retrievable blocks in the specified range then zero block count is returned with no error.
func (oracle *Oracle) resolveBlockRange(reqEnd rpctypes.BlockNumber, blocks uint64) (uint64, uint64, error) {
	var (
		headBlock *types2.BlockHeader
		err       error
	)

	// Get the chain's current head.
	meta, err := oracle.api.Chain().Meta()
	if err != nil {
		return 0, 0, err
	}
	if headBlock, err = oracle.api.Broker().GetBlockHeaderByNumber(meta.Height); err != nil {
		return 0, 0, err
	}
	head := rpctypes.BlockNumber(headBlock.Number)

	// Fail if request block is beyond the chain's current head.
	if head < reqEnd {
		return 0, 0, fmt.Errorf("%w: requested %d, head %d", errRequestBeyondHead, reqEnd, head)
	}

	// Resolve block tag.
	if reqEnd < 0 {
		var (
			resolved *types2.BlockHeader
		)
		if reqEnd == rpctypes.LatestBlockNumber || reqEnd == rpctypes.PendingBlockNumber {
			resolved = headBlock
		} else {
			return 0, 0, errors.New("unsupported block number")
		}
		// Absolute number resolved.
		reqEnd = rpctypes.BlockNumber(resolved.Number)
	}

	// If there are no blocks to return, short circuit.
	if blocks == 0 {
		return 0, 0, nil
	}
	// Ensure not trying to retrieve before genesis.
	if uint64(reqEnd+1) < blocks {
		blocks = uint64(reqEnd)
	} else {
		blocks--
	}
	return uint64(reqEnd), blocks, nil
}

type FeeHistoryResult struct {
	OldestBlock  rpctypes.BlockNumber `json:"oldestBlock"`
	Reward       [][]*hexutil.Big     `json:"reward,omitempty"`
	BaseFee      []*hexutil.Big       `json:"baseFeePerGas,omitempty"`
	GasUsedRatio []float64            `json:"gasUsedRatio"`
}

// feeHistory returns data relevant for fee estimation based on the specified range of blocks.
// The range can be specified either with absolute block numbers or ending with the latest
// or pending block. Backends may or may not support gathering data from the pending block
// or blocks older than a certain age (specified in maxHistory). The first block of the
// actually processed range is returned to avoid ambiguity when parts of the requested range
// are not available or when the head has changed during processing this request.
// Three arrays are returned based on the processed blocks:
//   - reward: the requested percentiles of effective priority fees per gas of transactions in each
//     block, sorted in ascending order and weighted by gas used.
//   - baseFee: base fee per gas in the given block
//   - gasUsedRatio: gasUsed/gasLimit in the given block
//
// Note: baseFee includes the next block after the newest of the returned range, because this
// value can be derived from the newest block.
// nolint
func (oracle *Oracle) feeHistory(blocks uint64, unresolvedLastBlock rpctypes.BlockNumber, rewardPercentiles []float64) (*FeeHistoryResult, error) {
	oracle.logger.Debug("eth_feeHistory")
	zeroRet := &FeeHistoryResult{
		OldestBlock: 0,
	}
	if blocks < 1 {
		return zeroRet, nil // returning with no data and no error means there are no retrievable blocks
	}
	maxFeeHistory := oracle.maxHeaderHistory
	if len(rewardPercentiles) != 0 {
		maxFeeHistory = oracle.maxBlockHistory
	}
	if blocks > maxFeeHistory {
		oracle.logger.Warnf("Sanitizing fee history length, requested: %d, truncated: %d", blocks, maxFeeHistory)
		blocks = maxFeeHistory
	}
	for i, p := range rewardPercentiles {
		if p < 0 || p > 100 {
			return zeroRet, fmt.Errorf("%w: %f", errInvalidPercentile, p)
		}
		if i > 0 && p < rewardPercentiles[i-1] {
			return zeroRet, fmt.Errorf("%w: #%d:%f > #%d:%f", errInvalidPercentile, i-1, rewardPercentiles[i-1], i, p)
		}
	}
	lastBlock, blocks, err := oracle.resolveBlockRange(unresolvedLastBlock, blocks)
	if err != nil || blocks == 0 {
		return zeroRet, err
	}
	oldestBlock := lastBlock + 1 - blocks

	var next atomic.Uint64
	next.Store(oldestBlock)
	results := make(chan *blockFees, blocks)

	percentileKey := make([]byte, 8*len(rewardPercentiles))
	for i, p := range rewardPercentiles {
		binary.LittleEndian.PutUint64(percentileKey[i*8:(i+1)*8], math.Float64bits(p))
	}
	for i := 0; i < maxBlockFetchers && i < int(blocks); i++ {
		go func() {
			for {
				// Retrieve the next block number to fetch with this goroutine
				blockNumber := next.Add(1) - 1
				if blockNumber > lastBlock {
					return
				}

				fees := &blockFees{blockNumber: blockNumber}
				cacheKey := cacheKey{number: blockNumber, percentiles: string(percentileKey)}

				if p, ok := oracle.historyCache.Get(cacheKey); ok {
					fees.results = p
					results <- fees
				} else {
					if len(rewardPercentiles) != 0 {
						fees.header, fees.err = oracle.api.Broker().GetBlockHeaderByNumber(blockNumber)
						if fees.header != nil && fees.err == nil {
							fees.transactions, fees.err = oracle.api.Broker().GetBlockTxList(blockNumber)
							if fees.transactions != nil && fees.err == nil {
								fees.receipts, fees.err = oracle.api.Broker().GetReceipts(fees.header.Number)
							}
						}
					} else {
						fees.header, fees.err = oracle.api.Broker().GetBlockHeaderByNumber(blockNumber)
					}
					if fees.header != nil && fees.err == nil {
						oracle.processBlock(fees, rewardPercentiles)
						if fees.err == nil {
							oracle.historyCache.Add(cacheKey, fees.results)
						}
					}
					// send to results even if empty to guarantee that blocks items are sent in total
					results <- fees
				}
			}
		}()
	}
	var (
		reward       = make([][]*hexutil.Big, blocks)
		baseFee      = make([]*hexutil.Big, blocks+1)
		gasUsedRatio = make([]float64, blocks)
		firstMissing = blocks
	)
	for ; blocks > 0; blocks-- {
		fees := <-results
		if fees.err != nil {
			return zeroRet, fees.err
		}
		i := fees.blockNumber - oldestBlock
		if fees.results.baseFee != nil {
			reward[i], baseFee[i], baseFee[i+1], gasUsedRatio[i] = fees.results.reward, fees.results.baseFee, fees.results.nextBaseFee, fees.results.gasUsedRatio
		} else {
			// getting no block and no error means we are requesting into the future (might happen because of a reorg)
			if i < firstMissing {
				firstMissing = i
			}
		}
	}
	if firstMissing == 0 {
		return zeroRet, nil
	}
	if len(rewardPercentiles) != 0 {
		reward = reward[:firstMissing]
	} else {
		reward = nil
	}
	baseFee, gasUsedRatio = baseFee[:firstMissing+1], gasUsedRatio[:firstMissing]
	return &FeeHistoryResult{
		OldestBlock:  rpctypes.BlockNumber(oldestBlock),
		Reward:       reward,
		BaseFee:      baseFee,
		GasUsedRatio: gasUsedRatio,
	}, nil
}

func (oracle *Oracle) FeeHistory(blockCount ethmath.HexOrDecimal64, lastBlock rpctypes.BlockNumber, rewardPercentiles []float64) (*FeeHistoryResult, error) {
	return oracle.feeHistory(uint64(blockCount), lastBlock, rewardPercentiles)
}
