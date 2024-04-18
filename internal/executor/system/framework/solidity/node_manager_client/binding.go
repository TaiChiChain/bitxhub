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
	ABI: "[{\"inputs\":[],\"name\":\"GetActiveValidatorSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"info\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"NodeID\",\"type\":\"uint64\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structConsensusVotingPower[]\",\"name\":\"votingPowers\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetCandidateSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetDataSyncerSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetExitedSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"GetNodeInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo\",\"name\":\"info\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64[]\",\"name\":\"nodeIDs\",\"type\":\"uint64[]\"}],\"name\":\"GetNodeInfos\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"info\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetPendingInactiveSet\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"ConsensusPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PPubKey\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PID\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"OperatorAddress\",\"type\":\"string\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"MetaData\",\"type\":\"tuple\"},{\"internalType\":\"enumStatus\",\"name\":\"Status\",\"type\":\"uint8\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GetTotalNodeCount\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"joinCandidateSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"}],\"name\":\"leaveValidatorSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"desc\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"imageURL\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"websiteURL\",\"type\":\"string\"}],\"internalType\":\"structNodeMetaData\",\"name\":\"metaData\",\"type\":\"tuple\"}],\"name\":\"updateMetaData\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"nodeID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"newOperatorAddress\",\"type\":\"string\"}],\"name\":\"updateOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// LeaveValidatorSet is a paid mutator transaction binding the contract method 0xede1f5d1.
//
// Solidity: function leaveValidatorSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactor) LeaveValidatorSet(opts *bind.TransactOpts, nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "leaveValidatorSet", nodeID)
}

// LeaveValidatorSet is a paid mutator transaction binding the contract method 0xede1f5d1.
//
// Solidity: function leaveValidatorSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractSession) LeaveValidatorSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.LeaveValidatorSet(&_BindingContract.TransactOpts, nodeID)
}

// LeaveValidatorSet is a paid mutator transaction binding the contract method 0xede1f5d1.
//
// Solidity: function leaveValidatorSet(uint64 nodeID) returns()
func (_BindingContract *BindingContractTransactorSession) LeaveValidatorSet(nodeID uint64) (*types.Transaction, error) {
	return _BindingContract.Contract.LeaveValidatorSet(&_BindingContract.TransactOpts, nodeID)
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
