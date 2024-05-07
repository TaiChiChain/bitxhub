// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package node_manager_client

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

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"status\",\"type\":\"uint8\"}],\"name\":\"incorrectStatus\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"pendingInactiveSetIsFull\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"JoinedCandidateSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"JoinedPendingInactiveSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"LeavedCandidateSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"indexed\":false,\"internalType\":\"structNodeMetaData\",\"name\":\"metaData\",\"type\":\"tuple\"}],\"name\":\"UpdateMetaData\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"newOperatorAddress\",\"type\":\"string\"}],\"name\":\"UpdateOperator\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"GetActiveValidatorSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"info\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"NodeID\",\"type\":\"uint64\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structConsensusVotingPower[]\",\"name\":\"votingPowers\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetCandidateSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetDataSyncerSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetExitedSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"GetNodeInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo\",\"name\":\"info\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"nodeIDs\",\"type\":\"uint64[]\"}],\"name\":\"GetNodeInfos\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"info\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetPendingInactiveSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetTotalNodeCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"joinCandidateSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"leaveValidatorOrCandidateSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"metaData\",\"type\":\"tuple\"}],\"name\":\"updateMetaData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"newOperatorAddress\",\"type\":\"string\"}],\"name\":\"updateOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// GetActiveValidatorSet is a free data retrieval call binding the contract method 0x59acaac4.
//
// Solidity: function GetActiveValidatorSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info, (uint64,int64)[] votingPowers)
func (_BindingContract *BindingContractCaller) GetActiveValidatorSet(opts *bind.CallOpts) (struct {
	Info         []NodeInfo
	VotingPowers []ConsensusVotingPower
}, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetActiveValidatorSet")

	outstruct := new(struct {
		Info         []NodeInfo
		VotingPowers []ConsensusVotingPower
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Info = *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)
	outstruct.VotingPowers = *abi.ConvertType(out[1], new([]ConsensusVotingPower)).(*[]ConsensusVotingPower)

	return *outstruct, err

}

// GetActiveValidatorSet is a free data retrieval call binding the contract method 0x59acaac4.
//
// Solidity: function GetActiveValidatorSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info, (uint64,int64)[] votingPowers)
func (_BindingContract *BindingContractSession) GetActiveValidatorSet() (struct {
	Info         []NodeInfo
	VotingPowers []ConsensusVotingPower
}, error) {
	return _BindingContract.Contract.GetActiveValidatorSet(&_BindingContract.CallOpts)
}

// GetActiveValidatorSet is a free data retrieval call binding the contract method 0x59acaac4.
//
// Solidity: function GetActiveValidatorSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info, (uint64,int64)[] votingPowers)
func (_BindingContract *BindingContractCallerSession) GetActiveValidatorSet() (struct {
	Info         []NodeInfo
	VotingPowers []ConsensusVotingPower
}, error) {
	return _BindingContract.Contract.GetActiveValidatorSet(&_BindingContract.CallOpts)
}

// GetCandidateSet is a free data retrieval call binding the contract method 0x84c3e579.
//
// Solidity: function GetCandidateSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCaller) GetCandidateSet(opts *bind.CallOpts) ([]NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetCandidateSet")

	if err != nil {
		return *new([]NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)

	return out0, err

}

// GetCandidateSet is a free data retrieval call binding the contract method 0x84c3e579.
//
// Solidity: function GetCandidateSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractSession) GetCandidateSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetCandidateSet(&_BindingContract.CallOpts)
}

// GetCandidateSet is a free data retrieval call binding the contract method 0x84c3e579.
//
// Solidity: function GetCandidateSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCallerSession) GetCandidateSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetCandidateSet(&_BindingContract.CallOpts)
}

// GetDataSyncerSet is a free data retrieval call binding the contract method 0x834b2b3b.
//
// Solidity: function GetDataSyncerSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCaller) GetDataSyncerSet(opts *bind.CallOpts) ([]NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetDataSyncerSet")

	if err != nil {
		return *new([]NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)

	return out0, err

}

