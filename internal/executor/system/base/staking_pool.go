package base

import (
	"math/big"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

const (
	CommissionRateDenominator = 10000
)

type LiquidStakingTokenRate struct {
	StakeAmount                   *big.Int
	TotalLiquidStakingTokenAmount *big.Int
}

type StakingPool interface {
	// NOTICE: Add stake and mint an LiquidStakingToken, pendingActive LiquidStakingToken can quick Withdraw and not need call UnlockStake
	// set PendingActiveStake = [PendingActiveStake + amount]
	// mint a LiquidStakingToken, set LiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
	// emit a AddStake event
	// emit a MintLiquidStakingToken event
	AddStake(owner string, amount *big.Int) error

	// NOTICE: Unlock the specified amount and enter the lockin period and add a record to UnlockingRecords to liquidStakingToken, will refresh the LiquidStakingToken.ActiveEpoch and LiquidStakingToken.Principal(first calculate reward)
	// NOTICE: Unable to obtain reward after calling unlock, and will revert when the LiquidStakingToken is pendingActive
	// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
	// set PendingInactiveStake = [PendingInactiveStake + amount]
	// append UnlockingRecord{Amount: amount, UnlockTimestamp: block.Timestamp} to LiquidStakingToken.UnlockingRecords
	UnlockStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error

	// NOTICE: Withdraw the unlocked amount from LiquidStakingToken.Unlocked
	// move unlocked amount from LiquidStakingToken.UnlockingRecords to LiquidStakingToken.Unlocked
	// set LiquidStakingToken.Unlocked = LiquidStakingToken.Unlocked - amount
	WithdrawStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error

	// Distribute reward to staking pool, and process stake status changes
	// calculate stakerReward =  reward * (1 - CommissionRate / CommissionRateDenominator)
	// calculate operatorCommission = reward * CommissionRate / CommissionRateDenominator
	// set ActiveStake = [ActiveStake + PendingActiveStake - PendingInactiveStake + stakerReward]
	// set TotalLiquidStakingToken = LiquidStakingTokenRateHistory[CurrentEpoch - 1].TotalLiquidStakingTokenAmount / LiquidStakingTokenRateHistory[CurrentEpoch - 1].TotalLiquidStakingTokenAmount * ActiveStake
	// set LiquidStakingTokenRateHistory[CurrentEpoch] = LiquidStakingTokenRate{StakeAmount: ActiveStake, TotalLiquidStakingTokenAmount: TotalLiquidStakingToken}
	// if OperatorLiquidStakingTokenID == 0 (mint a new LiquidStakingToken)
	//		set OperatorLiquidStakingTokenID = mint new LiquidStakingToken{ ActiveEpoch: CurrentEpoch+1, Principal: operatorCommission, UnlockingRecords: []UnlockingRecord{}}
	// else(merge LiquidStakingToken)
	//		calculate totalTokens = GetLockedToken(OperatorLiquidStakingTokenID)
	//      set OperatorLiquidStakingToken.Principal = totalTokens + operatorCommission
	//      set OperatorLiquidStakingToken.ActiveEpoch = CurrentEpoch + 1
	// set CommissionRate = NextEpochCommissionRate
	TurnIntoNewEpoch(reward *big.Int) error
	// set NextEpochCommissionRate = newCommissionRate
	UpdateCommissionRate(newCommissionRate uint64) error

	// set OperatorLiquidStakingTokenID = 0
	UpdateOperator(newOperator string) error

	GetActiveStake() *big.Int
	GetNextEpochActiveStake() *big.Int
	GetCommissionRate() uint64
	GetNextEpochCommissionRate() uint64
	GetCumulativeReward() *big.Int
}

type stakingPool struct {
	currentEpoch *rbft.EpochInfo
	stateLedger  ledger.StateLedger

	// PoolID is the staking pool id, which is also the Node ID
	PoolID uint64
	// if false, the node cannot be elected as a validator, and will not be added stake to StakingPool
	IsActive                      bool
	ActiveStake                   *big.Int
	TotalLiquidStakingToken       *big.Int
	LiquidStakingTokenRateHistory map[uint64]*LiquidStakingTokenRate
	PendingActiveStake            *big.Int
	PendingInactiveStake          *big.Int
	// 1000 = 10%, range: [0, CommissionRateDenominator]
	CommissionRate               uint64
	NextEpochCommissionRate      uint64
	CumulativeReward             *big.Int
	OperatorLiquidStakingTokenID *big.Int
}

func GetStakingPool(stateLedger ledger.StateLedger, currentEpoch *rbft.EpochInfo, poolID uint64) (StakingPool, error) {
	return nil, nil
}

func HasStakingPool(stateLedger ledger.StateLedger, poolId uint64) bool {
	return false
}

func CreateStakingPool(stateLedger ledger.StateLedger, currentEpoch *rbft.EpochInfo, poolID uint64, commissionRate uint64) (StakingPool, error) {
	return &stakingPool{}, nil
}

func (sp *stakingPool) AddStake(owner string, amount *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) UnlockStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) WithdrawStake(owner string, liquidStakingTokenID *big.Int, amount *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) TurnIntoNewEpoch(reward *big.Int) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) UpdateCommissionRate(newCommissionRate uint64) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) UpdateOperator(newOperator string) error {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) GetActiveStake() *big.Int {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) GetNextEpochActiveStake() *big.Int {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) GetCommissionRate() uint64 {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) GetNextEpochCommissionRate() uint64 {
	//TODO implement me
	panic("implement me")
}

func (sp *stakingPool) GetCumulativeReward() *big.Int {
	//TODO implement me
	panic("implement me")
}
