package framework

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework/solidity/node_manager_client"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	nextNodeIDStorageKey        = "nextNodeID"
	nodeRegistry                = "nodeRegistry"
	nodeP2PIDIndex              = "nodeP2PIDIndex"
	nodeConsensusPubKeyIndex    = "nodeConsensusPubKeyIndex"
	nodeNameIndex               = "nodeNameIndex"
	nextEpochUpdateNodes        = "nextEpochUpdateNodes"
	nextEpochUpdatedNodeIDSet   = "nextEpochUpdatedNodeIDSet"
	statusPrefix                = "status"
	activeValidatorVotingPowers = "activeValidatorVotingPowers"
)

var (
	ErrNodeStatusIncorrect     = errors.New("node status is incorrect")
	ErrNodeNotFound            = errors.New("node not found")
	ErrStatusSetNotFound       = errors.New("status set not found")
	ErrNodeNotFoundInStatusSet = errors.New("node not found in status set")
	ErrNodeAlreadyInStatusSet  = errors.New("node is already in status set")
	ErrPermissionDenied        = errors.New("permission denied")
)

var NodeManagerBuildConfig = &common.SystemContractBuildConfig[*NodeManager]{
	Name:    "framework_node_manager",
	Address: common.NodeManagerContractAddr,
	AbiStr:  node_manager_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *NodeManager {
		return &NodeManager{
			SystemContractBase: systemContractBase,
		}
	},
}

type NewNodeInfo struct {
	// Operator address, with permission to manage node (can update)
	OperatorAddress string

	// Meta data (can update)
	MetaData types.NodeMetaData
}

type ConsensusVotingPower struct {
	NodeID uint64

	// Consensus voting weight (calculated by active stake).
	// TODO: convert type to big.Int
	ConsensusVotingPower int64
}

type NodeManager struct {
	common.SystemContractBase

	nextNodeID *common.VMSlot[uint64]

	// store all registered node info (including exited nodes)
	nodeRegistry *common.VMMap[uint64, types.NodeInfo]

	// p2p id to node id
	nodeP2PIDIndex *common.VMMap[string, uint64]

	// node.ConsensusPubKey -> NodeID
	nodeConsensusPubKeyIndex *common.VMMap[string, uint64]

	// node.MetaData.Name -> NodeID
	nodeNameIndex *common.VMMap[string, uint64]

	// track all updated node id in the next epoch, will clean up in the next epoch
	nextEpochUpdatedNodeIDSet *common.VMSlot[[]uint64]

	// store all new node info in the next epoch, will clean up in the next epoch
	nextEpochUpdateNodes *common.VMMap[uint64, types.NodeInfo]

	// including Candidate and ActiveValidator
	activeValidatorVotingPowers *common.VMSlot[[]ConsensusVotingPower]
}

