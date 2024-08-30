// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package smart_account_proxy

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
	_ = packer.RevertError{}
)

// PassKey is an auto generated low-level Go binding around an user-defined struct.
type PassKey struct {
	PublicKey []byte
	Algo      uint8
}

// SessionKey is an auto generated low-level Go binding around an user-defined struct.
type SessionKey struct {
	Addr          common.Address
	SpendingLimit *big.Int
	SpentAmount   *big.Int
	ValidUntil    uint64
	ValidAfter    uint64
}

type SmartAccountProxy interface {

	// GetGuardian is a free data retrieval call binding the contract method 0xb301d902.
	//
	// Solidity: function getGuardian(address ) view returns(address)
	GetGuardian(arg0 common.Address) (common.Address, error)

	// GetOwner is a free data retrieval call binding the contract method 0xfa544161.
	//
	// Solidity: function getOwner(address ) view returns(address)
	GetOwner(arg0 common.Address) (common.Address, error)

	// GetPasskeys is a free data retrieval call binding the contract method 0xbc6f15ea.
	//
	// Solidity: function getPasskeys(address ) view returns((bytes,uint8)[])
	GetPasskeys(arg0 common.Address) ([]PassKey, error)

	// GetSessions is a free data retrieval call binding the contract method 0x13625e2e.
	//
	// Solidity: function getSessions(address ) view returns((address,uint256,uint256,uint64,uint64)[])
	GetSessions(arg0 common.Address) ([]SessionKey, error)

	// GetStatus is a free data retrieval call binding the contract method 0x30ccebb5.
	//
	// Solidity: function getStatus(address ) view returns(uint64)
	GetStatus(arg0 common.Address) (uint64, error)
}
