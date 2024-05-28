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

	"github.com/axiomesh/axiom-kit/types"
	rpctypes "github.com/axiomesh/axiom-ledger/api/jsonrpc/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/ethereum/go-ethereum/core/bloombits"
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
	matcher         *bloombits.Matcher
}

type bytesBacked interface {
	Bytes() []byte
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(api api.CoreAPI, begin, end int64, addresses []types.Address, topics [][]types.Hash, blockRangeLimit uint64) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter
	// system. Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filters [][][]byte
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filters = append(filters, filter)
	}
	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filters = append(filters, filter)
	}
	size, _ := api.Feed().BloomStatus()

	// Create a generic filter and convert it into a range filter
	filter := newFilter(api, addresses, topics)

	filter.matcher = bloombits.NewMatcher(size, filters)
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
		blockHeader, err := f.api.Broker().GetBlockHeaderByHash(f.block)
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
	var logs []*types.EvmLog
	size, sections := f.api.Feed().BloomStatus()
	fmt.Println("=============================")
	fmt.Println(sections)
	fmt.Println("=============================")

	if indexed := sections * size; indexed > uint64(f.begin) {
		if indexed > end {
			logs, err = f.indexedLogs(ctx, end)
		} else {
			logs, err = f.indexedLogs(ctx, indexed-1)
		}
		if err != nil {
			return logs, err
		}
	}

	return f.unindexedLogs(ctx, end)
}

// indexedLogs returns the logs matching the filter criteria based on the bloom
// bits indexed available locally or via the network.
func (f *Filter) indexedLogs(ctx context.Context, end uint64) ([]*types.EvmLog, error) {
	// Create a matcher session and request servicing from the backend
	matches := make(chan uint64, 64)

	session, err := f.matcher.Start(ctx, uint64(f.begin), end, matches)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	f.api.Feed().ServiceFilter(session)

	// Iterate over the matches until exhausted or context closed
	var logs []*types.EvmLog

	for {
		select {
		case number, ok := <-matches:
			// Abort if all matches have been fulfilled
			if !ok {
				err := session.Error()
				if err == nil {
					f.begin = int64(end) + 1
				}
				return logs, err
			}
			f.begin = int64(number) + 1

			// Retrieve the suggested block and pull any truly matching logs
			header, err := f.api.Broker().GetBlockHeaderByNumber(number)
			if header == nil || err != nil {
				return logs, err
			}
			found, err := f.checkMatches(ctx, number)
			if err != nil {
				return logs, err
			}
			logs = append(logs, found...)

		case <-ctx.Done():
			return logs, ctx.Err()
		}
	}
}

// unindexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *Filter) unindexedLogs(ctx context.Context, end uint64) ([]*types.EvmLog, error) {
	var logs []*types.EvmLog
	for ; f.begin <= int64(end); f.begin++ {
		blockHeader, err := f.api.Broker().GetBlockHeaderByNumber(uint64(f.begin))
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
		if addr.String() == a.String() {
			return true
		}
	}

	return false
}

// FilterLogs creates a slice of logs matching the given criteria.
func FilterLogs(logs []*types.EvmLog, fromBlock, toBlock *big.Int, addresses []types.Address, topics [][]types.Hash) []*types.EvmLog {
	var ret []*types.EvmLog
Logs:
	for _, log := range logs {
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
