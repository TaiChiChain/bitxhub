package governance

import (
	"encoding/json"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func initGovernance(t *testing.T) (*Governance, ledger.StateLedger) {
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

	err := InitCouncilMembers(stateLedger, []*repo.Admin{
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
	err = base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	return gov, stateLedger
}

func TestChainParam_RunForPropose(t *testing.T) {
	gov, stateLedger := initGovernance(t)

	testcases := []struct {
		Caller string
		Data   *ChainParamExtraArgs
		Err    error
		HasErr bool
	}{
		{
			Caller: admin1,
			Data:   nil,
			Err:    ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod: 0,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:     1,
				MaxValidatorNum: 3,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:     1,
				MaxValidatorNum: 4,
				BlockMaxTxNum:   0,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:             1,
				MaxValidatorNum:         4,
				BlockMaxTxNum:           10,
				AbnormalNodeExcludeView: 0,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:             1,
				MaxValidatorNum:         4,
				BlockMaxTxNum:           10,
				AbnormalNodeExcludeView: 1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 100,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:             1,
				MaxValidatorNum:         4,
				BlockMaxTxNum:           10,
				AbnormalNodeExcludeView: 1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 100,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:             1,
				MaxValidatorNum:         4,
				BlockMaxTxNum:           10,
				AbnormalNodeExcludeView: 1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 80,
				ContinuousNullRequestToleranceNumber:               0,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:             1,
				MaxValidatorNum:         4,
				BlockMaxTxNum:           10,
				AbnormalNodeExcludeView: 1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 80,
				ContinuousNullRequestToleranceNumber:               1,
				RebroadcastToleranceNumber:                         0,
			},
			Err: ErrChainParamInvalid,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:              1,
				MaxValidatorNum:          4,
				BlockMaxTxNum:            10,
				EnableTimedGenEmptyBlock: false,
				AbnormalNodeExcludeView:  1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 80,
				ContinuousNullRequestToleranceNumber:               1,
				RebroadcastToleranceNumber:                         1,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: &ChainParamExtraArgs{
				EpochPeriod:              1,
				MaxValidatorNum:          4,
				BlockMaxTxNum:            10,
				EnableTimedGenEmptyBlock: true,
				AbnormalNodeExcludeView:  1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 80,
				ContinuousNullRequestToleranceNumber:               1,
				RebroadcastToleranceNumber:                         1,
			},
			Err: ErrExistNotFinishedChainParamProposal,
		},
	}

	for _, test := range testcases {
		addr := types.NewAddressByStr(test.Caller).ETHAddress()
		logs := make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			CurrentUser:   &addr,
			CurrentHeight: 1,
			CurrentLogs:   &logs,
			StateLedger:   stateLedger,
		})

		data, err := json.Marshal(test.Data)
		assert.Nil(t, err)

		if test.Data == nil {
			data = []byte("")
		}

		err = gov.Propose(uint8(ChainParamUpgrade), "test", "test desc", 100, data)
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

func TestChainParam_VoteExecute(t *testing.T) {
	gov, stateLedger := initGovernance(t)

	data, err := json.Marshal(&ChainParamExtraArgs{
		EpochPeriod:              1,
		MaxValidatorNum:          4,
		BlockMaxTxNum:            10,
		EnableTimedGenEmptyBlock: true,
		AbnormalNodeExcludeView:  1,
		AgainProposeIntervalBlockInValidatorsNumPercentage: 80,
		ContinuousNullRequestToleranceNumber:               1,
		RebroadcastToleranceNumber:                         1,
	})
	assert.Nil(t, err)

	logs := make([]common.Log, 0)
	addr := types.NewAddressByStr(admin1).ETHAddress()
	gov.SetContext(&common.VMContext{
		CurrentUser:   &addr,
		CurrentHeight: 1,
		CurrentLogs:   &logs,
		StateLedger:   stateLedger,
	})

	err = gov.Propose(uint8(ChainParamUpgrade), "test", "test desc", 100, data)
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
			CurrentUser:   &addr,
			CurrentHeight: 1,
			CurrentLogs:   &logs,
			StateLedger:   stateLedger,
		})

		err = gov.Vote(test.ProposalID, uint8(test.Res))
		assert.Equal(t, test.Err, err)
	}
}
