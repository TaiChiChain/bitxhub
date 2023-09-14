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
	"github.com/axiomesh/axiom/internal/executor/system/common"
	"github.com/axiomesh/axiom/internal/ledger"
	"github.com/axiomesh/axiom/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

func TestNodeManager_RunForPropose(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "node_manager"))
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*NodeMember{
		{
			NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
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
						NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: NodeManagementProposalGas,
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data: generateNodeAddProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				Err: ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data: generateNodeAddProposeData(t, NodeExtraArgs{
				Nodes: []*NodeMember{
					{
						NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					},
					{
						NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
					},
				},
			}),
			Expected: vm.ExecutionResult{
				Err: ErrRepeatedNodeID,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(stateLedger)

		res, err := nm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})

		assert.Equal(t, test.Err, err)
		if res != nil {
			assert.Equal(t, uint64(NodeManagementProposalGas), res.UsedGas)
			assert.Equal(t, test.Expected.Err, res.Err)
		}
	}
}

func TestNodeManager_RunForVote(t *testing.T) {
	nm := NewNodeManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "node_manager"))
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

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
	err = InitNodeMembers(stateLedger, []*NodeMember{
		{
			NodeId: "16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
		},
	})

	// propose
	nm.Reset(stateLedger)
	_, err = nm.Run(&vm.Message{
		From: types.NewAddressByStr(admin1).ETHAddress(),
		Data: generateNodeAddProposeData(t, NodeExtraArgs{
			Nodes: []*NodeMember{
				{
					NodeId: "26Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
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
			Data:   generateNodeAddVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: NodeManagementVoteGas,
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: admin2,
			Data:   generateNodeAddVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: NodeManagementVoteGas,
			},
			Err: nil,
		},
		{
			Caller: admin2,
			Data:   generateNodeAddVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: NodeManagementVoteGas,
				Err:     ErrUseHasVoted,
			},
			Err: nil,
		},
		{
			Caller: "0x1000000000000000000000000000000000000000",
			Data:   generateNodeAddVoteData(t, nm.proposalID.GetID()-1, Pass),
			Expected: vm.ExecutionResult{
				UsedGas: NodeManagementVoteGas,
				Err:     ErrNotFoundCouncilMember,
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		nm.Reset(stateLedger)

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
	assert.Equal(t, NodeManagementProposalGas, gas)

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
	assert.Equal(t, NodeManagementVoteGas, gas)
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

func generateNodeAddVoteData(t *testing.T, proposalID uint64, voteResult VoteResult) []byte {
	gabi, err := GetABI()
	assert.Nil(t, err)

	data, err := gabi.Pack(VoteMethod, proposalID, uint8(voteResult), []byte(""))
	assert.Nil(t, err)

	return data
}
