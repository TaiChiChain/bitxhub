package governance

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func generateNodeRegisterExtraArgs(t *testing.T, p2pPrivateKey string, consensusPrivateKey string, metaData node_manager.NodeMetaData) *NodeRegisterExtraArgs {
	p2pKeystore, err := repo.GenerateP2PKeystore("", p2pPrivateKey, "test")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore("", consensusPrivateKey, "test")
	assert.Nil(t, err)

	nodeRegisterExtraArgsSignStruct := &NodeRegisterExtraArgsSignStruct{
		ConsensusPubKey: consensusKeystore.PublicKey.String(),
		P2PPubKey:       p2pKeystore.PublicKey.String(),
		MetaData:        metaData,
	}
	nodeRegisterExtraArgsSignStructBytes, err := json.Marshal(nodeRegisterExtraArgsSignStruct)
	assert.Nil(t, err)
	consensusPrivateKeySignature, err := consensusKeystore.PrivateKey.Sign(nodeRegisterExtraArgsSignStructBytes)
	assert.Nil(t, err)
	p2pPrivateKeySignature, err := p2pKeystore.PrivateKey.Sign(nodeRegisterExtraArgsSignStructBytes)
	assert.Nil(t, err)
	return &NodeRegisterExtraArgs{
		ConsensusPubKey:              nodeRegisterExtraArgsSignStruct.ConsensusPubKey,
		P2PPubKey:                    nodeRegisterExtraArgsSignStruct.P2PPubKey,
		MetaData:                     nodeRegisterExtraArgsSignStruct.MetaData,
		ConsensusPrivateKeySignature: hexutil.Encode(consensusPrivateKeySignature),
		P2PPrivateKeySignature:       hexutil.Encode(p2pPrivateKeySignature),
	}
}

func TestNodeManager_RunForNodeRegisterPropose(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	type NodeRegisterData struct {
		P2PPrivateKey       string
		ConsensusPrivateKey string
		MetaData            node_manager.NodeMetaData

		NodeRegisterExtraArgs *NodeRegisterExtraArgs
	}

	operator := ethcommon.HexToAddress("0x1200000000000000000000000000000000000000")
	testcases := []struct {
		Caller                      ethcommon.Address
		Data                        NodeRegisterData
		NodeRegisterExtraArgsSetter func(args *NodeRegisterExtraArgs)
		ErrHandler                  assert.ErrorAssertionFunc
	}{
		{
			Caller: operator,
			Data: NodeRegisterData{
				MetaData: node_manager.NodeMetaData{
					Name: "node5",
				},
			},
			ErrHandler: assert.NoError,
		},
		{
			Caller: operator,
			Data: NodeRegisterData{
				MetaData: node_manager.NodeMetaData{
					Name: "node6",
				},
			},
			NodeRegisterExtraArgsSetter: func(args *NodeRegisterExtraArgs) {
				args.P2PPrivateKeySignature = "err signature"
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "failed to verify p2p private key signature")
			},
		},
		{
			Caller: operator,
			Data: NodeRegisterData{
				MetaData: node_manager.NodeMetaData{
					Name: "node6",
				},
			},
			NodeRegisterExtraArgsSetter: func(args *NodeRegisterExtraArgs) {
				args.ConsensusPrivateKeySignature = "err signature"
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "failed to verify consensus private key signature")
			},
		},
		{
			Caller: operator,
			Data: NodeRegisterData{
				MetaData: node_manager.NodeMetaData{
					Name: "node1",
				},
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "name already registered")
			},
		},
		{
			Caller: operator,
			Data: NodeRegisterData{
				ConsensusPrivateKey: repo.MockConsensusKeys[0],
				MetaData: node_manager.NodeMetaData{
					Name: "node6",
				},
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "consensus public key already registered")
			},
		},
		{
			Caller: operator,
			Data: NodeRegisterData{
				P2PPrivateKey: repo.MockP2PKeys[0],
				MetaData: node_manager.NodeMetaData{
					Name: "node6",
				},
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "p2p public key already registered")
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				if test.Data.NodeRegisterExtraArgs == nil {
					test.Data.NodeRegisterExtraArgs = generateNodeRegisterExtraArgs(t, test.Data.P2PPrivateKey, test.Data.ConsensusPrivateKey, test.Data.MetaData)
				}

				if test.NodeRegisterExtraArgsSetter != nil {
					test.NodeRegisterExtraArgsSetter(test.Data.NodeRegisterExtraArgs)
				}

				data, err := json.Marshal(test.Data.NodeRegisterExtraArgs)
				assert.Nil(t, err)

				err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, data)
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}

