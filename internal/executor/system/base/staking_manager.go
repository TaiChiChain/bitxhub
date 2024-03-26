package base

import (
	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"math/big"
)

type stakingManager struct {
	currentEpoch *rbft.EpochInfo
	stateLedger  ledger.StateLedger

	AvailableStakingPools []uint64
	RewardRecordTable     map[uint64]*big.Int
}

type StakingManager interface {
	// Internal functions(called by System Contract)
	InternalTurnIntoNewEpoch() error
	// after node leave validator set or banded by governance
	InternalDisablePool(poolID uint64) error

	// include block gasfee reword + stake reward
	InternalRecordReward(poolID uint64, reward *big.Int)

	InternalCalculateStakeReward() *big.Int

	// Public functions(called by User)
	// Operator crate StakingPool for this node
	CrateStakingPool(id uint64) (err error)

	// call StakingPool.AddStake
	AddStake(poolID uint64, owner string, amount *big.Int) error
	// call StakingPool.UnlockStake
	UnlockStake(poolID uint64, owner string, amount *big.Int) error
	// call StakingPool.WithdrawStake
	WithdrawStake(poolID uint64, owner string, amount *big.Int) error

	// call StakingPool.UpdateCommissionRate
	UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error
	// call StakingPool.GetActiveStake
	GetPoolActiveStake(poolID uint64) *big.Int
	// call StakingPool.GetNextEpochActiveStake
	GetPoolNextEpochActiveStake(poolID uint64) *big.Int
	// call StakingPool.GetCommissionRate
	GetPoolCommissionRate(poolID uint64) uint64
	// call StakingPool.GetNextEpochCommissionRate
	GetPoolNextEpochCommissionRate(poolID uint64) uint64
	// call StakingPool.GetCumulativeReward
	GetPoolCumulativeReward(poolID uint64) *big.Int
}
