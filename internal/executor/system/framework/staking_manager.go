package framework

import (
	"math"
	"math/big"
	"sort"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/liquid_staking_token"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager_client"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrStakeValue = errors.New("total supply smaller than total stake")
)

const (
	blocksPerYear                          = uint64(63072000)
	rewardPerBlockStorageKey               = "stakeRewardPerBlock"
	availablePoolsStorageKey               = "availablePools"
	proposeBlockCountTableStorageKey       = "proposeBlockCountTable"
	gasRewardTableStorageKey               = "gasRewardTable"
	totalStakeStorageKey                   = "totalStake"
	lastEpochTotalStakeStorageKey          = "lsatEpochTotalStake"
	currentEpochTotalAddStakeStorageKey    = "currentEpochTotalAddStake"
	currentEpochTotalUnlockStakeStorageKey = "currentEpochTotalUnlockStake"
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

	availablePools               *common.VMSlot[[]uint64]
	rewardPerBlock               *common.VMSlot[*big.Int]
	proposeBlockCountTable       *common.VMMap[uint64, uint64]
	gasRewardTable               *common.VMMap[uint64, *big.Int]
	totalStake                   *common.VMSlot[*big.Int]
	lastEpochTotalStake          *common.VMSlot[*big.Int]
	currentEpochTotalAddStake    *common.VMSlot[*big.Int]
	currentEpochTotalUnlockStake *common.VMSlot[*big.Int]
}

func (s *StakingManager) GenesisInit(genesis *repo.GenesisConfig) error {
	// init staking pools
	totalStake := big.NewInt(0)
	var needCreateStakingPoolIDs []uint64
	for i, nodeCfg := range genesis.Nodes {
		if !nodeCfg.IsDataSyncer {
			nodeID := uint64(i + 1)
			stakeNumber := nodeCfg.StakeNumber.ToBigInt()
			if err := s.CreatePoolWithStake(nodeID, nodeCfg.CommissionRate, stakeNumber, true); err != nil {
				return errors.Wrapf(err, "failed to create staking pool for node %d", nodeID)
			}
			if err := s.LoadPool(nodeID).GenesisInit(genesis); err != nil {
				return errors.Wrapf(err, "failed to genesis init staking pool for node %d", nodeID)
			}

			totalStake = totalStake.Add(totalStake, stakeNumber)
			needCreateStakingPoolIDs = append(needCreateStakingPoolIDs, nodeID)
		}
	}

	s.StateAccount.AddBalance(totalStake)
	if err := s.totalStake.Put(totalStake); err != nil {
		return err
	}
	if err := s.lastEpochTotalStake.Put(totalStake); err != nil {
		return err
	}
	if err := s.currentEpochTotalAddStake.Put(new(big.Int)); err != nil {
		return err
	}
	if err := s.currentEpochTotalUnlockStake.Put(new(big.Int)); err != nil {
		return err
	}

	// init stake reward per block
	if err := s.UpdateStakeRewardPerBlock(); err != nil {
		return err
	}

	if err := s.availablePools.Put(needCreateStakingPoolIDs); err != nil {
		return err
	}

	return nil
}

func (s *StakingManager) SetContext(ctx *common.VMContext) {
	s.SystemContractBase.SetContext(ctx)

	s.availablePools = common.NewVMSlot[[]uint64](s.StateAccount, availablePoolsStorageKey)
	s.proposeBlockCountTable = common.NewVMMap[uint64, uint64](s.StateAccount, proposeBlockCountTableStorageKey, func(key uint64) string {
		return strconv.FormatUint(key, 10)
	})
	s.gasRewardTable = common.NewVMMap[uint64, *big.Int](s.StateAccount, gasRewardTableStorageKey, func(key uint64) string {
		return strconv.FormatUint(key, 10)
	})
	s.rewardPerBlock = common.NewVMSlot[*big.Int](s.StateAccount, rewardPerBlockStorageKey)
	s.totalStake = common.NewVMSlot[*big.Int](s.StateAccount, totalStakeStorageKey)

	s.lastEpochTotalStake = common.NewVMSlot[*big.Int](s.StateAccount, lastEpochTotalStakeStorageKey)
	s.currentEpochTotalAddStake = common.NewVMSlot[*big.Int](s.StateAccount, currentEpochTotalAddStakeStorageKey)
	s.currentEpochTotalUnlockStake = common.NewVMSlot[*big.Int](s.StateAccount, currentEpochTotalUnlockStakeStorageKey)
}

func (s *StakingManager) LoadPool(poolID uint64) *StakingPool {
	return NewStakingPool(s.SystemContractBase).Load(poolID)
}

