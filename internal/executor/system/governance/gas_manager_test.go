package governance

import (
	"encoding/json"
	"testing"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func prepareLedger(t *testing.T) ledger.StateLedger {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.EpochManagerContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	return stateLedger
}

func TestGasManager_RunForPropose(t *testing.T) {

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.GasManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	gm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

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
	err = InitGasMembers(stateLedger, &rbft.EpochInfo{
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
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data: generateGasUpdateProposeData(t, GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       50000000000,
				GasChangeRateValue: 1250,
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateProposeData(t, GasExtraArgs{
					MaxGasPrice:        10000000000000,
					MinGasPrice:        1000000000000,
					InitGasPrice:       50000000000,
					GasChangeRateValue: 1250,
				})),
				Err: ErrGasUpperOrLlower,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateGasUpdateProposeData(t, GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       0,
				GasChangeRateValue: 1250,
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateProposeData(t, GasExtraArgs{
					MaxGasPrice:        10000000000000,
					MinGasPrice:        1000000000000,
					InitGasPrice:       0,
					GasChangeRateValue: 1250,
				})),
				Err: ErrGasArgsType,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateGasUpdateProposeData(t, GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       5000000000000,
				GasChangeRateValue: 1250,
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateProposeData(t, GasExtraArgs{
					MaxGasPrice:        10000000000000,
					MinGasPrice:        1000000000000,
					InitGasPrice:       5000000000000,
					GasChangeRateValue: 1250,
				})),
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateGasUpdateProposeData(t, GasExtraArgs{
				MaxGasPrice:        10000000000000,
				MinGasPrice:        1000000000000,
				InitGasPrice:       6000000000000,
				GasChangeRateValue: 1250,
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateProposeData(t, GasExtraArgs{
					MaxGasPrice:        10000000000000,
					MinGasPrice:        1000000000000,
					InitGasPrice:       6000000000000,
					GasChangeRateValue: 1250,
				})),
			},
			Err: ErrExistNotFinishedGasProposal,
		},
	}

	for _, test := range testcases {
		gm.Reset(1, stateLedger)
		res, err := gm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})

		assert.Equal(t, test.Err, err)
		if res != nil {
			assert.Equal(t, test.Expected.UsedGas, res.UsedGas)
			assert.Equal(t, test.Expected.Err, res.Err)
		}
	}

}

func TestGasManager_RunForVote(t *testing.T) {

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.GasManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	gm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

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
	err = InitGasMembers(stateLedger, &rbft.EpochInfo{
		FinanceParams: rbft.FinanceParams{
			StartGasPriceAvailable: true,
			StartGasPrice:          5000000000000,
			MaxGasPrice:            10000000000000,
			MinGasPrice:            1000000000000,
			GasChangeRateValue:     1250,
		},
	})
	assert.Nil(t, err)

	// propose
	gm.Reset(1, stateLedger)
	_, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateGasUpdateProposeData(t, GasExtraArgs{
			MaxGasPrice:        10000000000000,
			MinGasPrice:        1000000000000,
			InitGasPrice:       5000000000000,
			GasChangeRateValue: 1250,
		}),
	})
	assert.Nil(t, err)

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data:   generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass)),
			},
			Err: ErrUseHasVoted,
		},
		{
			Caller: admin2,
			Data:   generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller: admin2,
			Data:   generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass)),
			},
			Err: ErrUseHasVoted,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data:   generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass)),
			},
			Err: ErrNotFoundCouncilMember,
		},
	}

	for _, test := range testcases {
		gm.Reset(1, stateLedger)

		result, err := gm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})

		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)
			assert.Equal(t, test.Expected.Err, result.Err)
		}
	}
}