func TestNodeManager_RunForNodeRemovePropose(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	testcases := []struct {
		Caller         ethcommon.Address
		Data           NodeRemoveExtraArgs
		PreProposeHook func(t *testing.T, ctx *common.VMContext)
		ErrHandler     assert.ErrorAssertionFunc
	}{
		{
			Caller: admin1,
			Data: NodeRemoveExtraArgs{
				NodeIDs: []uint64{1},
			},
			ErrHandler: assert.NoError,
		},
		{
			Caller: admin1,
			Data: NodeRemoveExtraArgs{
				NodeIDs: []uint64{1},
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "remove node already exist in other proposal")
			},
		},
		{
			Caller: ethcommon.HexToAddress("0x1200000000000000000000000000000000000000"),
			Data: NodeRemoveExtraArgs{
				NodeIDs: []uint64{3},
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrNotFoundCouncilMember.Error())
			},
		},
		{
			Caller: admin1,
			Data: NodeRemoveExtraArgs{
				NodeIDs: []uint64{2},
			},
			PreProposeHook: func(t *testing.T, ctx *common.VMContext) {
				nodeManagerContract := framework.NodeManagerBuildConfig.Build(ctx)
				nodeInfo, err := nodeManagerContract.GetInfo(2)
				assert.Nil(t, err)
				nodeInfo.Status = uint8(types.NodeStatusExited)
				err = nodeManagerContract.TestPutNodeInfo(&nodeInfo)
				assert.Nil(t, err)
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "node already exited")
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				if test.PreProposeHook != nil {
					test.PreProposeHook(t, gov.CrossCallSystemContractContext())
				}
				data, err := json.Marshal(test.Data)
				assert.Nil(t, err)

				err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, data)
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}

func TestNodeManager_RunForNodeUpgradePropose(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	testcases := []struct {
		Caller     ethcommon.Address
		Data       *NodeUpgradeExtraArgs
		ErrHandler assert.ErrorAssertionFunc
	}{
		{
			Caller: admin1,
			Data: &NodeUpgradeExtraArgs{
				DownloadUrls: []string{"1", "2"},
				CheckHash:    "1",
			},
			ErrHandler: assert.NoError,
		},
		{
			Caller: admin1,
			Data: &NodeUpgradeExtraArgs{
				DownloadUrls: []string{"1", "1"},
				CheckHash:    "1",
			},
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, "repeated download url")
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				data, err := json.Marshal(test.Data)
				assert.Nil(t, err)

				err = gov.Propose(uint8(NodeUpgrade), "test", "test desc", 100, data)
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}

func TestNodeManager_RunForRegisterVote(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		args := generateNodeRegisterExtraArgs(t, "", "", node_manager.NodeMetaData{Name: "node5"})

		data, err := json.Marshal(args)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	testcases := []struct {
		Caller     ethcommon.Address
		Res        VoteResult
		ErrHandler assert.ErrorAssertionFunc
	}{
		{
			Caller: admin1,
			Res:    Pass,
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrUseHasVoted.Error())
			},
		},
		{
			Caller:     admin2,
			Res:        Pass,
			ErrHandler: assert.NoError,
		},
		{
			Caller: admin2,
			Res:    Pass,
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrUseHasVoted.Error())
			},
		},
		{
			Caller: ethcommon.HexToAddress("0x1000000000000000000000000000000000000000"),
			Res:    Pass,
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrNotFoundCouncilMember.Error())
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(proposalID, uint8(test.Res))
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}

