package framework

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
)

func setEpochParam(nvm *common.TestNVM) {
	nvm.Rep.GenesisConfig.EpochInfo.StakeParams.EnablePartialUnlock = true
	nvm.Rep.GenesisConfig.EpochInfo.StakeParams.MinDelegateStake = types.CoinNumberByMol(1)
	nvm.Rep.GenesisConfig.EpochInfo.StakeParams.MinValidatorStake = types.CoinNumberByMol(1)
}

func TestStakingPool_ManageInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	setEpochParam(testNVM)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, nodeManagerContract, stakingManagerContract)

	operator := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")

	// query not exists pool
	testNVM.Call(stakingManagerContract, operator, func() {
		pool := stakingManagerContract.LoadPool(5)
		assert.False(t, pool.Exists())

		info, err := pool.Info()
		assert.Nil(t, err)
		assert.EqualValues(t, 0, info.ID)

		_, err = pool.MustGetInfo()
		assert.ErrorContains(t, err, "does not exist")
	})

	// create without value
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		pool := stakingManagerContract.LoadPool(5)
		_, err := pool.Create(operator, 1, 5000, big.NewInt(0))
		assert.Nil(t, err)

		// transfer coin to staking pool manager
		stakingManagerContract.StateAccount.AddBalance(big.NewInt(100))

		info, err := pool.Info()
		assert.Nil(t, err)
		assert.EqualValues(t, 5, info.ID)
		assert.True(t, info.IsActive)
		assert.EqualValues(t, 0, info.ActiveStake.Int64())
		assert.EqualValues(t, 0, info.TotalLiquidStakingToken.Int64())
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())
		assert.EqualValues(t, 5000, info.CommissionRate)
		assert.EqualValues(t, 5000, info.NextEpochCommissionRate)
		assert.Zero(t, info.CumulativeReward.Int64())
		assert.Zero(t, info.CumulativeCommission.Int64())
		assert.NotZero(t, info.OperatorLiquidStakingTokenID.Int64())
		return err
	})

	// create with value
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		pool := stakingManagerContract.LoadPool(6)
		_, err := pool.Create(operator, 1, 5000, big.NewInt(100))
		assert.Nil(t, err)

		// transfer coin to staking pool manager
		stakingManagerContract.StateAccount.AddBalance(big.NewInt(100))

		info, err := pool.Info()
		assert.Nil(t, err)
		assert.EqualValues(t, 6, info.ID)
		assert.True(t, info.IsActive)
		assert.EqualValues(t, 100, info.ActiveStake.Int64())
		assert.EqualValues(t, 100, info.TotalLiquidStakingToken.Int64())
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())
		assert.EqualValues(t, 5000, info.CommissionRate)
		assert.EqualValues(t, 5000, info.NextEpochCommissionRate)
		assert.Zero(t, info.CumulativeReward.Int64())
		assert.Zero(t, info.CumulativeCommission.Int64())
		assert.NotZero(t, info.OperatorLiquidStakingTokenID.Int64())

		rate, err := pool.HistoryLiquidStakingTokenRate(1)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, rate.StakeAmount.Int64())
		assert.EqualValues(t, 0, rate.LiquidStakingTokenAmount.Int64())

		return err
	})

	// UpdateCommissionRate
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		pool := stakingManagerContract.LoadPool(6)

		info, err := pool.MustGetInfo()
		assert.Nil(t, err)

		err = pool.UpdateCommissionRate(CommissionRateDenominator + 1)
		assert.ErrorContains(t, err, "invalid commission rate")

		err = pool.UpdateCommissionRate(CommissionRateDenominator - 1)
		assert.Nil(t, err)

		newInfo, err := pool.MustGetInfo()
		assert.Nil(t, err)
		assert.EqualValues(t, info.CommissionRate, newInfo.CommissionRate)
		assert.EqualValues(t, CommissionRateDenominator-1, newInfo.NextEpochCommissionRate)

		return err
	})
}

