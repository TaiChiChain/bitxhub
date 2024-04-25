package framework

import (
	"fmt"
	"math/big"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
)

const (
	CommissionRateDenominator                         = 10000
	StakingPoolNamespace                              = "stakingPool"
	StakingPoolInfoKeySuffix                          = "info"
	StakingPoolLiquidStakingTokenRateHistoryKeyPrefix = "liquidStakingTokenRateHistory"
)

type LiquidStakingTokenRate struct {
	StakeAmount                   *big.Int
	TotalLiquidStakingTokenAmount *big.Int
}

type StakingPoolInfo struct {
	PoolID                       uint64
	IsActive                     bool
	ActiveStake                  *big.Int
	TotalLiquidStakingToken      *big.Int
	PendingActiveStake           *big.Int
	PendingInactiveStake         *big.Int
	CommissionRate               uint64
	NextEpochCommissionRate      uint64
	CumulativeReward             *big.Int
	OperatorLiquidStakingTokenID *big.Int
}

type StakingPool struct {
	common.SystemContractBase

	poolID                           uint64
	info                             *common.VMSlot[StakingPoolInfo]
	HistoryLiquidStakingTokenRateMap *common.VMMap[uint64, LiquidStakingTokenRate]
}

func NewStakingPool(systemContractBase common.SystemContractBase) *StakingPool {
	return &StakingPool{
		SystemContractBase: systemContractBase,
	}
}

func (sp *StakingPool) Load(poolID uint64) *StakingPool {
	sp.poolID = poolID
	sp.info = common.NewVMSlot[StakingPoolInfo](sp.StateAccount, fmt.Sprintf("%s_%d_%s", StakingPoolNamespace, poolID, StakingPoolInfoKeySuffix))
	sp.HistoryLiquidStakingTokenRateMap = common.NewVMMap[uint64, LiquidStakingTokenRate](sp.StateAccount, fmt.Sprintf("%s_%d_%s", StakingPoolNamespace, poolID, StakingPoolLiquidStakingTokenRateHistoryKeyPrefix), func(key uint64) string {
		return strconv.FormatUint(key, 10)
	})
	return sp
}

func (sp *StakingPool) Create(poolID uint64, commissionRate uint64) error {
	zero := big.NewInt(0)

	stakingPoolInfo := &StakingPoolInfo{
		PoolID:                       poolID,
		IsActive:                     true,
		ActiveStake:                  zero,
		TotalLiquidStakingToken:      zero,
		PendingActiveStake:           zero,
		PendingInactiveStake:         zero,
		CommissionRate:               commissionRate,
		NextEpochCommissionRate:      commissionRate,
		CumulativeReward:             zero,
		OperatorLiquidStakingTokenID: zero,
	}

	return sp.createByStakingPool(stakingPoolInfo)
}

func (sp *StakingPool) createByStakingPool(stakingPoolInfo *StakingPoolInfo) error {
	sp.Load(stakingPoolInfo.PoolID)

	if sp.info.Has() {
		return errors.Errorf("staking pool %d already exists", stakingPoolInfo.PoolID)
	}

	if stakingPoolInfo.CommissionRate > CommissionRateDenominator {
		return errors.Errorf("invalid commission rate: %d, must be less than or equal %d", stakingPoolInfo.CommissionRate, CommissionRateDenominator)
	}

	return sp.info.Put(*stakingPoolInfo)
}

func (sp *StakingPool) Exists() bool {
	return sp.info.Has()
}

func (sp *StakingPool) MustLoadInfo() (*StakingPoolInfo, error) {
	exist, info, err := sp.info.Get()
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.Errorf("staking pool %d does not exist", sp.poolID)
	}

	return &info, nil
}

// AddStake Add stake and mint an LiquidStakingToken,
// pendingActive LiquidStakingToken can quick Withdraw and not need call UnlockStake
// set PendingActiveStake = [PendingActiveStake + amount] if pool is active
// set ActiveStake = [ActiveStake + amount] if pool is Candidate
// mint a LiquidStakingToken, set LiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
// emit a AddStake event
// emit a MintLiquidStakingToken event
func (sp *StakingPool) AddStake(owner ethcommon.Address, amount *big.Int) error {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return err
	}

	if !info.IsActive {
		return errors.Errorf("staking pool %d is inactive", info.PoolID)
	}

	// TODO: check received value is equal to amount, implement it after geth upgrade

	nodeManagerContract := NodeManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	nodeInfo, err := nodeManagerContract.GetNodeInfo(info.PoolID)
	if err != nil {
		return err
	}

	// add stake to pool
	switch types.NodeStatus(nodeInfo.Status) {
	case types.NodeStatusCandidate:
		info.ActiveStake = new(big.Int).Add(info.ActiveStake, amount)
	case types.NodeStatusActive:
		info.PendingActiveStake = new(big.Int).Add(info.PendingActiveStake, amount)
	default:
		return errors.Errorf("staking pool %d status does not support adding stake", info.PoolID)
	}

	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}

	// mint nft for owner
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	tokenID, err := liquidStakingTokenContract.InternalMint(owner, &LiquidStakingTokenInfo{
		PoolID:           info.PoolID,
		Principal:        amount,
		Unlocked:         new(big.Int),
		ActiveEpoch:      currentEpoch.Epoch + 1,
		UnlockingRecords: []UnlockingRecord{},
	})
	if err != nil {
		return err
	}

	sp.EmitEvent(&staking_manager.EventAddStake{
		PoolID:               info.PoolID,
		Owner:                owner,
		Amount:               amount,
		LiquidStakingTokenID: tokenID,
	})

	return nil
}

