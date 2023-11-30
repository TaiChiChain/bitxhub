package governance

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	// NodeProposalKey is key for NodeProposal storage
	NodeProposalKey = "nodeProposalKey"

	// NodeMembersKey is key for node member storage
	NodeMembersKey = "nodeMembersKey"
)

var (
	ErrNodeNumber              = errors.New("node members total count can't bigger than candidates count")
	ErrNotFoundNodeID          = errors.New("node id is not found")
	ErrNotFoundNodeName        = errors.New("node name is not found")
	ErrNotFoundNodeAddress     = errors.New("address is not found")
	ErrNodeExtraArgs           = errors.New("unmarshal node extra arguments error")
	ErrNodeProposalNumberLimit = errors.New("node proposal number limit, only allow one node proposal")
	ErrNotFoundNodeProposal    = errors.New("node proposal not found for the id")
	ErrUnKnownProposalArgs     = errors.New("unknown proposal args")
	ErrRepeatedNodeID          = errors.New("repeated node id")
	ErrRepeatedNodeName        = errors.New("repeated node name")
	ErrRepeatedNodeAddress     = errors.New("repeated address")
	ErrUpgradeExtraArgs        = errors.New("unmarshal node upgrade extra arguments error")
	ErrRepeatedDownloadUrl     = errors.New("repeated download url")
)

// NodeExtraArgs is Node proposal extra arguments
type NodeExtraArgs struct {
	Nodes []*NodeMember
}

// NodeProposalArgs is node proposal arguments
// For node add and remove
type NodeProposalArgs struct {
	BaseProposalArgs
	NodeExtraArgs
}

// NodeProposal is storage of node proposal
type NodeProposal struct {
	BaseProposal
	NodeExtraArgs
	UpgradeExtraArgs
}

type Node struct {
	Members []*NodeMember
}

type NodeMember struct {
	Name    string `mapstructure:"name" toml:"name"`
	NodeId  string `mapstructure:"node_id" toml:"node_id"`
	Address string `mapstructure:"address" toml:"address"`
	ID      uint64 `mapstructure:"id" toml:"id"`
}

type NodeVoteArgs struct {
	BaseVoteArgs
}

// UpgradeProposalArgs for node upgrade
type UpgradeProposalArgs struct {
	BaseProposalArgs
	UpgradeExtraArgs
}

type UpgradeExtraArgs struct {
	DownloadUrls []string
	CheckHash    string
}

var _ common.SystemContract = (*NodeManager)(nil)

type NodeManager struct {
	gov *Governance

	account                ledger.IAccount
	councilAccount         ledger.IAccount
	stateLedger            ledger.StateLedger
	currentLog             *common.Log
	proposalID             *ProposalID
	lastHeight             uint64
	notFinishedProposalMgr *NotFinishedProposalMgr
}

func NewNodeManager(cfg *common.SystemContractConfig) *NodeManager {
	gov, err := NewGov([]ProposalType{NodeUpgrade, NodeAdd, NodeRemove}, cfg.Logger)
	if err != nil {
		panic(err)
	}

	return &NodeManager{
		gov: gov,
	}
}

func (nm *NodeManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.NodeManagerContractAddr)
	nm.account = stateLedger.GetOrCreateAccount(addr)
	nm.stateLedger = stateLedger
	nm.currentLog = &common.Log{
		Address: addr,
	}
	nm.proposalID = NewProposalID(stateLedger)

	councilAddr := types.NewAddressByStr(common.CouncilManagerContractAddr)
	nm.councilAccount = stateLedger.GetOrCreateAccount(councilAddr)
	nm.notFinishedProposalMgr = NewNotFinishedProposalMgr(stateLedger)

	// check and update
	nm.checkAndUpdateState(lastHeight)
	nm.lastHeight = lastHeight
}