func TestStakingPool_StakeByUser(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	setEpochParam(testNVM)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.EnablePartialUnlock = false
	testNVM.Rep.GenesisConfig.Nodes[0].IsDataSyncer = false
	testNVM.Rep.GenesisConfig.Nodes[0].StakeNumber = types.CoinNumberByMol(10000)
	testNVM.Rep.GenesisConfig.Nodes[0].CommissionRate = 4000
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, nodeManagerContract, stakingManagerContract)

	operator := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")

	userA := ethcommon.HexToAddress("0xA000000000000000000000000000000000000000")
	userB := ethcommon.HexToAddress("0xB000000000000000000000000000000000000000")
	// add stake
	addStake := func(user ethcommon.Address, amount int64) *big.Int {
		var lstTokenID *big.Int
		testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
			epochManagerContract.SetContext(stakingManagerContract.Ctx)
			liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

			stakingManagerContract.Ctx.From = user
			stakingManagerContract.Ctx.Value = big.NewInt(amount)
			err := stakingManagerContract.AddStake(1, user, big.NewInt(amount))
			assert.Nil(t, err)
			lstTokenID = stakingManagerContract.Ctx.TestLogs[0].(*staking_manager.EventAddStake).LiquidStakingTokenID
			return err
		})
		return lstTokenID
	}
	turnIntoNewEpoch := func() {
		testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
			epochManagerContract.SetContext(stakingManagerContract.Ctx)
			liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
			pool := stakingManagerContract.LoadPool(1)

			oldEpoch, err := epochManagerContract.CurrentEpoch()
			assert.Nil(t, err)
			// turn into new epoch
			newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
			assert.Nil(t, err)
			err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(10000))
			assert.Nil(t, err)
			return err
		})
	}

	batchUnlock := func(user ethcommon.Address, liquidStakingTokenIDs []*big.Int) {
		testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
			epochManagerContract.SetContext(stakingManagerContract.Ctx)
			liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

			stakingManagerContract.Ctx.From = user
			amounts := make([]*big.Int, len(liquidStakingTokenIDs))
			var err error
			for i := range liquidStakingTokenIDs {
				amounts[i], err = liquidStakingTokenContract.GetTotalCoin(liquidStakingTokenIDs[i])
				assert.Nil(t, err)
			}
			err = stakingManagerContract.BatchUnlock(liquidStakingTokenIDs, amounts)
			assert.Nil(t, err)

			for i := range liquidStakingTokenIDs {
				lockedCoin, err := liquidStakingTokenContract.GetLockedCoin(liquidStakingTokenIDs[i])
				assert.Nil(t, err)
				assert.EqualValues(t, 0, lockedCoin.Int64())
				unlockingCoin, err := liquidStakingTokenContract.GetUnlockingCoin(liquidStakingTokenIDs[i])
				assert.Nil(t, err)
				assert.EqualValues(t, amounts[i].Int64(), unlockingCoin.Int64())
			}
			return err
		})
	}
	checkEmptyLockedCoin := func(liquidStakingTokenIDs []*big.Int) {
		testNVM.Call(stakingManagerContract, operator, func() {
			epochManagerContract.SetContext(stakingManagerContract.Ctx)
			liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

			for i := range liquidStakingTokenIDs {
				lockedCoin, err := liquidStakingTokenContract.GetLockedCoin(liquidStakingTokenIDs[i])
				assert.Nil(t, err)
				assert.EqualValues(t, 0, lockedCoin.Int64())
			}
		})
	}

	userALSTTokenID1 := addStake(userA, 30000)
	userALSTTokenID2 := addStake(userA, 30000)
	userALSTTokenID3 := addStake(userA, 30000)

	userBLSTTokenID1 := addStake(userB, 30000)
	userBLSTTokenID2 := addStake(userB, 30000)

	turnIntoNewEpoch()

	userALSTTokenID4 := addStake(userA, 30000)
	turnIntoNewEpoch()

	userALSTTokenID5 := addStake(userA, 30000)
	turnIntoNewEpoch()

	userALSTTokenID6 := addStake(userA, 30000)
	userALSTTokenID7 := addStake(userA, 30000)

	userBLSTTokenID3 := addStake(userB, 30000)
	turnIntoNewEpoch()

	turnIntoNewEpoch()
	turnIntoNewEpoch()

	turnIntoNewEpoch()

	batchUnlock(userA, []*big.Int{userALSTTokenID1, userALSTTokenID2, userALSTTokenID4})
	batchUnlock(userB, []*big.Int{userBLSTTokenID1, userBLSTTokenID3})
	turnIntoNewEpoch()
	checkEmptyLockedCoin([]*big.Int{userALSTTokenID1, userALSTTokenID2, userALSTTokenID4})
	checkEmptyLockedCoin([]*big.Int{userBLSTTokenID1, userBLSTTokenID3})

	batchUnlock(userA, []*big.Int{userALSTTokenID3, userALSTTokenID5, userALSTTokenID6})
	turnIntoNewEpoch()
	checkEmptyLockedCoin([]*big.Int{userALSTTokenID3, userALSTTokenID5, userALSTTokenID6})

	testNVM.Call(stakingManagerContract, userA, func() {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

		liquidStakingTokenIDs := []*big.Int{userBLSTTokenID2, userALSTTokenID7}
		stakingManagerContract.Ctx.From = userA
		amounts := make([]*big.Int, len(liquidStakingTokenIDs))
		var err error
		for i := range liquidStakingTokenIDs {
			amounts[i], err = liquidStakingTokenContract.GetTotalCoin(liquidStakingTokenIDs[i])
			assert.Nil(t, err)
		}
		err = stakingManagerContract.BatchUnlock(liquidStakingTokenIDs, amounts)
		assert.ErrorContains(t, err, "no permission")
	})

	testNVM.Call(stakingManagerContract, userA, func() {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

		liquidStakingTokenIDs := []*big.Int{userALSTTokenID7}
		stakingManagerContract.Ctx.From = userA
		amounts := make([]*big.Int, len(liquidStakingTokenIDs))
		var err error
		for i := range liquidStakingTokenIDs {
			amounts[i] = big.NewInt(1)
		}
		err = stakingManagerContract.BatchUnlock(liquidStakingTokenIDs, amounts)
		assert.ErrorContains(t, err, "not enable partial unlock")
	})

	turnIntoNewEpoch()

	batchUnlock(userA, []*big.Int{userALSTTokenID7})
	batchUnlock(userB, []*big.Int{userBLSTTokenID2})
	turnIntoNewEpoch()
	checkEmptyLockedCoin([]*big.Int{userALSTTokenID7, userBLSTTokenID2})

	turnIntoNewEpoch()
	turnIntoNewEpoch()
	checkEmptyLockedCoin([]*big.Int{userALSTTokenID1, userALSTTokenID2, userALSTTokenID3, userALSTTokenID4, userALSTTokenID5, userALSTTokenID6, userALSTTokenID7, userBLSTTokenID1, userBLSTTokenID2, userBLSTTokenID3})

	testNVM.RunSingleTX(stakingManagerContract, userA, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)

		beforeBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(types.NewAddress(userA.Bytes()))
		liquidStakingTokenIDs := []*big.Int{userALSTTokenID1, userALSTTokenID2, userALSTTokenID3, userALSTTokenID4, userALSTTokenID5, userALSTTokenID6, userALSTTokenID7}
		amounts := make([]*big.Int, len(liquidStakingTokenIDs))
		totalAmount := big.NewInt(0)
		var err error
		for i := range liquidStakingTokenIDs {
			amounts[i], err = liquidStakingTokenContract.GetUnlockedCoin(liquidStakingTokenIDs[i])
			assert.Nil(t, err)
			assert.NotEqualValues(t, 0, amounts[i])
			totalAmount = totalAmount.Add(totalAmount, amounts[i])
		}
		assert.NotEqualValues(t, 0, totalAmount.Uint64())

		err = stakingManagerContract.BatchWithdraw(liquidStakingTokenIDs, userA, amounts)
		assert.Nil(t, err)

		afterBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(types.NewAddress(userA.Bytes()))
		assert.Equal(t, totalAmount.Uint64(), new(big.Int).Sub(afterBalance, beforeBalance).Uint64())
		return err
	}, func(ctx *common.VMContext) {
		ctx.CurrentEVM.Context.Time = uint64(time.Now().Unix()) + testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.UnlockPeriod
	})
}

