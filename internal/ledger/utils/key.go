package utils

import (
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
)

const (
	BlockHashKey       = "block-hash-"
	BlockTxSetKey      = "block-tx-set-"
	TransactionMetaKey = "tx-meta-"
	ChainMetaKey       = "chain-meta"
	PruneJournalKey    = "prune-nodeInfo-"
	SnapshotKey        = "snap-"
	SnapshotMetaKey    = "snap-meta"
	RollbackBlockKey   = "rollback-block"
	RollbackStateKey   = "rollback-state"
)

const (
	MinHeightStr = "minHeight"
	MaxHeightStr = "maxHeight"
)

var keyCache = storagemgr.NewCacheWrapper(64, false)

func CompositeKey(prefix string, value any) []byte {
	return append([]byte(prefix), []byte(fmt.Sprintf("%v", value))...)
}

func CompositeAccountKey(addr *types.Address) []byte {
	k := addr.Bytes()
	if res, ok := keyCache.Get(k); ok {
		return res
	}
	res := hexutil.EncodeToNibbles(addr.String())
	keyCache.Set(k, res)
	return res
}

func CompositeStorageKey(addr *types.Address, key []byte) []byte {
	k := append(addr.Bytes(), key...)
	if res, ok := keyCache.Get(k); ok {
		return res
	}
	keyHash := sha256.Sum256(k)
	res := hexutil.EncodeToNibbles(types.NewHash(keyHash[:]).String())
	keyCache.Set(k, res)
	return res
}

func CompositeCodeKey(addr *types.Address, codeHash []byte) []byte {
	return append(addr.Bytes(), codeHash...)
}

func MarshalHeight(height uint64) []byte {
	return []byte(strconv.FormatUint(height, 10))
}

func UnmarshalHeight(data []byte) uint64 {
	height, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		panic(err)
	}

	return height
}