// GetDataSyncerSet is a free data retrieval call binding the contract method 0x834b2b3b.
//
// Solidity: function GetDataSyncerSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractSession) GetDataSyncerSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetDataSyncerSet(&_BindingContract.CallOpts)
}

// GetDataSyncerSet is a free data retrieval call binding the contract method 0x834b2b3b.
//
// Solidity: function GetDataSyncerSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCallerSession) GetDataSyncerSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetDataSyncerSet(&_BindingContract.CallOpts)
}

// GetExitedSet is a free data retrieval call binding the contract method 0x709de02e.
//
// Solidity: function GetExitedSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCaller) GetExitedSet(opts *bind.CallOpts) ([]NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetExitedSet")

	if err != nil {
		return *new([]NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)

	return out0, err

}

// GetExitedSet is a free data retrieval call binding the contract method 0x709de02e.
//
// Solidity: function GetExitedSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractSession) GetExitedSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetExitedSet(&_BindingContract.CallOpts)
}

// GetExitedSet is a free data retrieval call binding the contract method 0x709de02e.
//
// Solidity: function GetExitedSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCallerSession) GetExitedSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetExitedSet(&_BindingContract.CallOpts)
}

// GetNodeInfo is a free data retrieval call binding the contract method 0x7e6f8582.
//
// Solidity: function GetNodeInfo(uint64 nodeID) view returns((uint64,string,string,string,string,(string,string,string,string),uint8) info)
func (_BindingContract *BindingContractCaller) GetNodeInfo(opts *bind.CallOpts, nodeID uint64) (NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetNodeInfo", nodeID)

	if err != nil {
		return *new(NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(NodeInfo)).(*NodeInfo)

	return out0, err

}

// GetNodeInfo is a free data retrieval call binding the contract method 0x7e6f8582.
//
// Solidity: function GetNodeInfo(uint64 nodeID) view returns((uint64,string,string,string,string,(string,string,string,string),uint8) info)
func (_BindingContract *BindingContractSession) GetNodeInfo(nodeID uint64) (NodeInfo, error) {
	return _BindingContract.Contract.GetNodeInfo(&_BindingContract.CallOpts, nodeID)
}

// GetNodeInfo is a free data retrieval call binding the contract method 0x7e6f8582.
//
// Solidity: function GetNodeInfo(uint64 nodeID) view returns((uint64,string,string,string,string,(string,string,string,string),uint8) info)
func (_BindingContract *BindingContractCallerSession) GetNodeInfo(nodeID uint64) (NodeInfo, error) {
	return _BindingContract.Contract.GetNodeInfo(&_BindingContract.CallOpts, nodeID)
}

// GetNodeInfos is a free data retrieval call binding the contract method 0xc8c22194.
//
// Solidity: function GetNodeInfos(uint64[] nodeIDs) view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info)
func (_BindingContract *BindingContractCaller) GetNodeInfos(opts *bind.CallOpts, nodeIDs []uint64) ([]NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetNodeInfos", nodeIDs)

	if err != nil {
		return *new([]NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)

	return out0, err

}

// GetNodeInfos is a free data retrieval call binding the contract method 0xc8c22194.
//
// Solidity: function GetNodeInfos(uint64[] nodeIDs) view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info)
func (_BindingContract *BindingContractSession) GetNodeInfos(nodeIDs []uint64) ([]NodeInfo, error) {
	return _BindingContract.Contract.GetNodeInfos(&_BindingContract.CallOpts, nodeIDs)
}

// GetNodeInfos is a free data retrieval call binding the contract method 0xc8c22194.
//
// Solidity: function GetNodeInfos(uint64[] nodeIDs) view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] info)
func (_BindingContract *BindingContractCallerSession) GetNodeInfos(nodeIDs []uint64) ([]NodeInfo, error) {
	return _BindingContract.Contract.GetNodeInfos(&_BindingContract.CallOpts, nodeIDs)
}

// GetPendingInactiveSet is a free data retrieval call binding the contract method 0xa3295256.
//
// Solidity: function GetPendingInactiveSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCaller) GetPendingInactiveSet(opts *bind.CallOpts) ([]NodeInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetPendingInactiveSet")

	if err != nil {
		return *new([]NodeInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]NodeInfo)).(*[]NodeInfo)

	return out0, err

}

