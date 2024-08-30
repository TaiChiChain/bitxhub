package saccount

import (
	"math"
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func validateZeroNonce(t *testing.T, nm *NonceManager, sender ethcommon.Address, key *big.Int) {
	nonce, err := nm.GetNonce(sender, key)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nonce.Uint64())
}

func validateIncrementNonce(t *testing.T, nm *NonceManager, sender ethcommon.Address, key *big.Int) {
	nonce, err := nm.GetNonce(sender, key)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nonce.Uint64())

	nm.incrementNonce(sender, key)
	nonce, err = nm.GetNonce(sender, key)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), nonce.Uint64())
}

func validateUpdateNonce(t *testing.T, nm *NonceManager, sender ethcommon.Address, nonce *big.Int) {
	isValid, err := nm.validateAndUpdateNonce(sender, nonce)
	assert.Nil(t, err)
	assert.True(t, isValid)
	isValid, err = nm.validateAndUpdateNonce(sender, nonce)
	assert.Nil(t, err)
	assert.False(t, isValid)

	key := new(big.Int).Lsh(nonce, 64)
	nonce, _ = nm.GetNonce(sender, key)
	isValid, err = nm.validateAndUpdateNonce(sender, nonce)
	assert.Nil(t, err)
	assert.True(t, isValid)
}

func TestNonceManager_GetNonce(t *testing.T) {
	ethAddr := "0x1234567890123456789012345678901234567890"
	systemContractBase := common.SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract),
		EthAddress: ethcommon.HexToAddress(ethAddr),
		Abi:        abi.ABI{},
		Address:    types.NewAddressByStr(ethAddr),
	}
	nm := NewNonceManager(systemContractBase)

	sender := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")

	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)
	ctx := common.NewTestVMContext(lg.StateLedger, sender)
	nm.SetContext(ctx)

	validateZeroNonce(t, nm, sender, big.NewInt(0))
	validateZeroNonce(t, nm, sender, big.NewInt(1000))
	maxBigInt, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	validateZeroNonce(t, nm, sender, maxBigInt)
}

func TestNonceManager_incrementNonce(t *testing.T) {
	ethAddr := "0x1234567890123456789012345678901234567890"
	systemContractBase := common.SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract),
		EthAddress: ethcommon.HexToAddress(ethAddr),
		Abi:        abi.ABI{},
		Address:    types.NewAddressByStr(ethAddr),
	}
	nm := NewNonceManager(systemContractBase)

	sender := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")

	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)
	ctx := common.NewTestVMContext(lg.StateLedger, sender)
	nm.SetContext(ctx)

	validateIncrementNonce(t, nm, sender, big.NewInt(0))
	validateIncrementNonce(t, nm, sender, big.NewInt(9999))

	// add overflow
	key := new(big.Int).SetUint64(math.MaxUint64)
	nm.nonceSequenceNumber.Put(generateKey(sender, key), new(big.Int).SetUint64(math.MaxUint64))
	nonce, err := nm.GetNonce(sender, key)
	assert.Nil(t, err)
	assert.Equal(t, uint64(math.MaxUint64), nonce.Uint64())

	nm.incrementNonce(sender, key)
	nonce, err = nm.GetNonce(sender, key)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nonce.Uint64())
	assert.Equal(t, key.Uint64(), new(big.Int).Rsh(nonce, 64).Uint64())
}

func TestNonceManager_validateAndUpdateNonce(t *testing.T) {
	ethAddr := "0x1234567890123456789012345678901234567890"
	systemContractBase := common.SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract),
		EthAddress: ethcommon.HexToAddress(ethAddr),
		Abi:        abi.ABI{},
		Address:    types.NewAddressByStr(ethAddr),
	}
	nm := NewNonceManager(systemContractBase)

	sender := ethcommon.HexToAddress("0x82C6D3ed4cD33d8EC1E51d0B5Cc1d822Eaa0c3dC")

	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)
	ctx := common.NewTestVMContext(lg.StateLedger, sender)
	nm.SetContext(ctx)

	validateUpdateNonce(t, nm, sender, big.NewInt(0))
	key := big.NewInt(1000)
	nonce := new(big.Int).Lsh(key, 64)
	validateUpdateNonce(t, nm, sender, nonce)
}
