package framework

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/liquid_staking_token"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
)

func TestLiquidStakingToken_stakeInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.Nodes[0].IsDataSyncer = false
	testNVM.Rep.GenesisConfig.Nodes[0].StakeNumber = types.CoinNumberByMol(10000)
	testNVM.Rep.GenesisConfig.Nodes[0].CommissionRate = 10000
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
		err = pool.TurnIntoNewEpoch(oldEpoch, newEpoch, big.NewInt(10000))
		assert.Nil(t, err)

		// check
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
		lockedCoin, err := liquidStakingTokenContract.GetLockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 20000, lockedCoin.Int64())
		unlockingCoin, err := liquidStakingTokenContract.GetUnlockingCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, unlockingCoin.Int64())
		unlockedCoin, err := liquidStakingTokenContract.GetUnlockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, unlockedCoin.Int64())

		return err
	})

	// epoch 2, reward = 10000, operator get all rewards
	// operator unlock 10000
	// [operator] before: 20000, after: 20000
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

		err = stakingManagerContract.Unlock(info.OperatorLiquidStakingTokenID, big.NewInt(10000))
		assert.Nil(t, err)

		// turn into new epoch
		err = pool.TurnIntoNewEpoch(oldEpoch, newEpoch, big.NewInt(10000))
		assert.Nil(t, err)

		// check
		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assert.EqualValues(t, 20000, operatorLSTInfo.Principal.Int64())
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, newEpoch.Epoch, operatorLSTInfo.ActiveEpoch)
		assert.EqualValues(t, 1, len(operatorLSTInfo.UnlockingRecords))
		assert.EqualValues(t, 10000, operatorLSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())
		lockedCoin, err := liquidStakingTokenContract.GetLockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 20000, lockedCoin.Int64())
		unlockingCoin, err := liquidStakingTokenContract.GetUnlockingCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 10000, unlockingCoin.Int64())
		unlockedCoin, err := liquidStakingTokenContract.GetUnlockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, unlockedCoin.Int64())

		return err
	})

	// unlocking -> unlocked
	testNVM.RunSingleTX(stakingManagerContract, operator, func() error {
		epochManagerContract.SetContext(stakingManagerContract.Ctx)
		liquidStakingTokenContract.SetContext(stakingManagerContract.Ctx)
		pool := stakingManagerContract.LoadPool(1)
		info, err := pool.MustGetInfo()
		assert.Nil(t, err)

		operatorLSTInfo, err := liquidStakingTokenContract.GetInfo(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, operatorLSTInfo.PoolID)
		assert.EqualValues(t, 20000, operatorLSTInfo.Principal.Int64())
		assert.Zero(t, operatorLSTInfo.Unlocked.Int64())
		assert.EqualValues(t, 1, len(operatorLSTInfo.UnlockingRecords))
		assert.EqualValues(t, 10000, operatorLSTInfo.UnlockingRecords[0].Amount.Int64())
		totalCoin, err := liquidStakingTokenContract.GetTotalCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 30000, totalCoin.Int64())
		lockedReward, err := liquidStakingTokenContract.GetLockedReward(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.Zero(t, lockedReward.Int64())
		lockedCoin, err := liquidStakingTokenContract.GetLockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 20000, lockedCoin.Int64())
		unlockingCoin, err := liquidStakingTokenContract.GetUnlockingCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 0, unlockingCoin.Int64())
		unlockedCoin, err := liquidStakingTokenContract.GetUnlockedCoin(info.OperatorLiquidStakingTokenID)
		assert.Nil(t, err)
		assert.EqualValues(t, 10000, unlockedCoin.Int64())

		return err
	}, func(ctx *common.VMContext) {
		ctx.CurrentEVM.Context.Time = uint64(time.Now().Unix()) + testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.UnlockPeriod
	})
}

