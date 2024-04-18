// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package empty

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

type Empty interface {
	Receive() error
}
