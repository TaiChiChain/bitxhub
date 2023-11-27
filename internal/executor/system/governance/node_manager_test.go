package governance

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

func TestNodeManager_RunForPropose(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
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
			Caller: admin1,
			Data: generateNodeAddProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeAddProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data: generateNodeAddProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeAddProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
				Err: ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateNodeAddProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeAddProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
				Err: ErrRepeatedNodeID,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateNodeRemoveProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeRemoveProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data: generateNodeRemoveProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeRemoveProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
				Err: ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateNodeRemoveProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeRemoveProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
				Err: ErrRepeatedNodeID,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(1, stateLedger)

		res, err := nm.Run(&vm.Message{
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

func TestNodeManager_RunForNodeUpgradePropose(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
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
			Caller: admin1,
			Data: generateNodeUpgradeProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
						Address: admin1,
						Name:    "111",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeUpgradeProposeData(t, NodeExtraArgs{
					Nodes: []*NodeMember{
						{
							NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
							Address: admin1,
							Name:    "111",
						},
					},
				})),
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(1, stateLedger)

		res, err := nm.Run(&vm.Message{
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

func TestNodeManager_GetNodeMembers(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})

	members, err := GetNodeMembers(stateLedger)
	assert.Nil(t, err)
	t.Log("GetNodeMembers-members:", members)
}

func TestNodeManager_RunForAddVote(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeAddProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId:  "16Uiu2HAmTMVkvoGdwjHkqSEhdSM5P7L8ronFfnDePhmQSN6CvR8m",
					Address: admin1,
					Name:    "111",
				},
			},
		}),
	})
	assert.Nil(t, err)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	g.DataSyncerSet = append(g.DataSyncerSet, rbft.NodeInfo{
		ID:                   9,
		AccountAddress:       "0x88E9A1cE92b4D6e4d860CFBB5bB7aC44d9b548f8",
		P2PNodeID:            "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
		ConsensusVotingPower: 100,
	})
	err = base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: admin2,
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller: admin2,
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(1, stateLedger)

		result, err := nm.Run(&vm.Message{
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

func TestNodeManager_RunForAddVote_Approved(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})
	assert.Nil(t, err)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err = base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeAddProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId:  "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
					Address: admin2,
					Name:    "222",
				},
			},
		}),
	})
	assert.Nil(t, err)

	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin2).ETHAddress(),
		Data: generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin3).ETHAddress(),
		Data: generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)

	nodeProposal, err := nm.loadNodeProposal(nm.proposalID.GetID() - 1)
	assert.Nil(t, err)
	assert.Equal(t, Approved, nodeProposal.Status)

	_, data := nm.account.GetState([]byte(NodeMembersKey))
	nodeMembers := make([]*NodeMember, 0)
	err = json.Unmarshal(data, &nodeMembers)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(nodeMembers))
}

func TestNodeManager_RunForRemoveVote_Approved(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
		{
			ID:   4,
			Name: "444",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
			{
				ID:             4,
				AccountAddress: admin4,
				P2PNodeID:      "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
			},
		},
	})
	assert.Nil(t, err)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	g.DataSyncerSet = append(g.DataSyncerSet, rbft.NodeInfo{
		ID:                   9,
		AccountAddress:       admin4,
		P2PNodeID:            "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
		ConsensusVotingPower: 100,
	})
	err = base.InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeRemoveProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					ID:      9,
					NodeId:  "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
					Address: admin4,
					Name:    "444",
				},
			},
		}),
	})
	assert.Nil(t, err)

	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin2).ETHAddress(),
		Data: generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin3).ETHAddress(),
		Data: generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
	})
	assert.Nil(t, err)

	nodeProposal, err := nm.loadNodeProposal(nm.proposalID.GetID() - 1)
	assert.Nil(t, err)
	assert.Equal(t, Approved, nodeProposal.Status)

	_, data := nm.account.GetState([]byte(NodeMembersKey))
	nodeMembers := make([]*NodeMember, 0)
	err = json.Unmarshal(data, &nodeMembers)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(nodeMembers))
}

