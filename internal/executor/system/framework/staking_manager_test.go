package framework

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func Test_stakingManager_InternalCalculateStakeReward(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcBuildContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcBuildContract, epochManagerContract, nodeManagerContract, stakingManagerBuildContract)
	stakeValue, _ := new(big.Int).SetString("3000000000000000000", 10)
	errRatio := big.NewFloat(1e-3)

	// expect ErrStakeValue
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.totalStake.Put(types.CoinNumberByAxc(100000000000).ToBigInt())
		assert.Nil(t, err)
		err = stakingManagerBuildContract.UpdateStakeRewardPerBlock()
		assert.EqualError(t, ErrStakeValue, err.Error())
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 1, should have APY 3%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.totalStake.Put(repo.DefaultAXCBalance.ToBigInt())
		assert.Nil(t, err)
		err = stakingManagerBuildContract.UpdateStakeRewardPerBlock()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.rewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.03)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 0.1, should have APY 7.89%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.totalStake.Put(new(big.Int).Div(repo.DefaultAXCBalance.ToBigInt(), big.NewInt(10)))
		assert.Nil(t, err)
		err = stakingManagerBuildContract.UpdateStakeRewardPerBlock()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.rewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.0789)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 0, should have APY 12%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.UpdateStakeRewardPerBlock()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.rewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.12)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return err
	}, common.TestNVMRunOptionCallFromSystem())
}

func TestStakingManager_GenesisInit(t *testing.T) {
	testNVM := common.NewTestNVM(t)

	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, nodeManagerContract, liquidStakingTokenContract)

	stakingManagerContract.SetContext(common.NewVMContextByExecutor(testNVM.StateLedger).DisableRecordLogToLedger())
	testNVM.Rep.GenesisConfig.Nodes[0].CommissionRate = 1000000000
	err := stakingManagerContract.GenesisInit(testNVM.Rep.GenesisConfig)
	assert.ErrorContains(t, err, "invalid commission rate")

	genesis := repo.DefaultGenesisConfig()
	err = stakingManagerContract.GenesisInit(genesis)
	assert.Nil(t, err)
}

