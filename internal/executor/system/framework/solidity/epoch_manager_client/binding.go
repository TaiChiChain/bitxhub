// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package epoch_manager_client

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

// ConsensusParams is an auto generated low-level Go binding around an user-defined struct.
type ConsensusParams struct {
	ProposerElectionType                               string
	CheckpointPeriod                                   uint64
	HighWatermarkCheckpointPeriod                      uint64
	MaxValidatorNum                                    uint64
	BlockMaxTxNum                                      uint64
	EnableTimedGenEmptyBlock                           bool
	NotActiveWeight                                    int64
	AbnormalNodeExcludeView                            uint64
	AgainProposeIntervalBlockInValidatorsNumPercentage uint64
	ContinuousNullRequestToleranceNumber               uint64
	ReBroadcastToleranceNumber                         uint64
}

// EpochInfo is an auto generated low-level Go binding around an user-defined struct.
type EpochInfo struct {
	Version         uint64
	Epoch           uint64
	EpochPeriod     uint64
	StartBlock      uint64
	ConsensusParams ConsensusParams
	FinanceParams   FinanceParams
	MiscParams      MiscParams
}

// FinanceParams is an auto generated low-level Go binding around an user-defined struct.
type FinanceParams struct {
	GasLimit    uint64
	MinGasPrice uint64
}

// MiscParams is an auto generated low-level Go binding around an user-defined struct.
type MiscParams struct {
	TxMaxSize uint64
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"epochID\",\"type\":\"uint64\"}],\"name\":\"historyEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCaller) CurrentEpoch(opts *bind.CallOpts) (EpochInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractSession) CurrentEpoch() (EpochInfo, error) {
	return _BindingContract.Contract.CurrentEpoch(&_BindingContract.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCallerSession) CurrentEpoch() (EpochInfo, error) {
	return _BindingContract.Contract.CurrentEpoch(&_BindingContract.CallOpts)
}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCaller) HistoryEpoch(opts *bind.CallOpts, epochID uint64) (EpochInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "historyEpoch", epochID)

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractSession) HistoryEpoch(epochID uint64) (EpochInfo, error) {
	return _BindingContract.Contract.HistoryEpoch(&_BindingContract.CallOpts, epochID)
}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCallerSession) HistoryEpoch(epochID uint64) (EpochInfo, error) {
	return _BindingContract.Contract.HistoryEpoch(&_BindingContract.CallOpts, epochID)
}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCaller) NextEpoch(opts *bind.CallOpts) (EpochInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "nextEpoch")

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractSession) NextEpoch() (EpochInfo, error) {
	return _BindingContract.Contract.NextEpoch(&_BindingContract.CallOpts)
}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,(string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,uint64),(uint64)) epochInfo)
func (_BindingContract *BindingContractCallerSession) NextEpoch() (EpochInfo, error) {
	return _BindingContract.Contract.NextEpoch(&_BindingContract.CallOpts)
}
