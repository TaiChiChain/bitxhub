package framework

import (
	"fmt"
	"math/big"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/liquid_staking_token"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	CommissionRateDenominator                         = 10000
	StakingPoolNamespace                              = "stakingPool"
	StakingPoolInfoKeySuffix                          = "info"
	StakingPoolLiquidStakingTokenRateHistoryKeyPrefix = "liquidStakingTokenRateHistory"
)

type StakingPool struct {
	common.SystemContractBase

	poolID                           uint64
	info                             *common.VMSlot[staking_manager.PoolInfo]
	historyLiquidStakingTokenRateMap *common.VMMap[uint64, staking_manager.LiquidStakingTokenRate]
}

func NewStakingPool(systemContractBase common.SystemContractBase) *StakingPool {
	return &StakingPool{
		SystemContractBase: systemContractBase,
	}
}

func (sp *StakingPool) Load(poolID uint64) *StakingPool {
	sp.poolID = poolID
	sp.info = common.NewVMSlot[staking_manager.PoolInfo](sp.StateAccount, fmt.Sprintf("%s_%d_%s", StakingPoolNamespace, poolID, StakingPoolInfoKeySuffix))
	sp.historyLiquidStakingTokenRateMap = common.NewVMMap[uint64, staking_manager.LiquidStakingTokenRate](sp.StateAccount, fmt.Sprintf("%s_%d_%s", StakingPoolNamespace, poolID, StakingPoolLiquidStakingTokenRateHistoryKeyPrefix), func(key uint64) string {
		return strconv.FormatUint(key, 10)
	})
	return sp
}

