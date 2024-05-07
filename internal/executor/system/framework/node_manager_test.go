package framework

import (
	"fmt"
	"testing"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNodeManager_LifeCycleOfNode(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManagerContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManagerContract, epochManagerContract, nodeManagerContract, stakingManagerContract)

	nodes := make([]node_manager.NodeInfo, 5)
	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		totalNodeCount, err := nodeManagerContract.GetTotalNodeCount()
		assert.Nil(t, err)
		assert.EqualValues(t, 4, totalNodeCount)

		for i := uint64(0); i < 4; i++ {
			nodes[i], err = nodeManagerContract.GetNodeInfo(i + 1)
			assert.Nil(t, err)
			assert.EqualValues(t, i+1, nodes[i].ID)
			assert.EqualValues(t, fmt.Sprintf("node%d", i+1), nodes[i].MetaData.Name)
			assert.EqualValues(t, types.NodeStatusActive, nodes[i].Status)
		}

		activeValidatorSet, votingPowers, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, 4, len(activeValidatorSet))
		for i := 0; i < 4; i++ {
			assert.EqualValues(t, nodes[i], activeValidatorSet[i])
			assert.EqualValues(t, node_manager.ConsensusVotingPower{
				NodeID:               uint64(i + 1),
				ConsensusVotingPower: 1000,
			}, votingPowers[i])
		}
	})

	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	operatorAddress := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")

	// register permission denied
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		_, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           nodes[0].P2PID,
			OperatorAddress: operatorAddress.String(),
			MetaData: node_manager.NodeMetaData{
				Name:       "mockName",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.ErrorContains(t, err, ErrPermissionDenied.Error())
		return err
	})

	// error repeat register
	for i := 0; i < 4; i++ {
		testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
			_, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
				ConsensusPubKey: consensusKeystore.PublicKey.String(),
				P2PPubKey:       p2pKeystore.PublicKey.String(),
				P2PID:           nodes[i].P2PID,
				OperatorAddress: operatorAddress.String(),
				MetaData: node_manager.NodeMetaData{
					Name:       "mockName",
					Desc:       "mockDesc",
					ImageURL:   "https://example.com/image.png",
					WebsiteURL: "https://example.com/",
				},
			})
			assert.ErrorContains(t, err, "is already in use")
			return err
		}, common.TestNVMRunOptionCallFromSystem())
	}

	// register a node success
	var node5ID uint64
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		node5ID, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			OperatorAddress: operatorAddress.String(),
			MetaData: node_manager.NodeMetaData{
				Name:       "mockName",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.Nil(t, err)
		assert.EqualValues(t, 5, node5ID)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		nodes[4], err = nodeManagerContract.GetNodeInfo(5)
		assert.Nil(t, err)
		assert.Equal(t, node5ID, nodes[4].ID)
		assert.Equal(t, consensusKeystore.PublicKey.String(), nodes[4].ConsensusPubKey)
		assert.Equal(t, p2pKeystore.PublicKey.String(), nodes[4].P2PPubKey)
		assert.Equal(t, p2pKeystore.P2PID(), nodes[4].P2PID)
		assert.EqualValues(t, types.NodeStatusDataSyncer, nodes[4].Status)
		assert.Equal(t, "mockName", nodes[4].MetaData.Name)
		assert.Equal(t, operatorAddress, ethcommon.HexToAddress(nodes[4].OperatorAddress))

		dataSyncerSet, err := nodeManagerContract.GetDataSyncerSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[4]}, dataSyncerSet)
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.JoinCandidateSet(node5ID)
		assert.EqualError(t, err, ErrPermissionDenied.Error())
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, operatorAddress, func() error {
		err = nodeManagerContract.JoinCandidateSet(node5ID)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		dataSyncerSet, err := nodeManagerContract.GetDataSyncerSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{}, dataSyncerSet)

		candidateSet, err := nodeManagerContract.GetCandidateSet()
		assert.Nil(t, err)
		nodes[4].Status = uint8(types.NodeStatusCandidate)
		assert.Equal(t, []node_manager.NodeInfo{nodes[4]}, candidateSet)

		candidates, err := nodeManagerContract.InternalGetConsensusCandidateNodeIDs()
		assert.Nil(t, err)
		assert.Equal(t, []uint64{1, 2, 3, 4, 5}, candidates)
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.InternalUpdateActiveValidatorSet([]node_manager.ConsensusVotingPower{
			{
				NodeID:               1,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               2,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               3,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               5,
				ConsensusVotingPower: 100,
			},
		})
		assert.ErrorContains(t, err, ErrPermissionDenied.Error())
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.InternalUpdateActiveValidatorSet([]node_manager.ConsensusVotingPower{
			{
				NodeID:               1,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               2,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               3,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               5,
				ConsensusVotingPower: 100,
			},
		})
		assert.Nil(t, err)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		nodes[4].Status = uint8(types.NodeStatusActive)
		nodes[3].Status = uint8(types.NodeStatusCandidate)

		candidateSet, err := nodeManagerContract.GetCandidateSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[3]}, candidateSet)

		activeSet, returnVotingPowers, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[0], nodes[1], nodes[2], nodes[4]}, activeSet)
		assert.Equal(t, []node_manager.ConsensusVotingPower{
			{
				NodeID:               1,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               2,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               3,
				ConsensusVotingPower: 100,
			},
			{
				NodeID:               5,
				ConsensusVotingPower: 100,
			},
		}, returnVotingPowers)
	})

	testNVM.RunSingleTX(nodeManagerContract, operatorAddress, func() error {
		err = nodeManagerContract.LeaveValidatorOrCandidateSet(node5ID)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		nodes[4].Status = uint8(types.NodeStatusPendingInactive)
		activeSet, _, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[0], nodes[1], nodes[2], nodes[4]}, activeSet)

		pendingInactiveSet, err := nodeManagerContract.GetPendingInactiveSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[4]}, pendingInactiveSet)
	})

	testNVM.RunSingleTX(stakingManagerContract, types.NewAddressByStr(nodes[4].OperatorAddress).ETHAddress(), func() error {
		err = stakingManagerContract.CreateStakingPool(nodes[4].ID, 0)
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.InternalProcessNodeLeave()
		assert.Nil(t, err)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		pendingInactiveSet, err := nodeManagerContract.GetPendingInactiveSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{}, pendingInactiveSet)
		exitedSet, err := nodeManagerContract.GetExitedSet()
		assert.Nil(t, err)
		nodes[4].Status = uint8(types.NodeStatusExited)
		assert.EqualValues(t, []node_manager.NodeInfo{nodes[4]}, exitedSet)

		totalNodesNum, err := nodeManagerContract.GetTotalNodeCount()
		assert.Nil(t, err)
		assert.EqualValues(t, 5, totalNodesNum)
	})
}

