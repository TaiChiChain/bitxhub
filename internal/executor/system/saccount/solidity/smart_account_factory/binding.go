// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package smart_account_factory

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/bind"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = bind.Bind
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
)

type SmartAccountFactory interface {

	// CreateAccount is a paid mutator transaction binding the contract method 0x5fbfb9cf.
	//
	// Solidity: function createAccount(address owner, uint256 salt) returns(address ret)
	CreateAccount(owner common.Address, salt *big.Int) (common.Address, error)

	// AccountImplementation is a free data retrieval call binding the contract method 0x11464fbe.
	//
	// Solidity: function accountImplementation() view returns(address)
	AccountImplementation() (common.Address, error)

	// GetAddress is a free data retrieval call binding the contract method 0x8cb84e18.
	//
	// Solidity: function getAddress(address owner, uint256 salt) view returns(address)
	GetAddress(owner common.Address, salt *big.Int) (common.Address, error)
}