func (sp *StakingPool) Create(operator ethcommon.Address, epoch uint64, commissionRate uint64, stakeNumber *big.Int) (lstID *big.Int, err error) {
	zero := big.NewInt(0)

	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	liquidStakingTokenID, err := liquidStakingTokenContract.Mint(operator, &liquid_staking_token.LiquidStakingTokenInfo{
		PoolID:           sp.poolID,
		Principal:        stakeNumber,
		Unlocked:         zero,
		ActiveEpoch:      epoch,
		UnlockingRecords: []liquid_staking_token.UnlockingRecord{},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to mint liquid staking token for node %d", sp.poolID)
	}

	stakingPoolInfo := &staking_manager.PoolInfo{
		ID:                                      sp.poolID,
		IsActive:                                true,
		ActiveStake:                             stakeNumber,
		TotalLiquidStakingToken:                 stakeNumber,
		PendingActiveStake:                      zero,
		PendingInactiveStake:                    zero,
		PendingInactiveLiquidStakingTokenAmount: zero,
		CommissionRate:                          commissionRate,
		NextEpochCommissionRate:                 commissionRate,
		LastEpochReward:                         zero,
		LastEpochCommission:                     zero,
		CumulativeReward:                        zero,
		CumulativeCommission:                    zero,
		OperatorLiquidStakingTokenID:            liquidStakingTokenID,
		LastRateEpoch:                           0,
	}

	if sp.info.Has() {
		return nil, errors.Errorf("staking pool %d already exists", stakingPoolInfo.ID)
	}

	if stakingPoolInfo.CommissionRate > CommissionRateDenominator {
		return nil, errors.Errorf("invalid commission rate: %d, must be less than or equal %d", stakingPoolInfo.CommissionRate, CommissionRateDenominator)
	}
	if err := sp.info.Put(*stakingPoolInfo); err != nil {
		return nil, err
	}

	return liquidStakingTokenID, nil
}

func (sp *StakingPool) GenesisInit(genesis *repo.GenesisConfig) error {
	info, err := sp.MustGetInfo()
	if err != nil {
		return err
	}
	info.LastRateEpoch = genesis.EpochInfo.Epoch
	if err := sp.info.Put(*info); err != nil {
		return err
	}
	if err := sp.historyLiquidStakingTokenRateMap.Put(genesis.EpochInfo.Epoch, staking_manager.LiquidStakingTokenRate{
		StakeAmount:              info.ActiveStake,
		LiquidStakingTokenAmount: info.TotalLiquidStakingToken,
	}); err != nil {
		return err
	}

	return nil
}

func (sp *StakingPool) Exists() bool {
	return sp.info.Has()
}

func (sp *StakingPool) MustGetInfo() (*staking_manager.PoolInfo, error) {
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
// mint a LiquidStakingToken, set LiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
// emit a AddStake event
// emit a MintLiquidStakingToken event
func (sp *StakingPool) AddStake(owner ethcommon.Address, amount *big.Int) error {
	info, err := sp.MustGetInfo()
	if err != nil {
		return err
	}

	if !info.IsActive {
		return errors.Errorf("staking pool %d is inactive", info.ID)
	}

	if sp.Ctx.Value == nil || sp.Ctx.Value.Cmp(amount) < 0 {
		return errors.Errorf("received value %d is less than amount %d", sp.Ctx.Value, amount)
	}

	nodeManagerContract := NodeManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	nodeInfo, err := nodeManagerContract.GetInfo(info.ID)
	if err != nil {
		return err
	}

	if types.NodeStatus(nodeInfo.Status) != types.NodeStatusCandidate && types.NodeStatus(nodeInfo.Status) != types.NodeStatusActive {
		return errors.Errorf("staking pool %d status does not support adding stake", info.ID)
	}

	// add stake to pool
	info.PendingActiveStake = new(big.Int).Add(info.PendingActiveStake, amount)
	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}

	// mint nft for owner
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	tokenID, err := liquidStakingTokenContract.Mint(owner, &liquid_staking_token.LiquidStakingTokenInfo{
		PoolID:           info.ID,
		Principal:        amount,
		Unlocked:         new(big.Int),
		ActiveEpoch:      currentEpoch.Epoch + 1,
		UnlockingRecords: []liquid_staking_token.UnlockingRecord{},
	})
	if err != nil {
		return err
	}

	if err := sp.info.Put(*info); err != nil {
		return err
	}

	sp.EmitEvent(&staking_manager.EventAddStake{
		PoolID:               info.ID,
		Owner:                owner,
		Amount:               amount,
		LiquidStakingTokenID: tokenID,
	})

	return nil
}

// UnlockStake Unlock the specified amount and enter the lockin period and add a record to UnlockingRecords to liquidStakingToken, will refresh the LiquidStakingToken.ActiveEpoch and LiquidStakingToken.Principal(first calculate reward)
// NOTICE: Unable to obtain reward after calling unlock, and will revert when the LiquidStakingToken is pendingActive
// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
// cal lst principalAndReward and lstAmount, unlock all principalAndReward and restake remain
// set PendingInactiveStake = [PendingInactiveStake - principalAndReward]
// set PendingActiveStake = [PendingActiveStake + principalAndReward - unlock]
// set PendingInactiveLiquidStakingTokenAmount = [PendingInactiveLiquidStakingTokenAmount - lstAmount]
// append UnlockingRecord{Amount: amount, UnlockTimestamp: block.Timestamp} to LiquidStakingToken.UnlockingRecords
// emit a EventUnlock event
func (sp *StakingPool) UnlockStake(liquidStakingTokenID *big.Int, liquidStakingTokenInfo *liquid_staking_token.LiquidStakingTokenInfo, amount *big.Int) error {
	info, err := sp.MustGetInfo()
	if err != nil {
		return err
	}

	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}

	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	// token is not active, use Withdraw instead of Unlock
	if liquidStakingTokenInfo.ActiveEpoch-1 == currentEpoch.Epoch {
		return errors.Errorf("token is not active yet, use `WithdrawStake` instead")
	}

	updateLiquidStakingTokenUnlockingRecords(sp.Ctx.CurrentEVM.Context.Time, liquidStakingTokenInfo)

	var principalAndReward, pendingInactiveLiquidStakingTokenAmount *big.Int
	// calculate liquidStakingToken amount
	stakingRate, err := sp.HistoryLiquidStakingTokenRate(liquidStakingTokenInfo.ActiveEpoch - 1)
	if err != nil {
		return err
	}
	lstAmount := calculateLSTAmount(stakingRate, liquidStakingTokenInfo.Principal)
	// update Principal by reward
	if liquidStakingTokenInfo.ActiveEpoch < currentEpoch.Epoch {
		lastRate, err := sp.HistoryLiquidStakingTokenRate(currentEpoch.Epoch - 1)
		if err != nil {
			return err
		}
		principalAndReward = calculatePrincipalAndReward(stakingRate, lastRate, liquidStakingTokenInfo.Principal)
	} else {
		// no reward
		principalAndReward = new(big.Int).Set(liquidStakingTokenInfo.Principal)
	}
	if principalAndReward.Cmp(amount) < 0 {
		return errors.Errorf("amount is not enough, principalAndReward: %s, amount: %s", principalAndReward.String(), amount.String())
	}
	if !currentEpoch.StakeParams.EnablePartialUnlock && principalAndReward.Cmp(amount) > 0 {
		return errors.Errorf("not enable partial unlock, principalAndReward: %s, amount: %s", principalAndReward.String(), amount.String())
	}

	pendingInactiveLiquidStakingTokenAmount = new(big.Int).Set(lstAmount)
	info.PendingInactiveStake = new(big.Int).Add(info.PendingInactiveStake, principalAndReward)
	info.PendingInactiveLiquidStakingTokenAmount = new(big.Int).Add(info.PendingInactiveLiquidStakingTokenAmount, pendingInactiveLiquidStakingTokenAmount)
	unlockRecord := liquid_staking_token.UnlockingRecord{
		Amount:          amount,
		UnlockTimestamp: sp.Ctx.CurrentEVM.Context.Time + currentEpoch.StakeParams.UnlockPeriod,
	}
	liquidStakingTokenInfo.UnlockingRecords = append(liquidStakingTokenInfo.UnlockingRecords, unlockRecord)
	if uint64(len(liquidStakingTokenInfo.UnlockingRecords)) > currentEpoch.StakeParams.MaxUnlockingRecordNum {
		return errors.Errorf("unlocking record num exceed max num %d", currentEpoch.StakeParams.MaxUnlockingRecordNum)
	}

	// restake remain
	liquidStakingTokenInfo.Principal = new(big.Int).Sub(principalAndReward, amount)
	liquidStakingTokenInfo.ActiveEpoch = currentEpoch.Epoch + 1
	info.PendingActiveStake = new(big.Int).Add(info.PendingActiveStake, liquidStakingTokenInfo.Principal)

	if err := liquidStakingTokenContract.updateInfo(liquidStakingTokenID, liquidStakingTokenInfo); err != nil {
		return err
	}
	if err := sp.info.Put(*info); err != nil {
		return err
	}

	sp.EmitEvent(&staking_manager.EventUnlock{
		LiquidStakingTokenID: liquidStakingTokenID,
		Amount:               amount,
		UnlockTimestamp:      unlockRecord.UnlockTimestamp,
	})

	return nil
}