func TestNodeManager_RunForRegisterClean(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		args := generateNodeRegisterExtraArgs(t, "", "", node_manager.NodeMetaData{Name: "node5"})

		data, err := json.Marshal(args)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	// re propose used same data after before
	testNVM.Call(gov, admin1, func() {
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, proposal.Extra)
		assert.ErrorContains(t, err, "already registered")
	})

	// clean proposal, delete pending indexes
	testNVM.RunSingleTX(gov, admin1, func() error {
		handler, err := gov.getHandler(NodeRegister)
		assert.Nil(t, err)
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)
		err = handler.CleanProposal(proposal)
		assert.Nil(t, err)
		return err
	})

	// re propose used same data after clean
	testNVM.RunSingleTX(gov, admin1, func() error {
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, proposal.Extra)
		assert.Nil(t, err)
		return err
	})
}

func TestNodeManager_RunForRemoveClean(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		args := &NodeRemoveExtraArgs{
			NodeIDs: []uint64{1},
		}

		data, err := json.Marshal(args)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	// re propose used same data after before
	testNVM.Call(gov, admin1, func() {
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, proposal.Extra)
		assert.ErrorContains(t, err, "remove node already exist")
	})

	// clean proposal, delete pending indexes
	testNVM.RunSingleTX(gov, admin1, func() error {
		handler, err := gov.getHandler(NodeRegister)
		assert.Nil(t, err)
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)
		err = handler.CleanProposal(proposal)
		assert.Nil(t, err)
		return err
	})

	// re propose used same data after clean
	testNVM.RunSingleTX(gov, admin1, func() error {
		proposal, err := gov.Proposal(proposalID)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, proposal.Extra)
		assert.Nil(t, err)
		return err
	})
}

func TestNodeManager_RunForRegisterVote_Approved(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		args := generateNodeRegisterExtraArgs(t, "", "", node_manager.NodeMetaData{
			Name:       "node5",
			Desc:       "Desc",
			ImageURL:   "ImageURL",
			WebsiteURL: "WebsiteURL",
		})

		data, err := json.Marshal(args)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRegister), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	testcases := []struct {
		Caller ethcommon.Address
		Res    VoteResult
	}{
		{
			Caller: admin2,
			Res:    Pass,
		},
		{
			Caller: admin3,
			Res:    Pass,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(proposalID, uint8(test.Res))
				assert.Nil(t, err)
				return err
			})
		})
	}

	nodeInfo, err := nodeManager.GetInfo(5)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, nodeInfo.ID)
	assert.EqualValues(t, "node5", nodeInfo.MetaData.Name)
	assert.EqualValues(t, "Desc", nodeInfo.MetaData.Desc)
	assert.EqualValues(t, "ImageURL", nodeInfo.MetaData.ImageURL)
	assert.EqualValues(t, "WebsiteURL", nodeInfo.MetaData.WebsiteURL)
}

