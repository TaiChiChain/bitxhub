// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package governance_client

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// CouncilMember is an auto generated low-level Go binding around an user-defined struct.
type CouncilMember struct {
	Addr   common.Address
	Weight uint64
	Name   string
}

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
	ExecuteSuccess       bool
	ExecuteFailedMsg     string
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"proposalID\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"enumProposalType\",\"name\":\"proposalType\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"proposal\",\"type\":\"bytes\"}],\"name\":\"Propose\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"proposalID\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"enumProposalType\",\"name\":\"proposalType\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"proposal\",\"type\":\"bytes\"}],\"name\":\"Vote\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"getCouncilMembers\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"weight\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"internalType\":\"structCouncilMember[]\",\"name\":\"members\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestProposalID\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"proposalID\",\"type\":\"uint64\"}],\"name\":\"proposal\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"enumProposalType\",\"name\":\"Type\",\"type\":\"uint8\"},{\"internalType\":\"enumProposalStrategy\",\"name\":\"Strategy\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"Proposer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"Title\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"Desc\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"BlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"TotalVotes\",\"type\":\"uint64\"},{\"internalType\":\"string[]\",\"name\":\"PassVotes\",\"type\":\"string[]\"},{\"internalType\":\"string[]\",\"name\":\"RejectVotes\",\"type\":\"string[]\"},{\"internalType\":\"enumProposalStatus\",\"name\":\"Status\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"Extra\",\"type\":\"bytes\"},{\"internalType\":\"uint64\",\"name\":\"CreatedBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EffectiveBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"ExecuteSuccess\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"ExecuteFailedMsg\",\"type\":\"string\"}],\"internalType\":\"structProposal\",\"name\":\"proposal\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"enumProposalType\",\"name\":\"proposalType\",\"type\":\"uint8\"},{\"internalType\":\"string\",\"name\":\"title\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"deadlineBlockNumber\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"extra\",\"type\":\"bytes\"}],\"name\":\"propose\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"proposalID\",\"type\":\"uint64\"},{\"internalType\":\"enumVoteResult\",\"name\":\"voteResult\",\"type\":\"uint8\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BindingContractABI is the input ABI used to generate the binding from.
// Deprecated: Use BindingContractMetaData.ABI instead.
var BindingContractABI = BindingContractMetaData.ABI

// BindingContract is an auto generated Go binding around an Ethereum contract.
type BindingContract struct {
	BindingContractCaller     // Read-only binding to the contract
	BindingContractTransactor // Write-only binding to the contract
	BindingContractFilterer   // Log filterer for contract events
}

// BindingContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type BindingContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BindingContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BindingContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BindingContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BindingContractSession struct {
	Contract     *BindingContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BindingContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BindingContractCallerSession struct {
	Contract *BindingContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// BindingContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BindingContractTransactorSession struct {
	Contract     *BindingContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// BindingContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type BindingContractRaw struct {
	Contract *BindingContract // Generic contract binding to access the raw methods on
}

// BindingContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BindingContractCallerRaw struct {
	Contract *BindingContractCaller // Generic read-only contract binding to access the raw methods on
}

// BindingContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BindingContractTransactorRaw struct {
	Contract *BindingContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBindingContract creates a new instance of BindingContract, bound to a specific deployed contract.
func NewBindingContract(address common.Address, backend bind.ContractBackend) (*BindingContract, error) {
	contract, err := bindBindingContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BindingContract{BindingContractCaller: BindingContractCaller{contract: contract}, BindingContractTransactor: BindingContractTransactor{contract: contract}, BindingContractFilterer: BindingContractFilterer{contract: contract}}, nil
}

// NewBindingContractCaller creates a new read-only instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractCaller(address common.Address, caller bind.ContractCaller) (*BindingContractCaller, error) {
	contract, err := bindBindingContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BindingContractCaller{contract: contract}, nil
}

// NewBindingContractTransactor creates a new write-only instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractTransactor(address common.Address, transactor bind.ContractTransactor) (*BindingContractTransactor, error) {
	contract, err := bindBindingContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BindingContractTransactor{contract: contract}, nil
}