func (n *NodeManager) GenesisInit(genesis *repo.GenesisConfig) error {
	var activeValidators []uint64
	var pendingValidators []uint64
	var candidates []uint64
	var dataSyncers []uint64

	var needCreateStakingPoolIDs []uint64

	minValidatorStake := genesis.EpochInfo.StakeParams.MinValidatorStake.ToBigInt()
	// check node info
	nodeInfoMap := make(map[uint64]*types.NodeInfo, len(genesis.Nodes))
	nodeCfgMap := make(map[uint64]*repo.GenesisNodeInfo, len(genesis.Nodes))
	nodeStakeNumberMap := make(map[uint64]*big.Int, len(genesis.Nodes))
	for i, nodeCfg := range genesis.Nodes {
		nodeInfo := &types.NodeInfo{
			ID:              uint64(i + 1),
			ConsensusPubKey: nodeCfg.ConsensusPubKey,
			P2PPubKey:       nodeCfg.P2PPubKey,
			OperatorAddress: nodeCfg.OperatorAddress,
			MetaData: types.NodeMetaData{
				Name:       nodeCfg.MetaData.Name,
				Desc:       nodeCfg.MetaData.Desc,
				ImageURL:   nodeCfg.MetaData.ImageURL,
				WebsiteURL: nodeCfg.MetaData.WebsiteURL,
			},
		}

		err := func() error {
			if nodeCfg.CommissionRate > CommissionRateDenominator {
				return errors.Errorf("invalid commission rate %d, need <= %d", nodeCfg.CommissionRate, CommissionRateDenominator)
			}
			stakeNumber := nodeCfg.StakeNumber.ToBigInt()
			nodeStakeNumberMap[nodeInfo.ID] = stakeNumber

			_, _, p2pID, err := CheckNodeInfo(*nodeInfo)
			if err != nil {
				return err
			}
			nodeInfo.P2PID = p2pID
			return nil
		}()
		if err != nil {
			return errors.Wrapf(err, "failed to check genesis node %d", nodeInfo.ID)
		}

		nodeInfoMap[nodeInfo.ID] = nodeInfo
		nodeCfgMap[nodeInfo.ID] = &nodeCfg
		if nodeCfg.IsDataSyncer {
			dataSyncers = append(dataSyncers, nodeInfo.ID)
		} else {
			if minValidatorStake.Cmp(nodeStakeNumberMap[nodeInfo.ID]) >= 0 {
				pendingValidators = append(pendingValidators, nodeInfo.ID)
			} else {
				candidates = append(candidates, nodeInfo.ID)
			}

			needCreateStakingPoolIDs = append(needCreateStakingPoolIDs, nodeInfo.ID)
		}
	}

	// select active validators
	sort.Slice(pendingValidators, func(i, j int) bool {
		cmpRes := nodeStakeNumberMap[pendingValidators[i]].Cmp(nodeStakeNumberMap[pendingValidators[j]])
		if cmpRes == 0 {
			return nodeInfoMap[pendingValidators[i]].ID < nodeInfoMap[pendingValidators[j]].ID
		}
		return cmpRes == 1
	})

	if uint64(len(pendingValidators)) <= genesis.EpochInfo.ConsensusParams.MaxValidatorNum {
		activeValidators = append(activeValidators, pendingValidators...)
	} else {
		activeValidators = append(activeValidators, pendingValidators[:genesis.EpochInfo.ConsensusParams.MaxValidatorNum]...)
		candidates = append(candidates, pendingValidators[genesis.EpochInfo.ConsensusParams.MaxValidatorNum:]...)
	}

	// sort candidates
	sort.Slice(candidates, func(i, j int) bool {
		cmpRes := nodeStakeNumberMap[candidates[i]].Cmp(nodeStakeNumberMap[candidates[j]])
		if cmpRes == 0 {
			return nodeInfoMap[candidates[i]].ID < nodeInfoMap[candidates[j]].ID
		}
		return cmpRes == 1
	})
	// sort dataSyncers
	sort.Slice(dataSyncers, func(i, j int) bool {
		return dataSyncers[i] < dataSyncers[j]
	})

	// register node info
	registerNodesFn := func(nodeIDs []uint64, nodeStatus types.NodeStatus) error {
		for _, nodeID := range nodeIDs {
			nodeInfo := nodeInfoMap[nodeID]
			nodeInfo.Status = nodeStatus
			if _, err := n.registerNode(*nodeInfo, true); err != nil {
				return errors.Wrapf(err, "failed to register node %d", nodeID)
			}
		}
		return nil
	}
	if err := registerNodesFn(activeValidators, types.NodeStatusActive); err != nil {
		return err
	}
	if err := registerNodesFn(candidates, types.NodeStatusCandidate); err != nil {
		return err
	}
	if err := registerNodesFn(dataSyncers, types.NodeStatusDataSyncer); err != nil {
		return err
	}

	// update others
	if err := n.nextNodeID.Put(uint64(len(genesis.Nodes) + 1)); err != nil {
		return err
	}

	if err := n.getStatusSet(types.NodeStatusActive).Put(activeValidators); err != nil {
		return err
	}
	if err := n.getStatusSet(types.NodeStatusCandidate).Put(candidates); err != nil {
		return err
	}
	if err := n.getStatusSet(types.NodeStatusDataSyncer).Put(dataSyncers); err != nil {
		return err
	}

	stakingManagerContract := StakingManagerBuildConfig.Build(n.Ctx)
	liquidStakingTokenContract := LiquidStakingTokenBuildConfig.Build(n.Ctx)

	// init staking pools
	for _, nodeID := range needCreateStakingPoolIDs {
		nodeCfg := nodeCfgMap[nodeID]
		nodeInfo := nodeInfoMap[nodeID]
		stakeNumber := nodeStakeNumberMap[nodeID]
		zero := big.NewInt(0)

		liquidStakingTokenID, err := liquidStakingTokenContract.InternalMint(ethcommon.HexToAddress(nodeInfo.OperatorAddress), &LiquidStakingTokenInfo{
			PoolID:           nodeID,
			Principal:        stakeNumber,
			Unlocked:         zero,
			ActiveEpoch:      genesis.EpochInfo.Epoch,
			UnlockingRecords: []UnlockingRecord{},
		})
		if err != nil {
			return errors.Wrapf(err, "failed to mint liquid staking token for node %d", nodeID)
		}
		if err := stakingManagerContract.GetStakingPool(nodeID).createByStakingPool(&StakingPoolInfo{
			PoolID:                       nodeID,
			IsActive:                     true,
			ActiveStake:                  stakeNumber,
			TotalLiquidStakingToken:      stakeNumber,
			PendingActiveStake:           zero,
			PendingInactiveStake:         zero,
			CommissionRate:               nodeCfg.CommissionRate,
			NextEpochCommissionRate:      nodeCfg.CommissionRate,
			CumulativeReward:             zero,
			OperatorLiquidStakingTokenID: liquidStakingTokenID,
		}); err != nil {
			return errors.Wrapf(err, "failed to create staking pool for node %d", nodeID)
		}
	}

	if err := stakingManagerContract.InternalInitAvailableStakingPools(needCreateStakingPoolIDs); err != nil {
		return err
	}
	axcUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	activeValidatorVotingPowers := lo.Map(activeValidators, func(item uint64, index int) ConsensusVotingPower {
		stakeNumber := nodeStakeNumberMap[item]
		// convert unit `mol` to `axc`
		standardizedStakeNumber := stakeNumber.Div(stakeNumber, axcUnit)
		return ConsensusVotingPower{
			NodeID:               item,
			ConsensusVotingPower: standardizedStakeNumber.Int64(),
		}
	})
	if err := n.activeValidatorVotingPowers.Put(activeValidatorVotingPowers); err != nil {
		return err
	}

	return nil
}

