package framework

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

func Test_stakingManager_InternalCalculateStakeReward(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(epochManagerContract, nodeManagerContract, stakingManagerBuildContract)

	// TODO: add tests
	// t.Run("expect ErrZeroValue", func(t *testing.T) {
	//	err := mockStakingManager.InternalCalculateStakeReward()
	//	assert.EqualError(t, ErrZeroAXC, err.Error())
	// })
	//
	// t.Run("expect ErrStakeValue", func(t *testing.T) {
	//	axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	//	axcAccount.SetState([]byte(token.TotalSupplyKey), big.NewInt(100).Bytes())
	//	err := mockStakingManager.InternalCalculateStakeReward()
	//	assert.EqualError(t, ErrStakeValue, err.Error())
	// })
	//
	// t.Run("stake/totalSupply = 1, should have APY 3%", func(t *testing.T) {
	//	axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	//	axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Add(stakeValue, big.NewInt(1)).Bytes())
	//	err := mockStakingManager.InternalCalculateStakeReward()
	//	assert.Nil(t, err)
	//	isExist, value, err := slot.Get()
	//	assert.True(t, isExist)
	//	assert.Nil(t, err)
	//	rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
	//	ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
	//	expect := new(big.Float).SetFloat64(0.03)
	//	acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
	//	assert.True(t, acc.Cmp(errRatio) < 0)
	// })
	//
	// t.Run("stake/totalSupply = 0.1, should have APY 7.89%", func(t *testing.T) {
	//	axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	//	axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Mul(stakeValue, big.NewInt(10)).Bytes())
	//	err := mockStakingManager.InternalCalculateStakeReward()
	//	assert.Nil(t, err)
	//	isExist, value, err := slot.Get()
	//	assert.True(t, isExist)
	//	assert.Nil(t, err)
	//	rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
	//	ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
	//	expect := new(big.Float).SetFloat64(0.0789)
	//	acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
	//	assert.True(t, acc.Cmp(errRatio) < 0)
	// })
	//
	// t.Run("stake/totalSupply = 0, should have APY 12%", func(t *testing.T) {
	//	axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	//	axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Mul(stakeValue, big.NewInt(1000000)).Bytes())
	//	err := mockStakingManager.InternalCalculateStakeReward()
	//	assert.Nil(t, err)
	//	isExist, value, err := slot.Get()
	//	assert.True(t, isExist)
	//	assert.Nil(t, err)
	//	rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
	//	ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
	//	expect := new(big.Float).SetFloat64(0.12)
	//	acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
	//	assert.True(t, acc.Cmp(errRatio) < 0)
	// })
}
