// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package node_manager

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
	_ = packer.RevertError{}
)

// ConsensusVotingPower is an auto generated low-level Go binding around an user-defined struct.
type ConsensusVotingPower struct {
	NodeID               uint64
	ConsensusVotingPower int64
}

// NodeInfo is an auto generated low-level Go binding around an user-defined struct.
type NodeInfo struct {
	ID              uint64
	ConsensusPubKey string
	P2PPubKey       string
	P2PID           string
	OperatorAddress string
	MetaData        NodeMetaData
	Status          uint8
}

// NodeMetaData is an auto generated low-level Go binding around an user-defined struct.
type NodeMetaData struct {
	Name       string
	Desc       string
	ImageURL   string
	WebsiteURL string
}

type NodeManager interface {

	// JoinCandidateSet is a paid mutator transaction binding the contract method 0x2b27ec44.
	//
	// Solidity: function joinCandidateSet(uint64 nodeID) returns()
	JoinCandidateSet(nodeID uint64) error

	// LeaveValidatorSet is a paid mutator transaction binding the contract method 0xede1f5d1.
	//
	// Solidity: function leaveValidatorSet(uint64 nodeID) returns()
	LeaveValidatorSet(nodeID uint64) error

	// UpdateMetaData is a paid mutator transaction binding the contract method 0xee99437a.
	//
	// Solidity: function updateMetaData(uint64 nodeID, (string,string,string,string) metaData) returns()
	UpdateMetaData(nodeID uint64, metaData NodeMetaData) error

	// UpdateOperator is a paid mutator transaction binding the contract method 0xcee4d6f5.
	//
	// Solidity: function updateOperator(uint64 nodeID, string newOperatorAddress) returns()
	UpdateOperator(nodeID uint64, newOperatorAddress string) error

	// GetActiveValidatorSet is a free data retrieval call binding the contract method 0x59acaac4.
	//
	// Solidity: function GetActiveValidatorSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info, (uint64,int64)[] votingPowers)
	GetActiveValidatorSet() ([]NodeInfo, []ConsensusVotingPower, error)

	// GetCandidateSet is a free data retrieval call binding the contract method 0x84c3e579.
	//
	// Solidity: function GetCandidateSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
	GetCandidateSet() ([]NodeInfo, error)

	// GetDataSyncerSet is a free data retrieval call binding the contract method 0x834b2b3b.
	//
	// Solidity: function GetDataSyncerSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
	GetDataSyncerSet() ([]NodeInfo, error)

	// GetExitedSet is a free data retrieval call binding the contract method 0x709de02e.
	//
	// Solidity: function GetExitedSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
	GetExitedSet() ([]NodeInfo, error)

	// GetNodeInfo is a free data retrieval call binding the contract method 0x7e6f8582.
	//
	// Solidity: function GetNodeInfo(uint64 nodeID) view returns((uint64,string,string,string,string,(string,string,string,string),uint8) info)
	GetNodeInfo(nodeID uint64) (NodeInfo, error)

	// GetNodeInfos is a free data retrieval call binding the contract method 0xc8c22194.
	//
	// Solidity: function GetNodeInfos(uint64[] nodeIDs) view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info)
	GetNodeInfos(nodeIDs []uint64) ([]NodeInfo, error)

	// GetPendingInactiveSet is a free data retrieval call binding the contract method 0xa3295256.
	//
	// Solidity: function GetPendingInactiveSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
	GetPendingInactiveSet() ([]NodeInfo, error)

	// GetTotalNodeCount is a free data retrieval call binding the contract method 0x1a9efb39.
	//
	// Solidity: function GetTotalNodeCount() view returns(uint64)
	GetTotalNodeCount() (uint64, error)
}