func (nm *NodeManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer nm.gov.SaveLog(nm.stateLedger, nm.currentLog)

	// parse method and arguments from msg payload
	args, err := nm.gov.GetArgs(msg)
	if err != nil {
		return nil, err
	}

	var result *vm.ExecutionResult
	switch v := args.(type) {
	case *ProposalArgs:
		result, err = nm.propose(msg.From, v)
	case *VoteArgs:
		result, err = nm.vote(msg.From, v)
	case *GetProposalArgs:
		result, err = nm.getProposal(v.ProposalID)
	default:
		return nil, errors.New("unknown proposal args")
	}
	if result != nil {
		result.UsedGas = common.CalculateDynamicGas(msg.Data)
	}
	return result, err
}

func (nm *NodeManager) getNodeProposalArgs(args *ProposalArgs) (*NodeProposalArgs, error) {
	nodeArgs := &NodeProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}

	extraArgs := &NodeExtraArgs{}
	if err := json.Unmarshal(args.Extra, extraArgs); err != nil {
		return nil, ErrNodeExtraArgs
	}

	nodeArgs.NodeExtraArgs = *extraArgs
	return nodeArgs, nil
}

func (nm *NodeManager) getUpgradeArgs(args *ProposalArgs) (*UpgradeProposalArgs, error) {
	upgradeArgs := &UpgradeProposalArgs{
		BaseProposalArgs: args.BaseProposalArgs,
	}

	extraArgs := &UpgradeExtraArgs{}
	if err := json.Unmarshal(args.Extra, extraArgs); err != nil {
		return nil, ErrUpgradeExtraArgs
	}

	upgradeArgs.UpgradeExtraArgs = *extraArgs
	return upgradeArgs, nil
}

func (nm *NodeManager) propose(addr ethcommon.Address, args *ProposalArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	if args.ProposalType == uint8(NodeUpgrade) {
		upgradeArgs, err := nm.getUpgradeArgs(args)
		if err != nil {
			return nil, err
		}

		result.ReturnData, result.Err = nm.proposeUpgrade(addr, upgradeArgs)

		return result, nil
	}

	result.ReturnData, result.Err = nm.proposeNodeAddRemove(addr, args)

	return result, nil
}

func (nm *NodeManager) proposeNodeAddRemove(addr ethcommon.Address, args *ProposalArgs) ([]byte, error) {
	nodeArgs, err := nm.getNodeProposalArgs(args)
	if err != nil {
		return nil, err
	}

	members, err := GetNodeMembers(nm.stateLedger)
	if err != nil {
		return nil, err
	}
	// store node information in members into map
	memberNameMap := make(map[string]*NodeMember)
	for _, member := range members {
		key := member.Name
		memberNameMap[key] = member
	}
	memberAddressMap := make(map[string]*NodeMember)
	for _, member := range members {
		key := member.Address
		memberAddressMap[key] = member
	}
	memberNodeIdMap := make(map[string]*NodeMember)
	for _, member := range members {
		key := member.NodeId
		memberNodeIdMap[key] = member
	}

	isExist, _ := CheckInCouncil(nm.councilAccount, addr.String())
	nodeAddRole := false
	if args.ProposalType == uint8(NodeAdd) {
		// check addr if is exist in council
		if !isExist {
			return nil, ErrNotFoundCouncilMember
		}
		for _, node := range nodeArgs.NodeExtraArgs.Nodes {
			if _, found := memberNodeIdMap[node.NodeId]; found {
				return nil, ErrRepeatedNodeID
			}
			if _, found := memberAddressMap[node.Address]; found {
				return nil, ErrRepeatedNodeAddress
			}
			if _, found := memberNameMap[node.Name]; found {
				return nil, ErrRepeatedNodeName
			}
		}
	} else {
		// check whether the from address and node address are consistent
		for _, node := range nodeArgs.NodeExtraArgs.Nodes {
			if addr.String() != node.Address {
				// If the addresses are inconsistent, verify whether the address that initiated the proposal is a member of the committee.
				if !isExist {
					return nil, ErrNotFoundCouncilMember
				}
			} else if !isExist {
				nodeAddRole = true
			}

			if _, found := memberNodeIdMap[node.NodeId]; !found {
				return nil, ErrNotFoundNodeID
			}
			if _, found := memberAddressMap[node.Address]; !found {
				return nil, ErrNotFoundNodeAddress
			}
			if _, found := memberNameMap[node.Name]; !found {
				return nil, ErrNotFoundNodeName
			}
		}
	}
	council, isExistCouncil := GetCouncil(nm.councilAccount)
	if !isExistCouncil {
		return nil, ErrNotFoundCouncil
	}

	baseProposal, err := nm.gov.Propose(&addr, ProposalType(nodeArgs.ProposalType), nodeArgs.Title, nodeArgs.Desc, nodeArgs.BlockNumber, nm.lastHeight, nodeAddRole)
	if err != nil {
		return nil, err
	}

	// check proposal has repeated nodes
	if len(lo.Uniq[string](lo.Map[*NodeMember, string](nodeArgs.Nodes, func(item *NodeMember, index int) string {
		return item.NodeId
	}))) != len(nodeArgs.Nodes) {
		return nil, ErrRepeatedNodeID
	}

	// set proposal id
	proposal := &NodeProposal{
		BaseProposal: *baseProposal,
	}

	id, err := nm.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.Nodes = nodeArgs.Nodes

	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*CouncilMember, uint64](council.Members, func(item *CouncilMember, index int) uint64 {
		return item.Weight
	}))

	b, err := nm.saveNodeProposal(proposal)
	if err != nil {
		return nil, err
	}

	// propose generate not finished proposal
	if err = nm.notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  proposal.ID,
		DeadlineBlockNumber: proposal.BlockNumber,
		ContractAddr:        common.NodeManagerContractAddr,
	}); err != nil {
		return nil, err
	}

	returnData, err := nm.gov.PackOutputArgs(ProposeMethod, id)
	if err != nil {
		return nil, err
	}

	// record log
	nm.gov.RecordLog(nm.currentLog, ProposeMethod, &proposal.BaseProposal, b)

	return returnData, nil
}

