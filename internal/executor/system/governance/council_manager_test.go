package governance

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

func TestRunForPropose(t *testing.T) {
	cm := NewCouncilManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "node_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	err = InitCouncilMembers(stateLedger, []*repo.Admin{
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

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data: generateProposeData(t, CouncilExtraArgs{
				Candidates: []*CouncilMember{
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
			}),
			Expected: vm.ExecutionResult{},
			Err:      ErrMinCouncilMembersCount,
		},
		{
			Caller: admin1,
			Data: generateProposeData(t, CouncilExtraArgs{
				Candidates: []*CouncilMember{
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
			}),
			Expected: vm.ExecutionResult{},
			Err:      ErrRepeatedAddress,
		},
		{
			Caller: admin1,
			Data: generateProposeData(t, CouncilExtraArgs{
				Candidates: []*CouncilMember{
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
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateProposeData(t, CouncilExtraArgs{
					Candidates: []*CouncilMember{
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
				})),
				ReturnData: generateReturnData(t, cm.gov, 1),
			},
			Err: nil,
		},
		{
			Caller: "0xfff0000000000000000000000000000000000000",
			Data: generateProposeData(t, CouncilExtraArgs{
				Candidates: []*CouncilMember{
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
			}),
			Expected: vm.ExecutionResult{},
			Err:      ErrNotFoundCouncilMember,
		},
	}

	for _, test := range testcases {
		cm.Reset(1, stateLedger)

		result, err := cm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, nil, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)

			assert.EqualValues(t, test.Expected.ReturnData, result.ReturnData)
		}
	}
}

func TestRunForVote(t *testing.T) {
	cm := NewCouncilManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "node_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	err = InitCouncilMembers(stateLedger, []*repo.Admin{
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
	}, "10000000")
	assert.Nil(t, err)

	cm.Reset(1, stateLedger)

	_, err = cm.propose(types.NewAddressByStr(admin1).ETHAddress(), &CouncilProposalArgs{
		BaseProposalArgs: BaseProposalArgs{
			ProposalType: uint8(CouncilElect),
			Title:        "council elect",
			Desc:         "desc",
			BlockNumber:  2,
		},
		CouncilExtraArgs: CouncilExtraArgs{
			Candidates: []*CouncilMember{
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
		},
	})
	assert.Nil(t, err)

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin2,
			Data:   generateVoteData(t, cm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateVoteData(t, cm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller: admin3,
			Data:   generateVoteData(t, cm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateVoteData(t, cm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller:   "0xfff0000000000000000000000000000000000000",
			Data:     generateVoteData(t, cm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{},
			Err:      ErrNotFoundCouncilMember,
		},
	}

	for _, test := range testcases {
		cm.Reset(1, stateLedger)

		result, err := cm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, nil, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)
		}
	}
}

func TestRunForGetProposal(t *testing.T) {
	cm := NewCouncilManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "council_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	err = InitCouncilMembers(stateLedger, []*repo.Admin{
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
	}, "10000000")
	assert.Nil(t, err)

	cm.Reset(1, stateLedger)

	_, err = cm.propose(types.NewAddressByStr(admin1).ETHAddress(), &CouncilProposalArgs{
		BaseProposalArgs: BaseProposalArgs{
			ProposalType: uint8(CouncilElect),
			Title:        "council elect",
			Desc:         "desc",
			BlockNumber:  2,
		},
		CouncilExtraArgs: CouncilExtraArgs{
			Candidates: []*CouncilMember{
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
		},
	})
	assert.Nil(t, err)

	execResult, err := cm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err := cm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(ret))

	proposal := &CouncilProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.Equal(t, "desc", proposal.Desc)
	assert.EqualValues(t, 1, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))
	assert.Equal(t, 4, len(proposal.Candidates))

	_, err = cm.vote(types.NewAddressByStr(admin2).ETHAddress(), &CouncilVoteArgs{
		BaseVoteArgs: BaseVoteArgs{
			ProposalId: 1,
			VoteResult: uint8(Pass),
		},
	})
	assert.Nil(t, err)
	execResult, err = cm.Run(&vm.Message{
		From: types.NewAddressByStr(admin2).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err = cm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)

	proposal = &CouncilProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.EqualValues(t, 2, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))
}

func TestEstimateGas(t *testing.T) {
	cm := NewCouncilManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.CouncilManagerContractAddr).ETHAddress()
	data := hexutil.Bytes(generateProposeData(t, CouncilExtraArgs{
		Candidates: []*CouncilMember{
			{
				Address: admin1,
				Weight:  1,
			},
		},
	}))
	// test propose
	gas, err := cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(data), gas)

	// test vote
	data = generateVoteData(t, 1, Pass)
	gas, err = cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(data), gas)
}

func generateProposeData(t *testing.T, extraArgs CouncilExtraArgs) []byte {
	gabi, err := GetABI()

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(CouncilElect), title, desc, blockNumber, extra)
	assert.Nil(t, err)

	return data
}

func generateVoteData(t *testing.T, proposalID uint64, voteResult VoteResult) []byte {
	gabi, err := GetABI()

	data, err := gabi.Pack(VoteMethod, proposalID, voteResult, []byte(""))
	assert.Nil(t, err)

	return data
}

func generateProposalData(t *testing.T, proposalID uint64) []byte {
	gabi, err := GetABI()

	data, err := gabi.Pack(ProposalMethod, proposalID)
	assert.Nil(t, err)

	return data
}

func generateReturnData(t *testing.T, gov *Governance, id uint64) []byte {
	b, err := gov.PackOutputArgs(ProposeMethod, id)
	assert.Nil(t, err)

	return b
}