// NewBindingContractFilterer creates a new log filterer instance of BindingContract, bound to a specific deployed contract.
func NewBindingContractFilterer(address common.Address, filterer bind.ContractFilterer) (*BindingContractFilterer, error) {
	contract, err := bindBindingContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BindingContractFilterer{contract: contract}, nil
}

// bindBindingContract binds a generic wrapper to an already deployed contract.
func bindBindingContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BindingContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BindingContract *BindingContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BindingContract.Contract.BindingContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BindingContract *BindingContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.Contract.BindingContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BindingContract *BindingContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BindingContract.Contract.BindingContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BindingContract *BindingContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BindingContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BindingContract *BindingContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BindingContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BindingContract *BindingContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BindingContract.Contract.contract.Transact(opts, method, params...)
}

// GetCouncilMembers is a free data retrieval call binding the contract method 0x606a6b76.
//
// Solidity: function getCouncilMembers() view returns((address,uint64,string)[] members)
func (_BindingContract *BindingContractCaller) GetCouncilMembers(opts *bind.CallOpts) ([]CouncilMember, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getCouncilMembers")

	if err != nil {
		return *new([]CouncilMember), err
	}

	out0 := *abi.ConvertType(out[0], new([]CouncilMember)).(*[]CouncilMember)

	return out0, err

}

// GetCouncilMembers is a free data retrieval call binding the contract method 0x606a6b76.
//
// Solidity: function getCouncilMembers() view returns((address,uint64,string)[] members)
func (_BindingContract *BindingContractSession) GetCouncilMembers() ([]CouncilMember, error) {
	return _BindingContract.Contract.GetCouncilMembers(&_BindingContract.CallOpts)
}

// GetCouncilMembers is a free data retrieval call binding the contract method 0x606a6b76.
//
// Solidity: function getCouncilMembers() view returns((address,uint64,string)[] members)
func (_BindingContract *BindingContractCallerSession) GetCouncilMembers() ([]CouncilMember, error) {
	return _BindingContract.Contract.GetCouncilMembers(&_BindingContract.CallOpts)
}

// GetLatestProposalID is a free data retrieval call binding the contract method 0x6785bd6e.
//
// Solidity: function getLatestProposalID() view returns(uint64)
func (_BindingContract *BindingContractCaller) GetLatestProposalID(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getLatestProposalID")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetLatestProposalID is a free data retrieval call binding the contract method 0x6785bd6e.
//
// Solidity: function getLatestProposalID() view returns(uint64)
func (_BindingContract *BindingContractSession) GetLatestProposalID() (uint64, error) {
	return _BindingContract.Contract.GetLatestProposalID(&_BindingContract.CallOpts)
}

// GetLatestProposalID is a free data retrieval call binding the contract method 0x6785bd6e.
//
// Solidity: function getLatestProposalID() view returns(uint64)
func (_BindingContract *BindingContractCallerSession) GetLatestProposalID() (uint64, error) {
	return _BindingContract.Contract.GetLatestProposalID(&_BindingContract.CallOpts)
}

// Proposal is a free data retrieval call binding the contract method 0x7afa0aa3.
//
// Solidity: function proposal(uint64 proposalID) view returns((uint64,uint8,uint8,string,string,string,uint64,uint64,string[],string[],uint8,bytes,uint64,uint64,bool,string) proposal)
func (_BindingContract *BindingContractCaller) Proposal(opts *bind.CallOpts, proposalID uint64) (Proposal, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "proposal", proposalID)

	if err != nil {
		return *new(Proposal), err
	}

	out0 := *abi.ConvertType(out[0], new(Proposal)).(*Proposal)

	return out0, err

}

