package framework

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager_client"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrRewardPerBlockNotFound = errors.New("reward per block not found")
	ErrStakeValue             = errors.New("total supply smaller than total stake")
	ErrStakeNotFound          = errors.New("total stake not found")
)

const (
	blocksPerYear                   = uint64(63072000)
	rewardPerBlockStorageKey        = "stakeRewardPerBlock"
	availableStakingPoolsStorageKey = "availableStakingPools"
	rewardRecordsStorageKey         = "rewardRecords"
	gasRecordsStorageKey
	totalStakeStorageKey = "totalStake"
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

var _ staking_manager.StakingManager = (*StakingManager)(nil)

type StakingManager struct {
	common.SystemContractBase

	AvailableStakingPools *common.VMSlot[[]uint64]
	RewardRecordTable     *common.VMMap[uint64, uint64]
	RewardPerBlock        *common.VMSlot[*big.Int]
	GasRewardTable        *common.VMMap[uint64, *big.Int]
	TotalStake            *common.VMSlot[*big.Int]
}

func (s *StakingManager) GenesisInit(_ *repo.GenesisConfig) error {
	// init is already done in node manager
	return nil
}

func (s *StakingManager) SetContext(ctx *common.VMContext) {
	s.SystemContractBase.SetContext(ctx)

	s.AvailableStakingPools = common.NewVMSlot[[]uint64](s.StateAccount, availableStakingPoolsStorageKey)
	s.RewardRecordTable = common.NewVMMap[uint64, uint64](s.StateAccount, rewardRecordsStorageKey, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	s.RewardPerBlock = common.NewVMSlot[*big.Int](s.StateAccount, rewardPerBlockStorageKey)
	s.TotalStake = common.NewVMSlot[*big.Int](s.StateAccount, totalStakeStorageKey)
	s.GasRewardTable = common.NewVMMap[uint64, *big.Int](s.StateAccount, gasRecordsStorageKey, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
}

func (s *StakingManager) GetStakingPool(poolID uint64) *StakingPool {
	return NewStakingPool(s.SystemContractBase).Load(poolID)
}

func (s *StakingManager) InternalInitAvailableStakingPools(availableStakingPools []uint64) (err error) {
	if !s.Ctx.CallFromSystem {
		return ErrPermissionDenied
	}
	return s.AvailableStakingPools.Put(availableStakingPools)
}

func (s *StakingManager) InternalTurnIntoNewEpoch() error {
	if !s.Ctx.CallFromSystem {
		return ErrPermissionDenied
	}
	axc := token.AXCBuildConfig.Build(s.CrossCallSystemContractContext())
	pools, err := s.AvailableStakingPools.MustGet()
	if err != nil {
		return err
	}
	for _, poolID := range pools {
		exists, cnt, err := s.RewardRecordTable.Get(poolID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.Errorf("reward record of %d not found", poolID)
		}

		exists, rewardPerBlock, err := s.RewardPerBlock.Get()
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("reward per block not found")
		}

		exists, gasReward, err := s.GasRewardTable.Get(poolID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.Errorf("gas reward of %d not found", poolID)
		}

		reward := new(big.Int).Mul(new(big.Int).SetUint64(cnt), rewardPerBlock)
		// mint some token to the axc contract
		if err = axc.Mint(reward); err != nil {
			return err
		}
		// transfer token to staking manager contract
		axc.StateAccount.SubBalance(reward)
		s.StateAccount.AddBalance(reward)
		// add gas to the reward
		reward = reward.Add(reward, gasReward)
		if err = s.GetStakingPool(poolID).TurnIntoNewEpoch(reward); err != nil {
			return err
		}
		if err = s.RewardRecordTable.Put(poolID, 0); err != nil {
			return err
		}
		if err = s.GasRewardTable.Put(poolID, big.NewInt(0)); err != nil {
			return err
		}
	}
	return s.InternalCalculateStakeReward()
}

func (s *StakingManager) InternalDisablePool(poolID uint64) error {
	if !s.Ctx.CallFromSystem {
		return ErrPermissionDenied
	}
	isExist, curPools, err := s.AvailableStakingPools.Get()
	if err != nil {
		return err
	}
	if !isExist {
		return errors.New("available pools not found")
	}
	index := -1
	for i, id := range curPools {
		if id == poolID {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.Errorf("pool %d not found", poolID)
	}
	curPools = append(curPools[:index], curPools[index+1:]...)
	return s.AvailableStakingPools.Put(curPools)
}

func (s *StakingManager) InternalRecordReward(poolID uint64, gasReward *big.Int) (reward *big.Int, err error) {
	if !s.Ctx.CallFromSystem {
		return nil, ErrPermissionDenied
	}
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

	// record gas reward
	exists, gas, err := s.GasRewardTable.Get(poolID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("Gas reward not found")
	}
	newGasReward := new(big.Int).Add(gas, gasReward)
	if err = s.GasRewardTable.Put(poolID, newGasReward); err != nil {
		return nil, err
	}
	return reward, nil
}

func (s *StakingManager) InternalCalculateStakeReward() error {
	if !s.Ctx.CallFromSystem {
		return ErrPermissionDenied
	}
	axc := token.AXCBuildConfig.Build(s.CrossCallSystemContractContext())

	totalSupply, err := axc.TotalSupply()
	if err != nil {
		return err
	}
	exists, totalStake, err := s.TotalStake.Get()
	if err != nil {
		return err
	}
	if !exists {
		return ErrStakeNotFound
	}
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
	nodeManager := NodeManagerBuildConfig.Build(s.CrossCallSystemContractContext())
	info, err := nodeManager.GetNodeInfo(poolID)
	if err != nil {
		return err
	}
	if strings.ToLower(info.OperatorAddress) != strings.ToLower(s.Ctx.From.String()) {
		return errors.Errorf("operator address %s is not equal to caller %s", info.OperatorAddress, s.Ctx.From.String())
	}
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

func (s *StakingManager) AddStake(poolID uint64, owner ethcommon.Address, amount *big.Int) error {
	exist, stake, err := s.TotalStake.Get()
	if err != nil {
		return err
	}
	if !exist {
		return ErrStakeNotFound
	}
	stake = stake.Add(stake, amount)
	if err = s.TotalStake.Put(stake); err != nil {
		return err
	}
	return s.GetStakingPool(poolID).AddStake(owner, amount)
}

func (s *StakingManager) Unlock(poolID uint64, owner ethcommon.Address, liquidStakingTokenID *big.Int, amount *big.Int) error {
	exist, stake, err := s.TotalStake.Get()
	if err != nil {
		return err
	}
	if !exist {
		return ErrStakeNotFound
	}
	after := stake.Sub(stake, amount)
	if after.Cmp(big.NewInt(0)) < 0 {
		return errors.Errorf("total stake %s less than amount %s", stake.String(), amount.String())
	}
	if err = s.TotalStake.Put(after); err != nil {
		return err
	}
	return s.GetStakingPool(poolID).UnlockStake(owner, liquidStakingTokenID, amount)
}

func (s *StakingManager) Withdraw(poolID uint64, owner ethcommon.Address, liquidStakingTokenID *big.Int, amount *big.Int) error {
	return s.GetStakingPool(poolID).WithdrawStake(owner, liquidStakingTokenID, amount)
}

func (s *StakingManager) UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error {
	return s.GetStakingPool(poolID).UpdateCommissionRate(newCommissionRate)
}

func (s *StakingManager) GetPoolActiveStake(poolID uint64) (*big.Int, error) {
	return s.GetStakingPool(poolID).GetActiveStake()
}

func (s *StakingManager) GetPoolNextEpochActiveStake(poolID uint64) (*big.Int, error) {
	return s.GetStakingPool(poolID).GetNextEpochActiveStake()
}

func (s *StakingManager) GetPoolCommissionRate(poolID uint64) (uint64, error) {
	return s.GetStakingPool(poolID).GetCommissionRate()
}

func (s *StakingManager) GetPoolNextEpochCommissionRate(poolID uint64) (uint64, error) {
	return s.GetStakingPool(poolID).GetNextEpochCommissionRate()
}

func (s *StakingManager) GetPoolCumulativeReward(poolID uint64) (*big.Int, error) {
	return s.GetStakingPool(poolID).GetCumulativeReward()
}

func divideBigInt(a, b *big.Int) float64 {
	// Perform division
	result := new(big.Float).Quo(new(big.Float).SetInt(a), new(big.Float).SetInt(b))
	floatResult, _ := result.Float64()

	return floatResult
}
