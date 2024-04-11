package utils

import (
	"crypto/sha256"
	"fmt"
	"github.com/axiomesh/axiom-kit/hexutil"
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

func CompositeKey(prefix string, value any) []byte {
	return append([]byte(prefix), []byte(fmt.Sprintf("%v", value))...)
}

func CompositeAccountKey(addr *types.Address) []byte {
	return hexutil.EncodeToNibbles(addr.String())
}

func CompositeStorageKey(addr *types.Address, key []byte) []byte {
	keyHash := sha256.Sum256(append(addr.Bytes(), key...))
	return hexutil.EncodeToNibbles(types.NewHash(keyHash[:]).String())
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