func (n *NodeManager) SetContext(ctx *common.VMContext) {
	n.SystemContractBase.SetContext(ctx)

	n.nextNodeID = common.NewVMSlot[uint64](n.StateAccount, nextNodeIDStorageKey)
	n.nodeRegistry = common.NewVMMap[uint64, types.NodeInfo](n.StateAccount, nodeRegistry, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	n.nodeP2PIDIndex = common.NewVMMap[string, uint64](n.StateAccount, nodeP2PIDIndex, func(key string) string {
		return key
	})
	n.nodeConsensusPubKeyIndex = common.NewVMMap[string, uint64](n.StateAccount, nodeConsensusPubKeyIndex, func(key string) string {
		return key
	})
	n.nodeNameIndex = common.NewVMMap[string, uint64](n.StateAccount, nodeNameIndex, func(key string) string {
		return key
	})

	n.nextEpochUpdateNodes = common.NewVMMap[uint64, types.NodeInfo](n.StateAccount, nextEpochUpdateNodes, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	n.nextEpochUpdatedNodeIDSet = common.NewVMSlot[[]uint64](n.StateAccount, nextEpochUpdatedNodeIDSet)
	n.activeValidatorVotingPowers = common.NewVMSlot[[]ConsensusVotingPower](n.StateAccount, activeValidatorVotingPowers)
}

func (n *NodeManager) InternalRegisterNode(info types.NodeInfo) (id uint64, err error) {
	info.Status = types.NodeStatusDataSyncer
	return n.registerNode(info, false)
}

func (n *NodeManager) registerNode(info types.NodeInfo, isGenesisInit bool) (id uint64, err error) {
	info = StandardizationNodeInfo(info)
	if n.nodeP2PIDIndex.Has(info.P2PID) {
		return 0, errors.Errorf("p2p id %s is already in use", info.P2PID)
	}
	if n.nodeConsensusPubKeyIndex.Has(info.ConsensusPubKey) {
		return 0, errors.Errorf("consensus public key %s is already in use", info.ConsensusPubKey)
	}

	if n.nodeNameIndex.Has(info.MetaData.Name) {
		return 0, errors.Errorf("name %s is already in use", info.MetaData.Name)
	}

	if !isGenesisInit {
		exist, id, err := n.nextNodeID.Get()
		if err != nil {
			return 0, err
		}
		if !exist {
			id = 1
		}
		info.ID = id
		info.Status = types.NodeStatusDataSyncer

		if err = n.nextNodeID.Put(id + 1); err != nil {
			return 0, err
		}

		statusIDSet := n.getStatusSet(info.Status)
		isExist, set, err := statusIDSet.Get()
		if err != nil {
			return 0, err
		}
		if !isExist {
			set = []uint64{}
		}
		if err = statusIDSet.Put(append(set, id)); err != nil {
			return 0, err
		}
	}

	if err := n.nodeRegistry.Put(info.ID, info); err != nil {
		return 0, err
	}
	if err := n.nodeP2PIDIndex.Put(info.P2PID, info.ID); err != nil {
		return 0, err
	}
	if err := n.nodeConsensusPubKeyIndex.Put(info.ConsensusPubKey, info.ID); err != nil {
		return 0, err
	}
	if err := n.nodeNameIndex.Put(info.MetaData.Name, info.ID); err != nil {
		return 0, err
	}
	return info.ID, nil
}

func (n *NodeManager) InternalProcessNodeLeave() error {
	pendingInactiveSlot := n.getStatusSet(types.NodeStatusPendingInactive)
	// get pending inactive sets
	isExist, nodeIDs, err := pendingInactiveSlot.Get()
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}
	for _, id := range nodeIDs {
		// get the node
		info, err := n.GetNodeInfo(id)
		if err != nil {
			return err
		}
		// modify its status
		info.Status = types.NodeStatusExited
		// put it back
		if err = n.nodeRegistry.Put(id, info); err != nil {
			return err
		}
	}
	// get exited nodes set
	exitedSlot := n.getStatusSet(types.NodeStatusExited)
	isExist, exitedNodes, err := exitedSlot.Get()
	if err != nil {
		return err
	}
	if !isExist {
		exitedNodes = []uint64{}
	}
	// append all pending inactive nodes to exited nodes
	exitedNodes = append(exitedNodes, nodeIDs...)
	// clean up pending inactive set
	if err = pendingInactiveSlot.Delete(); err != nil {
		return err
	}
	// put exited nodes set
	return exitedSlot.Put(exitedNodes)
}

func (n *NodeManager) InternalGetConsensusCandidateNodeIDs() ([]uint64, error) {
	isExist, candidateIDs, err := n.getStatusSet(types.NodeStatusCandidate).Get()
	if err != nil {
		return nil, err
	}
	if !isExist {
		candidateIDs = []uint64{}
	}
	isExist, validateIDs, err := n.getStatusSet(types.NodeStatusActive).Get()
	if err != nil {
		return nil, err
	}
	if !isExist {
		validateIDs = []uint64{}
	}
	return append(candidateIDs, validateIDs...), nil
}

func (n *NodeManager) InternalUpdateActiveValidatorSet(ActiveValidatorVotingPowers []ConsensusVotingPower) error {
	allValidaNodes, err := n.InternalGetConsensusCandidateNodeIDs()
	if err != nil {
		return err
	}
	validatorMap := make(map[uint64]struct{}, len(ActiveValidatorVotingPowers))
	validators := make([]uint64, 0, len(ActiveValidatorVotingPowers))
	candidates := make([]uint64, 0, len(allValidaNodes)-len(ActiveValidatorVotingPowers))
	for _, votingPower := range ActiveValidatorVotingPowers {
		nodeId := votingPower.NodeID
		isExist, nodeInfo, err := n.nodeRegistry.Get(nodeId)
		if err != nil {
			return err
		}
		if !isExist {
			return ErrNodeNotFound
		}
		nodeInfo.Status = types.NodeStatusActive
		if err = n.nodeRegistry.Put(nodeId, nodeInfo); err != nil {
			return err
		}
		if _, ok := validatorMap[nodeId]; ok {
			return errors.New("duplicated node id")
		}
		validatorMap[nodeId] = struct{}{}
		validators = append(validators, nodeId)
	}
	for _, nodeId := range allValidaNodes {
		if _, ok := validatorMap[nodeId]; !ok {
			candidates = append(candidates, nodeId)
		}
	}
	if err = n.getStatusSet(types.NodeStatusCandidate).Put(candidates); err != nil {
		return err
	}
	if err = n.getStatusSet(types.NodeStatusActive).Put(validators); err != nil {
		return err
	}

	return n.activeValidatorVotingPowers.Put(ActiveValidatorVotingPowers)
}

func (n *NodeManager) JoinCandidateSet(nodeID uint64) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	// check permission
	if n.Ctx.From != ethcommon.HexToAddress(nodeInfo.OperatorAddress) {
		return ErrPermissionDenied
	}
	if nodeInfo.Status != types.NodeStatusDataSyncer {
		return ErrNodeStatusIncorrect
	}
	// TODO: support it
	panic("not support yet")
}

