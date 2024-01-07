package governance

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	WhiteListProvider1 = "0x1000000000000000000000000000000000000001"
	WhiteListProvider2 = "0x1000000000000000000000000000000000000002"
)

type TestWhiteListProviderProposal struct {
	ID          uint64
	Type        ProposalType
	Proposer    string
	TotalVotes  uint64
	PassVotes   []string
	RejectVotes []string
	Status      ProposalStatus
	Providers   []access.WhiteListProvider
}

func PrepareWhiteListProviderManager(t *testing.T) (*Governance, *mock_ledger.MockStateLedger) {
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
	}, "10")
	assert.Nil(t, err)

	return gov, stateLedger
}

func TestWhiteListProviderManager_RunForPropose(t *testing.T) {
	gov, stateLedger := PrepareWhiteListProviderManager(t)

	testcases := []struct {
		Caller string
		Data   []byte
		Err    error
		HasErr bool
	}{
		{
			Caller: WhiteListProvider1,
			Data: generateProviderProposeData(t, access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{
					{
						WhiteListProviderAddr: WhiteListProvider1,
					},
					{
						WhiteListProviderAddr: WhiteListProvider2,
					},
				},
			}),
			Err: ErrNotFoundCouncilMember,
		},
		{
			Caller: admin1,
			Data:   []byte{0, 1, 2, 3},
			Err:    errors.New("unmarshal provider extra arguments error"),
		},
		{
			Caller: admin1,
			Data: generateProviderProposeData(t, access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{
					{
						WhiteListProviderAddr: WhiteListProvider1,
					},
					{
						WhiteListProviderAddr: WhiteListProvider2,
					},
				},
			}),
			HasErr: false,
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

		err := gov.Propose(uint8(WhiteListProviderAdd), "test", "test desc", 100, test.Data)
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

func TestWhiteListProviderManager_RunForVoteAdd(t *testing.T) {
	gov, stateLedger := PrepareWhiteListProviderManager(t)

	addr := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		CurrentUser:   &addr,
		CurrentHeight: 1,
		CurrentLogs:   &logs,
		StateLedger:   stateLedger,
	})

	args, err := json.Marshal(access.WhiteListProviderArgs{
		Providers: []access.WhiteListProvider{
			{
				WhiteListProviderAddr: WhiteListProvider1,
			},
			{
				WhiteListProviderAddr: WhiteListProvider2,
			},
		},
	})
	assert.Nil(t, err)

	err = gov.Propose(uint8(WhiteListProviderAdd), "test title", "test desc", 100, args)
	assert.Nil(t, err)

	testcases := []struct {
		Caller     string
		ProposalID uint64
		Res        uint8
		Err        error
		HasErr     bool
	}{
		{
			Caller:     "0xfff0000000000000000000000000000000000000",
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        uint8(Pass),
			Err:        ErrNotFoundCouncilMember,
		},
		{
			Caller:     admin2,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        uint8(Pass),
			HasErr:     false,
		},
		{
			Caller:     admin3,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        uint8(Pass),
			HasErr:     false,
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

		err := gov.Vote(test.ProposalID, test.Res)
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

func TestWhiteListProviderManager_RunForVoteRemove(t *testing.T) {
	gov, stateLedger := PrepareWhiteListProviderManager(t)

	err := access.InitProvidersAndWhiteList(stateLedger, []string{WhiteListProvider1, WhiteListProvider2, admin1, admin2, admin3}, []string{WhiteListProvider1, WhiteListProvider2, admin1})
	assert.Nil(t, err)

	addr := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		CurrentUser:   &addr,
		CurrentHeight: 1,
		CurrentLogs:   &logs,
		StateLedger:   stateLedger,
	})

	args, err := json.Marshal(access.WhiteListProviderArgs{
		Providers: []access.WhiteListProvider{
			{
				WhiteListProviderAddr: WhiteListProvider1,
			},
			{
				WhiteListProviderAddr: WhiteListProvider2,
			},
		},
	})
	assert.Nil(t, err)

	err = gov.Propose(uint8(WhiteListProviderRemove), "test title", "test desc", 100, args)
	assert.Nil(t, err)

	testcases := []struct {
		Caller     string
		ProposalID uint64
		Res        uint8
		Err        error
		HasErr     bool
	}{
		{
			Caller:     admin2,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        uint8(Pass),
			HasErr:     false,
		},
		{
			Caller:     admin3,
			ProposalID: gov.proposalID.GetID() - 1,
			Res:        uint8(Pass),
			HasErr:     false,
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

		err := gov.Vote(test.ProposalID, test.Res)
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

func generateProviderProposeData(t *testing.T, extraArgs access.WhiteListProviderArgs) []byte {
	data, err := json.Marshal(extraArgs)
	assert.Nil(t, err)

	return data
}

func TestWhiteListProviderManager_propose(t *testing.T) {
	gov, stateLedger := PrepareWhiteListProviderManager(t)

	err := access.InitProvidersAndWhiteList(stateLedger, []string{admin1, admin2, admin3}, []string{admin1, admin2, admin3})
	assert.Nil(t, err)

	testcases := []struct {
		from     ethcommon.Address
		ptype    ProposalType
		args     *access.WhiteListProviderArgs
		expected error
	}{
		{
			from:  types.NewAddressByStr(admin1).ETHAddress(),
			ptype: WhiteListProviderAdd,
			args: &access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{
					{
						WhiteListProviderAddr: admin1,
					},
				},
			},
			expected: fmt.Errorf("provider already exists, %s", admin1),
		},
		{
			from:  types.NewAddressByStr(admin2).ETHAddress(),
			ptype: WhiteListProviderRemove,
			args: &access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{
					{
						WhiteListProviderAddr: "0x0000000000000000000000000000000000000000",
					},
				},
			},
			expected: fmt.Errorf("provider does not exist, %s", "0x0000000000000000000000000000000000000000"),
		},
		{
			from:  types.NewAddressByStr(admin3).ETHAddress(),
			ptype: WhiteListProviderAdd,
			args: &access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{},
			},
			expected: errors.New("empty providers"),
		},
		{
			from:  types.NewAddressByStr(admin3).ETHAddress(),
			ptype: WhiteListProviderAdd,
			args: &access.WhiteListProviderArgs{
				Providers: []access.WhiteListProvider{
					{
						WhiteListProviderAddr: "0x0000000000000000000000000000000000000000",
					},
					{
						WhiteListProviderAddr: "0x0000000000000000000000000000000000000000",
					},
				},
			},
			expected: errors.New("provider address repeated"),
		},
	}

	for _, testcase := range testcases {
		logs := make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			CurrentUser:   &testcase.from,
			CurrentHeight: 1,
			CurrentLogs:   &logs,
			StateLedger:   stateLedger,
		})

		data, err := json.Marshal(testcase.args)
		assert.Nil(t, err)

		err = gov.Propose(uint8(testcase.ptype), "test", "test desc", 100, data)
		assert.Equal(t, testcase.expected, err)
	}

	// test unfinished proposal
	addr := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		CurrentUser:   &addr,
		CurrentHeight: 1,
		CurrentLogs:   &logs,
		StateLedger:   stateLedger,
	})

	data, err := json.Marshal(access.WhiteListProviderArgs{
		Providers: []access.WhiteListProvider{
			{
				WhiteListProviderAddr: WhiteListProvider1,
			},
		},
	})
	assert.Nil(t, err)
	err = gov.Propose(uint8(WhiteListProviderAdd), "test", "test desc", 100, data)
	assert.Nil(t, err)

	err = gov.Propose(uint8(WhiteListProviderAdd), "test", "test desc", 100, data)
	assert.Equal(t, ErrExistVotingProposal, err)
}
