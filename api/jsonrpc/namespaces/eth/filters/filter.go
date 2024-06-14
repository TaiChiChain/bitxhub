// Copyright 2014 The go-ethereum Authors
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

package filters

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
)

// Filter can be used to retrieve and filter logs.
type Filter struct {
	api             api.CoreAPI
	addresses       []types.Address
	topics          [][]types.Hash
	block           *types.Hash // Block hash if filtering a single block
	begin           int64
	end             int64 // Range interval if filtering multiple blocks
	blockRangeLimit uint64
}

type bytesBacked interface {
	Bytes() []byte
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(api api.CoreAPI, begin, end int64, addresses []types.Address, topics [][]types.Hash, blockRangeLimit uint64) *Filter {
	// Create a generic filter and convert it into a range filter
	filter := newFilter(api, addresses, topics)

	filter.begin = begin
	filter.end = end
	filter.blockRangeLimit = blockRangeLimit

	return filter
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(api api.CoreAPI, block *types.Hash, addresses []types.Address, topics [][]types.Hash) *Filter {
	// Create a generic filter and convert it into a block filter
	filter := newFilter(api, addresses, topics)
	filter.block = block
	return filter
}

// newFilter creates a generic filter that can either filter based on a block hash,
// or based on range queries. The search criteria needs to be explicitly set.
func newFilter(api api.CoreAPI, addresses []types.Address, topics [][]types.Hash) *Filter {
	return &Filter{
		api:       api,
		addresses: addresses,
		topics:    topics,
	}
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(ctx context.Context) ([]*types.EvmLog, error) {
	// If we're doing singleton block filtering, execute and return
	if f.block != nil {
		blockHeader, err := f.api.Broker().GetBlockHeader("HASH", f.block.String())
		if err != nil {
			return nil, err
		}
		if blockHeader == nil {
			return nil, errors.New("unknown block")
		}
		return f.blockLogs(ctx, blockHeader)
	}
	// Figure out the limits of the filter range
	meta, err := f.api.Chain().Meta()
	if err != nil {
		return nil, err
	}

	head := meta.Height
	if f.begin == rpctypes.PendingBlockNumber.Int64() || f.begin == rpctypes.LatestBlockNumber.Int64() {
		f.begin = int64(head)
	}

	end := uint64(f.end)
	if f.end == rpctypes.PendingBlockNumber.Int64() || f.end == rpctypes.LatestBlockNumber.Int64() {
		end = head
	}
	if int64(end)-f.begin >= int64(f.blockRangeLimit) {
		return nil, fmt.Errorf("query block range needs to be less than or equal to %d", f.blockRangeLimit)
	}

	return f.unindexedLogs(ctx, end)
}

// unindexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *Filter) unindexedLogs(ctx context.Context, end uint64) ([]*types.EvmLog, error) {
	var logs []*types.EvmLog
	for ; f.begin <= int64(end); f.begin++ {
		blockHeader, err := f.api.Broker().GetBlockHeader("HEIGHT", fmt.Sprintf("%d", f.begin))
		if blockHeader == nil || err != nil {
			return logs, err
		}

		found, err := f.blockLogs(ctx, blockHeader)
		if err != nil {
			return logs, err
		}
		logs = append(logs, found...)
	}
	return logs, nil
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(ctx context.Context, header *types.BlockHeader) (logs []*types.EvmLog, err error) {
	if bloomFilter(header.Bloom, f.addresses, f.topics) {
		found, err := f.checkMatches(ctx, header.Number)
		if err != nil {
			return logs, err
		}
		logs = append(logs, found...)
	}
	return logs, nil
}

// checkMatches checks if the receipts belonging to the given header contain any log events that
// match the filter criteria. This function is called when the bloom filter signals a potential match.
func (f *Filter) checkMatches(_ context.Context, blockNum uint64) (logs []*types.EvmLog, err error) {
	// Get the logs of the block
	receipts, err := f.getBlockReceipts(blockNum)
	if err != nil {
		return nil, err
	}

	var unfiltered []*types.EvmLog
	for _, receipt := range receipts {
		unfiltered = append(unfiltered, receipt.EvmLogs...)
	}
	return FilterLogs(unfiltered, nil, nil, f.addresses, f.topics), nil
}

func (f *Filter) getBlockReceipts(blockNum uint64) ([]*types.Receipt, error) {
	var receipts []*types.Receipt

	blockHashList, err := f.api.Broker().GetBlockTxHashList(blockNum)
	if err != nil {
		return nil, err
	}

	for _, txHash := range blockHashList {
		receipt, err := f.api.Broker().GetReceipt(txHash)
		if err != nil {
			return nil, err
		}

		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

func includes(addresses []types.Address, a *types.Address) bool {
	// todo
	// temporary fix for panic caused by nil addresses or a
	if a == nil {
		return false
	}
	for _, addr := range addresses {
		if addr.ETHAddress() == a.ETHAddress() {
			return true
		}
	}

	return false
}

// FilterLogs creates a slice of logs matching the given criteria.
func FilterLogs(logs []*types.EvmLog, fromBlock, toBlock *big.Int, addresses []types.Address, topics [][]types.Hash) []*types.EvmLog {
	var ret []*types.EvmLog

	var cloneLogs []*types.EvmLog
	for _, log := range logs {
		cloneLogs = append(cloneLogs, &types.EvmLog{
			Address: types.NewAddress(log.Address.Bytes()),
			Topics: lo.Map(log.Topics, func(item *types.Hash, index int) *types.Hash {
				return types.NewHash(item.Bytes())
			}),
			Data:             log.Data,
			BlockNumber:      log.BlockNumber,
			TransactionHash:  types.NewHash(log.TransactionHash.Bytes()),
			TransactionIndex: log.TransactionIndex,
			BlockHash:        types.NewHash(log.BlockHash.Bytes()),
			LogIndex:         log.LogIndex,
			Removed:          log.Removed,
		})
	}

Logs:
	for _, log := range cloneLogs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}

		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i].String() == topic.String() {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func bloomFilter(bloom *types.Bloom, addresses []types.Address, topics [][]types.Hash) bool {
	if len(addresses) > 0 {
		var included bool
		for _, addr := range addresses {
			if BloomLookup(bloom, &addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if BloomLookup(bloom, &topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}

// BloomLookup is a convenience-method to check presence int he bloom filter
func BloomLookup(bin *types.Bloom, topic bytesBacked) bool {
	return bin.Test(topic.Bytes())
}
