package saccount

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
)

var _ interfaces.INonceManager = (*NonceManager)(nil)

// TODO save nonce
type NonceManager struct {
	// The next valid sequence number for a given nonce key
	nonceSequenceNumber map[ethcommon.Address]map[string]*big.Int
}

func NewNonceManager() *NonceManager {
	return &NonceManager{
		nonceSequenceNumber: make(map[ethcommon.Address]map[string]*big.Int),
	}
}

func (nm *NonceManager) GetNonce(sender ethcommon.Address, key *big.Int) *big.Int {
	ns, ok := nm.nonceSequenceNumber[sender]
	if ok {
		nonce, ok := ns[string(key.Bytes())]
		if ok {
			return nonce
		}
	} else {
		// initialize nonce
		nm.nonceSequenceNumber[sender] = make(map[string]*big.Int)
	}

	// key << 64
	newNonce := new(big.Int).Lsh(key, 64)
	nm.nonceSequenceNumber[sender][string(key.Bytes())] = newNonce
	return newNonce
}

// nolint
func (nm *NonceManager) incrementNonce(sender ethcommon.Address, key *big.Int) {
	nonce, ok := nm.nonceSequenceNumber[sender][string(key.Bytes())]
	if ok {
		nonce.Add(nonce, big.NewInt(1))

		// if nonce is out of range, reinit
		newNonce := new(big.Int).Lsh(key, 64)
		if nonce.Rsh(nonce, 64).Cmp(newNonce) == 1 {
			nonce.Set(newNonce)
		}
	}
}

func (nm *NonceManager) validateAndUpdateNonce(sender ethcommon.Address, nonce *big.Int) bool {
	key := nonce.Rsh(nonce, 64)
	n := nm.GetNonce(sender, key)

	if n.Cmp(nonce) == 0 {
		n.Add(n, big.NewInt(1))
		return true
	}

	return false
}
