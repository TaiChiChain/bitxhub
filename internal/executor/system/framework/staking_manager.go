package framework

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager_client"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrRewardPerBlockNotFound = errors.New("reward per block not found")
	ErrStakeValue             = errors.New("total supply smaller than total stake")
	ErrZeroAXC                = errors.New("AXC total supply is 0")
)

const (
	blocksPerYear                   = uint64(63072000)
	rewardPerBlockStorageKey        = "stakeRewardPerBlock"
	availableStakingPoolsStorageKey = "availableStakingPools"
	rewardRecordsStorageKey         = "rewardRecords"
)

var StakingManagerBuildConfig = &common.SystemContractBuildConfig[*StakingManager]{
	Name:    "framework_staking_manager",
	Address: common.StakingManagerContractAddr,
	AbiStr:  staking_manager_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *StakingManager {
		return &StakingManager{
			SystemContractBase: systemContractBase,
		}
	},
}

type StakingManager struct {
	common.SystemContractBase

	AvailableStakingPools *common.VMSlot[[]uint64]
	RewardRecordTable     *common.VMMap[uint64, uint64]
	RewardPerBlock        *common.VMSlot[*big.Int]
}

func (s *StakingManager) GenesisInit(genesis *repo.GenesisConfig) error {
	// do init in node manager
	return nil
}

func (s *StakingManager) SetContext(ctx *common.VMContext) {
	s.SystemContractBase.SetContext(ctx)

	s.AvailableStakingPools = common.NewVMSlot[[]uint64](s.StateAccount, availableStakingPoolsStorageKey)
	s.RewardRecordTable = common.NewVMMap[uint64, uint64](s.StateAccount, rewardRecordsStorageKey, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	s.RewardPerBlock = common.NewVMSlot[*big.Int](s.StateAccount, rewardPerBlockStorageKey)
}

func (s *StakingManager) GetStakingPool(poolID uint64) *StakingPool {
	return NewStakingPool(s.SystemContractBase).Load(poolID)
}

func (s *StakingManager) InternalInitAvailableStakingPools(availableStakingPools []uint64) (err error) {
	return s.AvailableStakingPools.Put(availableStakingPools)
}

func (s *StakingManager) InternalTurnIntoNewEpoch() error {
	pools, err := s.AvailableStakingPools.MustGet()
	if err != nil {
		return err
	}
	for _, poolID := range pools {
		_, cnt, err := s.RewardRecordTable.Get(poolID)
		if err != nil {
			return err
		}

		_, rewardPerBlock, err := s.RewardPerBlock.Get()
		if err != nil {
			return err
		}

		reward := new(big.Int).Mul(new(big.Int).SetUint64(cnt), rewardPerBlock)

		if err = s.GetStakingPool(poolID).TurnIntoNewEpoch(reward); err != nil {
			return err
		}
		if err = s.RewardRecordTable.Put(poolID, 0); err != nil {
			return err
		}
	}
	return nil
}

func (s *StakingManager) InternalDisablePool(poolID uint64) error {
	if !s.GetStakingPool(poolID).Exists() {
	}

	_, curPools, err := s.AvailableStakingPools.Get()
	if err != nil {
		return err
	}
	index := -1
	for i, id := range curPools {
		if id == poolID {
			index = i
			break
		}
	}
	curPools = append(curPools[:index], curPools[index+1:]...)
	return s.AvailableStakingPools.Put(curPools)
}

func (s *StakingManager) InternalRecordReward(poolID uint64) (reward *big.Int, err error) {
	exists, reward, err := s.RewardPerBlock.Get()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRewardPerBlockNotFound
	}
	exists, poolReward, err := s.RewardRecordTable.Get(poolID)
	if err != nil {
		return nil, err
	}
	if !exists {
		poolReward = 1
	} else {
		poolReward++
	}
	if err = s.RewardRecordTable.Put(poolID, poolReward); err != nil {
		return nil, err
	}
	return reward, nil
}

func (s *StakingManager) InternalCalculateStakeReward() error {
	axc := token.AXCBuildConfig.Build(s.CrossCallSystemContractContext())

	totalSupply, err := axc.TotalSupply()
	if err != nil {
		return err
	}
	if totalSupply.Cmp(big.NewInt(0)) == 0 {
		return ErrZeroAXC
	}
	totalStake := s.StateAccount.GetBalance()
	if totalSupply.Cmp(totalStake) < 0 {
		return ErrStakeValue
	}
	denominator := 5.5 * (1 + math.Exp(10*divideBigInt(totalStake, totalSupply)))
	ratio := 1/denominator + 0.03
	baseFloat := new(big.Float).SetInt(totalStake)
	multiplied := new(big.Float).Mul(
		baseFloat,
		new(big.Float).Quo(
			new(big.Float).SetFloat64(ratio),
			new(big.Float).SetUint64(blocksPerYear),
		),
	)
	multipliedInt, _ := multiplied.Int(nil)
	return s.RewardPerBlock.Put(multipliedInt)
}

func (s *StakingManager) CreateStakingPool(poolID uint64, commissionRate uint64) (err error) {
	if err := s.GetStakingPool(poolID).Create(poolID, commissionRate); err != nil {
		return err
	}
	_, curPools, err := s.AvailableStakingPools.Get()
	if err != nil {
		return err
	}
	curPools = append(curPools, poolID)
	sort.Slice(curPools, func(i, j int) bool { return curPools[i] < curPools[j] })
	if err := s.AvailableStakingPools.Put(curPools); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) AddStake(poolID uint64, owner string, amount *big.Int) error {
	return s.GetStakingPool(poolID).AddStake(owner, amount)
}

func (s *StakingManager) UnlockStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	return s.GetStakingPool(poolID).UnlockStake(owner, liquidStakingTokenID, amount)
}

func (s *StakingManager) WithdrawStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	return s.GetStakingPool(poolID).WithdrawStake(owner, liquidStakingTokenID, amount)
}

func (s *StakingManager) UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error {
	return s.GetStakingPool(poolID).UpdateCommissionRate(newCommissionRate)
}

func (s *StakingManager) GetPoolActiveStake(poolID uint64) *big.Int {
	return s.GetStakingPool(poolID).GetActiveStake()
}

func (s *StakingManager) GetPoolNextEpochActiveStake(poolID uint64) *big.Int {
	return s.GetStakingPool(poolID).GetNextEpochActiveStake()
}

func (s *StakingManager) GetPoolCommissionRate(poolID uint64) uint64 {
	return s.GetStakingPool(poolID).GetCommissionRate()
}

func (s *StakingManager) GetPoolNextEpochCommissionRate(poolID uint64) uint64 {
	return s.GetStakingPool(poolID).GetNextEpochCommissionRate()
}

func (s *StakingManager) GetPoolCumulativeReward(poolID uint64) *big.Int {
	return s.GetStakingPool(poolID).GetCumulativeReward()
}

func divideBigInt(a, b *big.Int) float64 {
	// Perform division
	result := new(big.Float).Quo(new(big.Float).SetInt(a), new(big.Float).SetInt(b))
	floatResult, _ := result.Float64()

	return floatResult
}
