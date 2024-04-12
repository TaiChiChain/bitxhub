package utils

import (
	"crypto/sha256"
	"fmt"
	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"strconv"

	"github.com/axiomesh/axiom-kit/types"
)

const (
	BlockHashKey       = "block-hash-"
	BlockTxSetKey      = "block-tx-set-"
	TransactionMetaKey = "tx-meta-"
	ChainMetaKey       = "chain-meta"
	TrieBlockHeaderKey = "trie-block-"
	TrieNodeInfoKey    = "trie-nodeInfo-"
	PruneJournalKey    = "prune-nodeInfo-"
	TrieNodeIdKey      = "trie-nodeId-"
	SnapshotKey        = "snap-"
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
