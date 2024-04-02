package base

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/stretchr/testify/assert"
)

func Test_stakingManager_InternalCalculateStakeReward(t *testing.T) {
	mockLedger := newMockLedger(t)
	mockStakingManager := NewStakingManager(nil)
	mockStakingManager.SetContext(&common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
	})
	account := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.StakingManagerContractAddr))
	slot := common.NewVMSlot[*big.Int](account, stakingPrefix+rewardPerBlockKey)
	stakeValue, _ := new(big.Int).SetString("3000000000000000000", 10)
	errRatio := big.NewFloat(1e-3)

	t.Run("expect ErrZeroValue", func(t *testing.T) {
		err := mockStakingManager.InternalCalculateStakeReward()
		assert.EqualError(t, ErrZeroAXC, err.Error())
	})

	t.Run("expect ErrStakeValue", func(t *testing.T) {
		axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
		axcAccount.SetState([]byte(token.TotalSupplyKey), big.NewInt(100).Bytes())
		err := mockStakingManager.InternalCalculateStakeReward()
		assert.EqualError(t, ErrStakeValue, err.Error())
	})

	t.Run("stake/totalSupply = 1, should have APY 3%", func(t *testing.T) {
		axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
		axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Add(stakeValue, big.NewInt(1)).Bytes())
		err := mockStakingManager.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := slot.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.03)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
	})

	t.Run("stake/totalSupply = 0.1, should have APY 7.89%", func(t *testing.T) {
		axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
		axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Mul(stakeValue, big.NewInt(10)).Bytes())
		err := mockStakingManager.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := slot.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.0789)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
	})

	t.Run("stake/totalSupply = 0, should have APY 12%", func(t *testing.T) {
		axcAccount := mockLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
		axcAccount.SetState([]byte(token.TotalSupplyKey), new(big.Int).Mul(stakeValue, big.NewInt(1000000)).Bytes())
		err := mockStakingManager.InternalCalculateStakeReward()
		assert.Nil(t, err)
		isExist, value, err := slot.Get()
		assert.True(t, isExist)
		assert.Nil(t, err)
		rewardsPerYear := new(big.Int).Mul(value, big.NewInt(int64(blocksPerYear)))
		ratio := new(big.Float).Quo(new(big.Float).SetInt(rewardsPerYear), new(big.Float).SetInt(stakeValue))
		expect := new(big.Float).SetFloat64(0.12)
		acc := new(big.Float).Quo(new(big.Float).Sub(expect, ratio), ratio)
		assert.True(t, acc.Cmp(errRatio) < 0)
	})
}
