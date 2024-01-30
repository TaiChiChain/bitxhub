package utils

import (
	"crypto/sha256"
	"fmt"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
)

const (
	BlockHashKey       = "block-hash-"
	BlockHeightKey     = "block-height-"
	BlockTxSetKey      = "block-tx-set-"
	InterChainMetaKey  = "interchain-meta-"
	TransactionMetaKey = "tx-meta-"
	ChainMetaKey       = "chain-meta"
	TrieBlockKey       = "trie-block-"
	TrieNodeInfoKey    = "trie-nodeInfo-"
	SnapshotKey        = "snap-"
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
