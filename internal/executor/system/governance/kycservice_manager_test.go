package governance

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"

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
	KycService1 = "0x1000000000000000000000000000000000000001"
	KycService2 = "0x1000000000000000000000000000000000000002"
)

type TestKycServiceProposal struct {
	ID          uint64
	Type        ProposalType
	Proposer    string
	TotalVotes  uint64
	PassVotes   []string
	RejectVotes []string
	Status      ProposalStatus
	Services    []*access.KycService
}

func TestKycServiceManager_RunForPropose(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kycservice_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycServiceContractAddr), ledger.NewChanger())

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
			Caller: KycService1,
			Data: generateKycProposeData(t, KycServiceAdd, access.KycServiceArgs{
				Services: []*access.KycService{
					{
						KycAddr: *types.NewAddressByStr(KycService1),
					},
					{
						KycAddr: *types.NewAddressByStr(KycService2),
					},
				},
			}),
			Expected: vm.ExecutionResult{},
			Err:      ErrNotFoundCouncilMember,
		},
		{
			Caller: admin1,
			Data: generateKycProposeData(t, KycServiceAdd, access.KycServiceArgs{
				Services: []*access.KycService{
					{
						KycAddr: *types.NewAddressByStr(KycService1),
					},
					{
						KycAddr: *types.NewAddressByStr(KycService2),
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: KycProposalGas,
				ReturnData: generateKycReturnData(t, &TestKycServiceProposal{
					ID:          1,
					Type:        KycServiceAdd,
					Proposer:    admin1,
					TotalVotes:  4,
					PassVotes:   []string{admin1},
					RejectVotes: nil,
					Status:      Voting,
					Services: []*access.KycService{
						{
							KycAddr: *types.NewAddressByStr(KycService1),
						},
						{
							KycAddr: *types.NewAddressByStr(KycService2),
						},
					},
				}),
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data:   []byte{0, 1, 2, 3},
			Expected: vm.ExecutionResult{
				UsedGas:    KycProposalGas,
				ReturnData: nil,
				Err:        ErrMethodName,
			},
			Err: ErrMethodName,
		},
	}

	for _, test := range testcases {
		ks.Reset(stateLedger)

		result, err := ks.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, test.Expected.Err, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)

			expectedProposal := &KycProposal{}
			err = json.Unmarshal(test.Expected.ReturnData, expectedProposal)
			assert.Nil(t, err)

			actualProposal := &KycProposal{}
			err = json.Unmarshal(result.ReturnData, actualProposal)
			assert.Nil(t, err)
			assert.Equal(t, *expectedProposal, *actualProposal)

			state, _ := account.GetState([]byte(fmt.Sprintf("%s%d", KycProposalKey, actualProposal.ID)))
			assert.Equal(t, true, state)
		}
	}
}

func TestKycServiceManager_RunForVoteAdd(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "kycservice_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycServiceContractAddr), ledger.NewChanger())

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

	ks.Reset(stateLedger)

	addr := types.NewAddressByStr(admin1).ETHAddress()
	ks.propose(&addr, &KycProposalArgs{
		BaseProposalArgs: BaseProposalArgs{
			ProposalType: uint8(KycServiceAdd),
			Title:        "title",
			Desc:         "desc",
			BlockNumber:  1000,
		},
		KycServiceArgs: access.KycServiceArgs{
			Services: []*access.KycService{
				{
					KycAddr: *types.NewAddressByStr(KycService1),
				},
				{
					KycAddr: *types.NewAddressByStr(KycService2),
				},
			},
		},
	})

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin2,
			Data:   generateKycVoteData(t, ks.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: KycVoteGas,
				ReturnData: generateKycReturnData(t, &TestKycServiceProposal{
					ID:          1,
					Type:        KycServiceAdd,
					Proposer:    admin1,
					TotalVotes:  4,
					PassVotes:   []string{admin1, admin2},
					RejectVotes: nil,
					Status:      Voting,
					Services: []*access.KycService{
						{
							KycAddr: *types.NewAddressByStr(KycService1),
						},
						{
							KycAddr: *types.NewAddressByStr(KycService2),
						},
					},
				}),
			},
			Err: nil,
		},
		{
			Caller: admin3,
			Data:   generateKycVoteData(t, ks.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: KycVoteGas,
				ReturnData: generateKycReturnData(t, &TestKycServiceProposal{
					ID:          1,
					Type:        KycServiceAdd,
					Proposer:    admin1,
					TotalVotes:  4,
					PassVotes:   []string{admin1, admin2, admin3},
					RejectVotes: nil,
					Status:      Approved,
					Services: []*access.KycService{
						{
							KycAddr: *types.NewAddressByStr(KycService1),
						},
						{
							KycAddr: *types.NewAddressByStr(KycService2),
						},
					},
				}),
			},
			Err: nil,
		},
		{
			Caller: "0xfff0000000000000000000000000000000000000",
			Data:   generateKycVoteData(t, ks.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: KycVoteGas,
				Err:     ErrNotFoundCouncilMember,
			},
			Err: ErrNotFoundCouncilMember,
		},
	}

	for _, test := range testcases {
		ks.Reset(stateLedger)

		result, err := ks.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, nil, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)

			expectedProposal := &KycProposal{}
			err = json.Unmarshal(test.Expected.ReturnData, expectedProposal)
			assert.Nil(t, err)

			actualProposal := &KycProposal{}
			err = json.Unmarshal(result.ReturnData, actualProposal)
			assert.Nil(t, err)
			assert.Equal(t, *expectedProposal, *actualProposal)
		}
	}
}