func (nm *NodeManager) proposeUpgrade(addr ethcommon.Address, args *UpgradeProposalArgs) ([]byte, error) {
	baseProposal, err := nm.gov.Propose(&addr, ProposalType(args.ProposalType), args.Title, args.Desc, args.BlockNumber, nm.lastHeight, false)
	if err != nil {
		return nil, err
	}

	// check proposal has repeated download url
	if len(lo.Uniq[string](args.DownloadUrls)) != len(args.DownloadUrls) {
		return nil, ErrRepeatedDownloadUrl
	}

	// check addr if is exist in council
	isExist, council := CheckInCouncil(nm.councilAccount, addr.String())
	if !isExist {
		return nil, ErrNotFoundCouncilMember
	}

	// set proposal id
	proposal := &NodeProposal{
		BaseProposal: *baseProposal,
	}

	id, err := nm.proposalID.GetAndAddID()
	if err != nil {
		return nil, err
	}
	proposal.ID = id
	proposal.DownloadUrls = args.DownloadUrls
	proposal.CheckHash = args.CheckHash
	proposal.TotalVotes = lo.Sum[uint64](lo.Map[*CouncilMember, uint64](council.Members, func(item *CouncilMember, index int) uint64 {
		return item.Weight
	}))

	b, err := nm.saveNodeProposal(proposal)
	if err != nil {
		return nil, err
	}

	// propose generate not finished proposal
	if err = nm.notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  proposal.ID,
		DeadlineBlockNumber: proposal.BlockNumber,
		ContractAddr:        common.NodeManagerContractAddr,
	}); err != nil {
		return nil, err
	}

	returnData, err := nm.gov.PackOutputArgs(ProposeMethod, id)
	if err != nil {
		return nil, err
	}

	// record log
	nm.gov.RecordLog(nm.currentLog, ProposeMethod, &proposal.BaseProposal, b)

	return returnData, nil
}

