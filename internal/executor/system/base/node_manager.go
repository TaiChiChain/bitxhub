package base

import (
	"fmt"
	"strconv"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Status byte

// status constants, the order is important
const (
	StatusSyncing Status = iota
	StatusCandidate
	StatusActive
	StatusPendingInactive
	StatusExited
	// StatusOutOfRange identities the last status, all status must add before StatusOutOfRange
	StatusOutOfRange
)

const (
	nodeManagerPrefix           = "nodeManager"
	idNonce                     = "idNonce"
	nodeRegistry                = "nodeRegistry"
	nextEpochUpdateNodes        = "nextEpochUpdateNodes"
	nextEpochUpdatedNodeIDSet   = "nextEpochUpdatedNodeIDSet"
	statusPrefix                = "Status"
	activeValidatorVotingPowers = "activeValidatorVotingPowers"
)

var (
	ErrNodeStatusIncorrect     = errors.New("node status is incorrect")
	ErrRepeatOperation         = errors.New("repeat operation, the operation has been executed")
	ErrNodeNotFound            = errors.New("node not found")
	ErrStatusSetNotFound       = errors.New("status set not found")
	ErrNodeNotFoundInStatusSet = errors.New("node not found in status set")
	ErrNodeAlreadyInStatusSet  = errors.New("node is already in status set")
	ErrPermissionDenied        = errors.New("permission denied")
)

type NodeMetaData struct {
	Name       string
	Desc       string
	ImageURL   string
	WebsiteURL string
}

type NewNodeInfo struct {
	// Operator address, with permission to manage node (can update)
	OperatorAddress string
	// Meta data (can update)
	MetaData NodeMetaData
}

type NodeInfo struct {
	// The node serial number is unique in the entire network.
	// Once allocated, it will not change.
	// It is allocated through the governance contract in a manner similar to the self-incrementing primary key.
	ID uint64
	// Node Consensus and Network signer public key
	NodePubKey string
	// Operator address, with permission to manage node (can update)
	OperatorAddress string
	// Meta data (can update)
	MetaData NodeMetaData
	// Node status
	Status Status
}

type ConsensusVotingPower struct {
	NodeID uint64
	// Consensus voting weight (calculated by active stake).
	ConsensusVotingPower int64
}

type NodeManager interface {
	// InternalRegisterNode is an Internal function(called by System Contract)
	// after RegisterNode proposal passed, node will be added to NodeRegister and NodeRegistry
	InternalRegisterNode(info NodeInfo) (id uint64, err error)
	// InternalProcessNodeLeave is an Internal function
	// after epoch change, mark all node in PendingInactiveValidatorIDSet as exited
	InternalProcessNodeLeave() error
	// InternalGetConsensusCandidateNodeIDs is an Internal function
	// after epoch change and after InternalProcessNodeLeave called, return CandidateIDSet and ActiveValidatorIDSet
	InternalGetConsensusCandidateNodeIDs() ([]uint64, error)
	// InternalUpdateActiveValidatorSet is an Internal function
	// after epoch change, set ActiveValidatorVotingPowers and ActiveValidatorIDSet, update CandidateIDSet
	InternalUpdateActiveValidatorSet(ActiveValidatorVotingPowers []ConsensusVotingPower) error

	// JoinCandidateSet is an Operator function
	// DataSyncer Node join candidate set
	JoinCandidateSet(nodeID uint64) error
	// LeaveValidatorSet is an Operator function
	// add node to PendingInactiveValidatorIDSet
	LeaveValidatorSet(nodeID uint64) error
	// UpdateMetaData own meta data
	UpdateMetaData(nodeID uint64, metaData NodeMetaData) error
	// UpdateOperator updates the operator address
	UpdateOperator(nodeID uint64, newOperatorAddress string) error

	// GetNodeInfo returns the detail info of a node
	GetNodeInfo(nodeID uint64) (info NodeInfo, err error)
	// GetTotalNodeCount returns the total number of nodes
	GetTotalNodeCount() int
	// GetNodeInfos returns the detail info of all nodes
	GetNodeInfos(nodeIDs []uint64) (infos []NodeInfo, err error)
	// GetActiveValidatorSet returns all active validator
	GetActiveValidatorSet() (infos []NodeInfo, votingPowers []ConsensusVotingPower, err error)
	// GetDataSyncerSet returns all data syncer
	GetDataSyncerSet() (infos []NodeInfo, err error)
	// GetCandidateSet returns all candidate
	GetCandidateSet() (infos []NodeInfo, err error)
	// GetPendingInactiveSet returns all pending inactive nodes
	GetPendingInactiveSet() (infos []NodeInfo, err error)
	// GetExitedSet returns all exited node
	GetExitedSet() (infos []NodeInfo, err error)

	SetContext(ctx *common.VMContext)
}

type nodeManager struct {
	currentEpoch     *rbft.EpochInfo
	stateLedger      ledger.StateLedger
	msgSender        *ethcommon.Address
	currentLogs      *[]common.Log
	abi              *abi.ABI
	account          ledger.IAccount
	innerNodeIDNonce *common.VMSlot[uint64]

	// store all registered node info (including exited nodes)
	NodeRegistry *common.VMMap[uint64, NodeInfo]
	// track all updated node id in the next epoch, will clean up in the next epoch
	NextEpochUpdatedNodeIDSet *common.VMSlot[[]uint64]
	// track all node at given status
	StatusIDSet func(Status) *common.VMSlot[[]uint64]
	// store all new node info in the next epoch, will clean up in the next epoch
	NextEpochUpdateNodes *common.VMMap[uint64, NodeInfo]
	// including Candidate and ActiveValidator
	ActiveValidatorVotingPowers *common.VMSlot[[]ConsensusVotingPower]
}

func NewNodeManager(cfg *common.SystemContractConfig) NodeManager {
	return &nodeManager{}
}

func (n *nodeManager) InternalRegisterNode(info NodeInfo) (id uint64, err error) {
	if n.msgSender.String() != common.EpochManagerContractAddr {
		return 0, ErrPermissionDenied
	}
	isExist := n.NodeRegistry.Has(info.ID)
	if isExist {
		return 0, ErrRepeatOperation
	}
	_, id, err = n.innerNodeIDNonce.Get()
	if err != nil {
		return 0, err
	}
	info.ID = id
	info.Status = StatusSyncing
	if err = n.innerNodeIDNonce.Put(id + 1); err != nil {
		return 0, err
	}
	syncingSlot := n.StatusIDSet(StatusSyncing)
	isExist, set, err := syncingSlot.Get()
	if err != nil {
		return 0, err
	}
	if !isExist {
		set = []uint64{}
	}
	if err = syncingSlot.Put(append(set, id)); err != nil {
		return 0, err
	}
	return id, n.NodeRegistry.Put(info.ID, info)
}

func (n *nodeManager) InternalProcessNodeLeave() error {
	if n.msgSender.String() != common.EpochManagerContractAddr {
		return ErrPermissionDenied
	}
	pendingInactiveSlot := n.StatusIDSet(StatusPendingInactive)
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
		info.Status = StatusExited
		// put it back
		if err = n.NodeRegistry.Put(id, info); err != nil {
			return err
		}
	}
	// get exited nodes set
	exitedSlot := n.StatusIDSet(StatusExited)
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

func (n *nodeManager) InternalGetConsensusCandidateNodeIDs() ([]uint64, error) {
	isExist, candidateIDs, err := n.StatusIDSet(StatusCandidate).Get()
	if err != nil {
		return nil, err
	}
	if !isExist {
		candidateIDs = []uint64{}
	}
	isExist, validateIDs, err := n.StatusIDSet(StatusActive).Get()
	if err != nil {
		return nil, err
	}
	if !isExist {
		validateIDs = []uint64{}
	}
	return append(candidateIDs, validateIDs...), nil
}

func (n *nodeManager) InternalUpdateActiveValidatorSet(ActiveValidatorVotingPowers []ConsensusVotingPower) error {
	if n.msgSender.String() != common.EpochManagerContractAddr {
		return ErrPermissionDenied
	}
	allValidaNodes, err := n.InternalGetConsensusCandidateNodeIDs()
	if err != nil {
		return err
	}
	validatorMap := make(map[uint64]struct{}, len(ActiveValidatorVotingPowers))
	validators := make([]uint64, 0, len(ActiveValidatorVotingPowers))
	candidates := make([]uint64, 0, len(allValidaNodes)-len(ActiveValidatorVotingPowers))
	for _, votingPower := range ActiveValidatorVotingPowers {
		nodeId := votingPower.NodeID
		isExist, nodeInfo, err := n.NodeRegistry.Get(nodeId)
		if err != nil {
			return err
		}
		if !isExist {
			return ErrNodeNotFound
		}
		nodeInfo.Status = StatusActive
		if err = n.NodeRegistry.Put(nodeId, nodeInfo); err != nil {
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
	if err = n.getStatusSet(StatusCandidate).Put(candidates); err != nil {
		return err
	}
	if err = n.getStatusSet(StatusActive).Put(validators); err != nil {
		return err
	}

	return n.ActiveValidatorVotingPowers.Put(ActiveValidatorVotingPowers)
}

func (n *nodeManager) JoinCandidateSet(nodeID uint64) error {
	return n.operatorTransferStatus(StatusSyncing, nodeID)
}

func (n *nodeManager) LeaveValidatorSet(nodeID uint64) error {
	return n.operatorTransferStatus(StatusActive, nodeID)
}

func (n *nodeManager) UpdateMetaData(nodeID uint64, metaData NodeMetaData) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	newNodeInfo := NewNodeInfo{
		OperatorAddress: nodeInfo.OperatorAddress,
		MetaData:        metaData,
	}
	return n.updateNodeInfo(nodeID, nodeInfo, newNodeInfo)
}

func (n *nodeManager) UpdateOperator(nodeID uint64, newOperatorAddress string) error {
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	newNodeInfo := NewNodeInfo{
		OperatorAddress: newOperatorAddress,
		MetaData:        nodeInfo.MetaData,
	}
	return n.updateNodeInfo(nodeID, nodeInfo, newNodeInfo)
}

func (n *nodeManager) updateNodeInfo(nodeID uint64, oldNodeInfo NodeInfo, newNodeInfo NewNodeInfo) error {
	if n.msgSender.String() != oldNodeInfo.OperatorAddress {
		return ErrPermissionDenied
	}
	nodeInfo := NodeInfo{
		ID:              oldNodeInfo.ID,
		NodePubKey:      oldNodeInfo.NodePubKey,
		OperatorAddress: newNodeInfo.OperatorAddress,
		MetaData:        newNodeInfo.MetaData,
		Status:          oldNodeInfo.Status,
	}
	return n.NodeRegistry.Put(nodeID, nodeInfo)
}

func (n *nodeManager) GetNodeInfo(nodeID uint64) (info NodeInfo, err error) {
	isExist, nodeInfo, err := n.NodeRegistry.Get(nodeID)
	if err != nil {
		return NodeInfo{}, err
	}
	if !isExist {
		return NodeInfo{}, ErrNodeNotFound
	}
	return nodeInfo, nil
}

func (n *nodeManager) GetTotalNodeCount() int {
	getNum := func(status Status) int {
		isExist, nodes, err := n.StatusIDSet(status).Get()
		if err != nil {
			return 0
		}
		if !isExist {
			return 0
		}
		return len(nodes)
	}
	cnt := 0
	for i := Status(0); i < StatusOutOfRange; i++ {
		cnt += getNum(i)
	}
	return cnt
}

func (n *nodeManager) GetNodeInfos(nodeIDs []uint64) (infos []NodeInfo, err error) {
	for _, id := range nodeIDs {
		info, err := n.GetNodeInfo(id)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return
}

func (n *nodeManager) GetActiveValidatorSet() (infos []NodeInfo, votingPowers []ConsensusVotingPower, err error) {
	isExists, nodes, err := n.StatusIDSet(StatusActive).Get()
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
	isExists, votingPowers, err = n.ActiveValidatorVotingPowers.Get()
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

func (n *nodeManager) GetDataSyncerSet() (infos []NodeInfo, err error) {
	isExists, nodes, err := n.StatusIDSet(StatusSyncing).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *nodeManager) GetCandidateSet() (infos []NodeInfo, err error) {
	isExists, nodes, err := n.StatusIDSet(StatusCandidate).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *nodeManager) GetPendingInactiveSet() (infos []NodeInfo, err error) {
	isExists, nodes, err := n.StatusIDSet(StatusPendingInactive).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *nodeManager) GetExitedSet() (infos []NodeInfo, err error) {
	isExists, nodes, err := n.StatusIDSet(StatusExited).Get()
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, nil
	}
	return n.GetNodeInfos(nodes)
}

func (n *nodeManager) operatorTransferStatus(from Status, nodeID uint64) error {
	if !(from < StatusOutOfRange-1) {
		return ErrNodeStatusIncorrect
	}
	nodeInfo, err := n.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	if nodeInfo.OperatorAddress != n.msgSender.String() {
		return ErrPermissionDenied
	}
	if nodeInfo.Status != from {
		return ErrNodeStatusIncorrect
	}
	nodeInfo.Status = from + 1
	if err = n.NodeRegistry.Put(nodeID, nodeInfo); err != nil {
		return err
	}
	fromSlot := n.StatusIDSet(from)
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
	toSlot := n.StatusIDSet(from + 1)
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

func (n *nodeManager) getStatusSet(status Status) *common.VMSlot[[]uint64] {
	return common.NewVMSlot[[]uint64](n.account, nodeManagerPrefix+statusPrefix+strconv.Itoa(int(status)))
}

func (n *nodeManager) SetContext(ctx *common.VMContext) {
	n.msgSender = ctx.CurrentUser
	n.abi = ctx.ABI
	n.stateLedger = ctx.StateLedger
	n.currentLogs = ctx.CurrentLogs

	n.account = ctx.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.NodeManagerContractAddr))
	n.innerNodeIDNonce = common.NewVMSlot[uint64](n.account, nodeManagerPrefix+idNonce)
	n.NodeRegistry = common.NewVMMap[uint64, NodeInfo](n.account, nodeManagerPrefix+nodeRegistry, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	n.NextEpochUpdateNodes = common.NewVMMap[uint64, NodeInfo](n.account, nodeManagerPrefix+nextEpochUpdateNodes, func(key uint64) string {
		return fmt.Sprintf("%d", key)
	})
	n.StatusIDSet = n.getStatusSet
	n.NextEpochUpdatedNodeIDSet = common.NewVMSlot[[]uint64](n.account, nodeManagerPrefix+nextEpochUpdatedNodeIDSet)
	n.ActiveValidatorVotingPowers = common.NewVMSlot[[]ConsensusVotingPower](n.account, nodeManagerPrefix+activeValidatorVotingPowers)
}
