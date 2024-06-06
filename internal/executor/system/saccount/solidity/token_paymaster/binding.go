// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package token_paymaster

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
}

type TokenPaymaster interface {

	// AddToken is a paid mutator transaction binding the contract method 0x5476bd72.
	//
	// Solidity: function addToken(address token, address oracle) returns()
	AddToken(token common.Address, oracle common.Address) error

	// PostOp is a paid mutator transaction binding the contract method 0xa9a23409.
	//
	// Solidity: function postOp(uint8 mode, bytes context, uint256 actualGasCost) returns()
	PostOp(mode uint8, context []byte, actualGasCost *big.Int) error

	// ValidatePaymasterUserOp is a paid mutator transaction binding the contract method 0xf465c77e.
	//
	// Solidity: function validatePaymasterUserOp((address,uint256,bytes,bytes,uint256,uint256,uint256,uint256,uint256,bytes,bytes) userOp, bytes32 userOpHash, uint256 maxCost) returns(bytes context, uint256 validationData)
	ValidatePaymasterUserOp(userOp UserOperation, userOpHash [32]byte, maxCost *big.Int) ([]byte, *big.Int, error)

	// GetToken is a free data retrieval call binding the contract method 0x59770438.
	//
	// Solidity: function getToken(address token) view returns(address)
	GetToken(token common.Address) (common.Address, error)
}
