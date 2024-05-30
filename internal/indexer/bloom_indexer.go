package indexer

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core/bloombits"
	types2 "github.com/ethereum/go-ethereum/core/types"
)

const (
	// bloomThrottling is the time to wait between processing two consecutive index
	// sections. It's useful during chain upgrades to prevent disk overload.
	bloomThrottling = 100 * time.Millisecond
)

var (
	bloomBitsPrefix = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits
)

// BloomIndexer implements a core.ChainIndexer, building up a rotated bloom bits index
// for the Ethereum header bloom filters, permitting blazing fast filtering.
type BloomIndexer struct {
	size    uint64                    // section size to generate bloombits for
	db      *storagemgr.CachedStorage // database instance to write index data and metadata into
	gen     *bloombits.Generator      // generator to rotate the bloom bits crating the bloom index
	section uint64                    // Section is the section number being processed currently
	head    *types.Hash               // Head is the hash of the last header processed
}

// NewBloomIndexer returns a chain indexer that generates bloom bits data for the
// canonical chain for fast logs filtering.
func NewBloomIndexer(chainDb *ledger.Ledger, db *storagemgr.CachedStorage, size, confirms uint64) *ChainIndexer {
	backend := &BloomIndexer{
		db:   db,
		size: size,
	}

	return NewChainIndexer(chainDb, db, backend, size, confirms, bloomThrottling, "bloombits")
}

// Reset implements core.ChainIndexerBackend, starting a new bloombits index
// section.
func (b *BloomIndexer) Reset(ctx context.Context, section uint64, lastSectionHead *types.Hash) error {
	gen, err := bloombits.NewGenerator(uint(b.size))
	b.gen, b.section, b.head = gen, section, &types.Hash{}
	return err
}

// Process implements core.ChainIndexerBackend, adding a new header's bloom into
// the index.
func (b *BloomIndexer) Process(ctx context.Context, header *types.BlockHeader) error {
	err := b.gen.AddBloom(uint(header.Number-b.section*b.size), types2.BytesToBloom(header.Bloom[:]))
	if err != nil {
		return err
	}
	b.head = header.Hash()
	return nil
}

// Commit implements core.ChainIndexerBackend, finalizing the bloom section and
// writing it out into the database.
func (b *BloomIndexer) Commit() error {
	batch := b.db.NewBatch()
	for i := 0; i < types2.BloomBitLength; i++ {
		bits, err := b.gen.Bitset(uint(i))
		if err != nil {
			return err
		}
		batch.Put(bloomBitsKey(uint(i), b.section, b.head), bitutil.CompressBytes(bits))
	}
	batch.Commit()

	return nil
}

// Prune returns an empty error since we don't support pruning here.
func (b *BloomIndexer) Prune(threshold uint64) error {
	return nil
}

// bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash
func bloomBitsKey(bit uint, section uint64, hash *types.Hash) []byte {
	key := append(append(bloomBitsPrefix, make([]byte, 10)...), hash.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	return key
}