// GetPendingInactiveSet is a free data retrieval call binding the contract method 0xa3295256.
//
// Solidity: function GetPendingInactiveSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractSession) GetPendingInactiveSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetPendingInactiveSet(&_BindingContract.CallOpts)
}

// GetPendingInactiveSet is a free data retrieval call binding the contract method 0xa3295256.
//
// Solidity: function GetPendingInactiveSet() view returns((uint64,string,string,string,string,(string,string,string,string),uint8)[] infos)
func (_BindingContract *BindingContractCallerSession) GetPendingInactiveSet() ([]NodeInfo, error) {
	return _BindingContract.Contract.GetPendingInactiveSet(&_BindingContract.CallOpts)
}

// GetTotalNodeCount is a free data retrieval call binding the contract method 0x1a9efb39.
//
// Solidity: function GetTotalNodeCount() view returns(uint64)
func (_BindingContract *BindingContractCaller) GetTotalNodeCount(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "GetTotalNodeCount")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetTotalNodeCount is a free data retrieval call binding the contract method 0x1a9efb39.
//
// Solidity: function GetTotalNodeCount() view returns(uint64)
func (_BindingContract *BindingContractSession) GetTotalNodeCount() (uint64, error) {
	return _BindingContract.Contract.GetTotalNodeCount(&_BindingContract.CallOpts)
}

// GetTotalNodeCount is a free data retrieval call binding the contract method 0x1a9efb39.
//
// Solidity: function GetTotalNodeCount() view returns(uint64)
func (_BindingContract *BindingContractCallerSession) GetTotalNodeCount() (uint64, error) {
	return _BindingContract.Contract.GetTotalNodeCount(&_BindingContract.CallOpts)
}

// JoinCandidateSet is a paid mutator transaction binding the contract method 0x2b27ec44.
//
// Solidity: function joinCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactor) JoinCandidateSet(opts *bind.TransactOpts, nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "joinCandidateSet", nodeID)
}

// JoinCandidateSet is a paid mutator transaction binding the contract method 0x2b27ec44.
//
// Solidity: function joinCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractSession) JoinCandidateSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.JoinCandidateSet(&_BindingContract.TransactOpts, nodeID)
}

// JoinCandidateSet is a paid mutator transaction binding the contract method 0x2b27ec44.
//
// Solidity: function joinCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactorSession) JoinCandidateSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.JoinCandidateSet(&_BindingContract.TransactOpts, nodeID)
}

// LeaveValidatorOrCandidateSet is a paid mutator transaction binding the contract method 0xe2aa7a23.
//
// Solidity: function leaveValidatorOrCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactor) LeaveValidatorOrCandidateSet(opts *bind.TransactOpts, nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "leaveValidatorOrCandidateSet", nodeID)
}

// LeaveValidatorOrCandidateSet is a paid mutator transaction binding the contract method 0xe2aa7a23.
//
// Solidity: function leaveValidatorOrCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractSession) LeaveValidatorOrCandidateSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.LeaveValidatorOrCandidateSet(&_BindingContract.TransactOpts, nodeID)
}

// LeaveValidatorOrCandidateSet is a paid mutator transaction binding the contract method 0xe2aa7a23.
//
// Solidity: function leaveValidatorOrCandidateSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactorSession) LeaveValidatorOrCandidateSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.LeaveValidatorOrCandidateSet(&_BindingContract.TransactOpts, nodeID)
}

// UpdateMetaData is a paid mutator transaction binding the contract method 0xee99437a.
//
// Solidity: function updateMetaData(uint64 nodeID, (string,string,string,string) metaData) returns()
func (_BindingContract *BindingContractTransactor) UpdateMetaData(opts *bind.TransactOpts, nodeID uint64, metaData NodeMetaData) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "updateMetaData", nodeID, metaData)
}