// Proposal is a free data retrieval call binding the contract method 0x7afa0aa3.
//
// Solidity: function proposal(uint64 proposalID) view returns((uint64,uint8,uint8,string,string,string,uint64,uint64,string[],string[],uint8,bytes,uint64,uint64,bool,string) proposal)
func (_BindingContract *BindingContractSession) Proposal(proposalID uint64) (Proposal, error) {
	return _BindingContract.Contract.Proposal(&_BindingContract.CallOpts, proposalID)
}

// Proposal is a free data retrieval call binding the contract method 0x7afa0aa3.
//
// Solidity: function proposal(uint64 proposalID) view returns((uint64,uint8,uint8,string,string,string,uint64,uint64,string[],string[],uint8,bytes,uint64,uint64,bool,string) proposal)
func (_BindingContract *BindingContractCallerSession) Proposal(proposalID uint64) (Proposal, error) {
	return _BindingContract.Contract.Proposal(&_BindingContract.CallOpts, proposalID)
}

// Propose is a paid mutator transaction binding the contract method 0xcee0ffe8.
//
// Solidity: function propose(uint8 proposalType, string title, string desc, uint64 deadlineBlockNumber, bytes extra) returns()
func (_BindingContract *BindingContractTransactor) Propose(opts *bind.TransactOpts, proposalType uint8, title string, desc string, deadlineBlockNumber uint64, extra []byte) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "propose", proposalType, title, desc, deadlineBlockNumber, extra)
}

// Propose is a paid mutator transaction binding the contract method 0xcee0ffe8.
//
// Solidity: function propose(uint8 proposalType, string title, string desc, uint64 deadlineBlockNumber, bytes extra) returns()
func (_BindingContract *BindingContractSession) Propose(proposalType uint8, title string, desc string, deadlineBlockNumber uint64, extra []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.Propose(&_BindingContract.TransactOpts, proposalType, title, desc, deadlineBlockNumber, extra)
}

// Propose is a paid mutator transaction binding the contract method 0xcee0ffe8.
//
// Solidity: function propose(uint8 proposalType, string title, string desc, uint64 deadlineBlockNumber, bytes extra) returns()
func (_BindingContract *BindingContractTransactorSession) Propose(proposalType uint8, title string, desc string, deadlineBlockNumber uint64, extra []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.Propose(&_BindingContract.TransactOpts, proposalType, title, desc, deadlineBlockNumber, extra)
}

// Vote is a paid mutator transaction binding the contract method 0xb040d166.
//
// Solidity: function vote(uint64 proposalID, uint8 voteResult) returns()
func (_BindingContract *BindingContractTransactor) Vote(opts *bind.TransactOpts, proposalID uint64, voteResult uint8) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "vote", proposalID, voteResult)
}

// Vote is a paid mutator transaction binding the contract method 0xb040d166.
//
// Solidity: function vote(uint64 proposalID, uint8 voteResult) returns()
func (_BindingContract *BindingContractSession) Vote(proposalID uint64, voteResult uint8) (*types.Transaction, error) {
	return _BindingContract.Contract.Vote(&_BindingContract.TransactOpts, proposalID, voteResult)
}

// Vote is a paid mutator transaction binding the contract method 0xb040d166.
//
// Solidity: function vote(uint64 proposalID, uint8 voteResult) returns()
func (_BindingContract *BindingContractTransactorSession) Vote(proposalID uint64, voteResult uint8) (*types.Transaction, error) {
	return _BindingContract.Contract.Vote(&_BindingContract.TransactOpts, proposalID, voteResult)
}

