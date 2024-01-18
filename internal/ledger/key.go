package ledger

import (
	"crypto/sha256"
	"fmt"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
)

const (
	blockKey           = "block-"
	blockHashKey       = "block-hash-"
	blockHeightKey     = "block-height-"
	blockTxSetKey      = "block-tx-set-"
	interchainMetaKey  = "interchain-meta-"
	transactionKey     = "tx-"
	transactionMetaKey = "tx-meta-"
	chainMetaKey       = "chain-meta"
	codeKey            = "code-"
	trieBlockKey       = "trie-block-"
	trieNodeInfoKey    = "trie-nodeInfo-"
)

func compositeKey(prefix string, value any) []byte {
	return append([]byte(prefix), []byte(fmt.Sprintf("%v", value))...)
}

func CompositeAccountKey(addr *types.Address) []byte {
	return hexutil.EncodeToNibbles(addr.String())
}

func CompositeStorageKey(addr *types.Address, key []byte) []byte {
	keyHash := sha256.Sum256(append(addr.Bytes(), key...))
	return hexutil.EncodeToNibbles(types.NewHash(keyHash[:]).String())
}

func compositeCodeKey(addr *types.Address, codeHash []byte) []byte {
	return append(addr.Bytes(), codeHash...)
}
