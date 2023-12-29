// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package binding

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
	ValidatorElectionType                              string
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
	Version                   uint64
	Epoch                     uint64
	EpochPeriod               uint64
	StartBlock                uint64
	P2PBootstrapNodeAddresses []string
	ConsensusParams           ConsensusParams
	FinanceParams             FinanceParams
	MiscParams                MiscParams
	ValidatorSet              []NodeInfo
	CandidateSet              []NodeInfo
	DataSyncerSet             []NodeInfo
}

// FinanceParams is an auto generated low-level Go binding around an user-defined struct.
type FinanceParams struct {
	GasLimit               uint64
	StartGasPriceAvailable bool
	StartGasPrice          uint64
	MaxGasPrice            uint64
	MinGasPrice            uint64
	GasChangeRateValue     uint64
	GasChangeRateDecimals  uint64
}

// MiscParams is an auto generated low-level Go binding around an user-defined struct.
type MiscParams struct {
	TxMaxSize uint64
}

// NodeInfo is an auto generated low-level Go binding around an user-defined struct.
type NodeInfo struct {
	ID                   uint64
	AccountAddress       string
	P2PNodeID            string
	ConsensusVotingPower int64
}

// EpochManagerMetaData contains all meta data concerning the EpochManager contract.
var EpochManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"internalType\":\"string[]\",\"name\":\"P2PBootstrapNodeAddresses\",\"type\":\"string[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ValidatorElectionType\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"StartGasPriceAvailable\",\"type\":\"bool\"},{\"internalType\":\"uint64\",\"name\":\"StartGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateValue\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateDecimals\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"ValidatorSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"CandidateSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"DataSyncerSet\",\"type\":\"tuple[]\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"epochID\",\"type\":\"uint64\"}],\"name\":\"historyEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"internalType\":\"string[]\",\"name\":\"P2PBootstrapNodeAddresses\",\"type\":\"string[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ValidatorElectionType\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"StartGasPriceAvailable\",\"type\":\"bool\"},{\"internalType\":\"uint64\",\"name\":\"StartGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateValue\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateDecimals\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"ValidatorSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"CandidateSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"DataSyncerSet\",\"type\":\"tuple[]\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextEpoch\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"Version\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"Epoch\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"EpochPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"StartBlock\",\"type\":\"uint64\"},{\"internalType\":\"string[]\",\"name\":\"P2PBootstrapNodeAddresses\",\"type\":\"string[]\"},{\"components\":[{\"internalType\":\"string\",\"name\":\"ValidatorElectionType\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"ProposerElectionType\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"CheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"HighWatermarkCheckpointPeriod\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxValidatorNum\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"BlockMaxTxNum\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"EnableTimedGenEmptyBlock\",\"type\":\"bool\"},{\"internalType\":\"int64\",\"name\":\"NotActiveWeight\",\"type\":\"int64\"},{\"internalType\":\"uint64\",\"name\":\"AbnormalNodeExcludeView\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"AgainProposeIntervalBlockInValidatorsNumPercentage\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ContinuousNullRequestToleranceNumber\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"ReBroadcastToleranceNumber\",\"type\":\"uint64\"}],\"internalType\":\"structConsensusParams\",\"name\":\"ConsensusParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"GasLimit\",\"type\":\"uint64\"},{\"internalType\":\"bool\",\"name\":\"StartGasPriceAvailable\",\"type\":\"bool\"},{\"internalType\":\"uint64\",\"name\":\"StartGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MaxGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"MinGasPrice\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateValue\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"GasChangeRateDecimals\",\"type\":\"uint64\"}],\"internalType\":\"structFinanceParams\",\"name\":\"FinanceParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"TxMaxSize\",\"type\":\"uint64\"}],\"internalType\":\"structMiscParams\",\"name\":\"MiscParams\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"ValidatorSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"CandidateSet\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"ID\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"AccountAddress\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"P2PNodeID\",\"type\":\"string\"},{\"internalType\":\"int64\",\"name\":\"ConsensusVotingPower\",\"type\":\"int64\"}],\"internalType\":\"structNodeInfo[]\",\"name\":\"DataSyncerSet\",\"type\":\"tuple[]\"}],\"internalType\":\"structEpochInfo\",\"name\":\"epochInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// EpochManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use EpochManagerMetaData.ABI instead.
var EpochManagerABI = EpochManagerMetaData.ABI

