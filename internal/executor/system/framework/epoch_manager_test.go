package framework

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

func TestEpochManager(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(epochManagerContract)

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		currentEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)
		assert.EqualValues(t, 1, currentEpoch.Epoch)
		assert.EqualValues(t, 0, currentEpoch.StartBlock)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		nextEpoch, err := epochManagerContract.NextEpoch()
		assert.Nil(t, err)
		assert.EqualValues(t, 2, nextEpoch.Epoch)
		assert.EqualValues(t, 100, nextEpoch.StartBlock)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		epoch1, err := epochManagerContract.HistoryEpoch(1)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, epoch1.Epoch)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		_, err := epochManagerContract.HistoryEpoch(2)
		assert.Error(t, err)
	})

	testNVM.RunSingleTX(epochManagerContract, ethcommon.Address{}, func() error {
		newCurrentEpoch, err := epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		assert.EqualValues(t, 2, newCurrentEpoch.Epoch)
		assert.EqualValues(t, 100, newCurrentEpoch.StartBlock)
		return err
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		currentEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)
		assert.EqualValues(t, 2, currentEpoch.Epoch)
		assert.EqualValues(t, 100, currentEpoch.StartBlock)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		nextEpoch, err := epochManagerContract.NextEpoch()
		assert.Nil(t, err)
		assert.EqualValues(t, 3, nextEpoch.Epoch)
		assert.EqualValues(t, 200, nextEpoch.StartBlock)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		epoch2, err := epochManagerContract.HistoryEpoch(2)
		assert.Nil(t, err)
		assert.EqualValues(t, 2, epoch2.Epoch)
	})

	testNVM.Call(epochManagerContract, ethcommon.Address{}, func() {
		_, err := epochManagerContract.HistoryEpoch(3)
		assert.Error(t, err)
	})
}
