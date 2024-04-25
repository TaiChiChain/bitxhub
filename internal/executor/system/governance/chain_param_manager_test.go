package governance

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestChainParam_RunForPropose(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testcases := []struct {
		Caller ethcommon.Address
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

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				data, err := json.Marshal(test.Data)
				assert.Nil(t, err)

				if test.Data == nil {
					data = []byte("")
				}

				err = gov.Propose(uint8(ChainParamUpgrade), "test", "test desc", 100, data)
				if test.Err != nil {
					assert.Contains(t, err.Error(), test.Err.Error())
				} else {
					if test.HasErr {
						assert.NotNil(t, err)
					} else {
						assert.Nil(t, err)
					}
				}
				return err
			})
		})
	}
}

func TestChainParam_VoteExecute(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.RunSingleTX(gov, admin1, func() error {
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

		err = gov.Propose(uint8(ChainParamUpgrade), "test", "test desc", 100, data)
		assert.Nil(t, err)

		return err
	})

	var proposalID uint64
	testNVM.Call(gov, admin1, func() {
		var err error
		proposalID, err = gov.GetLatestProposalID()
		assert.Nil(t, err)
	})

	testcases := []struct {
		Caller     ethcommon.Address
		ProposalID uint64
		Res        VoteResult
		Err        error
	}{
		{
			Caller:     admin1,
			ProposalID: proposalID,
			Res:        Pass,
			Err:        ErrUseHasVoted,
		},
		{
			Caller:     admin2,
			ProposalID: proposalID,
			Res:        Pass,
			Err:        nil,
		},
		{
			Caller:     admin3,
			ProposalID: proposalID,
			Res:        Pass,
			Err:        nil,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(test.ProposalID, uint8(test.Res))
				assert.Equal(t, test.Err, err)
				return err
			})
		})
	}
}
