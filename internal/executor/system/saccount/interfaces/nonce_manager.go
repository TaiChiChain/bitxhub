package interfaces

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type INonceManager interface {
	// GetNonce return the next nonce for this sender.
	// Within a given key, the nonce values are sequenced (starting with zero, and incremented by one on each userop)
	// But UserOp with different keys can come with arbitrary order.
	GetNonce(sender ethcommon.Address, key *big.Int) (nonce *big.Int)
}