// WithdrawStake Withdraw the unlocked amount from LiquidStakingToken.Unlocked
// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
// if LiquidStakingToken.Unlocked < amount, set LiquidStakingToken.Unlocked = LiquidStakingToken.Unlocked - amount
// else
//
//	if token is not active, Principal can directly withdraw
//
// emit a EventWithdraw event
func (sp *StakingPool) WithdrawStake(liquidStakingTokenID *big.Int, liquidStakingTokenInfo *liquid_staking_token.LiquidStakingTokenInfo, recipient ethcommon.Address, amount *big.Int) (principalWithdraw *big.Int, err error) {
	info, err := sp.MustGetInfo()
	if err != nil {
		return nil, err
	}

	epochManagerContract := EpochManagerBuildConfig.Build(sp.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return nil, err
	}

	principalWithdraw = new(big.Int)
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	updateLiquidStakingTokenUnlockingRecords(sp.Ctx.CurrentEVM.Context.Time, liquidStakingTokenInfo)

	if liquidStakingTokenInfo.Unlocked.Cmp(amount) >= 0 {
		liquidStakingTokenInfo.Unlocked = new(big.Int).Sub(liquidStakingTokenInfo.Unlocked, amount)
		if err := liquidStakingTokenContract.updateInfo(liquidStakingTokenID, liquidStakingTokenInfo); err != nil {
			return nil, err
		}
	} else {
		// token is not active, Principal can directly withdraw
		if liquidStakingTokenInfo.ActiveEpoch > currentEpoch.Epoch {
			totalAmount := new(big.Int).Add(liquidStakingTokenInfo.Principal, liquidStakingTokenInfo.Unlocked)
			if totalAmount.Cmp(amount) < 0 {
				return nil, errors.Errorf("unlocked amount is not enough, totalAmount: %s, amount: %s", totalAmount.String(), amount.String())
			}
			// withdraw unlocked amount first
			if liquidStakingTokenInfo.Unlocked.Cmp(amount) >= 0 {
				liquidStakingTokenInfo.Unlocked = new(big.Int).Sub(liquidStakingTokenInfo.Unlocked, amount)
			} else {
				principalWithdraw = new(big.Int).Sub(amount, liquidStakingTokenInfo.Unlocked)
				liquidStakingTokenInfo.Principal = new(big.Int).Sub(liquidStakingTokenInfo.Principal, principalWithdraw)
				liquidStakingTokenInfo.Unlocked = new(big.Int)
				info.PendingActiveStake = new(big.Int).Sub(info.PendingActiveStake, principalWithdraw)
				if err := sp.info.Put(*info); err != nil {
					return nil, err
				}
			}

			if err := liquidStakingTokenContract.updateInfo(liquidStakingTokenID, liquidStakingTokenInfo); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.Errorf("unlocked amount is not enough")
		}
	}

	if err := sp.Transfer(recipient, amount); err != nil {
		return nil, err
	}

	sp.EmitEvent(&staking_manager.EventWithdraw{
		LiquidStakingTokenID: liquidStakingTokenID,
		Recipient:            recipient,
		Amount:               amount,
	})
	return principalWithdraw, nil
}

