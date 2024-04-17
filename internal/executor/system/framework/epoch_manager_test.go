package framework

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func prepareLedger(t *testing.T) ledger.StateLedger {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.EpochManagerContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	return stateLedger
}

func TestEpochManager(t *testing.T) {
	stateLedger := prepareLedger(t)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	currentEpoch, err := GetCurrentEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, currentEpoch.Epoch)
	assert.EqualValues(t, 1, currentEpoch.StartBlock)

	nextEpoch, err := GetNextEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, nextEpoch.Epoch)
	assert.EqualValues(t, 101, nextEpoch.StartBlock)

	epoch1, err := GetEpochInfo(stateLedger, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, epoch1.Epoch)

	_, err = GetEpochInfo(stateLedger, 2)
	assert.Error(t, err)

	newCurrentEpoch, err := TurnIntoNewEpoch([]byte{}, stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, newCurrentEpoch.Epoch)
	assert.EqualValues(t, 101, newCurrentEpoch.StartBlock)

	currentEpoch, err = GetCurrentEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, currentEpoch.Epoch)
	assert.EqualValues(t, 101, currentEpoch.StartBlock)

	nextEpoch, err = GetNextEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, nextEpoch.Epoch)
	assert.EqualValues(t, 201, nextEpoch.StartBlock)

	epoch2, err := GetEpochInfo(stateLedger, 2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, epoch2.Epoch)

	_, err = GetEpochInfo(stateLedger, 3)
	assert.Error(t, err)

	epochMgr := NewEpochManager()
	epochMgr.SetContext(&common.VMContext{
		StateLedger: stateLedger,
		BlockNumber: 1,
	})

	t.Run("query currentEpoch", func(t *testing.T) {
		res, err := epochMgr.CurrentEpoch()
		assert.Nil(t, err)

		assert.EqualValues(t, currentEpoch, res)
	})

	t.Run("query nextEpoch", func(t *testing.T) {
		res, err := epochMgr.NextEpoch()
		assert.Nil(t, err)

		assert.EqualValues(t, nextEpoch, res)
	})

	t.Run("query historyEpoch", func(t *testing.T) {
		res, err := epochMgr.HistoryEpoch(2)
		assert.Nil(t, err)

		assert.EqualValues(t, epoch2, res)
	})
}