func TestGasManager_RunForVote_Approved(t *testing.T) {

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.GasManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	gm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

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
	err = InitGasMembers(stateLedger, &rbft.EpochInfo{
		FinanceParams: rbft.FinanceParams{
			StartGasPriceAvailable: true,
			StartGasPrice:          5000000000000,
			MaxGasPrice:            10000000000000,
			MinGasPrice:            1000000000000,
			GasChangeRateValue:     1250,
		},
	})
	assert.Nil(t, err)

	// propose
	gm.Reset(1, stateLedger)
	_, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateGasUpdateProposeData(t, GasExtraArgs{
			MaxGasPrice:        10000000000000,
			MinGasPrice:        1000000000000,
			InitGasPrice:       5000000000000,
			GasChangeRateValue: 1250,
		}),
	})
	assert.Nil(t, err)

	gm.Reset(1, stateLedger)
	_, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin2).ETHAddress(),
		Data: generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)
	_, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin3).ETHAddress(),
		Data: generateGasUpdateVoteData(t, gm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)

	gasProposal, err := gm.loadGasProposal(gm.proposalID.GetID() - 1)
	assert.Nil(t, err)
	assert.Equal(t, Approved, gasProposal.Status)
}

func TestGasManager_EstimateGas(t *testing.T) {
	nm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	gabi, err := GetABI()
	assert.Nil(t, err)

	data, err := gabi.Pack(ProposeMethod, uint8(GasUpdate), "title", "desc", uint64(1000), []byte(""))
	assert.Nil(t, err)

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.NodeManagerContractAddr).ETHAddress()
	dataBytes := hexutil.Bytes(data)

	// test propose
	gas, err := nm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &dataBytes,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(dataBytes), gas)

	// test vote
	data, err = gabi.Pack(VoteMethod, uint64(1), uint8(Pass), []byte(""))
	dataBytes = hexutil.Bytes(data)
	assert.Nil(t, err)
	gas, err = nm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &dataBytes,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(dataBytes), gas)
}

func TestGaseManager_GetProposal(t *testing.T) {

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.GasManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	gm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

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
	err = InitGasMembers(stateLedger, &rbft.EpochInfo{
		FinanceParams: rbft.FinanceParams{
			StartGasPriceAvailable: true,
			StartGasPrice:          5000000000000,
			MaxGasPrice:            10000000000000,
			MinGasPrice:            1000000000000,
			GasChangeRateValue:     1250,
		},
	})
	assert.Nil(t, err)

	// propose
	gm.Reset(1, stateLedger)
	_, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateGasUpdateProposeData(t, GasExtraArgs{
			MaxGasPrice:        10000000000000,
			MinGasPrice:        1000000000000,
			InitGasPrice:       5000000000000,
			GasChangeRateValue: 1250,
		}),
	})
	assert.Nil(t, err)

	execResult, err := gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err := gm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(ret))

	proposal := &GasProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.Equal(t, "desc", proposal.Desc)
	assert.EqualValues(t, 1, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))

	_, err = gm.vote(types.NewAddressByStr(admin2).ETHAddress(), &VoteArgs{
		BaseVoteArgs: BaseVoteArgs{
			ProposalId: 1,
			VoteResult: uint8(Pass),
		},
	})
	assert.Nil(t, err)
	execResult, err = gm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err = gm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)

	proposal = &GasProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.EqualValues(t, 2, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))
}

func TestGasManager_CheckAndUpdateState(t *testing.T) {
	gm := NewGasManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	addr := types.NewAddressByStr(common.GasManagerContractAddr)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.GasManagerContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	gm.account = stateLedger.GetOrCreateAccount(addr)
	gm.stateLedger = stateLedger
	gm.currentLog = &common.Log{
		Address: addr,
	}
	gm.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	gm.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)

	gm.checkAndUpdateState(1)
}

func generateGasUpdateProposeData(t *testing.T, extraArgs GasExtraArgs) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(GasUpdate), title, desc, blockNumber, extra)
	assert.Nil(t, err)
	return data
}

func generateGasUpdateVoteData(t *testing.T, proposalID uint64, voteResult VoteResult) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	data, err := gabi.Pack(VoteMethod, proposalID, uint8(voteResult), []byte(""))
	assert.Nil(t, err)

	return data
}