// UpdateMetaData is a paid mutator transaction binding the contract method 0xee99437a.
//
// Solidity: function updateMetaData(uint64 nodeID, (string,string,string,string) metaData) returns()
func (_BindingContract *BindingContractSession) UpdateMetaData(nodeID uint64, metaData NodeMetaData) (*types.Transaction, error) {
	return _BindingContract.Contract.UpdateMetaData(&_BindingContract.TransactOpts, nodeID, metaData)
}

// UpdateMetaData is a paid mutator transaction binding the contract method 0xee99437a.
//
// Solidity: function updateMetaData(uint64 nodeID, (string,string,string,string) metaData) returns()
func (_BindingContract *BindingContractTransactorSession) UpdateMetaData(nodeID uint64, metaData NodeMetaData) (*types.Transaction, error) {
	return _BindingContract.Contract.UpdateMetaData(&_BindingContract.TransactOpts, nodeID, metaData)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xcee4d6f5.
//
// Solidity: function updateOperator(uint64 nodeID, string newOperatorAddress) returns()
func (_BindingContract *BindingContractTransactor) UpdateOperator(opts *bind.TransactOpts, nodeID uint64, newOperatorAddress string) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "updateOperator", nodeID, newOperatorAddress)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xcee4d6f5.
//
// Solidity: function updateOperator(uint64 nodeID, string newOperatorAddress) returns()
func (_BindingContract *BindingContractSession) UpdateOperator(nodeID uint64, newOperatorAddress string) (*types.Transaction, error) {
	return _BindingContract.Contract.UpdateOperator(&_BindingContract.TransactOpts, nodeID, newOperatorAddress)
}

// UpdateOperator is a paid mutator transaction binding the contract method 0xcee4d6f5.
//
// Solidity: function updateOperator(uint64 nodeID, string newOperatorAddress) returns()
func (_BindingContract *BindingContractTransactorSession) UpdateOperator(nodeID uint64, newOperatorAddress string) (*types.Transaction, error) {
	return _BindingContract.Contract.UpdateOperator(&_BindingContract.TransactOpts, nodeID, newOperatorAddress)
}

