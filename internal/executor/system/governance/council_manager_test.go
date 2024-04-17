package governance

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

func TestRunForPropose(t *testing.T) {
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

	testcases := []struct {
		Caller   string
		ExtraArg *Candidates
		Err      error
		HasErr   bool
	}{
		{
			Caller:   admin1,
			ExtraArg: nil,
			HasErr:   true,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrMinCouncilMembersCount,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
					{
						Address: admin1,
						Weight:  1,
						Name:    "111",
					},
					{
						Address: admin1,
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
				},
			},
			Err: ErrRepeatedAddress,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
						Name:    "333",
					},
				},
			},
			Err: ErrRepeatedName,
		},
		{
			Caller: "0xfff0000000000000000000000000000000000000",
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrNotFoundCouncilMember,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			HasErr: false,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrExistNotFinishedProposal,
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

		data, err := json.Marshal(test.ExtraArg)
		assert.Nil(t, err)

		err = gov.Propose(uint8(CouncilElect), "test", "test desc", 100, data)
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

func TestVoteExecute(t *testing.T) {
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

	addr := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		From:        &addr,
		BlockNumber: 1,
		CurrentLogs: &logs,
		StateLedger: stateLedger,
	})

	data, err := json.Marshal(Candidates{
		Candidates: []CouncilMember{
			{
				Address: admin1,
				Weight:  2,
				Name:    "111",
			},
			{
				Address: admin2,
				Weight:  2,
				Name:    "222",
			},
			{
				Address: admin3,
				Weight:  2,
				Name:    "333",
			},
			{
				Address: admin4,
				Weight:  2,
				Name:    "444",
			},
		},
	})
	assert.Nil(t, err)

	err = gov.Propose(uint8(CouncilElect), "test", "test desc", 100, data)
	assert.Nil(t, err)

	testcases := []struct {
		Caller     string
		ProposalID uint64
		Res        VoteResult
		Err        error
	}{
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