// BindingContractProposeIterator is returned from FilterPropose and is used to iterate over the raw logs and unpacked data for Propose events raised by the BindingContract contract.
type BindingContractProposeIterator struct {
	Event *BindingContractPropose // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BindingContractProposeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractPropose)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BindingContractPropose)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BindingContractProposeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractProposeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractPropose represents a Propose event raised by the BindingContract contract.
type BindingContractPropose struct {
	ProposalID   uint64
	ProposalType uint8
	Proposer     common.Address
	Proposal     []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterPropose is a free log retrieval operation binding the contract event 0x59ce0d12a0cab5885441796f8e5bf1d1dde7cd00f7d33e3956af71e0018644e0.
//
// Solidity: event Propose(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) FilterPropose(opts *bind.FilterOpts, proposalID []uint64, proposalType []uint8, proposer []common.Address) (*BindingContractProposeIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}
	var proposalTypeRule []interface{}
	for _, proposalTypeItem := range proposalType {
		proposalTypeRule = append(proposalTypeRule, proposalTypeItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Propose", proposalIDRule, proposalTypeRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractProposeIterator{contract: _BindingContract.contract, event: "Propose", logs: logs, sub: sub}, nil
}

// WatchPropose is a free log subscription operation binding the contract event 0x59ce0d12a0cab5885441796f8e5bf1d1dde7cd00f7d33e3956af71e0018644e0.
//
// Solidity: event Propose(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) WatchPropose(opts *bind.WatchOpts, sink chan<- *BindingContractPropose, proposalID []uint64, proposalType []uint8, proposer []common.Address) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}
	var proposalTypeRule []interface{}
	for _, proposalTypeItem := range proposalType {
		proposalTypeRule = append(proposalTypeRule, proposalTypeItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Propose", proposalIDRule, proposalTypeRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractPropose)
				if err := _BindingContract.contract.UnpackLog(event, "Propose", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePropose is a log parse operation binding the contract event 0x59ce0d12a0cab5885441796f8e5bf1d1dde7cd00f7d33e3956af71e0018644e0.
//
// Solidity: event Propose(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) ParsePropose(log types.Log) (*BindingContractPropose, error) {
	event := new(BindingContractPropose)
	if err := _BindingContract.contract.UnpackLog(event, "Propose", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractVoteIterator is returned from FilterVote and is used to iterate over the raw logs and unpacked data for Vote events raised by the BindingContract contract.
type BindingContractVoteIterator struct {
	Event *BindingContractVote // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BindingContractVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractVote)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BindingContractVote)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BindingContractVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractVote represents a Vote event raised by the BindingContract contract.
type BindingContractVote struct {
	ProposalID   uint64
	ProposalType uint8
	Proposer     common.Address
	Proposal     []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterVote is a free log retrieval operation binding the contract event 0x526c6bb2256154a6052064cb768d67c2d39afa8f7d7967e8b79d392f33079249.
//
// Solidity: event Vote(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) FilterVote(opts *bind.FilterOpts, proposalID []uint64, proposalType []uint8, proposer []common.Address) (*BindingContractVoteIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}
	var proposalTypeRule []interface{}
	for _, proposalTypeItem := range proposalType {
		proposalTypeRule = append(proposalTypeRule, proposalTypeItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Vote", proposalIDRule, proposalTypeRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractVoteIterator{contract: _BindingContract.contract, event: "Vote", logs: logs, sub: sub}, nil
}

// WatchVote is a free log subscription operation binding the contract event 0x526c6bb2256154a6052064cb768d67c2d39afa8f7d7967e8b79d392f33079249.
//
// Solidity: event Vote(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) WatchVote(opts *bind.WatchOpts, sink chan<- *BindingContractVote, proposalID []uint64, proposalType []uint8, proposer []common.Address) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}
	var proposalTypeRule []interface{}
	for _, proposalTypeItem := range proposalType {
		proposalTypeRule = append(proposalTypeRule, proposalTypeItem)
	}
	var proposerRule []interface{}
	for _, proposerItem := range proposer {
		proposerRule = append(proposerRule, proposerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Vote", proposalIDRule, proposalTypeRule, proposerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractVote)
				if err := _BindingContract.contract.UnpackLog(event, "Vote", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVote is a log parse operation binding the contract event 0x526c6bb2256154a6052064cb768d67c2d39afa8f7d7967e8b79d392f33079249.
//
// Solidity: event Vote(uint64 indexed proposalID, uint8 indexed proposalType, address indexed proposer, bytes proposal)
func (_BindingContract *BindingContractFilterer) ParseVote(log types.Log) (*BindingContractVote, error) {
	event := new(BindingContractVote)
	if err := _BindingContract.contract.UnpackLog(event, "Vote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