func TestNodeManager_RunForRemoveVote(t *testing.T) {
	testNVM, gov := initGovernance(t, func(rep *repo.Repo) {
		node5 := generateNodeRegisterExtraArgs(t, "", "", node_manager.NodeMetaData{Name: "node5"})
		rep.GenesisConfig.Nodes = append(rep.GenesisConfig.Nodes, repo.GenesisNodeInfo{
			ConsensusPubKey: node5.ConsensusPubKey,
			P2PPubKey:       node5.P2PPubKey,
			OperatorAddress: admin1.String(),
			MetaData: repo.GenesisNodeMetaData{
				Name:       node5.MetaData.Name,
				Desc:       node5.MetaData.Desc,
				ImageURL:   node5.MetaData.ImageURL,
				WebsiteURL: node5.MetaData.WebsiteURL,
			},
			IsDataSyncer:   true,
			StakeNumber:    types.CoinNumberByAxc(0),
			CommissionRate: 0,
		})
	})
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(NodeRemoveExtraArgs{
			NodeIDs: []uint64{5},
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	testcases := []struct {
		Caller     ethcommon.Address
		Res        uint8
		ErrHandler assert.ErrorAssertionFunc
	}{
		{
			Caller:     admin2,
			Res:        uint8(Pass),
			ErrHandler: assert.NoError,
		},
		{
			Caller: admin1,
			Res:    uint8(Pass),
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrUseHasVoted.Error())
			},
		},
		{
			Caller: ethcommon.HexToAddress("0x1000000000000000000000000000000000000000"),
			Res:    uint8(Pass),
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrNotFoundCouncilMember.Error())
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(proposalID, test.Res)
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}

func TestNodeManager_RunForRemoveVote_Approved(t *testing.T) {
	testNVM, gov := initGovernance(t, func(rep *repo.Repo) {
		node5 := generateNodeRegisterExtraArgs(t, "", "", node_manager.NodeMetaData{Name: "node5"})
		rep.GenesisConfig.Nodes = append(rep.GenesisConfig.Nodes, repo.GenesisNodeInfo{
			ConsensusPubKey: node5.ConsensusPubKey,
			P2PPubKey:       node5.P2PPubKey,
			OperatorAddress: admin1.String(),
			MetaData: repo.GenesisNodeMetaData{
				Name:       node5.MetaData.Name,
				Desc:       node5.MetaData.Desc,
				ImageURL:   node5.MetaData.ImageURL,
				WebsiteURL: node5.MetaData.WebsiteURL,
			},
			IsDataSyncer:   true,
			StakeNumber:    types.CoinNumberByAxc(0),
			CommissionRate: 0,
		})
	})
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(NodeRemoveExtraArgs{
			NodeIDs: []uint64{5},
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeRemove), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	testcases := []struct {
		Caller ethcommon.Address
		Res    VoteResult
	}{
		{
			Caller: admin2,
			Res:    Pass,
		},
		{
			Caller: admin3,
			Res:    Pass,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(proposalID, uint8(test.Res))
				assert.Nil(t, err)
				return err
			})
		})
	}

	nodeInfo, err := nodeManager.GetInfo(5)
	assert.Nil(t, err)
	assert.EqualValues(t, types.NodeStatusExited, nodeInfo.Status)
}

func TestNodeManager_RunForUpgradeVote(t *testing.T) {
	testNVM, gov := initGovernance(t)
	nodeManager := framework.NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManager := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManager, nodeManager)

	// propose
	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(NodeUpgradeExtraArgs{
			DownloadUrls: []string{"http://127.0.0.1:9000"},
			CheckHash:    "hash",
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	proposalID, err := gov.GetLatestProposalID()
	assert.Nil(t, err)

	testcases := []struct {
		Caller     ethcommon.Address
		Res        uint8
		ErrHandler assert.ErrorAssertionFunc
		Err        error
		HasErr     bool
	}{
		{
			Caller:     admin2,
			Res:        uint8(Pass),
			ErrHandler: assert.NoError,
		},
		{
			Caller: admin1,
			Res:    uint8(Pass),
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrUseHasVoted.Error())
			},
		},
		{
			Caller: ethcommon.HexToAddress("0x1000000000000000000000000000000000000000"),
			Res:    uint8(Pass),
			ErrHandler: func(t assert.TestingT, err error, i ...any) bool {
				return assert.ErrorContains(t, err, ErrNotFoundCouncilMember.Error())
			},
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(proposalID, test.Res)
				test.ErrHandler(t, err)
				return err
			})
		})
	}
}