// EpochManager is an auto generated Go binding around an Ethereum contract.
type EpochManager struct {
	EpochManagerCaller     // Read-only binding to the contract
	EpochManagerTransactor // Write-only binding to the contract
	EpochManagerFilterer   // Log filterer for contract events
}

// EpochManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type EpochManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EpochManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EpochManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EpochManagerSession struct {
	Contract     *EpochManager     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EpochManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EpochManagerCallerSession struct {
	Contract *EpochManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// EpochManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EpochManagerTransactorSession struct {
	Contract     *EpochManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// EpochManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type EpochManagerRaw struct {
	Contract *EpochManager // Generic contract binding to access the raw methods on
}

// EpochManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EpochManagerCallerRaw struct {
	Contract *EpochManagerCaller // Generic read-only contract binding to access the raw methods on
}

// EpochManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EpochManagerTransactorRaw struct {
	Contract *EpochManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEpochManager creates a new instance of EpochManager, bound to a specific deployed contract.
func NewEpochManager(address common.Address, backend bind.ContractBackend) (*EpochManager, error) {
	contract, err := bindEpochManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EpochManager{EpochManagerCaller: EpochManagerCaller{contract: contract}, EpochManagerTransactor: EpochManagerTransactor{contract: contract}, EpochManagerFilterer: EpochManagerFilterer{contract: contract}}, nil
}

// NewEpochManagerCaller creates a new read-only instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerCaller(address common.Address, caller bind.ContractCaller) (*EpochManagerCaller, error) {
	contract, err := bindEpochManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EpochManagerCaller{contract: contract}, nil
}

// NewEpochManagerTransactor creates a new write-only instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*EpochManagerTransactor, error) {
	contract, err := bindEpochManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EpochManagerTransactor{contract: contract}, nil
}

// NewEpochManagerFilterer creates a new log filterer instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*EpochManagerFilterer, error) {
	contract, err := bindEpochManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EpochManagerFilterer{contract: contract}, nil
}

// bindEpochManager binds a generic wrapper to an already deployed contract.
func bindEpochManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EpochManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EpochManager *EpochManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EpochManager.Contract.EpochManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EpochManager *EpochManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.Contract.EpochManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EpochManager *EpochManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EpochManager.Contract.EpochManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EpochManager *EpochManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EpochManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EpochManager *EpochManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EpochManager *EpochManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EpochManager.Contract.contract.Transact(opts, method, params...)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCaller) CurrentEpoch(opts *bind.CallOpts) (EpochInfo, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerSession) CurrentEpoch() (EpochInfo, error) {
	return _EpochManager.Contract.CurrentEpoch(&_EpochManager.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCallerSession) CurrentEpoch() (EpochInfo, error) {
	return _EpochManager.Contract.CurrentEpoch(&_EpochManager.CallOpts)
}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCaller) HistoryEpoch(opts *bind.CallOpts, epochID uint64) (EpochInfo, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "historyEpoch", epochID)

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerSession) HistoryEpoch(epochID uint64) (EpochInfo, error) {
	return _EpochManager.Contract.HistoryEpoch(&_EpochManager.CallOpts, epochID)
}

// HistoryEpoch is a free data retrieval call binding the contract method 0x7ba5c50a.
//
// Solidity: function historyEpoch(uint64 epochID) view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCallerSession) HistoryEpoch(epochID uint64) (EpochInfo, error) {
	return _EpochManager.Contract.HistoryEpoch(&_EpochManager.CallOpts, epochID)
}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCaller) NextEpoch(opts *bind.CallOpts) (EpochInfo, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "nextEpoch")

	if err != nil {
		return *new(EpochInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(EpochInfo)).(*EpochInfo)

	return out0, err

}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerSession) NextEpoch() (EpochInfo, error) {
	return _EpochManager.Contract.NextEpoch(&_EpochManager.CallOpts)
}

// NextEpoch is a free data retrieval call binding the contract method 0xaea0e78b.
//
// Solidity: function nextEpoch() view returns((uint64,uint64,uint64,uint64,string[],(string,string,uint64,uint64,uint64,uint64,bool,int64,uint64,uint64,uint64,uint64),(uint64,bool,uint64,uint64,uint64,uint64,uint64),(uint64),(uint64,string,string,int64)[],(uint64,string,string,int64)[],(uint64,string,string,int64)[]) epochInfo)
func (_EpochManager *EpochManagerCallerSession) NextEpoch() (EpochInfo, error) {
	return _EpochManager.Contract.NextEpoch(&_EpochManager.CallOpts)
}