func TestNodeManager_RunForRemoveVote(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeRemoveProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					Address: admin1,
					Name:    "111",
				},
			},
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
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(1, stateLedger)

		result, err := nm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})

		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)
		}
	}
}

func TestNodeManager_RunForUpgradeVote(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeUpgradeProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					Address: admin1,
					Name:    "111",
				},
			},
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
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data:   generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateNodeVoteData(t, nm.proposalID.GetID()-1, Pass)),
				Err:     ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(1, stateLedger)

		result, err := nm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})

		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)
		}
	}
}

func TestNodeManager_EstimateGas(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	gabi, err := GetABI()
	assert.Nil(t, err)

	data, err := gabi.Pack(ProposeMethod, uint8(NodeAdd), "title", "desc", uint64(1000), []byte(""))
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

func TestNodeManager_GetProposal(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*repo.NodeName{
		{
			ID:   1,
			Name: "111",
		},
	}, &rbft.EpochInfo{
		ValidatorSet: []rbft.NodeInfo{
			{
				ID:             1,
				AccountAddress: admin1,
				P2PNodeID:      "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
			},
		},
	})
	assert.Nil(t, err)

	// propose
	nm.Reset(1, stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeUpgradeProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId:  "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					Address: admin1,
					Name:    "111",
				},
			},
		}),
	})
	assert.Nil(t, err)

	execResult, err := nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err := nm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(ret))

	proposal := &NodeProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.Equal(t, "desc", proposal.Desc)
	assert.EqualValues(t, 1, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))

	_, err = nm.vote(types.NewAddressByStr(admin2).ETHAddress(), &VoteArgs{
		BaseVoteArgs: BaseVoteArgs{
			ProposalId: 1,
			VoteResult: uint8(Pass),
		},
	})
	assert.Nil(t, err)
	execResult, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateProposalData(t, 1),
	})
	assert.Nil(t, err)
	ret, err = nm.gov.UnpackOutputArgs(ProposalMethod, execResult.ReturnData)
	assert.Nil(t, err)

	proposal = &NodeProposal{}
	err = json.Unmarshal(ret[0].([]byte), proposal)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, proposal.ID)
	assert.EqualValues(t, 2, len(proposal.PassVotes))
	assert.EqualValues(t, 0, len(proposal.RejectVotes))
}

func TestNodeManager_CheckAndUpdateState(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	addr := types.NewAddressByStr(common.NodeManagerContractAddr)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	nm.account = stateLedger.GetOrCreateAccount(addr)
	nm.stateLedger = stateLedger
	nm.currentLog = &common.Log{
		Address: addr,
	}
	nm.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	nm.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)

	nm.checkAndUpdateState(1)
}

func generateNodeAddProposeData(t *testing.T, extraArgs NodeExtraArgs) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(NodeAdd), title, desc, blockNumber, extra)
	assert.Nil(t, err)
	return data
}

func generateNodeVoteData(t *testing.T, proposalID uint64, voteResult VoteResult) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	data, err := gabi.Pack(VoteMethod, proposalID, uint8(voteResult), []byte(""))
	assert.Nil(t, err)

	return data
}

func generateNodeRemoveProposeData(t *testing.T, extraArgs NodeExtraArgs) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(NodeRemove), title, desc, blockNumber, extra)
	assert.Nil(t, err)
	return data
}

func generateNodeUpgradeProposeData(t *testing.T, extraArgs NodeExtraArgs) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	title := "title"
	desc := "desc"
	blockNumber := uint64(1000)
	extra, err := json.Marshal(extraArgs)
	assert.Nil(t, err)
	data, err := gabi.Pack(ProposeMethod, uint8(NodeUpgrade), title, desc, blockNumber, extra)
	assert.Nil(t, err)
	return data
}
