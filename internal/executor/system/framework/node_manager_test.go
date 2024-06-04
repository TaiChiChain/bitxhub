package framework

import (
	"fmt"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func resetGenesis(t *testing.T) *repo.GenesisConfig {
	genesis := repo.MockRepo(t).GenesisConfig
	genesis.EpochInfo.StakeParams.MaxPendingInactiveValidatorRatio = 10000
	return genesis
}

func TestNodeManager_GenesisInit(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.MaxPendingInactiveValidatorRatio = 10000
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))

	// a data syncer with error stake number
	testNVM.Rep.GenesisConfig.Nodes[0].IsDataSyncer = true
	errNum := types.CoinNumber(*big.NewInt(-1))
	testNVM.Rep.GenesisConfig.Nodes[0].StakeNumber = &errNum
	err := nodeManagerContract.GenesisInit(testNVM.Rep.GenesisConfig)
	assert.ErrorContains(t, err, "invalid stake number")

	// a node with error address format
	genesis := resetGenesis(t)
	genesis.Nodes[0].OperatorAddress = "123"
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "invalid operator address")

	// a node with error commission rate
	genesis = resetGenesis(t)
	genesis.Nodes[0].CommissionRate = CommissionRateDenominator + 1
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "invalid commission rate")

	// stake number is larger than maximum
	genesis = resetGenesis(t)
	errNum = types.CoinNumber(*new(big.Int).Add(genesis.EpochInfo.StakeParams.MaxValidatorStake.ToBigInt(), big.NewInt(1)))
	genesis.Nodes[0].StakeNumber = &errNum
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "invalid stake number")

	// err consensus key
	genesis = resetGenesis(t)
	genesis.Nodes[0].ConsensusPubKey += "123"
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "failed to unmarshal consensus public key")

	// err p2p key
	genesis = resetGenesis(t)
	genesis.Nodes[0].P2PPubKey += "123"
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "failed to unmarshal p2p public key")

	// repeat consensus key
	genesis = resetGenesis(t)
	genesis.Nodes[1].ConsensusPubKey = genesis.Nodes[0].ConsensusPubKey
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "failed to register node")

	genesis.Nodes[0].IsDataSyncer = true
	genesis.Nodes[0].StakeNumber = types.CoinNumberByAxc(0)
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "failed to register node")

	genesis = resetGenesis(t)
	genesis.Nodes[1].ConsensusPubKey = genesis.Nodes[0].ConsensusPubKey
	genesis.EpochInfo.ConsensusParams.MaxValidatorNum = 1
	err = nodeManagerContract.GenesisInit(genesis)
	assert.ErrorContains(t, err, "failed to register node")
}

