package governance

import (
	"encoding/json"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrNotFoundNodeID      = errors.New("node id is not found")
	ErrNotFoundNodeAddress = errors.New("address is not found")
	ErrRepeatedNodeID      = errors.New("repeated node id")
	ErrRepeatedNodeName    = errors.New("repeated node name")
	ErrRepeatedNodeAddress = errors.New("repeated address")
	ErrRegisterExtraArgs   = errors.New("unmarshal node register extra arguments error")
	ErrRemoveExtraArgs     = errors.New("unmarshal node remove extra arguments error")
	ErrUpgradeExtraArgs    = errors.New("unmarshal node upgrade extra arguments error")
	ErrRepeatedDownloadUrl = errors.New("repeated download url")
)

var _ ProposalHandler = (*NodeManager)(nil)

type NodeRegisterExtraArgsSignStruct struct {
	ConsensusPubKey string             `json:"consensus_pub_key"`
	P2PPubKey       string             `json:"p2p_pub_key"`
	MetaData        types.NodeMetaData `json:"meta_data"`
}

type NodeRegisterExtraArgs struct {
	ConsensusPubKey              string             `json:"consensus_pub_key"`
	P2PPubKey                    string             `json:"p2p_pub_key"`
	MetaData                     types.NodeMetaData `json:"meta_data"`
	ConsensusPrivateKeySignature string             `json:"consensus_private_key_signature"`
	P2PPrivateKeySignature       string             `json:"p2p_private_key_signature"`
}

type NodeRemoveExtraArgs struct {
	NodeID uint64 `json:"node_id"`
}

type NodeUpgradeExtraArgs struct {
	DownloadUrls []string `json:"download_urls"`
	CheckHash    string   `json:"check_hash"`
}

type NodeManager struct {
	gov *Governance
	DefaultProposalPermissionManager
}

func NewNodeManager(gov *Governance) *NodeManager {
	return &NodeManager{
		gov:                              gov,
		DefaultProposalPermissionManager: NewDefaultProposalPermissionManager(gov),
	}
}

func (nm *NodeManager) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (nm *NodeManager) SetContext(ctx *common.VMContext) {}

func (nm *NodeManager) ProposePermissionCheck(proposalType ProposalType, user ethcommon.Address) (has bool, err error) {
	switch proposalType {
	case NodeRegister:
		// any node can register
		return true, nil
	case NodeRemove:
		return nm.gov.isCouncilMember(user)
	case NodeUpgrade:
		return nm.gov.isCouncilMember(user)
	default:
		return false, errors.Errorf("unknown proposal type %d", proposalType)
	}
}

func (nm *NodeManager) ProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	switch proposalType {
	case NodeRegister:
		return nm.registerProposeArgsCheck(proposalType, title, desc, blockNumber, extra)
	case NodeRemove:
		return nm.removeProposeArgsCheck(proposalType, title, desc, blockNumber, extra)
	case NodeUpgrade:
		return nm.upgradeProposeArgsCheck(proposalType, title, desc, blockNumber, extra)
	default:
		return errors.Errorf("unknown proposal type %d", proposalType)
	}
}

func (nm *NodeManager) VotePassExecute(proposal *Proposal) error {
	switch proposal.Type {
	case NodeRegister:
		return nm.executeRegister(proposal)
	case NodeRemove:
		return nm.executeRemove(proposal)
	case NodeUpgrade:
		return nm.executeUpgrade(proposal)
	default:
		return errors.Errorf("unknown proposal type %d", proposal.Type)
	}
}