func TestStakingManager_LifeCycle(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.MinDelegateStake = types.CoinNumberByMol(1)

	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, liquidStakingTokenContract, nodeManagerContract, stakingManagerContract)
	operatorAddr := types.NewAddressByStr("0x4285B2E3E82a7eEd8d88A5081289a1Ee5e8F8888")

	stakingPools := []uint64{1, 2, 3, 4}

	// create staking pool
	// error repeat register
	testNVM.Call(stakingManagerContract, types.NewAddressByStr("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013").ETHAddress(), func() {
		err := stakingManagerContract.CreatePool(1, 10)
		assert.ErrorContains(t, err, "already exists")
	})
	// success
	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		// register the node first
		id, err := nodeManagerContract.Register(node_manager.NodeInfo{
			Operator:        operatorAddr.ETHAddress(),
			P2PID:           p2pKeystore.P2PID(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			MetaData: node_manager.NodeMetaData{
				Name:       "mockName",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, id, uint64(5))
		err = nodeManagerContract.JoinCandidateSet(id, 0)
		assert.Nil(t, err)
		exist, pools, err := stakingManagerContract.availablePools.Get()
		assert.Nil(t, err)
		assert.True(t, exist)
		assert.EqualValues(t, append(stakingPools, 5), pools)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	// add some stake
	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		before, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)
		stakingManagerContract.Ctx.Value = big.NewInt(200)
		stakingManagerContract.StateAccount.AddBalance(big.NewInt(200))

		err = stakingManagerContract.AddStake(5, operatorAddr.ETHAddress(), big.NewInt(200))
		assert.Nil(t, err)
		after, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)
		assert.Equal(t, uint64(200), new(big.Int).Sub(after, before).Uint64())
		return err
	})

	liquidityStakingTokenID := stakingManagerContract.Ctx.TestLogs[0].(*staking_manager.EventAddStake).LiquidStakingTokenID
	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		err = stakingManagerContract.Unlock(liquidityStakingTokenID, big.NewInt(100))
		assert.ErrorContains(t, err, "not active")
		return err
	})

	// in same epoch, can directly withdraw
	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		beforeBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(operatorAddr)
		before, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)

		err = stakingManagerContract.Withdraw(liquidityStakingTokenID, operatorAddr.ETHAddress(), big.NewInt(100))
		assert.Nil(t, err)

		afterBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(operatorAddr)
		after, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)
		assert.Equal(t, uint64(100), new(big.Int).Sub(after, before).Uint64())
		assert.Equal(t, uint64(100), new(big.Int).Sub(afterBalance, beforeBalance).Uint64())
		return err
	})

	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		_, err := EpochManagerBuildConfig.Build(stakingManagerContract.CrossCallSystemContractContext()).TurnIntoNewEpoch()
		assert.Nil(t, err)

		before, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)

		err = stakingManagerContract.Unlock(liquidityStakingTokenID, big.NewInt(100))
		assert.Nil(t, err)

		after, err := stakingManagerContract.totalStake.MustGet()
		assert.Nil(t, err)

		assert.Equal(t, uint64(100), new(big.Int).Sub(after, before).Uint64())
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		beforeBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(operatorAddr)

		err = stakingManagerContract.Withdraw(liquidityStakingTokenID, operatorAddr.ETHAddress(), big.NewInt(100))
		assert.Nil(t, err)

		afterBalance := stakingManagerContract.Ctx.StateLedger.GetBalance(operatorAddr)

		assert.Equal(t, uint64(100), new(big.Int).Sub(afterBalance, beforeBalance).Uint64())
		return err
	}, func(ctx *common.VMContext) {
		ctx.CurrentEVM.Context.Time = uint64(time.Now().Unix()) + testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.UnlockPeriod
	})

	testNVM.RunSingleTX(stakingManagerContract, operatorAddr.ETHAddress(), func() error {
		beforeGasReward, err := stakingManagerContract.gasRewardTable.MustGet(5)
		assert.Nil(t, err)
		beforeProposeBlockCount, err := stakingManagerContract.proposeBlockCountTable.MustGet(5)
		assert.Nil(t, err)

		_, err = stakingManagerContract.RecordReward(5, big.NewInt(100))
		assert.Nil(t, err)

		afterGasReward, err := stakingManagerContract.gasRewardTable.MustGet(5)
		assert.Nil(t, err)
		afterProposeBlockCount, err := stakingManagerContract.proposeBlockCountTable.MustGet(5)
		assert.Nil(t, err)

		assert.Equal(t, uint64(100), new(big.Int).Sub(afterGasReward, beforeGasReward).Uint64())
		assert.Equal(t, uint64(1), afterProposeBlockCount-beforeProposeBlockCount)

		return err
	}, common.TestNVMRunOptionCallFromSystem())
}

func TestStakingManager_TurnIntoNewEpoch(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.MinDelegateStake = types.CoinNumberByMol(1)

	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, liquidStakingTokenContract, nodeManagerContract, stakingManagerContract)

	epoch := testNVM.Rep.GenesisConfig.EpochInfo
	err := stakingManagerContract.TurnIntoNewEpoch(epoch, epoch)
	assert.Nil(t, err)
}

func TestStakingManager_UpdatePoolCommissionRate(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.MinDelegateStake = types.CoinNumberByMol(1)

	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcContract, epochManagerContract, liquidStakingTokenContract, nodeManagerContract, stakingManagerContract)

	err := stakingManagerContract.UpdatePoolCommissionRate(10, 1)
	assert.EqualError(t, err, "node not found")

	err = stakingManagerContract.UpdatePoolCommissionRate(1, 1)
	assert.EqualError(t, err, "no permission")

	info, err := nodeManagerContract.GetInfo(1)
	assert.Nil(t, err)
	newStakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, info.Operator))
	err = newStakingManagerContract.UpdatePoolCommissionRate(1, 1)
	assert.Nil(t, err)
}