func (s *StakingManager) TurnIntoNewEpoch(oldEpoch *types.EpochInfo, newEpoch *types.EpochInfo) error {
	axc := token.AXCBuildConfig.Build(s.CrossCallSystemContractContext())
	pools, err := s.availablePools.MustGet()
	if err != nil {
		return err
	}
	for _, poolID := range pools {
		rewardPerBlock, err := s.rewardPerBlock.MustGet()
		if err != nil {
			return err
		}

		cnt, err := s.proposeBlockCountTable.MustGet(poolID)
		if err != nil {
			return err
		}

		exists, gasReward, err := s.gasRewardTable.Get(poolID)
		if err != nil {
			return err
		}
		if !exists {
			gasReward = big.NewInt(0)
		}

		reward := new(big.Int).Mul(new(big.Int).SetUint64(cnt), rewardPerBlock)
		// mint some token to the axc contract
		if err = axc.Mint(reward); err != nil {
			return err
		}
		// add gas to the reward
		reward = reward.Add(reward, gasReward)

		// transfer token to staking manager contract
		axc.StateAccount.SubBalance(reward)
		s.StateAccount.AddBalance(reward)
		if err = s.LoadPool(poolID).TurnIntoNewEpoch(oldEpoch, newEpoch, reward); err != nil {
			return err
		}
		if err = s.proposeBlockCountTable.Put(poolID, 0); err != nil {
			return err
		}
		if err = s.gasRewardTable.Put(poolID, big.NewInt(0)); err != nil {
			return err
		}
	}

	// reset vars in epoch
	if err := s.currentEpochTotalAddStake.Put(new(big.Int)); err != nil {
		return err
	}
	if err := s.currentEpochTotalUnlockStake.Put(new(big.Int)); err != nil {
		return err
	}
	totalStake, err := s.totalStake.MustGet()
	if err != nil {
		return err
	}
	if err := s.lastEpochTotalStake.Put(totalStake); err != nil {
		return err
	}
	return s.UpdateStakeRewardPerBlock()
}

func (s *StakingManager) DisablePool(poolID uint64) error {
	if !s.LoadPool(poolID).Exists() {
		return errors.New("pool not exists")
	}
	if err := s.removeAvailableStakingPool(poolID); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) RecordReward(poolID uint64, gasReward *big.Int) (stakeReword *big.Int, err error) {
	poolReward, err := s.proposeBlockCountTable.MustGet(poolID)
	if err != nil {
		return nil, err
	}
	poolReward++
	if err = s.proposeBlockCountTable.Put(poolID, poolReward); err != nil {
		return nil, err
	}

	// record gas reward
	exist, oldGasReward, err := s.gasRewardTable.Get(poolID)
	if err != nil {
		return nil, err
	}
	if !exist {
		oldGasReward = big.NewInt(0)
	}
	newGasReward := new(big.Int).Add(oldGasReward, gasReward)
	if err = s.gasRewardTable.Put(poolID, newGasReward); err != nil {
		return nil, err
	}

	return s.rewardPerBlock.MustGet()
}

