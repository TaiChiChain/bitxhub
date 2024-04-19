// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package governance

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

// Proposal is an auto generated low-level Go binding around an user-defined struct.
type Proposal struct {
	ID                   uint64
	Type                 uint8
	Strategy             uint8
	Proposer             string
	Title                string
	Desc                 string
	BlockNumber          uint64
	TotalVotes           uint64
	PassVotes            []string
	RejectVotes          []string
	Status               uint8
	Extra                []byte
	CreatedBlockNumber   uint64
	EffectiveBlockNumber uint64
}

type Governance interface {

	// Propose is a paid mutator transaction binding the contract method 0xcee0ffe8.
	//
	// Solidity: function propose(uint8 proposalType, string title, string desc, uint64 deadlineBlockNumber, bytes extra) returns()
	Propose(proposalType uint8, title string, desc string, deadlineBlockNumber uint64, extra []byte) error

	// Vote is a paid mutator transaction binding the contract method 0xb040d166.
	//
	// Solidity: function vote(uint64 proposalID, uint8 voteResult) returns()
	Vote(proposalID uint64, voteResult uint8) error

	// GetLatestProposalID is a free data retrieval call binding the contract method 0x6785bd6e.
	//
	// Solidity: function getLatestProposalID() view returns(uint64)
	GetLatestProposalID() (uint64, error)

	// Proposal is a free data retrieval call binding the contract method 0x7afa0aa3.
	//
	// Solidity: function proposal(uint64 proposalID) view returns((uint64,uint8,uint8,string,string,string,uint64,uint64,string[],string[],uint8,bytes,uint64,uint64) proposal)
	Proposal(proposalID uint64) (Proposal, error)
}

// EventPropose represents a Propose event raised by the Governance contract.
type EventPropose struct {
	ProposalID   uint64
	ProposalType uint8
	Proposer     common.Address
	Proposal     []byte
}

func (_event *EventPropose) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Propose"])
}

// EventVote represents a Vote event raised by the Governance contract.
type EventVote struct {
	ProposalID   uint64
	ProposalType uint8
	Proposer     common.Address
	Proposal     []byte
}

func (_event *EventVote) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Vote"])
}