// TurnIntoNewEpoch update staking pool status
// Distribute reward to staking pool, and process stake status changes
// calculate stakerReward =  reward * (CommissionRateDenominator - CommissionRate / CommissionRateDenominator)
// calculate operatorCommission = reward - stakerReward
// set ActiveStake = [ActiveStake + PendingActiveStake - PendingInactiveStake + stakerReward]
// set TotalLiquidStakingToken = LiquidStakingTokenRateHistory[CurrentEpoch - 1].LiquidStakingTokenAmount / LiquidStakingTokenRateHistory[CurrentEpoch - 1].StakeAmount * ActiveStake
// set LiquidStakingTokenRateHistory[CurrentEpoch] = LiquidStakingTokenRate{StakeAmount: ActiveStake, LiquidStakingTokenAmount: TotalLiquidStakingToken}
// merge OperatorLiquidStakingToken:
//
//	calculate totalTokens = GetLockedToken(OperatorLiquidStakingTokenID)
//
// set OperatorLiquidStakingToken.Principal = totalTokens + operatorCommission
// set OperatorLiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
//
// set CommissionRate = NextEpochCommissionRate
func (sp *StakingPool) TurnIntoNewEpoch(oldEpoch *types.EpochInfo, newEpoch *types.EpochInfo, reward *big.Int) error {
	info, err := sp.MustGetInfo()
	if err != nil {
		return err
	}
	stakerReward := new(big.Int).Mul(reward, new(big.Int).SetUint64(CommissionRateDenominator-info.CommissionRate))
	stakerReward = new(big.Int).Div(stakerReward, new(big.Int).SetUint64(CommissionRateDenominator))
	operatorCommission := new(big.Int).Sub(reward, stakerReward)

	info.LastEpochReward = new(big.Int).Set(stakerReward)
	info.LastEpochCommission = new(big.Int).Set(operatorCommission)
	info.CumulativeReward = new(big.Int).Add(info.CumulativeReward, stakerReward)
	info.CumulativeCommission = new(big.Int).Add(info.CumulativeCommission, operatorCommission)

	// update ActiveStake
	info.ActiveStake = new(big.Int).Add(info.ActiveStake, stakerReward)
	info.ActiveStake = new(big.Int).Sub(info.ActiveStake, info.PendingInactiveStake)
	info.PendingInactiveStake = big.NewInt(0)
	info.TotalLiquidStakingToken = new(big.Int).Sub(info.TotalLiquidStakingToken, info.PendingInactiveLiquidStakingTokenAmount)
	info.PendingInactiveLiquidStakingTokenAmount = big.NewInt(0)

	oldActiveStake := new(big.Int).Set(info.ActiveStake)
	info.ActiveStake = new(big.Int).Add(info.ActiveStake, info.PendingActiveStake)
	info.PendingActiveStake = big.NewInt(0)

	info.TotalLiquidStakingToken = calculateLSTAmount(staking_manager.LiquidStakingTokenRate{
		StakeAmount:              oldActiveStake,
		LiquidStakingTokenAmount: info.TotalLiquidStakingToken,
	}, info.ActiveStake)

	// restake operatorCommission to operatorLiquidStakingToken
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(sp.CrossCallSystemContractContext())
	operatorLiquidStakingTokenInfo, err := liquidStakingTokenContract.mustGetInfo(info.OperatorLiquidStakingTokenID)
	if err != nil {
		return err
	}

	// withdraw
	// notice: operatorLiquidStakingTokenInfo.ActiveEpoch = oldEpoch.Epoch, so use rate[oldEpoch.Epoch]
	// calculate operator reward: stakerReward * (operatorLiquidStakingTokenInfo.Principal / lastActiveStake)
	lastRate, err := sp.HistoryLiquidStakingTokenRate(oldEpoch.Epoch - 1)
	if err != nil {
		return err
	}

	if operatorLiquidStakingTokenInfo.ActiveEpoch == oldEpoch.Epoch {
		operatorPrincipalAndReward := calculatePrincipalAndReward(lastRate, staking_manager.LiquidStakingTokenRate{
			StakeAmount:              info.ActiveStake,
			LiquidStakingTokenAmount: info.TotalLiquidStakingToken,
		}, operatorLiquidStakingTokenInfo.Principal)
		operatorLiquidStakingTokenInfo.Principal = new(big.Int).Add(operatorPrincipalAndReward, operatorCommission)
		operatorLiquidStakingTokenInfo.ActiveEpoch = newEpoch.Epoch
		if err := liquidStakingTokenContract.updateInfo(info.OperatorLiquidStakingTokenID, operatorLiquidStakingTokenInfo); err != nil {
			return err
		}
	} else {
		// operator lst has been unlocked in current epoch, not get stake reward in current epoch
		// add operatorCommission to operatorLiquidStakingToken
		operatorLiquidStakingTokenInfo.Principal = new(big.Int).Add(operatorLiquidStakingTokenInfo.Principal, operatorCommission)
		operatorLiquidStakingTokenInfo.ActiveEpoch = newEpoch.Epoch

		if err := liquidStakingTokenContract.updateInfo(info.OperatorLiquidStakingTokenID, operatorLiquidStakingTokenInfo); err != nil {
			return err
		}
	}

	// add stake
	oldActiveStake = new(big.Int).Set(info.ActiveStake)
	info.ActiveStake = new(big.Int).Add(info.ActiveStake, operatorCommission)
	info.TotalLiquidStakingToken = calculateLSTAmount(staking_manager.LiquidStakingTokenRate{
		StakeAmount:              oldActiveStake,
		LiquidStakingTokenAmount: info.TotalLiquidStakingToken,
	}, info.ActiveStake)

	// update commission
	info.CommissionRate = info.NextEpochCommissionRate
	info.LastRateEpoch = oldEpoch.Epoch
	if err := sp.info.Put(*info); err != nil {
		return err
	}

	// final rate formula:
	// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
	// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveStake)
	if err := sp.historyLiquidStakingTokenRateMap.Put(oldEpoch.Epoch, staking_manager.LiquidStakingTokenRate{
		StakeAmount:              info.ActiveStake,
		LiquidStakingTokenAmount: info.TotalLiquidStakingToken,
	}); err != nil {
		return err
	}

	return nil
}

