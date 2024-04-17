package framework

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
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

func (sp *StakingPool) AddStake(owner string, amount *big.Int) error {
	return nil
}

func (sp *StakingPool) UnlockStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	return nil
}

func (sp *StakingPool) WithdrawStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	return nil
}

func (sp *StakingPool) TurnIntoNewEpoch(reward *big.Int) error {
	return nil
}

func (sp *StakingPool) UpdateCommissionRate(newCommissionRate uint64) error {
	return nil
}

func (sp *StakingPool) UpdateOperator(newOperator string) error {
	return nil
}

func (sp *StakingPool) GetActiveStake() *big.Int {
	return nil
}

func (sp *StakingPool) GetNextEpochActiveStake() *big.Int {
	return nil
}

func (sp *StakingPool) GetCommissionRate() uint64 {
	return 0
}

func (sp *StakingPool) GetNextEpochCommissionRate() uint64 {
	// TODO implement me
	panic("implement me")
}

func (sp *StakingPool) GetCumulativeReward() *big.Int {
	// TODO implement me
	panic("implement me")
}
