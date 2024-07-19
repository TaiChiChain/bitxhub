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

	// AddDeposit is a paid mutator transaction binding the contract method 0x4a58db19.
	//
	// Solidity: function addDeposit() payable returns()
	AddDeposit() error

	// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
	//
	// Solidity: function execute(address dest, uint256 value, bytes func) returns()
	Execute(dest common.Address, value *big.Int, arg2 []byte) error

	// ExecuteBatch is a paid mutator transaction binding the contract method 0x18dfb3c7.
	//
	// Solidity: function executeBatch(address[] dest, bytes[] func) returns()
	ExecuteBatch(dest []common.Address, arg1 [][]byte) error

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

	// WithdrawDepositTo is a paid mutator transaction binding the contract method 0x4d44560d.
	//
	// Solidity: function withdrawDepositTo(address withdrawAddress, uint256 amount) returns()
	WithdrawDepositTo(withdrawAddress common.Address, amount *big.Int) error

	// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
	//
	// Solidity: function entryPoint() view returns(address)
	EntryPoint() (common.Address, error)

	// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
	//
	// Solidity: function getDeposit() view returns(uint256)
	GetDeposit() (*big.Int, error)

	// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
	//
	// Solidity: function getNonce() view returns(uint256)
	GetNonce() (*big.Int, error)

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
