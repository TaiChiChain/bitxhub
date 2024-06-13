// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package verifying_paymaster

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

type VerifyingPaymaster interface {

	// PostOp is a paid mutator transaction binding the contract method 0xa9a23409.
	//
	// Solidity: function postOp(uint8 mode, bytes context, uint256 actualGasCost) returns()
	PostOp(mode uint8, context []byte, actualGasCost *big.Int) error

	// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
	//
	// Solidity: function transferOwnership(address newOwner) returns()
	TransferOwnership(newOwner common.Address) error

	// ValidatePaymasterUserOp is a paid mutator transaction binding the contract method 0x7ba19417.
	//
	// Solidity: function validatePaymasterUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes,bytes,bytes) userOp, bytes32 userOpHash, uint256 maxCost) returns(bytes context, uint256 validationData)
	ValidatePaymasterUserOp(userOp UserOperation, userOpHash [32]byte, maxCost *big.Int) ([]byte, *big.Int, error)

	// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
	//
	// Solidity: function owner() view returns(address)
	Owner() (common.Address, error)
}

// EventOwnershipTransferred represents a OwnershipTransferred event raised by the VerifyingPaymaster contract.
type EventOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
}

func (_event *EventOwnershipTransferred) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["OwnershipTransferred"])
}

// ErrorOwnableInvalidOwner represents a OwnableInvalidOwner error raised by the VerifyingPaymaster contract.
type ErrorOwnableInvalidOwner struct {
	Owner common.Address
}

func (_error *ErrorOwnableInvalidOwner) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["OwnableInvalidOwner"])
}

// ErrorOwnableUnauthorizedAccount represents a OwnableUnauthorizedAccount error raised by the VerifyingPaymaster contract.
type ErrorOwnableUnauthorizedAccount struct {
	Account common.Address
}

func (_error *ErrorOwnableUnauthorizedAccount) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["OwnableUnauthorizedAccount"])
}
