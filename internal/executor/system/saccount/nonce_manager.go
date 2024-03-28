package saccount

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

var _ interfaces.INonceManager = (*NonceManager)(nil)

type NonceManager struct {
	// The next valid sequence number for a given nonce key
	nonceSequenceNumber *common.VMMap[string, *big.Int]
}

func NewNonceManager() *NonceManager {
	return &NonceManager{}
}

func (nm *NonceManager) Init(account ledger.IAccount) {
	nm.nonceSequenceNumber = common.NewVMMap[string, *big.Int](account, "nonce_manager", func(key string) string { return key })
}

// nolint
func (nm *NonceManager) GetNonce(sender ethcommon.Address, key *big.Int) *big.Int {
	k := generateKey(sender, key)
	exist, nonce, _ := nm.nonceSequenceNumber.Get(k)
	if exist {
		return nonce
	}

	// key << 64
	newNonce := new(big.Int).Lsh(key, 64)
	nm.nonceSequenceNumber.Put(k, newNonce)
	return newNonce
}

// nolint
func (nm *NonceManager) incrementNonce(sender ethcommon.Address, key *big.Int) {
	k := generateKey(sender, key)
	exist, nonce, _ := nm.nonceSequenceNumber.Get(k)
	if exist {
		nonce.Add(nonce, big.NewInt(1))

		// if nonce is out of range, reinit
		newNonce := new(big.Int).Lsh(key, 64)
		if nonce.Rsh(nonce, 64).Cmp(newNonce) == 1 {
			nonce.Set(newNonce)
		}
		nm.nonceSequenceNumber.Put(k, nonce)
	}
}

// nolint
func (nm *NonceManager) validateAndUpdateNonce(sender ethcommon.Address, nonce *big.Int) bool {
	key := new(big.Int).Rsh(nonce, 64)
	n := nm.GetNonce(sender, key)

	if n.Cmp(nonce) == 0 {
		n.Add(n, big.NewInt(1))
		k := generateKey(sender, key)
		nm.nonceSequenceNumber.Put(k, n)
		return true
	}

	return false
}

func generateKey(sender ethcommon.Address, key *big.Int) string {
	return fmt.Sprintf("%s_%s", string(sender.Bytes()), string(key.Bytes()))
}