func TestKycServiceManager_RunForVoteRemove(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "kycservice_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycServiceContractAddr), ledger.NewChanger())

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

	ks.Reset(stateLedger)

	addr := types.NewAddressByStr(admin1).ETHAddress()
	ks.propose(&addr, &KycProposalArgs{
		BaseProposalArgs: BaseProposalArgs{
			ProposalType: uint8(KycServiceRemove),
			Title:        "title",
			Desc:         "desc",
			BlockNumber:  1000,
		},
		KycServiceArgs: access.KycServiceArgs{
			Services: []*access.KycService{
				{
					KycAddr: *types.NewAddressByStr(KycService1),
				},
				{
					KycAddr: *types.NewAddressByStr(KycService2),
				},
			},
		},
	})

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin2,
			Data:   generateKycVoteData(t, ks.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: KycVoteGas,
				ReturnData: generateKycReturnData(t, &TestKycServiceProposal{
					ID:          1,
					Type:        KycServiceRemove,
					Proposer:    admin1,
					TotalVotes:  4,
					PassVotes:   []string{admin1, admin2},
					RejectVotes: nil,
					Status:      Voting,
					Services: []*access.KycService{
						{
							KycAddr: *types.NewAddressByStr(KycService1),
						},
						{
							KycAddr: *types.NewAddressByStr(KycService2),
						},
					},
				}),
			},
			Err: nil,
		},
		{
			Caller: admin3,
			Data:   generateKycVoteData(t, ks.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: KycVoteGas,
				ReturnData: generateKycReturnData(t, &TestKycServiceProposal{
					ID:          1,
					Type:        KycServiceRemove,
					Proposer:    admin1,
					TotalVotes:  4,
					PassVotes:   []string{admin1, admin2, admin3},
					RejectVotes: nil,
					Status:      Approved,
					Services: []*access.KycService{
						{
							KycAddr: *types.NewAddressByStr(KycService1),
						},
						{
							KycAddr: *types.NewAddressByStr(KycService2),
						},
					},
				}),
			},
			Err: fmt.Errorf("ACCESS ERROR: remove kyc services from an empty list"),
		},
	}

	for _, test := range testcases {
		ks.Reset(stateLedger)

		result, err := ks.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, nil, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)

			expectedProposal := &KycProposal{}
			err = json.Unmarshal(test.Expected.ReturnData, expectedProposal)
			assert.Nil(t, err)

			actualProposal := &KycProposal{}
			err = json.Unmarshal(result.ReturnData, actualProposal)
			assert.Nil(t, err)
			assert.Equal(t, *expectedProposal, *actualProposal)
		}
	}
}

func TestKycServiceManager_EstimateGas(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.KycServiceContractAddr).ETHAddress()
	data := hexutil.Bytes(generateKycProposeData(t, KycServiceAdd, access.KycServiceArgs{
		Services: []*access.KycService{
			{
				KycAddr: *types.NewAddressByStr(KycService1),
			},
			{
				KycAddr: *types.NewAddressByStr(KycService2),
			},
		},
	}))
	// test propose
	gas, err := ks.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, KycProposalGas, gas)

	// test vote
	data = hexutil.Bytes(generateKycVoteData(t, 1, Pass))
	gas, err = ks.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, KycVoteGas, gas)

	// test error args
	data = hexutil.Bytes([]byte{0, 1, 2, 3})
	gas, err = ks.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), gas)
}

func generateKycProposeData(t *testing.T, proposalType ProposalType, extraArgs access.KycServiceArgs) []byte {
	gabi, err := GetABI()

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(proposalType), title, desc, blockNumber, extra)
	assert.Nil(t, err)

	return data
}

func generateKycVoteData(t *testing.T, proposalID uint64, voteResult VoteResult) []byte {
	gabi, err := GetABI()

	data, err := gabi.Pack(VoteMethod, proposalID, voteResult, []byte(""))
	assert.Nil(t, err)

	return data
}

func generateKycReturnData(t *testing.T, testProposal *TestKycServiceProposal) []byte {
	proposal := &KycProposal{
		BaseProposal: BaseProposal{
			ID:          testProposal.ID,
			Type:        testProposal.Type,
			Strategy:    NowProposalStrategy,
			Proposer:    testProposal.Proposer,
			Title:       "title",
			Desc:        "desc",
			BlockNumber: uint64(1000),
			TotalVotes:  testProposal.TotalVotes,
			PassVotes:   testProposal.PassVotes,
			RejectVotes: testProposal.RejectVotes,
			Status:      testProposal.Status,
		},
		KycServiceArgs: access.KycServiceArgs{
			Services: testProposal.Services,
		},
	}

	b, err := json.Marshal(proposal)
	assert.Nil(t, err)

	return b
}

func TestKycServiceManager_CheckAndUpdateState(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	ks.CheckAndUpdateState(100, stateLedger)
}

func TestKycServiceManager_loadKycProposal(t *testing.T) {
	ks := NewKycServiceManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "kycservice_manager"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycServiceContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	ks.Reset(stateLedger)
	_, err = ks.loadKycProposal(1)
	assert.Equal(t, fmt.Errorf("node proposal not found for the id"), err)
}