// UpdateCommissionRate set NextEpochCommissionRate = newCommissionRate, NextEpochCommissionRate will be used in next epoch
func (sp *StakingPool) UpdateCommissionRate(newCommissionRate uint64) error {
	info, err := sp.MustGetInfo()
	if err != nil {
		return err
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

func (sp *StakingPool) Info() (staking_manager.PoolInfo, error) {
	_, info, err := sp.info.Get()
	return info, err
}

func (sp *StakingPool) HistoryLiquidStakingTokenRate(epoch uint64) (staking_manager.LiquidStakingTokenRate, error) {
	info, err := sp.MustGetInfo()
	if err != nil {
		return staking_manager.LiquidStakingTokenRate{}, err
	}
	if epoch > info.LastRateEpoch {
		epoch = info.LastRateEpoch
	}
	exist, rate, err := sp.historyLiquidStakingTokenRateMap.Get(epoch)
	if err != nil {
		return staking_manager.LiquidStakingTokenRate{}, err
	}
	if !exist {
		return staking_manager.LiquidStakingTokenRate{
			StakeAmount:              new(big.Int),
			LiquidStakingTokenAmount: new(big.Int),
		}, nil
	}
	return rate, err
}

func calculateLSTAmount(rate staking_manager.LiquidStakingTokenRate, stakeAmount *big.Int) *big.Int {
	if rate.StakeAmount.Sign() == 0 || rate.LiquidStakingTokenAmount.Sign() == 0 {
		return new(big.Int).Set(stakeAmount)
	}

	// lst = rate.LiquidStakingTokenAmount / rate.StakeAmount * stakeAmount
	lst := new(big.Int).Set(rate.LiquidStakingTokenAmount)
	lst = lst.Mul(lst, stakeAmount)
	lst = lst.Div(lst, rate.StakeAmount)
	return lst
}

func calculateStakeAmount(rate staking_manager.LiquidStakingTokenRate, lstAmount *big.Int) *big.Int {
	if rate.StakeAmount.Sign() == 0 || rate.LiquidStakingTokenAmount.Sign() == 0 {
		return new(big.Int).Set(lstAmount)
	}

	// lst:= rate.StakeAmount / rate.LiquidStakingTokenAmount * lstAmount
	res := new(big.Int).Set(rate.StakeAmount)
	res = res.Mul(res, lstAmount)
	res = res.Div(res, rate.LiquidStakingTokenAmount)
	return res
}

func calculatePrincipalAndReward(stakingRate, lastRate staking_manager.LiquidStakingTokenRate, principal *big.Int) *big.Int {
	lstAmount := calculateLSTAmount(stakingRate, principal)
	principalAndReward := calculateStakeAmount(lastRate, lstAmount)
	return principalAndReward
}