// Vote a proposal, return vote status
func (nm *NodeManager) vote(user ethcommon.Address, voteArgs *VoteArgs) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	// get proposal
	proposal, err := nm.loadNodeProposal(voteArgs.ProposalId)
	if err != nil {
		return nil, err
	}

	if proposal.Type == NodeUpgrade {
		result.Err = nm.voteUpgrade(user, proposal, &NodeVoteArgs{BaseVoteArgs: voteArgs.BaseVoteArgs})
		return result, nil
	}

	result.Err = nm.voteNodeAddRemove(user, proposal, &NodeVoteArgs{BaseVoteArgs: voteArgs.BaseVoteArgs})
	return result, nil
}

func (nm *NodeManager) voteNodeAddRemove(user ethcommon.Address, proposal *NodeProposal, voteArgs *NodeVoteArgs) error {
	// check user can vote
	isExist, _ := CheckInCouncil(nm.councilAccount, user.String())
	if !isExist {
		return ErrNotFoundCouncilMember
	}

	res := VoteResult(voteArgs.VoteResult)
	proposalStatus, err := nm.gov.Vote(&user, &proposal.BaseProposal, res)
	if err != nil {
		return err
	}
	proposal.Status = proposalStatus

	b, err := nm.saveNodeProposal(proposal)
	if err != nil {
		return err
	}

	// update not finished proposal
	if proposal.Status == Approved || proposal.Status == Rejected {
		if err := nm.notFinishedProposalMgr.RemoveProposal(proposal.ID); err != nil {
			return err
		}
	}

	// if proposal is approved, update the node members
	if proposal.Status == Approved {
		members, err := GetNodeMembers(nm.stateLedger)
		if err != nil {
			return err
		}

		if proposal.Type == NodeAdd {
			for _, node := range proposal.Nodes {
				newNodeID, err := base.AddNode(nm.stateLedger, rbft.NodeInfo{
					AccountAddress:       node.Address,
					P2PNodeID:            node.NodeId,
					ConsensusVotingPower: 100,
				})

				if err != nil {
					return err
				}
				node.ID = newNodeID
			}
			members = append(members, proposal.Nodes...)
		}

		if proposal.Type == NodeRemove {
			// https://github.com/samber/lo
			// Use the Associate method to create a map with the node's NodeId as the key and the NodeMember object as the value
			nodeIdToNodeMap := lo.Associate(proposal.Nodes, func(node *NodeMember) (string, *NodeMember) {
				return node.NodeId, node
			})

			// The members slice is updated to filteredMembers, which does not contain members with the same NodeId as proposalNodes
			filteredMembers := lo.Reject(members, func(member *NodeMember, _ int) bool {
				_, exists := nodeIdToNodeMap[member.NodeId]
				return exists
			})

			for _, node := range proposal.Nodes {
				err = base.RemoveNodeByP2PNodeID(nm.stateLedger, node.NodeId)
				if err != nil {
					return err
				}
			}
			members = filteredMembers
		}

		cb, err := json.Marshal(members)
		if err != nil {
			return err
		}
		nm.account.SetState([]byte(NodeMembersKey), cb)
	}

	// record log
	nm.gov.RecordLog(nm.currentLog, VoteMethod, &proposal.BaseProposal, b)

	// vote not return value
	return nil
}

func (nm *NodeManager) voteUpgrade(user ethcommon.Address, proposal *NodeProposal, voteArgs *NodeVoteArgs) error {
	// check user can vote
	isExist, _ := CheckInCouncil(nm.councilAccount, user.String())
	if !isExist {
		return ErrNotFoundCouncilMember
	}

	res := VoteResult(voteArgs.VoteResult)
	proposalStatus, err := nm.gov.Vote(&user, &proposal.BaseProposal, res)
	if err != nil {
		return err
	}
	proposal.Status = proposalStatus

	b, err := nm.saveNodeProposal(proposal)
	if err != nil {
		return err
	}

	// update not finished proposal
	if proposal.Status == Approved || proposal.Status == Rejected {
		if err := nm.notFinishedProposalMgr.RemoveProposal(proposal.ID); err != nil {
			return err
		}
	}

	// record log
	// if approved, guardian sync log, then update node and restart
	nm.gov.RecordLog(nm.currentLog, VoteMethod, &proposal.BaseProposal, b)

	// vote not return value
	return nil
}