func TestStakingPool_ManageStake(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	setEpochParam(testNVM)
	testNVM.Rep.GenesisConfig.Nodes[0].IsDataSyncer = false
	testNVM.Rep.GenesisConfig.Nodes[0].StakeNumber = types.CoinNumberByMol(10000)
	testNVM.Rep.GenesisConfig.Nodes[0].CommissionRate = 4000
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, nodeManagerContract, stakingManagerContract)

	operator := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")

	// epoch 1, reward = 10000, operator get all rewards
	// [operator] before: 10000, after: 20000
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)

		info, err := pool.MustGetInfo()
		assert.Nil(t, err)

		// turn into new epoch
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(10000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveStake)
		assert.EqualValues(t, 20000, info.ActiveStake.Int64())
		assert.EqualValues(t, 10000*20000/16000, info.TotalLiquidStakingToken.Int64())
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		rate, err := pool.historyLiquidStakingTokenRateMap.MustGet(oldEpoch.Epoch)
		assert.Nil(t, err)
		assert.EqualValues(t, info.ActiveStake.Int64(), rate.StakeAmount.Int64())
		assert.EqualValues(t, info.TotalLiquidStakingToken.Int64(), rate.LiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assert.EqualValues(t, 20000, operatorLSTInfo.Principal.Int64())
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.Empty(t, operatorLSTInfo.UnlockingRecords)
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 20000, totalCoin.Int64())
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID)

		return err
	})

	// epoch 2, reward = 10000
	// user A add stake = 30000
	// [active stake] before: 20000, after: 60000
	// [operator] before: 20000, after: 30000
	// [user A] before: 30000, after: 30000
	userA := ethcommon.HexToAddress("0xA000000000000000000000000000000000000000")
	var userALSTTokenID *big.Int
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		stakingManagerContract.Ctx.From = userA
		stakingManagerContract.Ctx.Value = big.NewInt(30000)
		err = stakingManagerContract.AddStake(1, userA, big.NewInt(30000))
		assert.Nil(t, err)
		userALSTTokenID = stakingManagerContract.Ctx.TestLogs[0].(*staking_manager.EventAddStake).LiquidStakingTokenID

		// check userA lst info
		userALSTInfo, err := liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assert.EqualValues(t, 30000, userALSTInfo.Principal.Int64())
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch+1, userALSTInfo.ActiveEpoch)
		assert.Empty(t, userALSTInfo.UnlockingRecords)
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check pool new info
		info, err := pool.MustGetInfo()
		assert.Nil(t, err)
		lastLST := info.TotalLiquidStakingToken.Int64()
		assert.EqualValues(t, 20000, info.ActiveStake.Int64())
		assert.EqualValues(t, 10000*20000/16000, info.TotalLiquidStakingToken.Int64())
		assert.EqualValues(t, 30000, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// turn into new epoch
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(10000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveStake)
		assert.EqualValues(t, 60000, info.ActiveStake.Int64())
		assertWithinGap(t, lastLST*60000/26000, info.TotalLiquidStakingToken.Int64(), 10)
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assert.EqualValues(t, 30000, operatorLSTInfo.Principal.Int64())
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.Empty(t, operatorLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check userA
		userALSTInfo, err = liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assert.EqualValues(t, 30000, userALSTInfo.Principal.Int64())
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, userALSTInfo.ActiveEpoch)
		assert.Empty(t, userALSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID)

		return err
	})

	// epoch 3, reward = 10000
	// user B add stake = 30000
	// [active stake] before: 60000, after: 100000
	// [operator] before: 30000, after: 37000 = 30000 + 6000/2 +4000
	// [user A] before: 30000, after: 33000 = 30000 + 6000/2
	// [user B] before: 30000, after: 30000
	userB := ethcommon.HexToAddress("0xB000000000000000000000000000000000000000")
	var userBLSTTokenID *big.Int
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		stakingManagerContract.Ctx.From = userB
		stakingManagerContract.Ctx.Value = big.NewInt(30000)
		err = stakingManagerContract.AddStake(1, userB, big.NewInt(30000))
		assert.Nil(t, err)
		userBLSTTokenID = stakingManagerContract.Ctx.TestLogs[0].(*staking_manager.EventAddStake).LiquidStakingTokenID

		// check userB lst info
		userBLSTInfo, err := liquidStakingTokenContract.GetInfo(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userBLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userBLSTInfo.Principal.Int64())
		assert.Zero(t, userBLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch+1, userBLSTInfo.ActiveEpoch)
		assert.Empty(t, userBLSTInfo.UnlockingRecords)
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(userBLSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check pool new info
		info, err := pool.MustGetInfo()
		assert.Nil(t, err)
		lastLST := info.TotalLiquidStakingToken.Int64()
		assert.EqualValues(t, 60000, info.ActiveStake.Int64())
		assert.EqualValues(t, 30000, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// turn into new epoch
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(10000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveLST)
		assert.EqualValues(t, 100000, info.ActiveStake.Int64())
		assertWithinGap(t, lastLST*100000/66000, info.TotalLiquidStakingToken.Int64(), 10)
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assert.EqualValues(t, 37000, operatorLSTInfo.Principal.Int64())
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.Empty(t, operatorLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 37000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check userA
		userALSTInfo, err := liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assert.EqualValues(t, 30000, userALSTInfo.Principal.Int64())
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 3, userALSTInfo.ActiveEpoch)
		assert.Empty(t, userALSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 33000, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 3000, lockedReward.Int64(), 10)

		// check userB
		userBLSTInfo, err = liquidStakingTokenContract.GetInfo(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userBLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userBLSTInfo.Principal.Int64())
		assert.Zero(t, userBLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 4, userBLSTInfo.ActiveEpoch)
		assert.Empty(t, userBLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userBLSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID, userBLSTTokenID)

		return err
	})

	// epoch 4, reward = 10000
	// [active stake] before: 100000, after: 110000
	// [operator] before: 37000, after: 43220 = 37000 + 6000 * 37000/100000 + 4000
	// [user A] before: 33000, after: 34980 = 33000 + 6000 * 33000/100000
	// [user B] before: 30000, after: 31800 = 30000 + 6000 * 30000/100000
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		info, err := pool.MustGetInfo()
		assert.Nil(t, err)
		lastLST := info.TotalLiquidStakingToken.Int64()

		// turn into new epoch
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(10000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveLST)
		assert.EqualValues(t, 110000, info.ActiveStake.Int64())
		assertWithinGap(t, lastLST*110000/106000, info.TotalLiquidStakingToken.Int64(), 10)
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assertWithinGap(t, 43220, operatorLSTInfo.Principal.Int64(), 10)
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.Empty(t, operatorLSTInfo.UnlockingRecords)
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 43220, totalCoin.Int64(), 10)
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check userA
		userALSTInfo, err := liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assert.EqualValues(t, 30000, userALSTInfo.Principal.Int64())
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 3, userALSTInfo.ActiveEpoch)
		assert.Empty(t, userALSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 34980, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 4980, lockedReward.Int64(), 10)

		// check userB
		userBLSTInfo, err := liquidStakingTokenContract.GetInfo(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userBLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userBLSTInfo.Principal.Int64())
		assert.Zero(t, userBLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 4, userBLSTInfo.ActiveEpoch)
		assert.Empty(t, userBLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 31800, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 1800, lockedReward.Int64(), 20)

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID, userBLSTTokenID)

		return err
	})

	// epoch 5, reward = 20000
	// user A unlock stake = 20000
	// user C add stake = 30000
	// [active stake] before: 110000, after: 140000 = 110000 - 20000 + 30000 +20000
	// [operator] before: 43220, after: 58133 = 43220 + 12000 * 43220/(110000-34980) + 8000
	// [user A] before: 34980, after: 14980 = 34980 - 20000
	// [user B] before: 31800, after: 36886 = 31800 + 12000 * 31800/(110000-34980)
	// [user C] before: 30000, after: 30000
	userC := ethcommon.HexToAddress("0xC000000000000000000000000000000000000000")
	var userCLSTTokenID *big.Int
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		info, err := pool.MustGetInfo()
		assert.Nil(t, err)
		lastLST := info.TotalLiquidStakingToken.Int64()

		// unlock stake
		stakingManagerContract.Ctx.From = userA
		err = stakingManagerContract.Unlock(userALSTTokenID, big.NewInt(20000))
		assert.Nil(t, err)
		// check userA lst info
		userALSTInfo, err := liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assertWithinGap(t, 14980, userALSTInfo.Principal.Int64(), 10)
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch+1, userALSTInfo.ActiveEpoch)
		assert.EqualValues(t, 1, len(userALSTInfo.UnlockingRecords))
		assert.EqualValues(t, 20000, userALSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 34980, totalCoin.Int64(), 10)
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check pool new info
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		assert.EqualValues(t, 110000, info.ActiveStake.Int64())
		assert.EqualValues(t, 14981, info.PendingActiveStake.Int64())
		assert.EqualValues(t, 34981, info.PendingInactiveStake.Int64())
		assert.NotZero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// add stake
		stakingManagerContract.Ctx.From = userC
		stakingManagerContract.Ctx.Value = big.NewInt(30000)
		err = stakingManagerContract.AddStake(1, userC, big.NewInt(30000))
		assert.Nil(t, err)
		userCLSTTokenID = stakingManagerContract.Ctx.TestLogs[1].(*staking_manager.EventAddStake).LiquidStakingTokenID

		// check userC lst info
		userCLSTInfo, err := liquidStakingTokenContract.GetInfo(userCLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userCLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userCLSTInfo.Principal.Int64())
		assert.Zero(t, userCLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch+1, userCLSTInfo.ActiveEpoch)
		assert.Empty(t, userCLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userCLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userCLSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check pool new info
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		assert.EqualValues(t, 110000, info.ActiveStake.Int64())
		assert.EqualValues(t, 44981, info.PendingActiveStake.Int64())
		assert.EqualValues(t, 34981, info.PendingInactiveStake.Int64())
		assert.NotZero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())
		pendingInactiveLiquidStakingTokenAmount := info.PendingInactiveLiquidStakingTokenAmount.Int64()

		// turn into new epoch
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(20000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveLST)
		assert.EqualValues(t, 140000, info.ActiveStake.Int64())
		assertWithinGap(t, (lastLST-pendingInactiveLiquidStakingTokenAmount)*140000/(110000+12000-34981), info.TotalLiquidStakingToken.Int64(), 10)
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assertWithinGap(t, 58133, operatorLSTInfo.Principal.Int64(), 10)
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.Empty(t, operatorLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 58133, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check userA
		// userA unlock, not get reward from epoch 5
		userALSTInfo, err = liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assertWithinGap(t, 14980, userALSTInfo.Principal.Int64(), 10)
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 6, userALSTInfo.ActiveEpoch)
		assert.EqualValues(t, 1, len(userALSTInfo.UnlockingRecords))
		assert.EqualValues(t, 20000, userALSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 34980, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 0, lockedReward.Int64(), 10)

		// check userB
		userBLSTInfo, err := liquidStakingTokenContract.GetInfo(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userBLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userBLSTInfo.Principal.Int64())
		assert.Zero(t, userBLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 4, userBLSTInfo.ActiveEpoch)
		assert.Empty(t, userBLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 36886, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 6886, lockedReward.Int64(), 10)

		// check userC
		userCLSTInfo, err = liquidStakingTokenContract.GetInfo(userCLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userCLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userCLSTInfo.Principal.Int64())
		assert.Zero(t, userCLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 6, userCLSTInfo.ActiveEpoch)
		assert.Empty(t, userCLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userCLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userCLSTTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID, userBLSTTokenID, userCLSTTokenID)

		return err
	})

	// epoch 6, reward = 20000
	// operator unlock 10000 (not get stake reward)
	// [active stake] before: 140000, after: 150000 = 140000 + 20000 - 10000
	// [operator] before: 58133, after: 56133 = 58133 - 10000 + 8000
	// [user A] before: 14980, after: 17175 = 14980 + 12000 * 14980/(140000-58133)
	// [user B] before: 36886, after: 42292 = 36886 + 12000 * 36886/(140000-58133)
	// [user C] before: 30000, after: 34397 = 30000 + 12000 * 30000/(140000-58133)
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)

		oldEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)

		info, err := pool.MustGetInfo()
		assert.Nil(t, err)
		lastLST := info.TotalLiquidStakingToken.Int64()

		// unlock stake
		stakingManagerContract.Ctx.From = operator
		err = stakingManagerContract.Unlock(info.OperatorLiquidStakingTokenID, big.NewInt(10000))
		assert.Nil(t, err)
		// check Operator lst info
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assertWithinGap(t, 48133, operatorLSTInfo.Principal.Int64(), 10)
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch+1, operatorLSTInfo.ActiveEpoch)
		assert.EqualValues(t, 1, len(operatorLSTInfo.UnlockingRecords))
		assert.EqualValues(t, 10000, operatorLSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 58133, totalCoin.Int64(), 10)
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check pool new info
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)
		assertWithinGap(t, 140000, info.ActiveStake.Int64(), 10)
		assertWithinGap(t, 48133, info.PendingActiveStake.Int64(), 10)
		assertWithinGap(t, 58133, info.PendingInactiveStake.Int64(), 10)
		assert.NotZero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())
		pendingInactiveLiquidStakingTokenAmount := info.PendingInactiveLiquidStakingTokenAmount.Int64()

		// turn into new epoch
		newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(20000))
		assert.Nil(t, err)

		// check auto restake
		info, err = pool.MustGetInfo()
		assert.Nil(t, err)

		// final rate formula:
		// stake = lastRateStake + reward - pendingInActiveStake + pendingActiveStake,
		// lst = (lastRateLST - pendingInActiveLST) * stake / (lastRateStake + (1-commissionRate)reward - pendingInActiveLST)
		assert.EqualValues(t, 150000, info.ActiveStake.Int64())
		assertWithinGap(t, (lastLST-pendingInactiveLiquidStakingTokenAmount)*150000/(140000+12000-58133), info.TotalLiquidStakingToken.Int64(), 10)
		assert.Zero(t, info.PendingActiveStake.Int64())
		assert.Zero(t, info.PendingInactiveStake.Int64())
		assert.Zero(t, info.PendingInactiveLiquidStakingTokenAmount.Int64())

		// check operator reward
		operatorLSTInfo, err = liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assertWithinGap(t, 56133, operatorLSTInfo.Principal.Int64(), 10)
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.EqualValues(t, 10000, operatorLSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 66133, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())

		// check userA
		userALSTInfo, err := liquidStakingTokenContract.GetInfo(userALSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userALSTInfo.PoolID)
		assertWithinGap(t, 14980, userALSTInfo.Principal.Int64(), 10)
		assert.Zero(t, userALSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 6, userALSTInfo.ActiveEpoch)
		assert.EqualValues(t, 1, len(userALSTInfo.UnlockingRecords))
		assert.EqualValues(t, 20000, userALSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 37175, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userALSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 17175-14980, lockedReward.Int64(), 20)

		// check userB
		userBLSTInfo, err := liquidStakingTokenContract.GetInfo(userBLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userBLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userBLSTInfo.Principal.Int64())
		assert.Zero(t, userBLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 4, userBLSTInfo.ActiveEpoch)
		assert.Empty(t, userBLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 42292, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userBLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 12292, lockedReward.Int64(), 10)

		// check userC
		userCLSTInfo, err := liquidStakingTokenContract.GetInfo(userCLSTTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, userCLSTInfo.PoolID)
		assert.EqualValues(t, 30000, userCLSTInfo.Principal.Int64())
		assert.Zero(t, userCLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, oldEpoch.Epoch, userCLSTInfo.ActiveEpoch)
		assert.Empty(t, userCLSTInfo.UnlockingRecords)
		totalCoin, err = liquidStakingTokenContract.GetTotalCoin(userCLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 34397, totalCoin.Int64(), 10)
		lockedReward, err = liquidStakingTokenContract.GetLockedReward(userCLSTTokenID)
		assert.Nil(t, err)
		assertWithinGap(t, 4397, lockedReward.Int64(), 10)

		assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID, userBLSTTokenID, userCLSTTokenID)

		return err
	})

	// check long time run, pool coin is enough
	for i := 0; i < 10000; i++ {
		testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
			epochManagerContract.SetContext(stakingManagerContract.Ctx)
			liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
			pool := stakingManagerContract.LoadPool(1)

			oldEpoch, err := epochManagerContract.CurrentEpoch()
			assert.Nil(t, err)

			// turn into new epoch
			newEpoch, err := epochManagerContract.TurnIntoNewEpoch()
			assert.Nil(t, err)
			err = pool.TurnIntoNewEpoch(oldEpoch.ToTypesEpoch(), newEpoch.ToTypesEpoch(), big.NewInt(20000))
			assert.Nil(t, err)

			info, err := pool.MustGetInfo()
			assert.Nil(t, err)
			assertPoolCoinIsEnough(t, stakingManagerContract, liquidStakingTokenContract, 1, info.OperatorLiquidStakingTokenID, userALSTTokenID, userBLSTTokenID, userCLSTTokenID)
			return err
		})
	}
}