func (s *StakingManager) UpdateStakeRewardPerBlock() error {
	axc := token.AXCBuildConfig.Build(s.CrossCallSystemContractContext())

	totalSupply, err := axc.TotalSupply()
	if err != nil {
		return err
	}
	totalStake, err := s.totalStake.MustGet()
	if err != nil {
		return err
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
	return s.rewardPerBlock.Put(multipliedInt)
}

func (s *StakingManager) CreatePool(poolID uint64, commissionRate uint64) (err error) {
	return s.CreatePoolWithStake(poolID, commissionRate, big.NewInt(0), false)
}

func (s *StakingManager) CreatePoolWithStake(poolID uint64, commissionRate uint64, stake *big.Int, isGenesisInit bool) (err error) {
	epochManagerContract := EpochManagerBuildConfig.Build(s.CrossCallSystemContractContext())
	nodeManagerContract := NodeManagerBuildConfig.Build(s.CrossCallSystemContractContext())

	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}
	nodeInfo, err := nodeManagerContract.GetInfo(poolID)
	if err != nil {
		return err
	}

	if err := s.LoadPool(poolID).Create(nodeInfo.Operator, currentEpoch.Epoch, commissionRate, stake); err != nil {
		return err
	}

	if !isGenesisInit {
		if err := s.addAvailableStakingPool(poolID); err != nil {
			return err
		}
	}

	if err = s.proposeBlockCountTable.Put(poolID, 0); err != nil {
		return err
	}
	if err = s.gasRewardTable.Put(poolID, big.NewInt(0)); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) AddStake(poolID uint64, owner ethcommon.Address, amount *big.Int) error {
	epochManagerContract := EpochManagerBuildConfig.Build(s.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}
	if currentEpoch.StakeParams.MinDelegateStake.Cmp(amount) > 0 {
		return errors.Errorf("amount %s less than min stake %s", amount.String(), currentEpoch.StakeParams.MinDelegateStake.String())
	}

	// check add stake limit
	currentEpochTotalAddStake, err := s.currentEpochTotalAddStake.MustGet()
	if err != nil {
		return err
	}
	lastEpochTotalStake, err := s.lastEpochTotalStake.MustGet()
	if err != nil {
		return err
	}
	currentEpochTotalAddStake = new(big.Int).Add(currentEpochTotalAddStake, amount)
	limit := lastEpochTotalStake.Mul(lastEpochTotalStake, big.NewInt(int64(currentEpoch.StakeParams.MaxAddStakeRatio)))
	limit = limit.Div(limit, big.NewInt(types.RatioLimit))
	if currentEpochTotalAddStake.Cmp(limit) > 0 {
		return errors.Errorf("add stake reach epoch limit %s", limit.String())
	}

	if err := s.updateTotalStake(true, amount); err != nil {
		return err
	}

	if err := s.LoadPool(poolID).AddStake(owner, amount); err != nil {
		return err
	}

	if err := s.currentEpochTotalAddStake.Put(currentEpochTotalAddStake); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) checkLiquidStakingTokenPermission(liquidStakingTokenID *big.Int) (*liquid_staking_token.LiquidStakingTokenInfo, error) {
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(s.CrossCallSystemContractContext())
	liquidStakingTokenInfo, err := liquidStakingTokenContract.mustGetInfo(liquidStakingTokenID)
	if err != nil {
		return nil, err
	}
	owner, err := liquidStakingTokenContract.ownerOf(liquidStakingTokenID)
	if err != nil {
		return nil, err
	}
	if s.Ctx.From != owner {
		return nil, errors.New("no permission")
	}
	return liquidStakingTokenInfo, nil
}

func (s *StakingManager) checkPoolPermission(poolID uint64) error {
	nodeManagerContract := NodeManagerBuildConfig.Build(s.CrossCallSystemContractContext())
	nodeInfo, err := nodeManagerContract.GetInfo(poolID)
	if err != nil {
		return err
	}

	if s.Ctx.From != nodeInfo.Operator {
		return errors.New("no permission")
	}
	return nil
}

func (s *StakingManager) Unlock(liquidStakingTokenID *big.Int, amount *big.Int) error {
	return s.BatchUnlock([]*big.Int{liquidStakingTokenID}, []*big.Int{amount})
}

func (s *StakingManager) Withdraw(liquidStakingTokenID *big.Int, recipient ethcommon.Address, amount *big.Int) error {
	return s.BatchWithdraw([]*big.Int{liquidStakingTokenID}, recipient, []*big.Int{amount})
}

func (s *StakingManager) BatchUnlock(liquidStakingTokenIDs []*big.Int, amounts []*big.Int) error {
	if len(liquidStakingTokenIDs) != len(amounts) || len(liquidStakingTokenIDs) == 0 {
		return errors.New("invalid args")
	}

	// check add stake limit
	currentEpochTotalUnlockStake, err := s.currentEpochTotalUnlockStake.MustGet()
	if err != nil {
		return err
	}
	lastEpochTotalStake, err := s.lastEpochTotalStake.MustGet()
	if err != nil {
		return err
	}

	totalAmount := big.NewInt(0)
	for _, amount := range amounts {
		totalAmount = new(big.Int).Add(totalAmount, amount)
	}
	currentEpochTotalUnlockStake = new(big.Int).Add(currentEpochTotalUnlockStake, totalAmount)
	epochManagerContract := EpochManagerBuildConfig.Build(s.CrossCallSystemContractContext())
	currentEpoch, err := epochManagerContract.CurrentEpoch()
	if err != nil {
		return err
	}
	limit := lastEpochTotalStake.Mul(lastEpochTotalStake, big.NewInt(int64(currentEpoch.StakeParams.MaxUnlockStakeRatio)))
	limit = limit.Div(limit, big.NewInt(types.RatioLimit))
	if currentEpochTotalUnlockStake.Cmp(limit) > 0 {
		return errors.Errorf("unlock stake reach epoch limit %s", limit.String())
	}

	for i, liquidStakingTokenID := range liquidStakingTokenIDs {
		liquidStakingTokenInfo, err := s.checkLiquidStakingTokenPermission(liquidStakingTokenID)
		if err != nil {
			return err
		}
		if err := s.LoadPool(liquidStakingTokenInfo.PoolID).UnlockStake(liquidStakingTokenID, liquidStakingTokenInfo, amounts[i]); err != nil {
			return err
		}
	}

	if err := s.updateTotalStake(false, totalAmount); err != nil {
		return err
	}

	if err := s.currentEpochTotalUnlockStake.Put(currentEpochTotalUnlockStake); err != nil {
		return err
	}

	return nil
}

