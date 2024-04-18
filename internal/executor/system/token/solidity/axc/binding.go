// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package axc

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

type Axc interface {

	// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
	//
	// Solidity: function totalSupply() view returns(uint256)
	TotalSupply() (*big.Int, error)
}
