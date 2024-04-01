package base

import (
	"fmt"
	"math"
	"math/big"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	ErrStakingPoolNotExist    = errors.New("staking pool not exist")
	ErrRewardPerBlockNotFound = errors.New("reward per block not found")
)

const (
	blocksPerYear        = uint64(63072000)
	stakingPrefix        = "staking"
	rewardPerBlockKey    = "stakeRewardPerBlock"
	availableStakingPool = "availableStakingPools"
	rewardRecords        = "rewardRecords"
)

type stakingManager struct {
	currentEpoch  *rbft.EpochInfo
	stateLedger   ledger.StateLedger
	account       ledger.IAccount
	abi           *abi.ABI
	currentLogs   *[]common.Log
	currentHeight uint64
	config        *config

	AvailableStakingPools *common.VMSlot[[]uint64]
	RewardRecordTable     *common.VMMap[uint64, uint64]
	RewardPerBlock        *common.VMSlot[*big.Int]
}

func NewStakingManager(cfg *common.SystemContractConfig) StakingManager {
	return &stakingManager{
		config: &config{
			CommissionRate: 5000,
		},
	}
}

func (s *stakingManager) InternalTurnIntoNewEpoch() error {
	exists, pools, err := s.AvailableStakingPools.Get()
	if err != nil {
		return err
	}
	if !exists {
		return ErrStakingPoolNotExist
	}
	for _, poolID := range pools {
		exists, cnt, err := s.RewardRecordTable.Get(poolID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrStakingPoolNotExist
		}
		exists, rewardPerBlock, err := s.RewardPerBlock.Get()
		if err != nil {
			return err
		}
		if !exists {
			return ErrRewardPerBlockNotFound
		}
		reward := new(big.Int).Mul(new(big.Int).SetUint64(cnt), rewardPerBlock)
		pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
		if err != nil {
			return err
		}
		if err = pool.TurnIntoNewEpoch(reward); err != nil {
			return err
		}
		if err = s.RewardRecordTable.Put(poolID, 0); err != nil {
			return err
		}
	}
	return nil
}

func (s *stakingManager) InternalDisablePool(poolID uint64) error {
	if !HasStakingPool(s.stateLedger, poolID) {
		return ErrStakingPoolNotExist
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

func (s *stakingManager) InternalRecordReward(poolID uint64) (reward *big.Int, err error) {
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

func (s *stakingManager) InternalCalculateStakeReward() error {
	axc := GetAXCContract(s.stateLedger, s.currentHeight, s.currentLogs)
	totalSupply := axc.TotalSupply()
	totalStake := s.stateLedger.GetBalance(types.NewAddressByStr(common.StakingManagerContractAddr))
	denominator := 0.18 * (1 + math.Exp(10*divideBigInt(totalStake, totalSupply)))
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

func (s *stakingManager) CreateStakingPool(poolID uint64) (err error) {
	_, err = CreateStakingPool(s.stateLedger, s.currentEpoch, poolID, s.config.CommissionRate)
	return
}

func (s *stakingManager) AddStake(poolID uint64, owner string, amount *big.Int) error {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return common.NewRevertStringError(err.Error())
	}
	return pool.AddStake(owner, amount)
}

func (s *stakingManager) UnlockStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return common.NewRevertStringError(err.Error())
	}
	return pool.UnlockStake(owner, liquidStakingTokenID, amount)
}

func (s *stakingManager) WithdrawStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return common.NewRevertStringError(err.Error())
	}
	return pool.WithdrawStake(owner, liquidStakingTokenID, amount)
}

func (s *stakingManager) UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return err
	}
	return pool.UpdateCommissionRate(newCommissionRate)
}

func (s *stakingManager) GetPoolActiveStake(poolID uint64) *big.Int {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return big.NewInt(0)
	}
	return pool.GetActiveStake()
}

func (s *stakingManager) GetPoolNextEpochActiveStake(poolID uint64) *big.Int {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return big.NewInt(0)
	}
	return pool.GetNextEpochActiveStake()
}

func (s *stakingManager) GetPoolCommissionRate(poolID uint64) uint64 {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return 0
	}
	return pool.GetCommissionRate()
}

