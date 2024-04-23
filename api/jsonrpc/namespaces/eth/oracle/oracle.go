// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package oracle

import (
	"math/big"
	"sync"

	types2 "github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const sampleNumber = 3 // Number of transactions sampled in a block

var (
	DefaultMaxPrice    = big.NewInt(500 * params.GWei)
	DefaultIgnorePrice = big.NewInt(2 * params.Wei)
	DefaultChainID     = uint64(1356)
	DefaultGasLimit    = uint64(50000000)
)

type Config struct {
	GasLimit         uint64
	ChainID          uint64
	Blocks           int
	Percentile       int
	MaxHeaderHistory uint64
	MaxBlockHistory  uint64
	Default          *big.Int `toml:",omitempty"`
	MaxPrice         *big.Int `toml:",omitempty"`
	IgnorePrice      *big.Int `toml:",omitempty"`
}

// Oracle recommends gas prices based on the content of recent
// blocks. Suitable for both light and full clients.
type Oracle struct {
	api         api.CoreAPI
	logger      logrus.FieldLogger
	gasLimit    uint64
	lastHead    *types2.Hash
	lastPrice   *big.Int
	maxPrice    *big.Int
	ignorePrice *big.Int
	chainID     uint64
	cacheLock   sync.RWMutex
	fetchLock   sync.Mutex

	checkBlocks, percentile           int
	maxHeaderHistory, maxBlockHistory uint64

	historyCache *lru.Cache[cacheKey, processedFees]
}

func generateConfig(rep *repo.Repo) Config {
	config := Config{
		GasLimit:         rep.GenesisConfig.EpochInfo.FinanceParams.GasLimit,
		ChainID:          rep.GenesisConfig.ChainID,
		Blocks:           2, // default traverse block nums
		Percentile:       60,
		MaxHeaderHistory: 300,
		MaxBlockHistory:  5,
		MaxPrice:         DefaultMaxPrice,
		IgnorePrice:      DefaultIgnorePrice,
	}
	//config := oracle.Config{
	//	GasLimit:         rep.GenesisConfig.EpochInfo.FinanceParams.GasLimit,
	//	ChainID:          rep.GenesisConfig.ChainID,
	//	Blocks:           20, // default traverse block nums
	//	Percentile:       60,
	//	MaxHeaderHistory: 1024,
	//	MaxBlockHistory:  1024,
	//	MaxPrice:         oracle.DefaultMaxPrice,
	//	IgnorePrice:      oracle.DefaultIgnorePrice,
	//}
	return config
}

// NewOracle returns a new gasprice oracle which can recommend suitable
// gasprice for newly created transaction.
func NewOracle(rep *repo.Repo, api api.CoreAPI, logger logrus.FieldLogger) *Oracle {
	config := generateConfig(rep)
	blocks := config.Blocks
	if blocks < 1 {
		blocks = 1
		logger.Warnf("Sanitizing invalid gasprice oracle sample blocks, provided: %d, updated: %d", config.Blocks, blocks)
	}
	percent := config.Percentile
	if percent < 0 {
		percent = 0
		logger.Warnf("Sanitizing invalid gasprice oracle sample percentile, provided: %d, updated: %d", config.Percentile, percent)
	} else if percent > 100 {
		percent = 100
		logger.Warnf("Sanitizing invalid gasprice oracle sample percentile, provided: %d, updated: %d", config.Percentile, percent)
	}
	maxPrice := config.MaxPrice
	if maxPrice == nil || maxPrice.Int64() <= 0 {
		maxPrice = DefaultMaxPrice
		logger.Warnf("Sanitizing invalid gasprice oracle price cap, provided: %d, updated: %d", config.MaxPrice, maxPrice)
	}
	ignorePrice := config.IgnorePrice
	if ignorePrice == nil || ignorePrice.Int64() <= 0 {
		ignorePrice = DefaultIgnorePrice
		logger.Warnf("Sanitizing invalid gasprice oracle ignore price, provided: %d, updated: %d", config.IgnorePrice, ignorePrice)
	} else if ignorePrice.Int64() > 0 {
		logger.Infof("Gasprice oracle is ignoring threshold set threshold: %d", ignorePrice)
	}
	maxHeaderHistory := config.MaxHeaderHistory
	if maxHeaderHistory < 1 {
		maxHeaderHistory = 1
		logger.Warnf("Sanitizing invalid gasprice oracle max header history, provided: %d, updated: %d", config.MaxHeaderHistory, maxHeaderHistory)
	}
	maxBlockHistory := config.MaxBlockHistory
	if maxBlockHistory < 1 {
		maxBlockHistory = 1
		logger.Warnf("Sanitizing invalid gasprice oracle max block history, provided: %d, updated: %d", config.MaxBlockHistory, maxBlockHistory)
	}
	chainID := config.ChainID
	if chainID <= 0 {
		chainID = DefaultChainID
		logger.Warnf("Sanitizing invalid gasprice oracle chain id, provided: %d, updated: %d", config.ChainID, chainID)
	} else if ignorePrice.Int64() > 0 {
		logger.Infof("Gasprice oracle is ignoring threshold set, threshold: %d", chainID)
	}
	gasLimit := config.GasLimit
	if gasLimit <= 0 {
		gasLimit = DefaultGasLimit
		logger.Warnf("Sanitizing invalid gasprice oracle gas limit, provided: %d, updated: %d", config.GasLimit, gasLimit)
	}

	cache := lru.NewCache[cacheKey, processedFees](2048)
	headEvent := make(chan events.ExecutedEvent, 1)
	api.Feed().SubscribeNewBlockEvent(headEvent)
	go func() {
		var lastHead *types2.Hash
		for ev := range headEvent {
			if ev.Block.Header.ParentHash != lastHead {
				cache.Purge()
			}
			lastHead = ev.Block.Header.Hash()
		}
	}()

	return &Oracle{
		api:              api,
		gasLimit:         gasLimit,
		logger:           logger,
		lastPrice:        config.Default,
		maxPrice:         maxPrice,
		ignorePrice:      ignorePrice,
		chainID:          chainID,
		checkBlocks:      blocks,
		percentile:       percent,
		maxHeaderHistory: maxHeaderHistory,
		maxBlockHistory:  maxBlockHistory,
		historyCache:     cache,
	}
}

// SuggestTipCap returns a tip cap so that newly created transaction can have a
// very high chance to be included in the following blocks.
//
// Note, for legacy transactions and the legacy eth_gasPrice RPC call, it will be
// necessary to add the basefee to the returned number to fall back to the legacy
// behavior.
func (oracle *Oracle) SuggestTipCap() (*big.Int, error) {
	meta, err := oracle.api.Chain().Meta()
	if err != nil {
		return nil, err
	}
	head, _ := oracle.api.Broker().GetBlockHeaderByNumber(meta.Height)
	headHash := head.Hash()

	// If the latest gasprice is still available, return it.
	oracle.cacheLock.RLock()
	lastHead, lastPrice := oracle.lastHead, oracle.lastPrice
	oracle.cacheLock.RUnlock()
	if headHash == lastHead {
		return new(big.Int).Set(lastPrice), nil
	}
	oracle.fetchLock.Lock()
	defer oracle.fetchLock.Unlock()

	// Try checking the cache again, maybe the last fetch fetched what we need
	oracle.cacheLock.RLock()
	lastHead, lastPrice = oracle.lastHead, oracle.lastPrice
	oracle.cacheLock.RUnlock()
	if headHash == lastHead {
		return new(big.Int).Set(lastPrice), nil
	}
	var (
		sent, exp int
		number    = head.Number
		result    = make(chan results, oracle.checkBlocks)
		quit      = make(chan struct{})
		results   []*big.Int
	)
	for sent < oracle.checkBlocks && number > 0 {
		go oracle.getBlockValues(number, sampleNumber, oracle.ignorePrice, result, quit)
		sent++
		exp++
		number--
	}
	for exp > 0 {
		res := <-result
		if res.err != nil {
			close(quit)
			return new(big.Int).Set(lastPrice), res.err
		}
		exp--
		// Nothing returned. There are two special cases here:
		// - The block is empty
		// - All the transactions included are sent by the miner itself.
		// In these cases, use the latest calculated price for sampling.
		if len(res.values) == 0 {
			res.values = []*big.Int{lastPrice}
		}
		// Besides, in order to collect enough data for sampling, if nothing
		// meaningful returned, try to query more blocks. But the maximum
		// is 2*checkBlocks.
		if len(res.values) == 1 && len(results)+1+exp < oracle.checkBlocks*2 && number > 0 {
			go oracle.getBlockValues(number, sampleNumber, oracle.ignorePrice, result, quit)
			sent++
			exp++
			number--
		}
		results = append(results, res.values...)
	}
	price := lastPrice
	if len(results) > 0 {
		slices.SortFunc(results, func(a, b *big.Int) int { return a.Cmp(b) })
		price = results[(len(results)-1)*oracle.percentile/100]
	}
	if price.Cmp(oracle.maxPrice) > 0 {
		price = new(big.Int).Set(oracle.maxPrice)
	}
	oracle.cacheLock.Lock()
	oracle.lastHead = headHash
	oracle.lastPrice = price
	oracle.cacheLock.Unlock()

	return new(big.Int).Set(price), nil
}

type results struct {
	values []*big.Int
	err    error
}

// getBlockValues calculates the lowest transaction gas price in a given block
// and sends it to the result channel. If the block is empty or all transactions
// are sent by the miner itself(it doesn't make any sense to include this kind of
// transaction prices for sampling), nil gasprice is returned.
func (oracle *Oracle) getBlockValues(blockNum uint64, limit int, ignoreUnder *big.Int, result chan results, quit chan struct{}) {
	block, err := oracle.api.Broker().GetBlockHeaderByNumber(blockNum)
	if block == nil {
		select {
		case result <- results{nil, err}:
		case <-quit:
		}
		return
	}

	// Sort the transaction by effective tip in ascending sort.
	// todo: handle the error
	txs, _ := oracle.api.Broker().GetBlockTxList(blockNum)
	sortedTxs := make([]*types2.Transaction, len(txs))
	copy(sortedTxs, txs)
	slices.SortFunc(sortedTxs, func(a, b *types2.Transaction) int {
		tip1 := a.GetGasTipCap()
		tip2 := b.GetGasTipCap()
		return tip1.Cmp(tip2)
	})

	var prices []*big.Int
	for _, tx := range sortedTxs {
		tip := tx.GetGasTipCap()
		if ignoreUnder != nil && tip.Cmp(ignoreUnder) == -1 {
			continue
		}
		prices = append(prices, tip)
		if len(prices) >= limit {
			break
		}
	}
	select {
	case result <- results{prices, nil}:
	case <-quit:
	}
}
