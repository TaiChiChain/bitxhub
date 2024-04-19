// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package staking_manager

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

type StakingManager interface {

	// Stake is a paid mutator transaction binding the contract method 0xa328a677.
	//
	// Solidity: function stake(uint64 poolID, address owner, uint256 amount) payable returns()
	Stake(poolID uint64, owner common.Address, amount *big.Int) error

	// Unlock is a paid mutator transaction binding the contract method 0x7a94f25b.
	//
	// Solidity: function unlock(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
	Unlock(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) error

	// Withdraw is a paid mutator transaction binding the contract method 0xcdfc4800.
	//
	// Solidity: function withdraw(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
	Withdraw(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) error
}

// EventStake represents a Stake event raised by the StakingManager contract.
type EventStake struct {
	PoolID uint64
	Owner  common.Address
	Amount *big.Int
}

func (_event *EventStake) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Stake"])
}

// EventUnlock represents a Unlock event raised by the StakingManager contract.
type EventUnlock struct {
	PoolID               uint64
	Owner                common.Address
	LiquidStakingTokenID *big.Int
	Amount               *big.Int
}

func (_event *EventUnlock) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Unlock"])
}

// EventWithdraw represents a Withdraw event raised by the StakingManager contract.
type EventWithdraw struct {
	PoolID               uint64
	Owner                common.Address
	LiquidStakingTokenID *big.Int
	Amount               *big.Int
}

func (_event *EventWithdraw) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Withdraw"])
}