func (n *NodeManager) LeaveValidatorSet(nodeID uint64) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	// governance has permission
	if !n.Ctx.CallFromSystem {
		// check permission
		if n.Ctx.From != ethcommon.HexToAddress(nodeInfo.OperatorAddress) {
			return ErrPermissionDenied
		}
	}
	// TODO: support it
	panic("not support yet")
}

func (n *NodeManager) UpdateMetaData(nodeID uint64, metaData types.NodeMetaData) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}

	// check permission
	if n.Ctx.From != ethcommon.HexToAddress(nodeInfo.OperatorAddress) {
		return ErrPermissionDenied
	}

	newNodeInfo := NewNodeInfo{
		OperatorAddress: nodeInfo.OperatorAddress,
		MetaData:        metaData,
	}

	// rebuild name index
	if err := n.nodeNameIndex.Delete(nodeInfo.MetaData.Name); err != nil {
		return err
	}
	if err := n.nodeNameIndex.Put(newNodeInfo.MetaData.Name, nodeID); err != nil {
		return err
	}
	return n.updateNodeInfo(nodeID, nodeInfo, newNodeInfo)
}

func (n *NodeManager) UpdateOperator(nodeID uint64, newOperatorAddress string) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	// check permission
	if n.Ctx.From != ethcommon.HexToAddress(nodeInfo.OperatorAddress) {
		return ErrPermissionDenied
	}
	newNodeInfo := NewNodeInfo{
		OperatorAddress: newOperatorAddress,
		MetaData:        nodeInfo.MetaData,
	}
	return n.updateNodeInfo(nodeID, nodeInfo, newNodeInfo)
}