func (nm *NodeManager) saveNodeProposal(proposal *NodeProposal) ([]byte, error) {
	return saveNodeProposal(nm.stateLedger, proposal)
}

func (nm *NodeManager) loadNodeProposal(proposalID uint64) (*NodeProposal, error) {
	return loadNodeProposal(nm.stateLedger, proposalID)
}

// getProposal view proposal details
func (nm *NodeManager) getProposal(proposalID uint64) (*vm.ExecutionResult, error) {
	result := &vm.ExecutionResult{}

	isExist, b := nm.account.GetState([]byte(fmt.Sprintf("%s%d", NodeProposalKey, proposalID)))
	if isExist {
		packed, err := nm.gov.PackOutputArgs(ProposalMethod, b)
		if err != nil {
			return nil, err
		}
		result.ReturnData = packed
		return result, nil
	}
	return nil, ErrNotFoundNodeProposal
}

func (nm *NodeManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := nm.gov.GetArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}

	return common.CalculateDynamicGas(*callArgs.Data), nil
}

func (nm *NodeManager) checkAndUpdateState(lastHeight uint64) {
	if err := CheckAndUpdateState(lastHeight, nm.stateLedger); err != nil {
		nm.gov.logger.Errorf("check and update state error: %s", err)
	}
}

func InitNodeMembers(lg ledger.StateLedger, members []*repo.NodeName, epochInfo *rbft.EpochInfo) error {
	// read member config, write to ViewLedger
	nodeMembers := mergeData(members, epochInfo)
	c, err := json.Marshal(nodeMembers)
	if err != nil {
		return err
	}
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.NodeManagerContractAddr))
	account.SetState([]byte(NodeMembersKey), c)
	return nil
}

func mergeData(members []*repo.NodeName, epochInfo *rbft.EpochInfo) []*NodeMember {
	var result []*NodeMember
	nodeMap := make(map[uint64]*repo.NodeName)
	for _, node := range members {
		nodeMap[node.ID] = node
	}

	// traverse epochInfo and merge data
	fillNodeInfo := func(nodes []rbft.NodeInfo) {
		for _, nodeInfo := range nodes {
			nodeMember := NodeMember{
				NodeId:  nodeInfo.P2PNodeID,
				Address: nodeInfo.AccountAddress,
				ID:      nodeInfo.ID,
			}
			if member, ok := nodeMap[nodeInfo.ID]; ok {
				nodeMember.Name = member.Name
			}
			result = append(result, &nodeMember)
		}
	}
	fillNodeInfo(epochInfo.ValidatorSet)
	fillNodeInfo(epochInfo.CandidateSet)
	fillNodeInfo(epochInfo.DataSyncerSet)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func GetNodeMembers(lg ledger.StateLedger) ([]*NodeMember, error) {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.NodeManagerContractAddr))
	success, data := account.GetState([]byte(NodeMembersKey))
	if success {
		var members []*NodeMember
		if err := json.Unmarshal(data, &members); err != nil {
			return nil, err
		}
		return members, nil
	}
	return nil, errors.New("node member should be initialized in genesis")
}

func saveNodeProposal(stateLedger ledger.StateLedger, proposal ProposalObject) ([]byte, error) {
	addr := types.NewAddressByStr(common.NodeManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	b, err := json.Marshal(proposal)
	if err != nil {
		return nil, err
	}
	// save proposal
	account.SetState([]byte(fmt.Sprintf("%s%d", NodeProposalKey, proposal.GetID())), b)

	return b, nil
}

func loadNodeProposal(stateLedger ledger.StateLedger, proposalID uint64) (*NodeProposal, error) {
	addr := types.NewAddressByStr(common.NodeManagerContractAddr)
	account := stateLedger.GetOrCreateAccount(addr)

	isExist, data := account.GetState([]byte(fmt.Sprintf("%s%d", NodeProposalKey, proposalID)))
	if !isExist {
		return nil, ErrNotFoundNodeProposal
	}

	proposal := &NodeProposal{}
	if err := json.Unmarshal(data, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}