// BindingContractJoinedCandidateSetIterator is returned from FilterJoinedCandidateSet and is used to iterate over the raw logs and unpacked data for JoinedCandidateSet events raised by the BindingContract contract.
type BindingContractJoinedCandidateSetIterator struct {
	Event *BindingContractJoinedCandidateSet // Event containing the contract specifics and raw log

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
func (it *BindingContractJoinedCandidateSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractJoinedCandidateSet)
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
		it.Event = new(BindingContractJoinedCandidateSet)
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
func (it *BindingContractJoinedCandidateSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractJoinedCandidateSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractJoinedCandidateSet represents a JoinedCandidateSet event raised by the BindingContract contract.
type BindingContractJoinedCandidateSet struct {
	NodeID uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterJoinedCandidateSet is a free log retrieval operation binding the contract event 0xa2c8eb967942364d035cf63cfe32b84bd71472f0a44504053c0a7881236e9419.
//
// Solidity: event JoinedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) FilterJoinedCandidateSet(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractJoinedCandidateSetIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "JoinedCandidateSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractJoinedCandidateSetIterator{contract: _BindingContract.contract, event: "JoinedCandidateSet", logs: logs, sub: sub}, nil
}

// WatchJoinedCandidateSet is a free log subscription operation binding the contract event 0xa2c8eb967942364d035cf63cfe32b84bd71472f0a44504053c0a7881236e9419.
//
// Solidity: event JoinedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) WatchJoinedCandidateSet(opts *bind.WatchOpts, sink chan<- *BindingContractJoinedCandidateSet, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "JoinedCandidateSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractJoinedCandidateSet)
				if err := _BindingContract.contract.UnpackLog(event, "JoinedCandidateSet", log); err != nil {
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

// ParseJoinedCandidateSet is a log parse operation binding the contract event 0xa2c8eb967942364d035cf63cfe32b84bd71472f0a44504053c0a7881236e9419.
//
// Solidity: event JoinedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) ParseJoinedCandidateSet(log types.Log) (*BindingContractJoinedCandidateSet, error) {
	event := new(BindingContractJoinedCandidateSet)
	if err := _BindingContract.contract.UnpackLog(event, "JoinedCandidateSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractJoinedPendingInactiveSetIterator is returned from FilterJoinedPendingInactiveSet and is used to iterate over the raw logs and unpacked data for JoinedPendingInactiveSet events raised by the BindingContract contract.
type BindingContractJoinedPendingInactiveSetIterator struct {
	Event *BindingContractJoinedPendingInactiveSet // Event containing the contract specifics and raw log

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
func (it *BindingContractJoinedPendingInactiveSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractJoinedPendingInactiveSet)
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
		it.Event = new(BindingContractJoinedPendingInactiveSet)
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
func (it *BindingContractJoinedPendingInactiveSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractJoinedPendingInactiveSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractJoinedPendingInactiveSet represents a JoinedPendingInactiveSet event raised by the BindingContract contract.
type BindingContractJoinedPendingInactiveSet struct {
	NodeID uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterJoinedPendingInactiveSet is a free log retrieval operation binding the contract event 0x273f44946e2870fd0742ac56cd27a3d4b0741c914f9a5798c57011283ded0f4d.
//
// Solidity: event JoinedPendingInactiveSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) FilterJoinedPendingInactiveSet(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractJoinedPendingInactiveSetIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "JoinedPendingInactiveSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractJoinedPendingInactiveSetIterator{contract: _BindingContract.contract, event: "JoinedPendingInactiveSet", logs: logs, sub: sub}, nil
}

// WatchJoinedPendingInactiveSet is a free log subscription operation binding the contract event 0x273f44946e2870fd0742ac56cd27a3d4b0741c914f9a5798c57011283ded0f4d.
//
// Solidity: event JoinedPendingInactiveSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) WatchJoinedPendingInactiveSet(opts *bind.WatchOpts, sink chan<- *BindingContractJoinedPendingInactiveSet, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "JoinedPendingInactiveSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractJoinedPendingInactiveSet)
				if err := _BindingContract.contract.UnpackLog(event, "JoinedPendingInactiveSet", log); err != nil {
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

// ParseJoinedPendingInactiveSet is a log parse operation binding the contract event 0x273f44946e2870fd0742ac56cd27a3d4b0741c914f9a5798c57011283ded0f4d.
//
// Solidity: event JoinedPendingInactiveSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) ParseJoinedPendingInactiveSet(log types.Log) (*BindingContractJoinedPendingInactiveSet, error) {
	event := new(BindingContractJoinedPendingInactiveSet)
	if err := _BindingContract.contract.UnpackLog(event, "JoinedPendingInactiveSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractLeavedCandidateSetIterator is returned from FilterLeavedCandidateSet and is used to iterate over the raw logs and unpacked data for LeavedCandidateSet events raised by the BindingContract contract.
type BindingContractLeavedCandidateSetIterator struct {
	Event *BindingContractLeavedCandidateSet // Event containing the contract specifics and raw log

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
func (it *BindingContractLeavedCandidateSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractLeavedCandidateSet)
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
		it.Event = new(BindingContractLeavedCandidateSet)
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
func (it *BindingContractLeavedCandidateSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractLeavedCandidateSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractLeavedCandidateSet represents a LeavedCandidateSet event raised by the BindingContract contract.
type BindingContractLeavedCandidateSet struct {
	NodeID uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterLeavedCandidateSet is a free log retrieval operation binding the contract event 0xb09a084aa1219ff08727ae777592b11ab6f8ebea0e3870d79d50f6165cca0537.
//
// Solidity: event LeavedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) FilterLeavedCandidateSet(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractLeavedCandidateSetIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "LeavedCandidateSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractLeavedCandidateSetIterator{contract: _BindingContract.contract, event: "LeavedCandidateSet", logs: logs, sub: sub}, nil
}

// WatchLeavedCandidateSet is a free log subscription operation binding the contract event 0xb09a084aa1219ff08727ae777592b11ab6f8ebea0e3870d79d50f6165cca0537.
//
// Solidity: event LeavedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) WatchLeavedCandidateSet(opts *bind.WatchOpts, sink chan<- *BindingContractLeavedCandidateSet, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "LeavedCandidateSet", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractLeavedCandidateSet)
				if err := _BindingContract.contract.UnpackLog(event, "LeavedCandidateSet", log); err != nil {
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

// ParseLeavedCandidateSet is a log parse operation binding the contract event 0xb09a084aa1219ff08727ae777592b11ab6f8ebea0e3870d79d50f6165cca0537.
//
// Solidity: event LeavedCandidateSet(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) ParseLeavedCandidateSet(log types.Log) (*BindingContractLeavedCandidateSet, error) {
	event := new(BindingContractLeavedCandidateSet)
	if err := _BindingContract.contract.UnpackLog(event, "LeavedCandidateSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractRegisteredIterator is returned from FilterRegistered and is used to iterate over the raw logs and unpacked data for Registered events raised by the BindingContract contract.
type BindingContractRegisteredIterator struct {
	Event *BindingContractRegistered // Event containing the contract specifics and raw log

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
func (it *BindingContractRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractRegistered)
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
		it.Event = new(BindingContractRegistered)
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
func (it *BindingContractRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractRegistered represents a Registered event raised by the BindingContract contract.
type BindingContractRegistered struct {
	NodeID uint64
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRegistered is a free log retrieval operation binding the contract event 0xd13e027f8072af75628a0ccb98abc6ad980639af9497f2ee7b8ad1d0ab6385e1.
//
// Solidity: event Registered(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) FilterRegistered(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractRegisteredIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Registered", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractRegisteredIterator{contract: _BindingContract.contract, event: "Registered", logs: logs, sub: sub}, nil
}

// WatchRegistered is a free log subscription operation binding the contract event 0xd13e027f8072af75628a0ccb98abc6ad980639af9497f2ee7b8ad1d0ab6385e1.
//
// Solidity: event Registered(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) WatchRegistered(opts *bind.WatchOpts, sink chan<- *BindingContractRegistered, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Registered", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractRegistered)
				if err := _BindingContract.contract.UnpackLog(event, "Registered", log); err != nil {
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

// ParseRegistered is a log parse operation binding the contract event 0xd13e027f8072af75628a0ccb98abc6ad980639af9497f2ee7b8ad1d0ab6385e1.
//
// Solidity: event Registered(uint64 indexed nodeID)
func (_BindingContract *BindingContractFilterer) ParseRegistered(log types.Log) (*BindingContractRegistered, error) {
	event := new(BindingContractRegistered)
	if err := _BindingContract.contract.UnpackLog(event, "Registered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractUpdateMetaDataIterator is returned from FilterUpdateMetaData and is used to iterate over the raw logs and unpacked data for UpdateMetaData events raised by the BindingContract contract.
type BindingContractUpdateMetaDataIterator struct {
	Event *BindingContractUpdateMetaData // Event containing the contract specifics and raw log

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
func (it *BindingContractUpdateMetaDataIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUpdateMetaData)
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
		it.Event = new(BindingContractUpdateMetaData)
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
func (it *BindingContractUpdateMetaDataIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUpdateMetaDataIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUpdateMetaData represents a UpdateMetaData event raised by the BindingContract contract.
type BindingContractUpdateMetaData struct {
	NodeID   uint64
	MetaData NodeMetaData
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterUpdateMetaData is a free log retrieval operation binding the contract event 0x808f1ed07a386a6f6e717905ba5391e9fa1b8bbad1da2374216aa7a1bbd85a15.
//
// Solidity: event UpdateMetaData(uint64 indexed nodeID, (string,string,string,string) metaData)
func (_BindingContract *BindingContractFilterer) FilterUpdateMetaData(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractUpdateMetaDataIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "UpdateMetaData", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractUpdateMetaDataIterator{contract: _BindingContract.contract, event: "UpdateMetaData", logs: logs, sub: sub}, nil
}

// WatchUpdateMetaData is a free log subscription operation binding the contract event 0x808f1ed07a386a6f6e717905ba5391e9fa1b8bbad1da2374216aa7a1bbd85a15.
//
// Solidity: event UpdateMetaData(uint64 indexed nodeID, (string,string,string,string) metaData)
func (_BindingContract *BindingContractFilterer) WatchUpdateMetaData(opts *bind.WatchOpts, sink chan<- *BindingContractUpdateMetaData, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "UpdateMetaData", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUpdateMetaData)
				if err := _BindingContract.contract.UnpackLog(event, "UpdateMetaData", log); err != nil {
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

// ParseUpdateMetaData is a log parse operation binding the contract event 0x808f1ed07a386a6f6e717905ba5391e9fa1b8bbad1da2374216aa7a1bbd85a15.
//
// Solidity: event UpdateMetaData(uint64 indexed nodeID, (string,string,string,string) metaData)
func (_BindingContract *BindingContractFilterer) ParseUpdateMetaData(log types.Log) (*BindingContractUpdateMetaData, error) {
	event := new(BindingContractUpdateMetaData)
	if err := _BindingContract.contract.UnpackLog(event, "UpdateMetaData", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractUpdateOperatorIterator is returned from FilterUpdateOperator and is used to iterate over the raw logs and unpacked data for UpdateOperator events raised by the BindingContract contract.
type BindingContractUpdateOperatorIterator struct {
	Event *BindingContractUpdateOperator // Event containing the contract specifics and raw log

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
func (it *BindingContractUpdateOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUpdateOperator)
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
		it.Event = new(BindingContractUpdateOperator)
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
func (it *BindingContractUpdateOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUpdateOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUpdateOperator represents a UpdateOperator event raised by the BindingContract contract.
type BindingContractUpdateOperator struct {
	NodeID             uint64
	NewOperatorAddress string
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterUpdateOperator is a free log retrieval operation binding the contract event 0x5d84f8a70af0f297030b553a76963d8084018a96aec001413bfbcff1d2be51b0.
//
// Solidity: event UpdateOperator(uint64 indexed nodeID, string newOperatorAddress)
func (_BindingContract *BindingContractFilterer) FilterUpdateOperator(opts *bind.FilterOpts, nodeID []uint64) (*BindingContractUpdateOperatorIterator, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "UpdateOperator", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractUpdateOperatorIterator{contract: _BindingContract.contract, event: "UpdateOperator", logs: logs, sub: sub}, nil
}

// WatchUpdateOperator is a free log subscription operation binding the contract event 0x5d84f8a70af0f297030b553a76963d8084018a96aec001413bfbcff1d2be51b0.
//
// Solidity: event UpdateOperator(uint64 indexed nodeID, string newOperatorAddress)
func (_BindingContract *BindingContractFilterer) WatchUpdateOperator(opts *bind.WatchOpts, sink chan<- *BindingContractUpdateOperator, nodeID []uint64) (event.Subscription, error) {

	var nodeIDRule []interface{}
	for _, nodeIDItem := range nodeID {
		nodeIDRule = append(nodeIDRule, nodeIDItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "UpdateOperator", nodeIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUpdateOperator)
				if err := _BindingContract.contract.UnpackLog(event, "UpdateOperator", log); err != nil {
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

// ParseUpdateOperator is a log parse operation binding the contract event 0x5d84f8a70af0f297030b553a76963d8084018a96aec001413bfbcff1d2be51b0.
//
// Solidity: event UpdateOperator(uint64 indexed nodeID, string newOperatorAddress)
func (_BindingContract *BindingContractFilterer) ParseUpdateOperator(log types.Log) (*BindingContractUpdateOperator, error) {
	event := new(BindingContractUpdateOperator)
	if err := _BindingContract.contract.UnpackLog(event, "UpdateOperator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