func TestNodeManager_UpdateInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManagerContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManagerContract, epochManagerContract, nodeManagerContract, stakingManagerBuildContract)

	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	operatorAddress := "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"

	var node5ID uint64
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		node5ID, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			OperatorAddress: operatorAddress,
			MetaData: node_manager.NodeMetaData{
				Name:       "mockName",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.Nil(t, err)
		assert.EqualValues(t, 5, node5ID)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.UpdateOperator(node5ID, common.ZeroAddress)
		assert.EqualError(t, err, ErrPermissionDenied.Error())
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.HexToAddress(operatorAddress), func() error {
		err = nodeManagerContract.UpdateOperator(node5ID, common.ZeroAddress)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		info, err := nodeManagerContract.GetNodeInfo(node5ID)
		assert.Nil(t, err)
		assert.Equal(t, common.ZeroAddress, info.OperatorAddress)
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.HexToAddress(operatorAddress), func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{})
		assert.EqualError(t, err, ErrPermissionDenied.Error())
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "",
		})
		assert.ErrorContains(t, err, "name cannot be empty")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "node1",
		})
		assert.ErrorContains(t, err, "already in use")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "newName",
		})
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		info, err := nodeManagerContract.GetNodeInfo(node5ID)
		assert.Nil(t, err)
		assert.Equal(t, "newName", info.MetaData.Name)
	})
}

func TestCheckNodeInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManagerContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManagerContract, epochManagerContract, nodeManagerContract, stakingManagerBuildContract)

	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	operatorAddress := "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"

	var node5ID uint64
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		node5ID, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			OperatorAddress: operatorAddress,
			MetaData: node_manager.NodeMetaData{
				Name:       "mockName",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.Nil(t, err)
		assert.EqualValues(t, 5, node5ID)
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		nodeId, err := nodeManagerContract.GetNodeIDByP2PID(p2pKeystore.P2PID())
		assert.Nil(t, err)
		assert.EqualValues(t, 5, nodeId)

		exist := nodeManagerContract.ExistNodeByP2PID(p2pKeystore.P2PID())
		assert.True(t, exist)

		exist = nodeManagerContract.ExistNodeByConsensusPubKey(consensusKeystore.PublicKey.String())
		assert.True(t, exist)

		exist = nodeManagerContract.ExistNodeByName("node5")
		assert.False(t, exist)

		exist = nodeManagerContract.ExistNodeByName("mockName")
		assert.True(t, exist)

		nodes, err := nodeManagerContract.GetNodeInfos([]uint64{7})
		assert.ErrorContains(t, err, ErrNodeNotFound.Error())

		nodes, err = nodeManagerContract.GetNodeInfos([]uint64{1, 2, 3, 4, 5})
		assert.Nil(t, err)
		for i := 0; i < 5; i++ {
			assert.EqualValues(t, i+1, nodes[i].ID)
		}
	})
}
