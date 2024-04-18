// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package node_manager

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/bind"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = bind.Bind
	_ = common.Big1
	_ = types.AxcUnit
	_ = abi.ConvertType
)

// NodeManagerConsensusVotingPower is an auto generated low-level Go binding around an user-defined struct.
type NodeManagerConsensusVotingPower struct {
	NodeID               *big.Int
	ConsensusVotingPower *big.Int
}

// NodeManagerNodeInfo is an auto generated low-level Go binding around an user-defined struct.
type NodeManagerNodeInfo struct {
	ID              uint64
	NodePubKey      string
	OperatorAddress string
	MetaData        NodeManagerNodeMetaData
	Status          uint8
}

// NodeManagerNodeMetaData is an auto generated low-level Go binding around an user-defined struct.
type NodeManagerNodeMetaData struct {
	Name       string
	Desc       string
	ImageURL   string
	WebsiteURL string
}

type NodeManager interface {

	// JoinCandidate is a paid mutator transaction binding the contract method 0x0b30e031.
	//
	// Solidity: function joinCandidate(uint256 nodeID) returns()
	JoinCandidate(nodeID *big.Int) error

	// LeaveValidatorSet is a paid mutator transaction binding the contract method 0xde85f8cb.
	//
	// Solidity: function leaveValidatorSet(uint256 nodeID) returns()
	LeaveValidatorSet(nodeID *big.Int) error

	// UpdateMetaData is a paid mutator transaction binding the contract method 0x98a850ac.
	//
	// Solidity: function updateMetaData(uint256 nodeID, (string,string,string,string) metaData) returns()
	UpdateMetaData(nodeID *big.Int, metaData NodeManagerNodeMetaData) error

	// UpdateOperator is a paid mutator transaction binding the contract method 0xc08a631c.
	//
	// Solidity: function updateOperator(uint256 nodeID, string newOperatorAddress) returns()
	UpdateOperator(nodeID *big.Int, newOperatorAddress string) error

	// GetActiveValidatorSet is a free data retrieval call binding the contract method 0x59acaac4.
	//
	// Solidity: function GetActiveValidatorSet() view returns((uint64,string,string,(string,string,string,string),uint8)[] info, (uint256,uint256)[] votingPowers)
	GetActiveValidatorSet() (struct {
		Info         []NodeManagerNodeInfo
		VotingPowers []NodeManagerConsensusVotingPower
	}, error)

	// GetCandidateSet is a free data retrieval call binding the contract method 0x84c3e579.
	//
	// Solidity: function GetCandidateSet() view returns((uint64,string,string,(string,string,string,string),uint8)[] infos)
	GetCandidateSet() ([]NodeManagerNodeInfo, error)

	// GetDataSyncerSet is a free data retrieval call binding the contract method 0x834b2b3b.
	//
	// Solidity: function GetDataSyncerSet() view returns((uint64,string,string,(string,string,string,string),uint8)[] infos)
	GetDataSyncerSet() ([]NodeManagerNodeInfo, error)

	// GetExitedSet is a free data retrieval call binding the contract method 0x709de02e.
	//
	// Solidity: function GetExitedSet() view returns((uint64,string,string,(string,string,string,string),uint8)[] infos)
	GetExitedSet() ([]NodeManagerNodeInfo, error)

	// GetNodeInfo is a free data retrieval call binding the contract method 0xeadf26a7.
	//
	// Solidity: function GetNodeInfo(uint256 nodeID) view returns((uint64,string,string,(string,string,string,string),uint8) info)
	GetNodeInfo(nodeID *big.Int) (NodeManagerNodeInfo, error)

	// GetNodeInfos is a free data retrieval call binding the contract method 0xae24b4b2.
	//
	// Solidity: function GetNodeInfos(uint256[] nodeIDs) view returns((uint64,string,string,(string,string,string,string),uint8)[] info)
	GetNodeInfos(nodeIDs []*big.Int) ([]NodeManagerNodeInfo, error)

	// GetPendingInactiveSet is a free data retrieval call binding the contract method 0xa3295256.
	//
	// Solidity: function GetPendingInactiveSet() view returns((uint64,string,string,(string,string,string,string),uint8)[] infos)
	GetPendingInactiveSet() ([]NodeManagerNodeInfo, error)

	// GetTotalNodeCount is a free data retrieval call binding the contract method 0x1a9efb39.
	//
	// Solidity: function GetTotalNodeCount() view returns(uint256)
	GetTotalNodeCount() (*big.Int, error)
}