func TestLiquidStakingToken_erc721(t *testing.T) {
	user1 := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D011")
	user2 := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D012")
	user3 := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(epochManagerContract, liquidStakingTokenContract)

	var lstID1, lstID2, lstID3 *big.Int
	var err error
	testNVM.RunSingleTX(liquidStakingTokenContract, user1, func() error {
		lstID1, err = liquidStakingTokenContract.Mint(user1, &liquid_staking_token.LiquidStakingTokenInfo{})
		assert.Nil(t, err)
		assert.EqualValues(t, 1, lstID1.Int64())

		lstID2, err = liquidStakingTokenContract.Mint(user1, &liquid_staking_token.LiquidStakingTokenInfo{})
		assert.Nil(t, err)
		assert.EqualValues(t, 2, lstID2.Int64())

		lstID3, err = liquidStakingTokenContract.Mint(user1, &liquid_staking_token.LiquidStakingTokenInfo{})
		assert.Nil(t, err)
		assert.EqualValues(t, 3, lstID3.Int64())

		return err
	})

	t.Run("OwnerOf", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user1, func() {
			_, err := liquidStakingTokenContract.OwnerOf(big.NewInt(0))
			assert.ErrorContains(t, err, "token not exist")

			_, err = liquidStakingTokenContract.OwnerOf(big.NewInt(1000))
			assert.ErrorContains(t, err, "token not exist")

			owner, err := liquidStakingTokenContract.OwnerOf(lstID1)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)

			owner, err = liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)

			owner, err = liquidStakingTokenContract.OwnerOf(lstID3)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)
		})
	})

	t.Run("BalanceOf", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user1, func() {
			balance, err := liquidStakingTokenContract.BalanceOf(user1)
			assert.Nil(t, err)
			assert.EqualValues(t, 3, balance.Int64())

			balance, err = liquidStakingTokenContract.BalanceOf(user2)
			assert.Nil(t, err)
			assert.EqualValues(t, 0, balance.Int64())

			balance, err = liquidStakingTokenContract.BalanceOf(user3)
			assert.Nil(t, err)
			assert.EqualValues(t, 0, balance.Int64())

			_, err = liquidStakingTokenContract.BalanceOf(ethcommon.Address{})
			assert.ErrorContains(t, err, "invalid owner")
		})
	})

	t.Run("Burn", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user2, func() {
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)

			err = liquidStakingTokenContract.Burn(lstID2)
			assert.ErrorContains(t, err, "insufficient approval")

			owner, err = liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)
		})

		testNVM.Call(liquidStakingTokenContract, ethcommon.Address{}, func() {
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)

			err = liquidStakingTokenContract.Burn(lstID2)
			assert.Nil(t, err)

			_, err = liquidStakingTokenContract.OwnerOf(lstID2)
			assert.ErrorContains(t, err, "token not exist")
		})

		testNVM.Call(liquidStakingTokenContract, user1, func() {
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)

			err = liquidStakingTokenContract.Burn(lstID2)
			assert.Nil(t, err)

			_, err = liquidStakingTokenContract.OwnerOf(lstID2)
			assert.ErrorContains(t, err, "token not exist")
		})
	})

	t.Run("SafeTransferFrom", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user1, func() {
			user1BeforeBalance, err := liquidStakingTokenContract.BalanceOf(user1)
			assert.Nil(t, err)
			user2BeforeBalance, err := liquidStakingTokenContract.BalanceOf(user2)
			assert.Nil(t, err)

			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID2)
			assert.Nil(t, err)

			user1AfterBalance, err := liquidStakingTokenContract.BalanceOf(user1)
			assert.Nil(t, err)
			assert.EqualValues(t, 1, user1BeforeBalance.Int64()-user1AfterBalance.Int64())
			user2AfterBalance, err := liquidStakingTokenContract.BalanceOf(user2)
			assert.Nil(t, err)
			assert.EqualValues(t, 1, user2AfterBalance.Int64()-user2BeforeBalance.Int64())

			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID2)
			assert.ErrorContains(t, err, "insufficient approval")

			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, owner)
		})

		testNVM.Call(liquidStakingTokenContract, user1, func() {
			err := liquidStakingTokenContract.SafeTransferFrom(user2, user3, lstID2)
			assert.ErrorContains(t, err, "incorrect owner")

			err = liquidStakingTokenContract.SafeTransferFrom(user1, ethcommon.Address{}, lstID2)
			assert.ErrorContains(t, err, "invalid receiver")
		})

		testNVM.Call(liquidStakingTokenContract, user1, func() {
			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID2)
			assert.Nil(t, err)
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, owner)

			liquidStakingTokenContract.Ctx.From = user2
			err = liquidStakingTokenContract.SafeTransferFrom(user2, user1, lstID2)
			assert.Nil(t, err)
			owner, err = liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user1, owner)
		})
	})

	t.Run("Approve", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user1, func() {
			approved, err := liquidStakingTokenContract.GetApproved(big.NewInt(100))
			assert.Nil(t, err)
			assert.EqualValues(t, ethcommon.Address{}, approved)

			approved, err = liquidStakingTokenContract.GetApproved(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, ethcommon.Address{}, approved)

			err = liquidStakingTokenContract.Approve(user2, lstID2)
			assert.Nil(t, err)

			approved, err = liquidStakingTokenContract.GetApproved(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, approved)

			liquidStakingTokenContract.Ctx.From = user2
			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID2)
			assert.Nil(t, err)
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, owner)

			approved, err = liquidStakingTokenContract.GetApproved(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, ethcommon.Address{}, approved)

			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID3)
			assert.ErrorContains(t, err, "insufficient approval")
		})
	})

	t.Run("AetApprovalForAll", func(t *testing.T) {
		testNVM.Call(liquidStakingTokenContract, user1, func() {
			isApprovedForAll, err := liquidStakingTokenContract.IsApprovedForAll(user1, user2)
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(user1, ethcommon.Address{})
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(ethcommon.Address{}, user1)
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(user2, ethcommon.Address{})
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			err = liquidStakingTokenContract.SetApprovalForAll(user2, true)
			assert.Nil(t, err)
			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(user1, user2)
			assert.Nil(t, err)
			assert.True(t, isApprovedForAll)
			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(user2, user1)
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			liquidStakingTokenContract.Ctx.From = user2
			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID2)
			assert.Nil(t, err)
			owner, err := liquidStakingTokenContract.OwnerOf(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, owner)

			approved, err := liquidStakingTokenContract.GetApproved(lstID2)
			assert.Nil(t, err)
			assert.EqualValues(t, ethcommon.Address{}, approved)

			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID3)
			assert.Nil(t, err)
			owner, err = liquidStakingTokenContract.OwnerOf(lstID3)
			assert.Nil(t, err)
			assert.EqualValues(t, user2, owner)

			approved, err = liquidStakingTokenContract.GetApproved(lstID3)
			assert.Nil(t, err)
			assert.EqualValues(t, ethcommon.Address{}, approved)

			liquidStakingTokenContract.Ctx.From = user1
			err = liquidStakingTokenContract.SetApprovalForAll(user2, false)
			assert.Nil(t, err)
			isApprovedForAll, err = liquidStakingTokenContract.IsApprovedForAll(user1, user2)
			assert.Nil(t, err)
			assert.False(t, isApprovedForAll)

			liquidStakingTokenContract.Ctx.From = user2
			err = liquidStakingTokenContract.SafeTransferFrom(user1, user2, lstID1)
			assert.ErrorContains(t, err, "insufficient approval")
		})
	})
}
