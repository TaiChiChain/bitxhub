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

// LiquidStakingTokenRate is an auto generated low-level Go binding around an user-defined struct.
type LiquidStakingTokenRate struct {
	StakeAmount              *big.Int
	LiquidStakingTokenAmount *big.Int
}

// PoolInfo is an auto generated low-level Go binding around an user-defined struct.
type PoolInfo struct {
	ID                                      uint64
	IsActive                                bool
	ActiveStake                             *big.Int
	TotalLiquidStakingToken                 *big.Int
	PendingActiveStake                      *big.Int
	PendingInactiveStake                    *big.Int
	PendingInactiveLiquidStakingTokenAmount *big.Int
	CommissionRate                          uint64
	NextEpochCommissionRate                 uint64
	LastEpochReward                         *big.Int
	LastEpochCommission                     *big.Int
	CumulativeReward                        *big.Int
	CumulativeCommission                    *big.Int
	OperatorLiquidStakingTokenID            *big.Int
	LastRateEpoch                           uint64
}

type StakingManager interface {

	// AddStake is a paid mutator transaction binding the contract method 0xad899a39.
	//
	// Solidity: function addStake(uint64 poolID, address owner, uint256 amount) payable returns()
	AddStake(poolID uint64, owner common.Address, amount *big.Int) error

	// BatchUnlock is a paid mutator transaction binding the contract method 0x49e091fa.
	//
	// Solidity: function batchUnlock(uint256[] liquidStakingTokenIDs, uint256[] amounts) returns()
	BatchUnlock(liquidStakingTokenIDs []*big.Int, amounts []*big.Int) error

	// BatchWithdraw is a paid mutator transaction binding the contract method 0x024b762c.
	//
	// Solidity: function batchWithdraw(uint256[] liquidStakingTokenIDs, address recipient, uint256[] amounts) returns()
	BatchWithdraw(liquidStakingTokenIDs []*big.Int, recipient common.Address, amounts []*big.Int) error

	// Unlock is a paid mutator transaction binding the contract method 0x5bfadb24.
	//
	// Solidity: function unlock(uint256 liquidStakingTokenID, uint256 amount) returns()
	Unlock(liquidStakingTokenID *big.Int, amount *big.Int) error

	// UpdatePoolCommissionRate is a paid mutator transaction binding the contract method 0xe1d7afb3.
	//
	// Solidity: function updatePoolCommissionRate(uint64 poolID, uint64 newCommissionRate) returns()
	UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error

	// Withdraw is a paid mutator transaction binding the contract method 0xe63697c8.
	//
	// Solidity: function withdraw(uint256 liquidStakingTokenID, address recipient, uint256 amount) returns()
	Withdraw(liquidStakingTokenID *big.Int, recipient common.Address, amount *big.Int) error

	// GetPoolHistoryLiquidStakingTokenRate is a free data retrieval call binding the contract method 0x222b3405.
	//
	// Solidity: function getPoolHistoryLiquidStakingTokenRate(uint64 poolID, uint64 epoch) view returns((uint256,uint256) poolHistoryLiquidStakingTokenRate)
	GetPoolHistoryLiquidStakingTokenRate(poolID uint64, epoch uint64) (LiquidStakingTokenRate, error)

	// GetPoolHistoryLiquidStakingTokenRates is a free data retrieval call binding the contract method 0x844f1bb2.
	//
	// Solidity: function getPoolHistoryLiquidStakingTokenRates(uint64[] poolIDs, uint64 epoch) view returns((uint256,uint256)[] poolHistoryLiquidStakingTokenRate)
	GetPoolHistoryLiquidStakingTokenRates(poolIDs []uint64, epoch uint64) ([]LiquidStakingTokenRate, error)

	// GetPoolInfo is a free data retrieval call binding the contract method 0xf2347366.
	//
	// Solidity: function getPoolInfo(uint64 poolID) view returns((uint64,bool,uint256,uint256,uint256,uint256,uint256,uint64,uint64,uint256,uint256,uint256,uint256,uint256,uint64) poolInfo)
	GetPoolInfo(poolID uint64) (PoolInfo, error)

	// GetPoolInfos is a free data retrieval call binding the contract method 0xba37e6d6.
	//
	// Solidity: function getPoolInfos(uint64[] poolIDs) view returns((uint64,bool,uint256,uint256,uint256,uint256,uint256,uint64,uint64,uint256,uint256,uint256,uint256,uint256,uint64)[] poolInfos)
	GetPoolInfos(poolIDs []uint64) ([]PoolInfo, error)
}

// EventAddStake represents a AddStake event raised by the StakingManager contract.
type EventAddStake struct {
	PoolID               uint64
	Owner                common.Address
	Amount               *big.Int
	LiquidStakingTokenID *big.Int
}

func (_event *EventAddStake) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["AddStake"])
}

// EventUnlock represents a Unlock event raised by the StakingManager contract.
type EventUnlock struct {
	LiquidStakingTokenID *big.Int
	Amount               *big.Int
	UnlockTimestamp      uint64
}

func (_event *EventUnlock) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Unlock"])
}

// EventWithdraw represents a Withdraw event raised by the StakingManager contract.
type EventWithdraw struct {
	LiquidStakingTokenID *big.Int
	Recipient            common.Address
	Amount               *big.Int
}

func (_event *EventWithdraw) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Withdraw"])
}

// ErrorAddStakeReachEpochLimit represents a AddStakeReachEpochLimit error raised by the StakingManager contract.
type ErrorAddStakeReachEpochLimit struct {
	Remain *big.Int
}

func (_error *ErrorAddStakeReachEpochLimit) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["AddStakeReachEpochLimit"])
}

// ErrorUnlockStakeReachEpochLimit represents a UnlockStakeReachEpochLimit error raised by the StakingManager contract.
type ErrorUnlockStakeReachEpochLimit struct {
	Remain *big.Int
}

func (_error *ErrorUnlockStakeReachEpochLimit) Pack(abi abi.ABI) error {
	return packer.PackError(_error, abi.Errors["UnlockStakeReachEpochLimit"])
}
