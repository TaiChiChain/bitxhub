package governance

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestGasManager_RunForPropose(t *testing.T) {
	logger := logrus.New()
	gov := NewGov(&common.SystemContractConfig{
		Logger: logger,
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	err := InitCouncilMembers(stateLedger, []*repo.CouncilMember{
		{
			Address: admin1,
			Weight:  1,
			Name:    "111",
		},
		{
			Address: admin2,
			Weight:  1,
			Name:    "222",
		},
		{
			Address: admin3,
			Weight:  1,
			Name:    "333",
		},
		{
			Address: admin4,
			Weight:  1,
			Name:    "444",
		},
	})
	assert.Nil(t, err)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err = framework.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	err = InitGasParam(stateLedger, &types.EpochInfo{
		FinanceParams: rbft.FinanceParams{
			StartGasPriceAvailable: true,
			StartGasPrice:          5000000000000,
			MaxGasPrice:            10000000000000,
			MinGasPrice:            1000000000000,
			GasChangeRateValue:     1250,
		},
	})
	assert.Nil(t, err)

	testcases := []struct {
		Caller string
		Data   *GasExtraArgs
		Err    error
		HasErr bool
	}{
		{
			Caller: admin1,
			Data:   nil,
			Err:    ErrGasExtraArgs,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       50000000000,
				GasChangeRateValue: 1250,
			},
			Err: ErrGasUpperOrLlower,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       50000000000,
				GasChangeRateValue: 0,
			},
			Err: ErrGasArgsType,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       100000000000000,
				GasChangeRateValue: 1250,
			},
			Err: ErrGasUpperOrLlower,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       5000000000000,
				GasChangeRateValue: 1250,
			},
			Err: ErrRepeatedGasInfo,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       2000000000000,
				GasChangeRateValue: 1250,
			},
			HasErr: true,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       2000000000000,
				GasChangeRateValue: 1250,
			},
			HasErr: false,
		},
		{
			Caller: admin1,
			Data: &GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       6000000000000,
				GasChangeRateValue: 1250,
			},
			Err: ErrExistNotFinishedGasProposal,
		},
	}

	for _, test := range testcases {
		addr := types.NewAddressByStr(test.Caller).ETHAddress()
		logs := make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			From:        &addr,
			BlockNumber: 1,
			CurrentLogs: &logs,
			StateLedger: stateLedger,
		})

		data, err := json.Marshal(test.Data)
		assert.Nil(t, err)

		if test.Data == nil {
			data = []byte("")
		}

		err = gov.Propose(uint8(GasUpdate), "test", "test desc", 100, data)
		if test.Err != nil {
			assert.Equal(t, test.Err, err)
		} else {
			if test.HasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestGasManager_VoteExecute(t *testing.T) {
	logger := logrus.New()
	gov := NewGov(&common.SystemContractConfig{
		Logger: logger,
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	err := InitCouncilMembers(stateLedger, []*repo.CouncilMember{
		{
			Address: admin1,
			Weight:  1,
			Name:    "111",
		},
		{
			Address: admin2,
			Weight:  1,
			Name:    "222",
		},
		{
			Address: admin3,
			Weight:  1,
			Name:    "333",
		},
		{
			Address: admin4,
			Weight:  1,
			Name:    "444",
		},
	})
	assert.Nil(t, err)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err = framework.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	err = InitGasParam(stateLedger, &types.EpochInfo{
		FinanceParams: rbft.FinanceParams{
			StartGasPriceAvailable: true,
			StartGasPrice:          5000000000000,
			MaxGasPrice:            10000000000000,
			MinGasPrice:            1000000000000,
			GasChangeRateValue:     1250,
		},
	})
	assert.Nil(t, err)

	data, err := json.Marshal(GasExtraArgs{
		MaxGasPrice:        10000000000000,
		MinGasPrice:        1000000000000,
		InitGasPrice:       2000000000000,
		GasChangeRateValue: 1250,
	})
	assert.Nil(t, err)

	logs := make([]common.Log, 0)
	addr := types.NewAddressByStr(admin1).ETHAddress()
	gov.SetContext(&common.VMContext{
		From:        &addr,
		BlockNumber: 1,
		CurrentLogs: &logs,
		StateLedger: stateLedger,
	})

	err = gov.Propose(uint8(GasUpdate), "test", "test desc", 100, data)
	assert.Nil(t, err)

	testcases := []struct {
		Caller     string
		ProposalID uint64
		Res        VoteResult
		Err        error
	}{
		{
			Caller:     admin1,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        Pass,
			Err:        ErrUseHasVoted,
		},
		{
			Caller:     admin2,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        Pass,
			Err:        nil,
		},
		{
			Caller:     admin3,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        Pass,
			Err:        nil,
		},
	}

	for _, test := range testcases {
		addr := types.NewAddressByStr(test.Caller).ETHAddress()
		logs := make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			From:        &addr,
			BlockNumber: 1,
			CurrentLogs: &logs,
			StateLedger: stateLedger,
		})

		err = gov.Vote(test.ProposalID, uint8(test.Res))
		assert.Equal(t, test.Err, err)
	}
}
