/*
 * @Date: 2024-03-14 14:18:23
 * @LastEditors: levi9311 790890362@qq.com
 * @LastEditTime: 2024-05-28 14:43:09
 * @FilePath: /axiom-ledger/internal/coreapi/feed.go
 */
package coreapi

import (
	"time"

	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/event"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/coreapi/api"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type FeedAPI CoreAPI

var _ api.FeedAPI = (*FeedAPI)(nil)

var emptyTxFeed event.Feed

// todo: subscribe from txpool?
func (api *FeedAPI) SubscribeNewTxEvent(ch chan<- []*types.Transaction) event.Subscription {
	if api.axiomLedger.Repo.StartArgs.ReadonlyMode {
		return emptyTxFeed.Subscribe(ch)
	}
	return api.axiomLedger.Consensus.SubscribeTxEvent(ch)
}

func (api *FeedAPI) SubscribeNewBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return api.axiomLedger.BlockExecutor.SubscribeBlockEventForRemote(ch)
}

func (api *FeedAPI) SubscribeLogsEvent(ch chan<- []*types.EvmLog) event.Subscription {
	return api.axiomLedger.BlockExecutor.SubscribeLogsEvent(ch)
}

// TODO: check it
func (api *FeedAPI) BloomStatus() (uint64, uint64) {
	if !api.axiomLedger.Repo.Config.Ledger.EnableIndexer {
		return repo.BloomBitsBlocks, 0
	}
	sections, _, _ := api.axiomLedger.Indexer.Sections()
	return repo.BloomBitsBlocks, sections
}

const (
	// bloomServiceThreads is the number of goroutines used globally by an Ethereum
	// instance to service bloombits lookups for all running filters.
	// bloomServiceThreads = 16

	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	bloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	bloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	bloomRetrievalWait = time.Duration(0)
)

func (api *FeedAPI) ServiceFilter(session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, api.axiomLedger.BloomRequests)
	}
}