func (nm *NodeManager) registerProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	nodeExtraArgs := &NodeRegisterExtraArgs{}
	if err := json.Unmarshal(extra, nodeExtraArgs); err != nil {
		return ErrRegisterExtraArgs
	}

	consensusPubKey, p2pPubKey, p2pID, err := framework.CheckNodeInfo(types.NodeInfo{
		ConsensusPubKey: nodeExtraArgs.ConsensusPubKey,
		P2PPubKey:       nodeExtraArgs.P2PPubKey,
		OperatorAddress: nm.gov.Ctx.From.String(),
		MetaData:        nodeExtraArgs.MetaData,
	})

	// verify signature
	nodeRegisterExtraArgsSignStruct := &NodeRegisterExtraArgsSignStruct{
		ConsensusPubKey: nodeExtraArgs.ConsensusPubKey,
		P2PPubKey:       nodeExtraArgs.P2PPubKey,
		MetaData:        nodeExtraArgs.MetaData,
	}
	nodeRegisterExtraArgsSignStructBytes, err := json.Marshal(nodeRegisterExtraArgsSignStruct)
	if err != nil {
		return errors.Wrap(err, "failed to marshal node register extra args sign struct")
	}

	if !p2pPubKey.Verify(nodeRegisterExtraArgsSignStructBytes, hexutil.Decode(nodeExtraArgs.P2PPrivateKeySignature)) {
		return errors.New("failed to verify p2p private key signature")
	}
	if !consensusPubKey.Verify(nodeRegisterExtraArgsSignStructBytes, hexutil.Decode(nodeExtraArgs.ConsensusPrivateKeySignature)) {
		return errors.New("failed to verify consensus private key signature")
	}

	nodeManagerContract := framework.NodeManagerBuildConfig.Build(nm.gov.CrossCallSystemContractContext())

	// check  unique index
	if nodeManagerContract.ExistNodeByConsensusPubKey(nodeExtraArgs.ConsensusPubKey) {
		return errors.Errorf("consensus public key %s already registered", nodeExtraArgs.ConsensusPubKey)
	}
	if nodeManagerContract.ExistNodeByP2PID(p2pID) {
		return errors.Errorf("p2p public key %s already registered", nodeExtraArgs.P2PPubKey)
	}
	if nodeManagerContract.ExistNodeByName(nodeExtraArgs.MetaData.Name) {
		return errors.Errorf("name %s already registered", nodeExtraArgs.ConsensusPubKey)
	}

	return nil
}

func (nm *NodeManager) removeProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	nodeExtraArgs := &NodeRemoveExtraArgs{}
	if err := json.Unmarshal(extra, nodeExtraArgs); err != nil {
		return ErrRemoveExtraArgs
	}

	nodeManagerContract := framework.NodeManagerBuildConfig.Build(nm.gov.CrossCallSystemContractContext())

	nodeInfo, err := nodeManagerContract.GetNodeInfo(nodeExtraArgs.NodeID)
	if err != nil {
		return errors.Wrapf(err, "failed to get node info %d", nodeExtraArgs.NodeID)
	}
	if nodeInfo.Status == types.NodeStatusExited {
		return errors.Errorf("node %d [%s] already exited", nodeExtraArgs.NodeID, nodeInfo.MetaData.Name)
	}

	return nil
}

func (nm *NodeManager) upgradeProposeArgsCheck(proposalType ProposalType, title, desc string, blockNumber uint64, extra []byte) error {
	upgradeExtraArgs := &NodeUpgradeExtraArgs{}
	if err := json.Unmarshal(extra, upgradeExtraArgs); err != nil {
		return ErrUpgradeExtraArgs
	}

	// check proposal has repeated download url
	if len(lo.Uniq[string](upgradeExtraArgs.DownloadUrls)) != len(upgradeExtraArgs.DownloadUrls) {
		return ErrRepeatedDownloadUrl
	}

	return nil
}

func (nm *NodeManager) executeRegister(proposal *Proposal) error {
	nodeExtraArgs := &NodeRegisterExtraArgs{}
	if err := json.Unmarshal(proposal.Extra, nodeExtraArgs); err != nil {
		return ErrRegisterExtraArgs
	}

	nodeManagerContract := framework.NodeManagerBuildConfig.Build(nm.gov.CrossCallSystemContractContext())
	_, err := nodeManagerContract.InternalRegisterNode(types.NodeInfo{
		ConsensusPubKey: nodeExtraArgs.ConsensusPubKey,
		P2PPubKey:       nodeExtraArgs.P2PPubKey,
		OperatorAddress: proposal.Proposer,
		MetaData:        nodeExtraArgs.MetaData,
		Status:          types.NodeStatusDataSyncer,
	})
	if err != nil {
		return err
	}

	// TODO: emit event

	return nil
}

func (nm *NodeManager) executeRemove(proposal *Proposal) error {
	nodeExtraArgs := &NodeRemoveExtraArgs{}
	if err := json.Unmarshal(proposal.Extra, nodeExtraArgs); err != nil {
		return ErrRemoveExtraArgs
	}

	nodeManagerContract := framework.NodeManagerBuildConfig.Build(nm.gov.CrossCallSystemContractContext())
	if err := nodeManagerContract.LeaveValidatorSet(nodeExtraArgs.NodeID); err != nil {
		return err
	}

	// TODO: emit event
	return nil
}

func (nm *NodeManager) executeUpgrade(proposal *Proposal) error {
	return nil
}