func (n *NodeManager) updateNodeInfo(nodeID uint64, oldNodeInfo types.NodeInfo, newNodeInfo NewNodeInfo) error {
	if n.Ctx.From.String() != oldNodeInfo.OperatorAddress {
		return ErrPermissionDenied
	}
	oldNodeInfo.OperatorAddress = newNodeInfo.OperatorAddress
	oldNodeInfo.MetaData = newNodeInfo.MetaData
	return n.nodeRegistry.Put(nodeID, oldNodeInfo)
}

func (n *NodeManager) GetNodeInfo(nodeID uint64) (info types.NodeInfo, err error) {
	isExist, nodeInfo, err := n.nodeRegistry.Get(nodeID)
	if err != nil {
		return types.NodeInfo{}, err
	}
	if !isExist {
		return types.NodeInfo{}, ErrNodeNotFound
	}
	return nodeInfo, nil
}

func (n *NodeManager) GetNodeIDByP2PID(p2pID string) (uint64, error) {
	isExist, nodeID, err := n.nodeP2PIDIndex.Get(p2pID)
	if err != nil {
		return 0, err
	}
	if !isExist {
		return 0, ErrNodeNotFound
	}
	return nodeID, nil
}

func (n *NodeManager) ExistNodeByP2PID(p2pID string) bool {
	return n.nodeP2PIDIndex.Has(p2pID)
}

