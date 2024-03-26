package base

import (
	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type NodeMetaData struct {
	Name       string `mapstructure:"name" toml:"name" json:"name"`
	Desc       string `mapstructure:"desc" toml:"desc" json:"desc"`
	ImageURL   string `mapstructure:"image_url" toml:"image_url" json:"image_url"`
	WebsiteURL string `mapstructure:"website_url" toml:"website_url" json:"website_url"`
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
}

type ConsensusVotingPower struct {
	NodeID uint64
	// Consensus voting weight (calculated by active stake).
	ConsensusVotingPower int64
}

type NodeManager interface {
	// Internal functions(called by System Contract)
	// after RegisterNode proposal passed, node will be added to NodeRegister and NodeRegistry
	InternalRegisterNode(info NodeInfo) (id uint64, err error)
	// after epoch change, mark all node in PendingInactiveValidatorIDSet as inactive
	InternalProcessNodeLeave() error
	// after epoch change and after InternalProcessNodeLeave called, return CandidateIDSet and ActiveValidatorIDSet
	InternalGetConsensusCandidateNodeIDs() ([]uint64, error)
	// after epoch change, set ActiveValidatorVotingPowers and ActiveValidatorIDSet, update CandidateIDSet
	InternalUpdateActiveValidatorSet(ActiveValidatorVotingPowers []ConsensusVotingPower) error

	// Operator functions
	// register node to NodeRegistry, and add it to DataSyncerIDSet
	JoinValidatorSet(nodeID uint64) error
	// add node to PendingInactiveValidatorIDSet
	LeaveValidatorSet(nodeID uint64) error
	UpdateMetaData(nodeID uint64, metaData NodeMetaData) error
	UpdateOperator(nodeID uint64, newOperatorAddress string) error

	// Query functions
	GetNodeInfo(nodeID uint64) (info NodeInfo, err error)
	GetTotalNodeCount() int
	GetNodeInfos(nodeIDs []uint64) (infos []NodeInfo, err error)
	GetActiveValidatorSet() (infos []NodeInfo, votingPowers []ConsensusVotingPower, err error)
	GetDataSyncerSet() (infos []NodeInfo, err error)
	GetCandidateSet() (infos []NodeInfo, err error)
	GetExitedSet() (infos []NodeInfo, err error)
}

type nodeManager struct {
	currentEpoch *rbft.EpochInfo
	stateLedger  ledger.StateLedger

	// store all registered node info (including exited nodes)
	NodeRegistry map[uint64]NodeInfo
	// store all new node info in the next epoch, will clean up in the next epoch
	NextEpochUpdatedNodes map[uint64]NewNodeInfo

	// track all updated node id in the next epoch, will clean up in the next epoch
	NextEpochUpdatedNodeIDSet []uint64
	// DataSyncer requires the Operator to send a JoinValidatorSet transaction to become a candidate/validator
	DataSyncerIDSet []uint64
	// node stake is not enough, or is not Top N
	CandidateIDSet []uint64
	// node stake is enough, and is Top N
	ActiveValidatorIDSet []uint64
	// Operator send a LeaveValidatorSet transaction
	PendingInactiveValidatorIDSet []uint64
	// including Candidate and ActiveValidator
	ActiveValidatorVotingPowers []ConsensusVotingPower
}