func (s *StakingManager) BatchWithdraw(liquidStakingTokenIDs []*big.Int, recipient ethcommon.Address, amounts []*big.Int) error {
	if len(liquidStakingTokenIDs) != len(amounts) || len(liquidStakingTokenIDs) == 0 {
		return errors.New("invalid args")
	}

	totalPrincipalWithdraw := big.NewInt(0)
	for i, liquidStakingTokenID := range liquidStakingTokenIDs {
		liquidStakingTokenInfo, err := s.checkLiquidStakingTokenPermission(liquidStakingTokenID)
		if err != nil {
			return err
		}
		principalWithdraw, err := s.LoadPool(liquidStakingTokenInfo.PoolID).WithdrawStake(liquidStakingTokenID, liquidStakingTokenInfo, recipient, amounts[i])
		if err != nil {
			return err
		}
		totalPrincipalWithdraw = totalPrincipalWithdraw.Add(totalPrincipalWithdraw, principalWithdraw)
	}
	if err := s.updateTotalStake(false, totalPrincipalWithdraw); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) UpdatePoolCommissionRate(poolID uint64, newCommissionRate uint64) error {
	if err := s.checkPoolPermission(poolID); err != nil {
		return err
	}
	return s.LoadPool(poolID).UpdateCommissionRate(newCommissionRate)
}

func (s *StakingManager) PoolActiveStake(poolID uint64) (*big.Int, error) {
	info, err := s.LoadPool(poolID).MustGetInfo()
	if err != nil {
		return nil, err
	}
	return new(big.Int).Set(info.ActiveStake), nil
}

func (s *StakingManager) GetPoolInfo(poolID uint64) (staking_manager.PoolInfo, error) {
	return s.LoadPool(poolID).Info()
}

func (s *StakingManager) GetPoolHistoryLiquidStakingTokenRate(poolID uint64, epoch uint64) (staking_manager.LiquidStakingTokenRate, error) {
	return s.LoadPool(poolID).HistoryLiquidStakingTokenRate(epoch)
}

func (s *StakingManager) updateTotalStake(isAdd bool, amount *big.Int) error {
	if amount == nil {
		return nil
	}
	if amount.Sign() == 0 {
		return nil
	}
	totalStake, err := s.totalStake.MustGet()
	if err != nil {
		return err
	}
	if isAdd {
		totalStake = totalStake.Add(totalStake, amount)
	} else {
		if totalStake.Cmp(amount) < 0 {
			return errors.Errorf("total stake %s less than amount %s", totalStake.String(), amount.String())
		}
		totalStake = totalStake.Sub(totalStake, amount)
	}

	if err = s.totalStake.Put(totalStake); err != nil {
		return err
	}
	return nil
}

func (s *StakingManager) addAvailableStakingPool(poolID uint64) error {
	pools, err := s.availablePools.MustGet()
	if err != nil {
		return err
	}
	_, found := lo.Find(pools, func(id uint64) bool {
		return id == poolID
	})
	if found {
		return errors.Errorf("pool %d already available", poolID)
	}
	pools = append(pools, poolID)
	sort.Slice(pools, func(i, j int) bool { return pools[i] < pools[j] })
	return s.availablePools.Put(pools)
}

func (s *StakingManager) removeAvailableStakingPool(poolID uint64) error {
	pools, err := s.availablePools.MustGet()
	if err != nil {
		return err
	}
	_, index, found := lo.FindIndexOf(pools, func(id uint64) bool {
		return id == poolID
	})
	if !found {
		return errors.Errorf("pool %d not available", poolID)
	}
	pools = append(pools[:index], pools[index+1:]...)
	sort.Slice(pools, func(i, j int) bool { return pools[i] < pools[j] })
	return s.availablePools.Put(pools)
}

func divideBigInt(a, b *big.Int) float64 {
	// Perform division
	result := new(big.Float).Quo(new(big.Float).SetInt(a), new(big.Float).SetInt(b))
	floatResult, _ := result.Float64()

	return floatResult
}