func TestNodeManager_LifeCycleOfNode(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.StakeParams.MaxPendingInactiveValidatorRatio = 10000
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManagerContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManagerContract, epochManagerContract, nodeManagerContract, stakingManagerContract)

	nodes := make([]node_manager.NodeInfo, 5)
	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		totalNodeCount, err := nodeManagerContract.GetTotalCount()
		assert.Nil(t, err)
		assert.EqualValues(t, 4, totalNodeCount)

		for i := uint64(0); i < 4; i++ {
			nodes[i], err = nodeManagerContract.GetInfo(i + 1)
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

	// error repeat register
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		_, err = nodeManagerContract.Register(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			Operator:        operatorAddress,
			MetaData: node_manager.NodeMetaData{
				Name:       "node1",
				Desc:       "mockDesc",
				ImageURL:   "https://example.com/image.png",
				WebsiteURL: "https://example.com/",
			},
		})
		assert.ErrorContains(t, err, "is already in use")
		return err
	}, common.TestNVMRunOptionCallFromSystem())

	// register a node success
	var node5ID uint64
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		node5ID, err = nodeManagerContract.Register(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			Operator:        operatorAddress,
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
		nodes[4], err = nodeManagerContract.GetInfo(5)
		assert.Nil(t, err)
		assert.Equal(t, node5ID, nodes[4].ID)
		assert.Equal(t, consensusKeystore.PublicKey.String(), nodes[4].ConsensusPubKey)
		assert.Equal(t, p2pKeystore.PublicKey.String(), nodes[4].P2PPubKey)
		assert.Equal(t, p2pKeystore.P2PID(), nodes[4].P2PID)
		assert.EqualValues(t, types.NodeStatusDataSyncer, nodes[4].Status)
		assert.Equal(t, "mockName", nodes[4].MetaData.Name)
		assert.Equal(t, operatorAddress, nodes[4].Operator)

		dataSyncerSet, err := nodeManagerContract.GetDataSyncerSet()
		assert.Nil(t, err)
		assert.Equal(t, []node_manager.NodeInfo{nodes[4]}, dataSyncerSet)
	})

	// exit request from arbitrary account will revert
	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		err = nodeManagerContract.Exit(node5ID)
		assert.ErrorContains(t, err, "no permission")
	})

	testNVM.Call(nodeManagerContract, operatorAddress, func() {
		node, err := nodeManagerContract.GetInfo(node5ID)
		assert.Nil(t, err)
		node.Status = uint8(types.NodeStatusExited)
		err = nodeManagerContract.nodeRegistry.Put(node5ID, node)
		assert.Nil(t, err)
		err = nodeManagerContract.Exit(node5ID)
		assert.ErrorContains(t, err, "IncorrectStatus")
	})

	testNVM.Call(nodeManagerContract, operatorAddress, func() {
		err = nodeManagerContract.Exit(node5ID)
		assert.Nil(t, err)
		exitedSet, err := nodeManagerContract.GetExitedSet()
		assert.Nil(t, err)
		assert.Equal(t, nodes[4].ID, exitedSet[0].ID)
		assert.Equal(t, 1, len(exitedSet))
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.JoinCandidateSet(node5ID, 0)
		assert.ErrorContains(t, err, "no permission")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, operatorAddress, func() error {
		err = nodeManagerContract.JoinCandidateSet(node5ID, 0)
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

		candidates, err := nodeManagerContract.getConsensusCandidateNodeIDs()
		assert.Nil(t, err)
		assert.Equal(t, []uint64{1, 2, 3, 4, 5}, candidates)
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		err = nodeManagerContract.updateActiveValidatorSet([]node_manager.ConsensusVotingPower{
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

	testNVM.Call(nodeManagerContract, operatorAddress, func() {
		curEpoch, err := epochManagerContract.CurrentEpoch()
		assert.Nil(t, err)
		curEpoch.StakeParams.MaxPendingInactiveValidatorRatio = 0
		err = epochManagerContract.UpdateNextEpoch(curEpoch)
		assert.Nil(t, err)
		_, err = epochManagerContract.TurnIntoNewEpoch()
		assert.Nil(t, err)
		err = nodeManagerContract.Exit(node5ID)
		assert.ErrorContains(t, err, "PendingInactiveSetIsFull")
	})

	testNVM.RunSingleTX(nodeManagerContract, operatorAddress, func() error {
		err = nodeManagerContract.Exit(node5ID)
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

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.Address{}, func() error {
		exitedIDs, err := nodeManagerContract.processNodeLeave()
		assert.Nil(t, err)
		assert.EqualValues(t, []uint64{5}, exitedIDs)
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

		totalNodesNum, err := nodeManagerContract.GetTotalCount()
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
		node5ID, err = nodeManagerContract.Register(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			Operator:        ethcommon.HexToAddress(operatorAddress),
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

	zeroAddr := ethcommon.Address{}
	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		err = nodeManagerContract.UpdateOperator(node5ID, zeroAddr)
		assert.ErrorContains(t, err, "no permission")

		nodeManagerContract.Ctx.From = ethcommon.HexToAddress(operatorAddress)
		err = nodeManagerContract.UpdateOperator(node5ID, zeroAddr)
		assert.ErrorContains(t, err, "new operator cannot be zero address")
	})

	newOperator := ethcommon.HexToAddress("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D012")
	testNVM.RunSingleTX(nodeManagerContract, ethcommon.HexToAddress(operatorAddress), func() error {
		err = nodeManagerContract.UpdateOperator(node5ID, newOperator)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		info, err := nodeManagerContract.GetInfo(node5ID)
		assert.Nil(t, err)
		assert.Equal(t, newOperator, info.Operator)
	})

	testNVM.RunSingleTX(nodeManagerContract, ethcommon.HexToAddress(operatorAddress), func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{})
		assert.ErrorContains(t, err, "no permission")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, newOperator, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "",
		})
		assert.ErrorContains(t, err, "name cannot be empty")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, newOperator, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "node1",
		})
		assert.ErrorContains(t, err, "already in use")
		return err
	})

	testNVM.RunSingleTX(nodeManagerContract, newOperator, func() error {
		err = nodeManagerContract.UpdateMetaData(node5ID, node_manager.NodeMetaData{
			Name: "newName",
		})
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(nodeManagerContract, ethcommon.Address{}, func() {
		info, err := nodeManagerContract.GetInfo(node5ID)
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
		node5ID, err = nodeManagerContract.Register(node_manager.NodeInfo{
			ConsensusPubKey: consensusKeystore.PublicKey.String(),
			P2PPubKey:       p2pKeystore.PublicKey.String(),
			P2PID:           p2pKeystore.P2PID(),
			Operator:        ethcommon.HexToAddress(operatorAddress),
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

		nodes, err := nodeManagerContract.GetInfos([]uint64{7})
		assert.ErrorContains(t, err, ErrNodeNotFound.Error())

		nodes, err = nodeManagerContract.GetInfos([]uint64{1, 2, 3, 4, 5})
		assert.Nil(t, err)
		for i := 0; i < 5; i++ {
			assert.EqualValues(t, i+1, nodes[i].ID)
		}
	})
}

func TestNodeManager_TurnIntoNewEpoch(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	epochManagerContract := EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	nodeManagerContract := NodeManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	stakingManagerBuildContract := StakingManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	axcManagerContract := token.AXCBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(axcManagerContract, epochManagerContract, nodeManagerContract, stakingManagerBuildContract)

	newEpoch := repo.GenesisEpochInfo()
	err := nodeManagerContract.TurnIntoNewEpoch(nil, newEpoch)
	assert.ErrorContains(t, err, "not enough validators")

	newEpoch.ConsensusParams.MaxValidatorNum = 3
	newEpoch.StakeParams.MinValidatorStake = types.CoinNumberByAxc(1)
	err = nodeManagerContract.TurnIntoNewEpoch(nil, newEpoch)
	assert.ErrorContains(t, err, "not enough validators")

	newEpoch.ConsensusParams.MinValidatorNum = 3
	err = nodeManagerContract.TurnIntoNewEpoch(nil, newEpoch)
	assert.Nil(t, err)
	nodes, _, err := nodeManagerContract.GetActiveValidatorSet()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(nodes))
}
