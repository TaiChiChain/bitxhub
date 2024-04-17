package governance

import (
	"encoding/json"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestGovernance_Propose(t *testing.T) {
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
		input struct {
			user         string
			proposalType ProposalType
			title        string
			desc         string
			BlockNumber  uint64
		}
		err    error
		hasErr bool
	}{
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         admin1,
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			hasErr: false,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: ProposalType(250),
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			err: ErrProposalType,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			err: ErrTitle,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "",
				BlockNumber:  10000,
			},
			err: ErrDesc,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  0,
			},
			err: ErrBlockNumber,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  2,
			},
			err: ErrNotFoundCouncilMember,
		},
	}

	for _, test := range testcases {
		addr := types.NewAddressByStr(test.input.user)
		ethaddr := addr.ETHAddress()

		logs := make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			StateLedger: stateLedger,
			BlockNumber: 1,
			From:        &ethaddr,
			CurrentLogs: &logs,
		})

		arg := NodeUpgradeExtraArgs{
			DownloadUrls: []string{"http://127.0.0.1:10000"},
			CheckHash:    "",
		}
		data, err := json.Marshal(arg)
		assert.Nil(t, err)

		err = gov.Propose(uint8(test.input.proposalType), test.input.title, test.input.desc, test.input.BlockNumber, data)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			if test.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestGovernance_Vote(t *testing.T) {
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

	admin1Addr := types.NewAddressByStr(admin1).ETHAddress()
	admin2Addr := types.NewAddressByStr(admin2).ETHAddress()
	admin3Addr := types.NewAddressByStr(admin3).ETHAddress()
	admin4Addr := types.NewAddressByStr(admin4).ETHAddress()
	addr := types.NewAddressByStr("0x10010000000").ETHAddress()

	arg := NodeUpgradeExtraArgs{
		DownloadUrls: []string{"http://127.0.0.1:10000"},
		CheckHash:    "",
	}
	data, err := json.Marshal(arg)
	assert.Nil(t, err)

	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		From:        &admin1Addr,
		BlockNumber: 1,
		StateLedger: stateLedger,
		CurrentLogs: &logs,
	})
	err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
	assert.Nil(t, err)

	testcases := []struct {
		from   *ethcommon.Address
		id     uint64
		result uint8
		err    error
		hasErr bool
	}{
		{
			from:   &admin2Addr,
			id:     100,
			result: uint8(Reject),
			err:    ErrNotFoundProposal,
		},
		{
			from:   &admin2Addr,
			id:     1,
			result: uint8(100),
			err:    ErrVoteResult,
		},
		{
			from:   &admin1Addr,
			id:     1,
			result: uint8(Pass),
			err:    ErrUseHasVoted,
		},
		{
			from:   &addr,
			id:     1,
			result: uint8(Pass),
			err:    ErrNotFoundCouncilMember,
		},
		{
			// vote success
			from:   &admin2Addr,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			from:   &admin3Addr,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			// has finish the proposal
			from:   &admin4Addr,
			id:     1,
			result: uint8(Pass),
			err:    ErrProposalFinished,
		},
	}

	for _, test := range testcases {
		logs = make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			StateLedger: stateLedger,
			BlockNumber: 1,
			CurrentLogs: &logs,
			From:        test.from,
		})

		err := gov.Vote(test.id, test.result)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			if test.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestGovernance_GetLatestProposalID(t *testing.T) {
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

	admin1Addr := types.NewAddressByStr(admin1).ETHAddress()

	arg := NodeUpgradeExtraArgs{
		DownloadUrls: []string{"http://127.0.0.1:10000"},
		CheckHash:    "",
	}
	data, err := json.Marshal(arg)
	assert.Nil(t, err)

	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		From:        &admin1Addr,
		BlockNumber: 1,
		StateLedger: stateLedger,
		CurrentLogs: &logs,
	})

	proposalID := gov.GetLatestProposalID()
	assert.EqualValues(t, 0, proposalID)

	err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
	assert.Nil(t, err)

	proposalID = gov.GetLatestProposalID()
	assert.EqualValues(t, 1, proposalID)
}

func TestGovernance_Proposal(t *testing.T) {
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

	admin1Addr := types.NewAddressByStr(admin1).ETHAddress()
	admin2Addr := types.NewAddressByStr(admin2).ETHAddress()
	admin3Addr := types.NewAddressByStr(admin3).ETHAddress()

	arg := NodeUpgradeExtraArgs{
		DownloadUrls: []string{"http://127.0.0.1:10000"},
		CheckHash:    "",
	}
	data, err := json.Marshal(arg)
	assert.Nil(t, err)

	logs := make([]common.Log, 0)
	gov.SetContext(&common.VMContext{
		From:        &admin1Addr,
		BlockNumber: 1,
		StateLedger: stateLedger,
		CurrentLogs: &logs,
	})
	err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
	assert.Nil(t, err)

	proposal, err := gov.Proposal(1)
	assert.Nil(t, err)
	assert.EqualValues(t, Voting, proposal.Status)
	assert.EqualValues(t, 1, proposal.CreatedBlockNumber)
	assert.Equal(t, 1, len(logs))

	testcases := []struct {
		from   *ethcommon.Address
		id     uint64
		result uint8
		err    error
		hasErr bool
	}{
		{
			// vote success
			from:   &admin2Addr,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			from:   &admin3Addr,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
	}

	for _, test := range testcases {
		logs = make([]common.Log, 0)
		gov.SetContext(&common.VMContext{
			StateLedger: stateLedger,
			BlockNumber: 2,
			CurrentLogs: &logs,
			From:        test.from,
		})

		err := gov.Vote(test.id, test.result)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			if test.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}

	proposal, err = gov.Proposal(1)
	assert.Nil(t, err)
	assert.EqualValues(t, Approved, proposal.Status)
	assert.EqualValues(t, 1, proposal.CreatedBlockNumber)
	assert.EqualValues(t, 2, proposal.EffectiveBlockNumber)
	assert.Equal(t, 1, len(logs))
}