// |value2 - value1| / value1 <= gapRatio/10000
// |value2 - value1| * 10000 <= gapRatio * value1
func assertWithinGap(t testing.TB, value1, value2, gapRatio int64) {
	var dif int64
	if value2 > value1 {
		dif = value2 - value1
	} else if value2 < value1 {
		dif = value1 - value2
	}
	if dif == 0 {
		return
	}

	assert.True(t, dif*10000 <= gapRatio*value1, "expect %d, got %d, allowed gapRatio %d/10000", value1, value2, gapRatio)
}

func assertPoolCoinIsEnough(t testing.TB, stakingManagerContract *StakingManager, liquidStakingTokenContract *LiquidStakingToken, poolID uint64, tokenIDs ...*big.Int) {
	pool, err := stakingManagerContract.GetPoolInfo(poolID)
	assert.Nil(t, err)

	totalCoinFromTokens := new(big.Int)
	for _, tokenID := range tokenIDs {
		coin, err := liquidStakingTokenContract.GetLockedCoin(tokenID)
		if err != nil {
			assert.Nil(t, err)
		}
		totalCoinFromTokens = new(big.Int).Add(totalCoinFromTokens, coin)
	}
	assert.True(t, pool.ActiveStake.Uint64() >= totalCoinFromTokens.Uint64(), "pool activeStake %d, totalCoinFromTokens %d", pool.ActiveStake.Uint64(), totalCoinFromTokens.Uint64())
	assertWithinGap(t, pool.ActiveStake.Int64(), totalCoinFromTokens.Int64(), 10)
}
