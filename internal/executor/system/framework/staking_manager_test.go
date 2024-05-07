package framework

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/staking_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
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

	// expect ErrPermissionDenied
	testNVM.Call(stakingManagerBuildContract, ethcommon.Address{}, func() {
		err := stakingManagerBuildContract.TotalStake.Put(types.CoinNumberByAxc(100000000000).ToBigInt())
		assert.Nil(t, err)
		err = stakingManagerBuildContract.InternalCalculateStakeReward()
		assert.EqualError(t, ErrPermissionDenied, err.Error())
	})

	// expect ErrStakeValue
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.TotalStake.Put(types.CoinNumberByAxc(100000000000).ToBigInt())
		assert.Nil(t, err)
		err = stakingManagerBuildContract.InternalCalculateStakeReward()
		assert.EqualError(t, ErrStakeValue, err.Error())
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 1, should have APY 3%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.TotalStake.Put(repo.DefaultAXCBalance.ToBigInt())
		assert.Nil(t, err)
		err = stakingManagerBuildContract.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.RewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.03)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 0.1, should have APY 7.89%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.TotalStake.Put(new(big.Int).Div(repo.DefaultAXCBalance.ToBigInt(), big.NewInt(10)))
		assert.Nil(t, err)
		err = stakingManagerBuildContract.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.RewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.0789)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	// stake/totalSupply = 0, should have APY 12%
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := stakingManagerBuildContract.RewardPerBlock.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.12)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())
}

func TestStakingManager_LifeCycle(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcBuildContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcBuildContract, epochManagerContract, nodeManagerContract, stakingManagerBuildContract)
	operatorAddr := types.NewAddressByStr("0x4285B2E3E82a7eEd8d88A5081289a1Ee5e8F8888")

	stakingPools := []uint64{1, 2, 3, 4}
	// register staking pools
	testNVM.Call(stakingManagerBuildContract, ethcommon.Address{}, func() {
		err := stakingManagerBuildContract.InternalInitAvailableStakingPools(stakingPools)
		assert.EqualError(t, err, ErrPermissionDenied.Error())
	})
	testNVM.RunSingleTX(stakingManagerBuildContract, ethcommon.Address{}, func() error {
		err := stakingManagerBuildContract.InternalInitAvailableStakingPools(stakingPools)
		assert.Nil(t, err)
		exist, pools, err := stakingManagerBuildContract.AvailableStakingPools.Get()
		assert.Nil(t, err)
		assert.True(t, exist)
		assert.EqualValues(t, stakingPools, pools)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	// create staking pool
	// error sender
	testNVM.Call(stakingManagerBuildContract, ethcommon.Address{}, func() {
		err := stakingManagerBuildContract.CreateStakingPool(1, 10)
		assert.ErrorContains(t, err, "not equal to caller")
	})
	// error repeat register
	testNVM.Call(stakingManagerBuildContract, types.NewAddressByStr("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013").ETHAddress(), func() {
		err := stakingManagerBuildContract.CreateStakingPool(1, 10)
		assert.ErrorContains(t, err, "already exists")
	})
	// success
	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		// register the node first
		id, err := nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			OperatorAddress: operatorAddr.String(),
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
		err = nodeManagerContract.JoinCandidateSet(id)
		assert.Nil(t, err)
		err = stakingManagerBuildContract.CreateStakingPool(5, 10)
		assert.Nil(t, err)
		exist, pools, err := stakingManagerBuildContract.AvailableStakingPools.Get()
		assert.Nil(t, err)
		assert.True(t, exist)
		assert.EqualValues(t, append(stakingPools, 5), pools)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	// add some stake
	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		exists, before, err := stakingManagerBuildContract.TotalStake.Get()
		assert.Nil(t, err)
		assert.True(t, exists)
		// todo: check balance
		err = stakingManagerBuildContract.AddStake(5, operatorAddr.ETHAddress(), big.NewInt(100))
		assert.Nil(t, err)
		exists, after, err := stakingManagerBuildContract.TotalStake.Get()
		assert.Nil(t, err)
		assert.True(t, exists)
		assert.Equal(t, new(big.Int).Sub(after, before).Uint64(), uint64(100))
		return nil
	})

	liquidityStakingTokenID := stakingManagerBuildContract.Ctx.TestLogs[0].(*staking_manager.EventAddStake).LiquidStakingTokenID

	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		err = stakingManagerBuildContract.Withdraw(5, operatorAddr.ETHAddress(), liquidityStakingTokenID, big.NewInt(100))
		assert.ErrorContains(t, err, "not active")
		return nil
	})

	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		_, err = epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = stakingManagerBuildContract.Withdraw(5, operatorAddr.ETHAddress(), liquidityStakingTokenID, big.NewInt(100))
		assert.Nil(t, err)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		err = stakingManagerBuildContract.Unlock(5, operatorAddr.ETHAddress(), liquidityStakingTokenID, big.NewInt(100))
		assert.Nil(t, err)
		// todo: check balance after unlock
		return nil
	})

	// permission denied
	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		_, err = stakingManagerBuildContract.InternalRecordReward(5, big.NewInt(100))
		assert.EqualError(t, err, ErrPermissionDenied.Error())
		return nil
	})

	testNVM.RunSingleTX(stakingManagerBuildContract, operatorAddr.ETHAddress(), func() error {
		_, err = stakingManagerBuildContract.InternalRecordReward(5, big.NewInt(100))
		assert.Nil(t, err)
		return nil
	}, common.TestNVMRunOptionCallFromSystem())
}
