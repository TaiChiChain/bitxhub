// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package smart_account

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
	PubKeyX *big.Int
	PubKeyY *big.Int
	Algo    uint8
}

// SessionKey is an auto generated low-level Go binding around an user-defined struct.
type SessionKey struct {
	Addr          common.Address
	SpendingLimit *big.Int
	SpentAmount   *big.Int
	ValidUntil    uint64
	ValidAfter    uint64
}

// UserOperation is an auto generated low-level Go binding around an user-defined struct.
type UserOperation struct {
	Sender               common.Address
	Nonce                *big.Int
	InitCode             []byte
	CallData             []byte
	CallGasLimit         *big.Int
	VerificationGasLimit *big.Int
	PreVerificationGas   *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	PaymasterAndData     []byte
	Signature            []byte
	AuthData             []byte
	ClientData           []byte
}

type SmartAccount interface {

	// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
	//
	// Solidity: function execute(address dest, uint256 value, bytes func) returns()
	Execute(dest common.Address, value *big.Int, arg2 []byte) error

	// ExecuteBatch is a paid mutator transaction binding the contract method 0x18dfb3c7.
	//
	// Solidity: function executeBatch(address[] dest, bytes[] func) returns()
	ExecuteBatch(dest []common.Address, arg1 [][]byte) error

	// ExecuteBatch0 is a paid mutator transaction binding the contract method 0x47e1da2a.
	//
	// Solidity: function executeBatch(address[] dest, uint256[] value, bytes[] func) returns()
	ExecuteBatch0(dest []common.Address, value []*big.Int, arg2 [][]byte) error

	// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
	//
	// Solidity: function initialize(address anOwner) returns()
	Initialize(anOwner common.Address) error

	// ResetOwner is a paid mutator transaction binding the contract method 0x73cc802a.
	//
	// Solidity: function resetOwner(address ) returns()
	ResetOwner(arg0 common.Address) error

	// SetGuardian is a paid mutator transaction binding the contract method 0x8a0dac4a.
	//
	// Solidity: function setGuardian(address ) returns()
	SetGuardian(arg0 common.Address) error

	// SetPasskey is a paid mutator transaction binding the contract method 0x39d8bab5.
	//
	// Solidity: function setPasskey(bytes , uint8 ) returns()
	SetPasskey(arg0 []byte, arg1 uint8) error

	// SetSession is a paid mutator transaction binding the contract method 0xa3ca5fe3.
	//
	// Solidity: function setSession(address , uint256 , uint64 , uint64 ) returns()
	SetSession(arg0 common.Address, arg1 *big.Int, arg2 uint64, arg3 uint64) error

	// ValidateUserOp is a paid mutator transaction binding the contract method 0xb0fff5ca.
	//
	// Solidity: function validateUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
	ValidateUserOp(userOp UserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*big.Int, error)

	// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
	//
	// Solidity: function entryPoint() view returns(address)
	EntryPoint() (common.Address, error)

	// GetGuardian is a free data retrieval call binding the contract method 0xa75b87d2.
	//
	// Solidity: function getGuardian() view returns(address)
	GetGuardian() (common.Address, error)

	// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
	//
	// Solidity: function getNonce() view returns(uint256)
	GetNonce() (*big.Int, error)

	// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
	//
	// Solidity: function getOwner() view returns(address)
	GetOwner() (common.Address, error)

	// GetPasskeys is a free data retrieval call binding the contract method 0xe4093c22.
	//
	// Solidity: function getPasskeys() view returns((uint256,uint256,uint8)[])
	GetPasskeys() ([]PassKey, error)

	// GetSessions is a free data retrieval call binding the contract method 0x61503e45.
	//
	// Solidity: function getSessions() view returns((address,uint256,uint256,uint64,uint64)[])
	GetSessions() ([]SessionKey, error)

	// GetStatus is a free data retrieval call binding the contract method 0x4e69d560.
	//
	// Solidity: function getStatus() view returns(uint64)
	GetStatus() (uint64, error)

	// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
	//
	// Solidity: function owner() view returns(address)
	Owner() (common.Address, error)

	Receive() error
}

// EventSmartAccountInitialized represents a SmartAccountInitialized event raised by the SmartAccount contract.
type EventSmartAccountInitialized struct {
	EntryPoint common.Address
	Owner      common.Address
}

func (_event *EventSmartAccountInitialized) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["SmartAccountInitialized"])
}

// EventUserLocked represents a UserLocked event raised by the SmartAccount contract.
type EventUserLocked struct {
	Sender     common.Address
	NewOwner   common.Address
	LockedTime *big.Int
}

func (_event *EventUserLocked) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["UserLocked"])
}