// UnlockStake Unlock the specified amount and enter the lockin period and add a record to UnlockingRecords to liquidStakingToken, will refresh the LiquidStakingToken.ActiveEpoch and LiquidStakingToken.Principal(first calculate reward)
// NOTICE: Unable to obtain reward after calling unlock, and will revert when the LiquidStakingToken is pendingActive
// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
// set PendingInactiveStake = [PendingInactiveStake + amount]
// append UnlockingRecord{Amount: amount, UnlockTimestamp: block.Timestamp} to LiquidStakingToken.UnlockingRecords
func (sp *StakingPool) UnlockStake(owner ethcommon.Address, liquidStakingTokenID *big.Int, amount *big.Int) error {
	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}

	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	liquidStakingTokenInfo, err := liquidStakingTokenContract.mustGetInfo(liquidStakingTokenID)
	if err != nil {
		return err
	}

	// token is not active, use Withdraw instead of Unlock
	if liquidStakingTokenInfo.ActiveEpoch > currentEpoch.Epoch {
		return errors.Errorf("token is not active yet, use `WithdrawStake` instead")
	}

	if liquidStakingTokenInfo.ActiveEpoch == currentEpoch.Epoch {
	}

	return nil
}

// WithdrawStake Withdraw the unlocked amount from LiquidStakingToken.Unlocked
// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
// set LiquidStakingToken.Unlocked = LiquidStakingToken.Unlocked - amount
func (sp *StakingPool) WithdrawStake(owner ethcommon.Address, liquidStakingTokenID *big.Int, amount *big.Int) error {
	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}

	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	liquidStakingTokenInfo, err := liquidStakingTokenContract.mustGetInfo(liquidStakingTokenID)
	if err != nil {
		return err
	}

	if liquidStakingTokenInfo.ActiveEpoch > currentEpoch.Epoch {
		return errors.Errorf("staking pool %d is not active", liquidStakingTokenInfo.PoolID)
	}

	return nil
}

// TurnIntoNewEpoch update staking pool status
// Distribute reward to staking pool, and process stake status changes
// calculate stakerReward =  reward * (1 - CommissionRate / CommissionRateDenominator)
// calculate operatorCommission = reward * CommissionRate / CommissionRateDenominator
// set ActiveStake = [ActiveStake + PendingActiveStake - PendingInactiveStake + stakerReward]
// set TotalLiquidStakingToken = LiquidStakingTokenRateHistory[CurrentEpoch - 1].TotalLiquidStakingTokenAmount / LiquidStakingTokenRateHistory[CurrentEpoch - 1].TotalLiquidStakingTokenAmount * ActiveStake
// set LiquidStakingTokenRateHistory[CurrentEpoch] = LiquidStakingTokenRate{StakeAmount: ActiveStake, TotalLiquidStakingTokenAmount: TotalLiquidStakingToken}
// if OperatorLiquidStakingTokenID == 0 (mint a new LiquidStakingToken)
//
//	set OperatorLiquidStakingTokenID = mint new LiquidStakingToken{ ActiveEpoch: CurrentEpoch+1, Principal: operatorCommission, UnlockingRecords: []UnlockingRecord{}}
//
// else(merge LiquidStakingToken)
//
//			calculate totalTokens = GetLockedToken(OperatorLiquidStakingTokenID)
//	     set OperatorLiquidStakingToken.Principal = totalTokens + operatorCommission
//	     set OperatorLiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
//
// set CommissionRate = NextEpochCommissionRate
func (sp *StakingPool) TurnIntoNewEpoch(reward *big.Int) error {
	return nil
}

// UpdateCommissionRate set NextEpochCommissionRate = newCommissionRate, NextEpochCommissionRate will be used in next epoch
func (sp *StakingPool) UpdateCommissionRate(newCommissionRate uint64) error {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return err
	}

	nodeManagerContract := NodeManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	nodeInfo, err := nodeManagerContract.GetNodeInfo(info.PoolID)
	if err != nil {
		return err
	}

	if sp.Ctx.From != ethcommon.HexToAddress(nodeInfo.OperatorAddress) {
		return ErrPermissionDenied
	}

	if newCommissionRate > CommissionRateDenominator {
		return errors.Errorf("invalid commission rate %d, need <= %d", newCommissionRate, CommissionRateDenominator)
	}
	info.NextEpochCommissionRate = newCommissionRate
	if err := sp.info.Put(*info); err != nil {
		return err
	}
	return nil
}

func (sp *StakingPool) GetActiveStake() (*big.Int, error) {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return nil, err
	}

	return info.ActiveStake, nil
}

func (sp *StakingPool) GetNextEpochActiveStake() (*big.Int, error) {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return nil, err
	}

	nextEpochActiveStake := new(big.Int).Set(info.ActiveStake)
	nextEpochActiveStake = nextEpochActiveStake.Add(nextEpochActiveStake, info.PendingActiveStake)
	nextEpochActiveStake = nextEpochActiveStake.Sub(nextEpochActiveStake, info.PendingInactiveStake)
	return nextEpochActiveStake, nil
}

func (sp *StakingPool) GetCommissionRate() (uint64, error) {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return 0, err
	}

	return info.CommissionRate, nil
}

func (sp *StakingPool) GetNextEpochCommissionRate() (uint64, error) {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return 0, err
	}

	return info.NextEpochCommissionRate, nil
}

func (sp *StakingPool) GetCumulativeReward() (*big.Int, error) {
	info, err := sp.MustLoadInfo()
	if err != nil {
		return nil, err
	}

	return info.CumulativeReward, nil
}