func (n *NodeManager) ExistNodeByConsensusPubKey(consensusPubKey string) bool {
	return n.nodeConsensusPubKeyIndex.Has(hexutil.Encode(hexutil.Decode(consensusPubKey)))
}

func (n *NodeManager) ExistNodeByName(nodeName string) bool {
	return n.nodeNameIndex.Has(nodeName)
}

func (n *NodeManager) GetTotalNodeCount() (int, error) {
	_, id, err := n.nextNodeID.Get()
	if err != nil {
		return 0, err
	}
	return int(id - 1), nil
}

func (n *NodeManager) GetNodeInfos(nodeIDs []uint64) (infos []types.NodeInfo, err error) {
	for _, id := range nodeIDs {
		info, err := n.GetNodeInfo(id)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return
}

func (n *NodeManager) GetActiveValidatorSet() (infos []types.NodeInfo, votingPowers []ConsensusVotingPower, err error) {
	isExists, nodes, err := n.getStatusSet(types.NodeStatusActive).Get()
	if err != nil {
		return nil, nil, err
	}
	if !isExists {
		return nil, nil, nil
	}
	infos, err = n.GetNodeInfos(nodes)
	if err != nil {
		return nil, nil, err
	}
	isExists, votingPowers, err = n.activeValidatorVotingPowers.Get()
	if err != nil {
		return nil, nil, err
	}
	// if voting powers not exists but active validators exists, it is a bug
	// if voting powers are not as long as active validators, it is a bug
	if !isExists && !(len(votingPowers) == len(nodes)) {
		return nil, nil, errors.New("active validators and voting powers length not match")
	}
	return infos, votingPowers, nil
}

func (n *NodeManager) GetDataSyncerSet() (infos []types.NodeInfo, err error) {
	isExists, nodes, err := n.getStatusSet(types.NodeStatusDataSyncer).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *NodeManager) GetCandidateSet() (infos []types.NodeInfo, err error) {
	isExists, nodes, err := n.getStatusSet(types.NodeStatusCandidate).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *NodeManager) GetPendingInactiveSet() (infos []types.NodeInfo, err error) {
	isExists, nodes, err := n.getStatusSet(types.NodeStatusPendingInactive).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *NodeManager) GetExitedSet() (infos []types.NodeInfo, err error) {
	isExists, nodes, err := n.getStatusSet(types.NodeStatusExited).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *NodeManager) GetActiveValidatorVotingPowers() (votingPowers []ConsensusVotingPower, err error) {
	votingPowers, err = n.activeValidatorVotingPowers.MustGet()
	if err != nil {
		return nil, err
	}
	return votingPowers, nil
}

func (n *NodeManager) operatorTransferStatus(from types.NodeStatus, nodeID uint64) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	if nodeInfo.OperatorAddress != n.Ctx.From.String() {
		return ErrPermissionDenied
	}
	if nodeInfo.Status != from {
		return ErrNodeStatusIncorrect
	}
	nodeInfo.Status = from + 1
	if err = n.nodeRegistry.Put(nodeID, nodeInfo); err != nil {
		return err
	}
	fromSlot := n.getStatusSet(from)
	isExist, fromSets, err := fromSlot.Get()
	if err != nil {
		return err
	}
	if !isExist {
		return ErrStatusSetNotFound
	}
	index := -1
	for i, id := range fromSets {
		if id == nodeID {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrNodeNotFoundInStatusSet
	}
	fromSets = append(fromSets[:index], fromSets[index+1:]...)
	if err = fromSlot.Put(fromSets); err != nil {
		return err
	}
	toSlot := n.getStatusSet(from + 1)
	isExist, toSets, err := toSlot.Get()
	if err != nil {
		return err
	}
	if !isExist {
		toSets = []uint64{}
	}
	for _, id := range toSets {
		if id == nodeID {
			return ErrNodeAlreadyInStatusSet
		}
	}
	return toSlot.Put(append(toSets, nodeID))
}

func (n *NodeManager) getStatusSet(status types.NodeStatus) *common.VMSlot[[]uint64] {
	return common.NewVMSlot[[]uint64](n.StateAccount, statusPrefix+strconv.Itoa(int(status)))
}

func StandardizationNodeInfo(info types.NodeInfo) types.NodeInfo {
	return types.NodeInfo{
		ID:              info.ID,
		ConsensusPubKey: hexutil.Encode(hexutil.Decode(info.ConsensusPubKey)),
		P2PPubKey:       hexutil.Encode(hexutil.Decode(info.P2PPubKey)),
		P2PID:           info.P2PID,
		OperatorAddress: hexutil.Encode(hexutil.Decode(info.OperatorAddress)),
		MetaData: types.NodeMetaData{
			Name:       info.MetaData.Name,
			Desc:       info.MetaData.Desc,
			ImageURL:   info.MetaData.ImageURL,
			WebsiteURL: info.MetaData.WebsiteURL,
		},
		Status: info.Status,
	}
}

func CheckNodeInfo(nodeInfo types.NodeInfo) (consensusPubKey *crypto.Bls12381PublicKey, p2pPubKey *crypto.Ed25519PublicKey, p2pID string, err error) {
	if nodeInfo.MetaData.Name == "" {
		return nil, nil, "", errors.New("node name cannot be empty")
	}

	if !ethcommon.IsHexAddress(nodeInfo.OperatorAddress) {
		return nil, nil, "", errors.New("invalid operator address")
	}

	consensusPubKey = &crypto.Bls12381PublicKey{}
	if err := consensusPubKey.Unmarshal(hexutil.Decode(nodeInfo.ConsensusPubKey)); err != nil {
		return nil, nil, "", errors.Wrap(err, "failed to unmarshal consensus public key")
	}

	p2pPubKey = &crypto.Ed25519PublicKey{}
	if err := p2pPubKey.Unmarshal(hexutil.Decode(nodeInfo.P2PPubKey)); err != nil {
		return nil, nil, "", errors.Wrap(err, "failed to unmarshal p2p public key")
	}
	p2pID, err = repo.P2PPubKeyToID(p2pPubKey)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "failed to calculate p2p id from p2p public key")
	}

	return consensusPubKey, p2pPubKey, p2pID, nil
}
