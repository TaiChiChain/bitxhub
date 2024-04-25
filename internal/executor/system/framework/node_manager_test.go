package framework

import (
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNodeManager_LifeCycleOfNode(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(epochManagerContract, nodeManagerContract, stakingManagerContract)

	var node1, node2, node3, node4, node5 node_manager.NodeInfo
	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		totalNodeCount, err := nodeManagerContract.GetTotalNodeCount()
		assert.Nil(t, err)
		assert.EqualValues(t, 4, totalNodeCount)

		node1, err = nodeManagerContract.GetNodeInfo(1)
		assert.Nil(t, err)
		node2, err = nodeManagerContract.GetNodeInfo(2)
		assert.Nil(t, err)
		node3, err = nodeManagerContract.GetNodeInfo(3)
		assert.Nil(t, err)
		node4, err = nodeManagerContract.GetNodeInfo(4)
		assert.Nil(t, err)

		activeValidatorSet, votingPowers, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, 4, len(activeValidatorSet))
		assert.EqualValues(t, node1, activeValidatorSet[0])
		assert.EqualValues(t, node_manager.ConsensusVotingPower{
			NodeID:               1,
			ConsensusVotingPower: 1000,
		}, votingPowers[0])
		assert.EqualValues(t, node2, activeValidatorSet[1])
		assert.EqualValues(t, node_manager.ConsensusVotingPower{
			NodeID:               2,
			ConsensusVotingPower: 1000,
		}, votingPowers[1])
		assert.EqualValues(t, node3, activeValidatorSet[2])
		assert.EqualValues(t, node_manager.ConsensusVotingPower{
			NodeID:               3,
			ConsensusVotingPower: 1000,
		}, votingPowers[2])
		assert.EqualValues(t, node4, activeValidatorSet[3])
		assert.EqualValues(t, node_manager.ConsensusVotingPower{
			NodeID:               4,
			ConsensusVotingPower: 1000,
		}, votingPowers[3])
	})

	p2pKeystore, err := repo.GenerateP2PKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	consensusKeystore, err := repo.GenerateConsensusKeystore(testNVM.Rep.RepoRoot, "", "")
	assert.Nil(t, err)
	operatorAddress := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013")

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		_, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           node1.P2PID,
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
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		_, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: node1.ConsensusPubKey,
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
		assert.ErrorContains(t, err, "is already in use")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		_, err = nodeManagerContract.InternalRegisterNode(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			OperatorAddress: operatorAddress.String(),
			MetaData: node_manager.NodeMetaData{
				Name:       node1.MetaData.Name,
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.ErrorContains(t, err, "is already in use")
		return err
	})

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
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		node5, err = nodeManagerContract.GetNodeInfo(5)
		assert.Nil(t, err)
		assert.Equal(t, node5ID, node5.ID)
		assert.Equal(t, consensusKeystore.PublicKey.String(), node5.ConsensusPubKey)
		assert.Equal(t, p2pKeystore.PublicKey.String(), node5.P2PPubKey)
		assert.Equal(t, p2pKeystore.P2PID(), node5.P2PID)
		assert.EqualValues(t, types.NodeStatusDataSyncer, node5.Status)
		assert.Equal(t, "mockName", node5.MetaData.Name)
		assert.Equal(t, operatorAddress, ethcommon.HexToAddress(node5.OperatorAddress))

		dataSyncerSet, err := nodeManagerContract.GetDataSyncerSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{node5}, dataSyncerSet)
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
		node5.Status = uint8(types.NodeStatusCandidate)
		assert.Equal(t, []node_manager.NodeInfo{node5}, candidateSet)

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
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		node5.Status = uint8(types.NodeStatusActive)
		node4.Status = uint8(types.NodeStatusCandidate)

		candidateSet, err := nodeManagerContract.GetCandidateSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{node4}, candidateSet)

		activeSet, returnVotingPowers, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{node1, node2, node3, node5}, activeSet)
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
		node5.Status = uint8(types.NodeStatusPendingInactive)
		activeSet, _, err := nodeManagerContract.GetActiveValidatorSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{node1, node2, node3, node5}, activeSet)

		pendingInactiveSet, err := nodeManagerContract.GetPendingInactiveSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{node5}, pendingInactiveSet)
	})

	testNVM.RunSingleTX(stakingManagerContract, ethcommon.Address{}, func() error {
		err = stakingManagerContract.CreateStakingPool(node5.ID, 0)
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
		node5.Status = uint8(types.NodeStatusExited)
		assert.EqualValues(t, []node_manager.NodeInfo{node5}, exitedSet)

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
	testNVM.GenesisInit(epochManagerContract, nodeManagerContract, stakingManagerBuildContract)

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
	})

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
	type args struct {
		nodeInfo node_manager.NodeInfo
	}
	tests := []struct {
		name                string
		args                args
		wantConsensusPubKey *crypto.Bls12381PublicKey
		wantP2pPubKey       *crypto.Ed25519PublicKey
		wantP2pID           string
		wantErr             assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConsensusPubKey, gotP2pPubKey, gotP2pID, err := CheckNodeInfo(tt.args.nodeInfo)
			if !tt.wantErr(t, err, fmt.Sprintf("CheckNodeInfo(%v)", tt.args.nodeInfo)) {
				return
			}
			assert.Equalf(t, tt.wantConsensusPubKey, gotConsensusPubKey, "CheckNodeInfo(%v)", tt.args.nodeInfo)
			assert.Equalf(t, tt.wantP2pPubKey, gotP2pPubKey, "CheckNodeInfo(%v)", tt.args.nodeInfo)
			assert.Equalf(t, tt.wantP2pID, gotP2pID, "CheckNodeInfo(%v)", tt.args.nodeInfo)
		})
	}
}
