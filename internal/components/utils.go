package components

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core"

	"github.com/axiomesh/axiom-kit/types"
)

func VerifyInsufficientBalance[T any, Constraint types.TXConstraint[T]](tx *T, getBalanceFn func(address string) *big.Int) error {
	// 1. account has enough balance to cover transaction fee(gaslimit * gasprice), gasprice is the chain's latest gas price
	txGasLimit := Constraint(tx).RbftGetGasLimit()
	txValue := Constraint(tx).RbftGetValue()
	txFrom := Constraint(tx).RbftGetFrom()
	txTo := Constraint(tx).RbftGetTo()
	txData := Constraint(tx).RbftGetData()
	txAccessList := Constraint(tx).RbftGetAccessList()

	mgval := new(big.Int).SetUint64(txGasLimit)
	mgval = mgval.Mul(mgval, Constraint(tx).RbftGetGasPrice())
	balanceCheck := mgval
	balanceRemaining := new(big.Int).Set(getBalanceFn(txFrom))
	if have, want := balanceRemaining, balanceCheck; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", core.ErrInsufficientFunds, txFrom, have, want)
	}

	// sub gas fee temporarily
	balanceRemaining.Sub(balanceRemaining, mgval)

	gasRemaining := txGasLimit

	var isContractCreation bool
	if txTo == "" {
		isContractCreation = true
	}

	// 2.1 the purchased gas is enough to cover intrinsic usage
	// 2.2 there is no overflow when calculating intrinsic gas
	gas, err := core.IntrinsicGas(txData, txAccessList.ToEthAccessList(), isContractCreation, true, true, true)
	if err != nil {
		return err
	}
	if gasRemaining < gas {
		return fmt.Errorf("%w: have %d, want %d", core.ErrIntrinsicGas, gasRemaining, gas)
	}

	// 3. account has enough balance to cover asset transfer for **topmost** call
	if txValue.Sign() > 0 && balanceRemaining.Cmp(txValue) < 0 {
		return fmt.Errorf("%w: address %v", core.ErrInsufficientFundsForTransfer, txFrom)
	}
	return nil
}
