package saccount

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
)

var _ interfaces.INonceManager = (*NonceManager)(nil)

type NonceManager struct {
	common.SystemContractBase

	// The next valid sequence number for a given nonce key
	nonceSequenceNumber *common.VMMap[string, *big.Int]
}

func NewNonceManager(systemContractBase common.SystemContractBase) *NonceManager {
	return &NonceManager{
		SystemContractBase: systemContractBase,
	}
}

func (nm *NonceManager) SetContext(context *common.VMContext) {
	nm.SystemContractBase.SetContext(context)

	nm.nonceSequenceNumber = common.NewVMMap[string, *big.Int](nm.StateAccount, "nonce_manager", func(key string) string { return key })
}

// nolint
func (nm *NonceManager) GetNonce(sender ethcommon.Address, key *big.Int) (*big.Int, error) {
	k := generateKey(sender, key)
	exist, nonce, err := nm.nonceSequenceNumber.Get(k)
	if err != nil {
		return nil, err
	}
	if exist {
		return nonce, nil
	}

	// key << 64
	newNonce := new(big.Int).Lsh(key, 64)
	if err := nm.nonceSequenceNumber.Put(k, newNonce); err != nil {
		return nil, err
	}
	return newNonce, nil
}

// nolint
func (nm *NonceManager) incrementNonce(sender ethcommon.Address, key *big.Int) {
	nonce, err := nm.GetNonce(sender, key)
	if err != nil {
		nm.Logger.Errorf("incrementNonce error: %v", err)
		return
	}

	nonce.Add(nonce, big.NewInt(1))

	// if nonce is out of range, reinit
	newNonce := new(big.Int).Lsh(key, 64)
	if new(big.Int).Rsh(nonce, 64).Cmp(key) != 0 {
		nonce.Set(newNonce)
	}
	k := generateKey(sender, key)
	nm.nonceSequenceNumber.Put(k, nonce)
}

// nolint
func (nm *NonceManager) validateAndUpdateNonce(sender ethcommon.Address, nonce *big.Int) (bool, error) {
	key := new(big.Int).Rsh(nonce, 64)
	n, err := nm.GetNonce(sender, key)
	if err != nil {
		return false, err
	}
	if n.Cmp(nonce) == 0 {
		nm.incrementNonce(sender, key)
		return true, nil
	}

	return false, nil
}

func generateKey(sender ethcommon.Address, key *big.Int) string {
	return fmt.Sprintf("%s_%s", string(sender.Bytes()), string(key.Bytes()))
}