func (s *stakingManager) GetPoolNextEpochCommissionRate(poolID uint64) uint64 {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return 0
	}
	return pool.GetNextEpochCommissionRate()
}

func (s *stakingManager) GetPoolCumulativeReward(poolID uint64) *big.Int {
	pool, err := GetStakingPool(s.stateLedger, s.currentEpoch, poolID)
	if err != nil {
		return big.NewInt(0)
	}
	return pool.GetCumulativeReward()
}

func GetAXCContract(stateLedger ledger.StateLedger, currentHeight uint64, currentLogs *[]common.Log) *token.Manager {
	cfg := &common.SystemContractConfig{
		Logger: nil,
	}
	axc := token.New(cfg)
	contractAddr := common2.HexToAddress(common.AXCContractAddr)
	axc.SetContext(&common.VMContext{
		StateLedger:   stateLedger,
		CurrentHeight: currentHeight,
		CurrentUser:   &contractAddr,
		CurrentLogs:   currentLogs,
	})
	return axc
}

func (s *stakingManager) SetContext(ctx *common.VMContext) {
	s.stateLedger = ctx.StateLedger
	s.account = ctx.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.StakingManagerContractAddr))
	s.abi = ctx.ABI
	s.currentLogs = ctx.CurrentLogs
	s.currentHeight = ctx.CurrentHeight

	s.AvailableStakingPools = common.NewVMSlot[[]uint64](s.account, stakingPrefix+availableStakingPool)
	s.RewardRecordTable = common.NewVMMap[uint64, uint64](s.account, stakingPrefix+rewardRecords, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	s.RewardPerBlock = common.NewVMSlot[*big.Int](s.account, stakingPrefix+rewardPerBlockKey)
}

type config struct {
	CommissionRate uint64
}

type StakingManager interface {
	// InternalTurnIntoNewEpoch turn into new epoch, switch some parameter
	// Internal functions(called by System Contract)
	InternalTurnIntoNewEpoch() error

	// InternalDisablePool disables a staking pool
	// after node leave validator set or banded by governance
	InternalDisablePool(poolID uint64) error

	// InternalRecordReward records reward
	// include block gas fee reward + stake reward
	InternalRecordReward(poolID uint64) (reward *big.Int, err error)

	// InternalCalculateStakeReward calculates stake reward in each epoch when a block produced
	InternalCalculateStakeReward() error

	// CreateStakingPool creates a stake pool
	// Public functions(called by User)
	// Operator crate StakingPool for this node
	CreateStakingPool(poolID uint64) (err error)

	// AddStake adds stake to staking pool for user
	// call StakingPool.AddStake
	AddStake(poolID uint64, owner string, amount *big.Int) error

	// UnlockStake unlock stake from staking pool for user
	// call StakingPool.UnlockStake
	UnlockStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error

	// WithdrawStake withdraw stake from staking pool for user
	// call StakingPool.WithdrawStake
	WithdrawStake(poolID uint64, owner string, liquidStakingTokenID *big.Int, amount *big.Int) error

	// UpdatePoolCommissionRate updates commission rate for staking pool
	// call StakingPool.UpdateCommissionRate
	UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error

	// GetPoolActiveStake returns active stake for staking pool
	// call StakingPool.GetActiveStake
	GetPoolActiveStake(poolID uint64) *big.Int

	// GetPoolNextEpochActiveStake returns next epoch active stake for staking pool
	// call StakingPool.GetNextEpochActiveStake
	GetPoolNextEpochActiveStake(poolID uint64) *big.Int

	// GetPoolCommissionRate returns current commission rate for staking pool
	// call StakingPool.GetCommissionRate
	GetPoolCommissionRate(poolID uint64) uint64

	// GetPoolNextEpochCommissionRate returns next epoch commission rate for staking pool
	// call StakingPool.GetNextEpochCommissionRate
	GetPoolNextEpochCommissionRate(poolID uint64) uint64

	// GetPoolCumulativeReward returns cumulative reward for staking pool
	// call StakingPool.GetCumulativeReward
	GetPoolCumulativeReward(poolID uint64) *big.Int

	SetContext(ctx *common.VMContext)
}
