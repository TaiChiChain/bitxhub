package indexer

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/blockfile"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/storage/kv/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainIndexerSingle(t *testing.T) {
	for i := 0; i < 10; i++ {
		testChainIndexer(t, 1)
	}
}

// Runs multiple tests with randomized parameters and different number of
// chain backends.
func TestChainIndexerWithChildren(t *testing.T) {
	for i := 2; i < 8; i++ {
		testChainIndexer(t, i)
	}
}

// testChainIndexer runs a test with either a single chain indexer or a chain of
// multiple backends. The section size and required confirmation count parameters
// are randomized.
func testChainIndexer(t *testing.T, count int) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	lSnapshotStorage, err := leveldb.New(filepath.Join(repoRoot, "lSnapshot"), nil)
	assert.Nil(t, err)

	logger := log.NewWithModule("indexer_test")
	blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, "indexer"), logger)
	assert.Nil(t, err)
	l, err := ledger.NewLedgerWithStores(createMockRepo(t), lBlockStorage, lStateStorage, lSnapshotStorage, blockFile)
	require.Nil(t, err)
	require.NotNil(t, l)

	c := storagemgr.NewCachedStorage(kv.NewMemory(), 10).(*storagemgr.CachedStorage)

	// Create a chain of indexers and ensure they all report empty
	backends := make([]*testChainIndexBackend, count)
	for i := 0; i < count; i++ {
		var (
			sectionSize = uint64(rand.Intn(20) + 1)
			confirmsReq = uint64(rand.Intn(5))
		)
		backends[i] = &testChainIndexBackend{t: t, processCh: make(chan uint64)}
		backends[i].indexer = NewChainIndexer(l, c, backends[i], sectionSize, confirmsReq, 0, fmt.Sprintf("indexer-%d", i))

		if sections, _, _ := backends[i].indexer.Sections(); sections != 0 {
			t.Fatalf("Canonical section count mismatch: have %v, want %v", sections, 0)
		}
		if i > 0 {
			backends[i-1].indexer.AddChildIndexer(backends[i].indexer)
		}
	}
	defer backends[0].indexer.Close() // parent indexer shuts down children
	// notify pings the root indexer about a new head or reorg, then expect
	// processed blocks if a section is processable
	// notify := func(headNum, failNum uint64, reorg bool) {
	// 	backends[0].indexer.newHead(headNum, reorg)
	// 	if reorg {
	// 		for _, backend := range backends {
	// 			headNum = backend.reorg(headNum)
	// 			backend.assertSections()
	// 		}
	// 		return
	// 	}
	// 	var cascade bool
	// 	for _, backend := range backends {
	// 		headNum, cascade = backend.assertBlocks(headNum, failNum)
	// 		if !cascade {
	// 			break
	// 		}
	// 		backend.assertSections()
	// 	}
	// }
	// inject inserts a new random canonical header into the database directly
	inject := func(number uint64) {
		header := &types.BlockHeader{Number: number}
		if number > 0 {
			prevHeader, err := l.ChainLedger.GetBlockHeader(number - 1)
			assert.Nil(t, err)
			header.ParentHash = prevHeader.Hash()
		}
		l.ChainLedger.PersistExecutionResult(&types.Block{Header: header}, []*types.Receipt{})
	}
	// Start indexer with an already existing chain
	for i := uint64(0); i <= 30; i++ {
		fmt.Println("inject: ", i)
		inject(i)
	}
	// notify(30, 30, false)

}

// testChainIndexBackend implements ChainIndexerBackend
type testChainIndexBackend struct {
	t                          *testing.T
	indexer                    *ChainIndexer
	section, headerCnt, stored uint64
	processCh                  chan uint64
}

// assertSections verifies if a chain indexer has the correct number of section.
func (b *testChainIndexBackend) assertSections() {
	// Keep trying for 3 seconds if it does not match
	var sections uint64
	for i := 0; i < 6000; i++ {
		sections, _, _ = b.indexer.Sections()
		if sections == b.stored {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	b.t.Fatalf("Canonical section count mismatch: have %v, want %v", sections, b.stored)
}

// assertBlocks expects processing calls after new blocks have arrived. If the
// failNum < headNum then we are simulating a scenario where a reorg has happened
// after the processing has started and the processing of a section fails.
func (b *testChainIndexBackend) assertBlocks(headNum, failNum uint64) (uint64, bool) {
	var sections uint64
	if headNum >= b.indexer.confirmsReq {
		sections = (headNum + 1 - b.indexer.confirmsReq) / b.indexer.sectionSize
		if sections > b.stored {
			// expect processed blocks
			for expectd := b.stored * b.indexer.sectionSize; expectd < sections*b.indexer.sectionSize; expectd++ {
				if expectd > failNum {
					// rolled back after processing started, no more process calls expected
					// wait until updating is done to make sure that processing actually fails
					var updating bool
					for i := 0; i < 300; i++ {
						b.indexer.lock.Lock()
						updating = b.indexer.knownSections > b.indexer.storedSections
						b.indexer.lock.Unlock()
						if !updating {
							break
						}
						time.Sleep(100 * time.Millisecond)
					}
					if updating {
						b.t.Fatalf("update did not finish")
					}
					sections = expectd / b.indexer.sectionSize
					break
				}
				select {
				case <-time.After(10 * time.Second):
					b.t.Fatalf("Expected processed block #%d, got nothing", expectd)
				case processed := <-b.processCh:
					if processed != expectd {
						b.t.Errorf("Expected processed block #%d, got #%d", expectd, processed)
					}
				}
			}
			b.stored = sections
		}
	}
	if b.stored == 0 {
		return 0, false
	}
	return b.stored*b.indexer.sectionSize - 1, true
}

func (b *testChainIndexBackend) reorg(headNum uint64) uint64 {
	firstChanged := (headNum + 1) / b.indexer.sectionSize
	if firstChanged < b.stored {
		b.stored = firstChanged
	}
	return b.stored * b.indexer.sectionSize
}

func (b *testChainIndexBackend) Reset(ctx context.Context, section uint64, prevHead *types.Hash) error {
	b.section = section
	b.headerCnt = 0
	return nil
}

func (b *testChainIndexBackend) Process(ctx context.Context, header *types.BlockHeader) error {
	b.headerCnt++
	if b.headerCnt > b.indexer.sectionSize {
		b.t.Error("Processing too many headers")
	}
	//t.processCh <- header.Number.Uint64()
	select {
	case <-time.After(10 * time.Second):
		b.t.Error("Unexpected call to Process")
		// Can't use Fatal since this is not the test's goroutine.
		// Returning error stops the chainIndexer's updateLoop
		return errors.New("Unexpected call to Process")
	case b.processCh <- header.Number:
	}
	return nil
}

func (b *testChainIndexBackend) Commit() error {
	if b.headerCnt != b.indexer.sectionSize {
		b.t.Error("Not enough headers processed")
	}
	return nil
}

func (b *testChainIndexBackend) Prune(threshold uint64) error {
	return nil
}

func createMockRepo(t *testing.T) *repo.Repo {
	return repo.MockRepo(t)
}
