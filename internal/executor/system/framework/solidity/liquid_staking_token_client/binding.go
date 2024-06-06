// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package liquid_staking_token_client

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

// LiquidStakingTokenInfo is an auto generated low-level Go binding around an user-defined struct.
type LiquidStakingTokenInfo struct {
	PoolID           uint64
	Principal        *big.Int
	Unlocked         *big.Int
	ActiveEpoch      uint64
	UnlockingRecords []UnlockingRecord
}

// UnlockingRecord is an auto generated low-level Go binding around an user-defined struct.
type UnlockingRecord struct {
	Amount          *big.Int
	UnlockTimestamp uint64
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_newPrincipal\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_newUnlocked\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"_newActiveEpoch\",\"type\":\"uint64\"}],\"name\":\"UpdateInfo\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"PoolID\",\"type\":\"uint64\"},{\"internalType\":\"uint256\",\"name\":\"Principal\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Unlocked\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"ActiveEpoch\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"Amount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"UnlockTimestamp\",\"type\":\"uint64\"}],\"internalType\":\"structUnlockingRecord[]\",\"name\":\"UnlockingRecords\",\"type\":\"tuple[]\"}],\"internalType\":\"structLiquidStakingTokenInfo\",\"name\":\"info\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"_tokenIds\",\"type\":\"uint256[]\"}],\"name\":\"getInfos\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"PoolID\",\"type\":\"uint64\"},{\"internalType\":\"uint256\",\"name\":\"Principal\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Unlocked\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"ActiveEpoch\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"Amount\",\"type\":\"uint256\"},{\"internalType\":\"uint64\",\"name\":\"UnlockTimestamp\",\"type\":\"uint64\"}],\"internalType\":\"structUnlockingRecord[]\",\"name\":\"UnlockingRecords\",\"type\":\"tuple[]\"}],\"internalType\":\"structLiquidStakingTokenInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getLockedCoin\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getLockedReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getTotalCoin\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getUnlockedCoin\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"getUnlockingCoin\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"_approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_BindingContract *BindingContractCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "balanceOf", _owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_BindingContract *BindingContractSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _BindingContract.Contract.BalanceOf(&_BindingContract.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _BindingContract.Contract.BalanceOf(&_BindingContract.CallOpts, _owner)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractCaller) GetApproved(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getApproved", _tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _BindingContract.Contract.GetApproved(&_BindingContract.CallOpts, _tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractCallerSession) GetApproved(_tokenId *big.Int) (common.Address, error) {
	return _BindingContract.Contract.GetApproved(&_BindingContract.CallOpts, _tokenId)
}

// GetInfo is a free data retrieval call binding the contract method 0x1a3cd59a.
//
// Solidity: function getInfo(uint256 _tokenId) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[]) info)
func (_BindingContract *BindingContractCaller) GetInfo(opts *bind.CallOpts, _tokenId *big.Int) (LiquidStakingTokenInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getInfo", _tokenId)

	if err != nil {
		return *new(LiquidStakingTokenInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(LiquidStakingTokenInfo)).(*LiquidStakingTokenInfo)

	return out0, err

}

// GetInfo is a free data retrieval call binding the contract method 0x1a3cd59a.
//
// Solidity: function getInfo(uint256 _tokenId) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[]) info)
func (_BindingContract *BindingContractSession) GetInfo(_tokenId *big.Int) (LiquidStakingTokenInfo, error) {
	return _BindingContract.Contract.GetInfo(&_BindingContract.CallOpts, _tokenId)
}

// GetInfo is a free data retrieval call binding the contract method 0x1a3cd59a.
//
// Solidity: function getInfo(uint256 _tokenId) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[]) info)
func (_BindingContract *BindingContractCallerSession) GetInfo(_tokenId *big.Int) (LiquidStakingTokenInfo, error) {
	return _BindingContract.Contract.GetInfo(&_BindingContract.CallOpts, _tokenId)
}

// GetInfos is a free data retrieval call binding the contract method 0xb8944000.
//
// Solidity: function getInfos(uint256[] _tokenIds) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[])[] infos)
func (_BindingContract *BindingContractCaller) GetInfos(opts *bind.CallOpts, _tokenIds []*big.Int) ([]LiquidStakingTokenInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getInfos", _tokenIds)

	if err != nil {
		return *new([]LiquidStakingTokenInfo), err
	}

	out0 := *abi.ConvertType(out[0], new([]LiquidStakingTokenInfo)).(*[]LiquidStakingTokenInfo)

	return out0, err

}

// GetInfos is a free data retrieval call binding the contract method 0xb8944000.
//
// Solidity: function getInfos(uint256[] _tokenIds) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[])[] infos)
func (_BindingContract *BindingContractSession) GetInfos(_tokenIds []*big.Int) ([]LiquidStakingTokenInfo, error) {
	return _BindingContract.Contract.GetInfos(&_BindingContract.CallOpts, _tokenIds)
}

// GetInfos is a free data retrieval call binding the contract method 0xb8944000.
//
// Solidity: function getInfos(uint256[] _tokenIds) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[])[] infos)
func (_BindingContract *BindingContractCallerSession) GetInfos(_tokenIds []*big.Int) ([]LiquidStakingTokenInfo, error) {
	return _BindingContract.Contract.GetInfos(&_BindingContract.CallOpts, _tokenIds)
}

// GetLockedCoin is a free data retrieval call binding the contract method 0xd5691067.
//
// Solidity: function getLockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCaller) GetLockedCoin(opts *bind.CallOpts, _tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getLockedCoin", _tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLockedCoin is a free data retrieval call binding the contract method 0xd5691067.
//
// Solidity: function getLockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractSession) GetLockedCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetLockedCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetLockedCoin is a free data retrieval call binding the contract method 0xd5691067.
//
// Solidity: function getLockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetLockedCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetLockedCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetLockedReward is a free data retrieval call binding the contract method 0x803ea356.
//
// Solidity: function getLockedReward(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCaller) GetLockedReward(opts *bind.CallOpts, _tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getLockedReward", _tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLockedReward is a free data retrieval call binding the contract method 0x803ea356.
//
// Solidity: function getLockedReward(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractSession) GetLockedReward(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetLockedReward(&_BindingContract.CallOpts, _tokenId)
}

// GetLockedReward is a free data retrieval call binding the contract method 0x803ea356.
//
// Solidity: function getLockedReward(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetLockedReward(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetLockedReward(&_BindingContract.CallOpts, _tokenId)
}

// GetTotalCoin is a free data retrieval call binding the contract method 0xf5a0aace.
//
// Solidity: function getTotalCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCaller) GetTotalCoin(opts *bind.CallOpts, _tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getTotalCoin", _tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalCoin is a free data retrieval call binding the contract method 0xf5a0aace.
//
// Solidity: function getTotalCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractSession) GetTotalCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetTotalCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetTotalCoin is a free data retrieval call binding the contract method 0xf5a0aace.
//
// Solidity: function getTotalCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetTotalCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetTotalCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetUnlockedCoin is a free data retrieval call binding the contract method 0xf064dd20.
//
// Solidity: function getUnlockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCaller) GetUnlockedCoin(opts *bind.CallOpts, _tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getUnlockedCoin", _tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnlockedCoin is a free data retrieval call binding the contract method 0xf064dd20.
//
// Solidity: function getUnlockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractSession) GetUnlockedCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetUnlockedCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetUnlockedCoin is a free data retrieval call binding the contract method 0xf064dd20.
//
// Solidity: function getUnlockedCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetUnlockedCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetUnlockedCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetUnlockingCoin is a free data retrieval call binding the contract method 0xa906c0ae.
//
// Solidity: function getUnlockingCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCaller) GetUnlockingCoin(opts *bind.CallOpts, _tokenId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "getUnlockingCoin", _tokenId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnlockingCoin is a free data retrieval call binding the contract method 0xa906c0ae.
//
// Solidity: function getUnlockingCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractSession) GetUnlockingCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetUnlockingCoin(&_BindingContract.CallOpts, _tokenId)
}

// GetUnlockingCoin is a free data retrieval call binding the contract method 0xa906c0ae.
//
// Solidity: function getUnlockingCoin(uint256 _tokenId) view returns(uint256)
func (_BindingContract *BindingContractCallerSession) GetUnlockingCoin(_tokenId *big.Int) (*big.Int, error) {
	return _BindingContract.Contract.GetUnlockingCoin(&_BindingContract.CallOpts, _tokenId)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address _owner, address _operator) view returns(bool)
func (_BindingContract *BindingContractCaller) IsApprovedForAll(opts *bind.CallOpts, _owner common.Address, _operator common.Address) (bool, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "isApprovedForAll", _owner, _operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address _owner, address _operator) view returns(bool)
func (_BindingContract *BindingContractSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _BindingContract.Contract.IsApprovedForAll(&_BindingContract.CallOpts, _owner, _operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address _owner, address _operator) view returns(bool)
func (_BindingContract *BindingContractCallerSession) IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error) {
	return _BindingContract.Contract.IsApprovedForAll(&_BindingContract.CallOpts, _owner, _operator)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractCaller) OwnerOf(opts *bind.CallOpts, _tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "ownerOf", _tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _BindingContract.Contract.OwnerOf(&_BindingContract.CallOpts, _tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 _tokenId) view returns(address)
func (_BindingContract *BindingContractCallerSession) OwnerOf(_tokenId *big.Int) (common.Address, error) {
	return _BindingContract.Contract.OwnerOf(&_BindingContract.CallOpts, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactor) Approve(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "approve", _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Approve(&_BindingContract.TransactOpts, _to, _tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactorSession) Approve(_to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Approve(&_BindingContract.TransactOpts, _to, _tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactor) SafeTransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "safeTransferFrom", _from, _to, _tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.SafeTransferFrom(&_BindingContract.TransactOpts, _from, _to, _tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactorSession) SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.SafeTransferFrom(&_BindingContract.TransactOpts, _from, _to, _tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes data) payable returns()
func (_BindingContract *BindingContractTransactor) SafeTransferFrom0(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "safeTransferFrom0", _from, _to, _tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes data) payable returns()
func (_BindingContract *BindingContractSession) SafeTransferFrom0(_from common.Address, _to common.Address, _tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.SafeTransferFrom0(&_BindingContract.TransactOpts, _from, _to, _tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes data) payable returns()
func (_BindingContract *BindingContractTransactorSession) SafeTransferFrom0(_from common.Address, _to common.Address, _tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _BindingContract.Contract.SafeTransferFrom0(&_BindingContract.TransactOpts, _from, _to, _tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address _operator, bool _approved) returns()
func (_BindingContract *BindingContractTransactor) SetApprovalForAll(opts *bind.TransactOpts, _operator common.Address, _approved bool) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "setApprovalForAll", _operator, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address _operator, bool _approved) returns()
func (_BindingContract *BindingContractSession) SetApprovalForAll(_operator common.Address, _approved bool) (*types.Transaction, error) {
	return _BindingContract.Contract.SetApprovalForAll(&_BindingContract.TransactOpts, _operator, _approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address _operator, bool _approved) returns()
func (_BindingContract *BindingContractTransactorSession) SetApprovalForAll(_operator common.Address, _approved bool) (*types.Transaction, error) {
	return _BindingContract.Contract.SetApprovalForAll(&_BindingContract.TransactOpts, _operator, _approved)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "transferFrom", _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.TransferFrom(&_BindingContract.TransactOpts, _from, _to, _tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _tokenId) payable returns()
func (_BindingContract *BindingContractTransactorSession) TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.TransferFrom(&_BindingContract.TransactOpts, _from, _to, _tokenId)
}

// BindingContractApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the BindingContract contract.
type BindingContractApprovalIterator struct {
	Event *BindingContractApproval // Event containing the contract specifics and raw log

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
func (it *BindingContractApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractApproval)
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
		it.Event = new(BindingContractApproval)
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
func (it *BindingContractApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractApproval represents a Approval event raised by the BindingContract contract.
type BindingContractApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _approved, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (*BindingContractApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _approvedRule []interface{}
	for _, _approvedItem := range _approved {
		_approvedRule = append(_approvedRule, _approvedItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractApprovalIterator{contract: _BindingContract.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _approved, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BindingContractApproval, _owner []common.Address, _approved []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _approvedRule []interface{}
	for _, _approvedItem := range _approved {
		_approvedRule = append(_approvedRule, _approvedItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Approval", _ownerRule, _approvedRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractApproval)
				if err := _BindingContract.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed _owner, address indexed _approved, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) ParseApproval(log types.Log) (*BindingContractApproval, error) {
	event := new(BindingContractApproval)
	if err := _BindingContract.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the BindingContract contract.
type BindingContractApprovalForAllIterator struct {
	Event *BindingContractApprovalForAll // Event containing the contract specifics and raw log

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
func (it *BindingContractApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractApprovalForAll)
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
		it.Event = new(BindingContractApprovalForAll)
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
func (it *BindingContractApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractApprovalForAll represents a ApprovalForAll event raised by the BindingContract contract.
type BindingContractApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed _owner, address indexed _operator, bool _approved)
func (_BindingContract *BindingContractFilterer) FilterApprovalForAll(opts *bind.FilterOpts, _owner []common.Address, _operator []common.Address) (*BindingContractApprovalForAllIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractApprovalForAllIterator{contract: _BindingContract.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed _owner, address indexed _operator, bool _approved)
func (_BindingContract *BindingContractFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *BindingContractApprovalForAll, _owner []common.Address, _operator []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "ApprovalForAll", _ownerRule, _operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractApprovalForAll)
				if err := _BindingContract.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed _owner, address indexed _operator, bool _approved)
func (_BindingContract *BindingContractFilterer) ParseApprovalForAll(log types.Log) (*BindingContractApprovalForAll, error) {
	event := new(BindingContractApprovalForAll)
	if err := _BindingContract.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the BindingContract contract.
type BindingContractTransferIterator struct {
	Event *BindingContractTransfer // Event containing the contract specifics and raw log

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
func (it *BindingContractTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractTransfer)
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
		it.Event = new(BindingContractTransfer)
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
func (it *BindingContractTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractTransfer represents a Transfer event raised by the BindingContract contract.
type BindingContractTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (*BindingContractTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractTransferIterator{contract: _BindingContract.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BindingContractTransfer, _from []common.Address, _to []common.Address, _tokenId []*big.Int) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}
	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule, _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractTransfer)
				if err := _BindingContract.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed _from, address indexed _to, uint256 indexed _tokenId)
func (_BindingContract *BindingContractFilterer) ParseTransfer(log types.Log) (*BindingContractTransfer, error) {
	event := new(BindingContractTransfer)
	if err := _BindingContract.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractUpdateInfoIterator is returned from FilterUpdateInfo and is used to iterate over the raw logs and unpacked data for UpdateInfo events raised by the BindingContract contract.
type BindingContractUpdateInfoIterator struct {
	Event *BindingContractUpdateInfo // Event containing the contract specifics and raw log

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
func (it *BindingContractUpdateInfoIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUpdateInfo)
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
		it.Event = new(BindingContractUpdateInfo)
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
func (it *BindingContractUpdateInfoIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUpdateInfoIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUpdateInfo represents a UpdateInfo event raised by the BindingContract contract.
type BindingContractUpdateInfo struct {
	TokenId        *big.Int
	NewPrincipal   *big.Int
	NewUnlocked    *big.Int
	NewActiveEpoch uint64
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpdateInfo is a free log retrieval operation binding the contract event 0x7430a4e0a12fcc6a5b4d0da4e6ca2beb748928beaaac0e1207e4be8723de0635.
//
// Solidity: event UpdateInfo(uint256 indexed _tokenId, uint256 _newPrincipal, uint256 _newUnlocked, uint64 _newActiveEpoch)
func (_BindingContract *BindingContractFilterer) FilterUpdateInfo(opts *bind.FilterOpts, _tokenId []*big.Int) (*BindingContractUpdateInfoIterator, error) {

	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "UpdateInfo", _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractUpdateInfoIterator{contract: _BindingContract.contract, event: "UpdateInfo", logs: logs, sub: sub}, nil
}

// WatchUpdateInfo is a free log subscription operation binding the contract event 0x7430a4e0a12fcc6a5b4d0da4e6ca2beb748928beaaac0e1207e4be8723de0635.
//
// Solidity: event UpdateInfo(uint256 indexed _tokenId, uint256 _newPrincipal, uint256 _newUnlocked, uint64 _newActiveEpoch)
func (_BindingContract *BindingContractFilterer) WatchUpdateInfo(opts *bind.WatchOpts, sink chan<- *BindingContractUpdateInfo, _tokenId []*big.Int) (event.Subscription, error) {

	var _tokenIdRule []interface{}
	for _, _tokenIdItem := range _tokenId {
		_tokenIdRule = append(_tokenIdRule, _tokenIdItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "UpdateInfo", _tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUpdateInfo)
				if err := _BindingContract.contract.UnpackLog(event, "UpdateInfo", log); err != nil {
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

// ParseUpdateInfo is a log parse operation binding the contract event 0x7430a4e0a12fcc6a5b4d0da4e6ca2beb748928beaaac0e1207e4be8723de0635.
//
// Solidity: event UpdateInfo(uint256 indexed _tokenId, uint256 _newPrincipal, uint256 _newUnlocked, uint64 _newActiveEpoch)
func (_BindingContract *BindingContractFilterer) ParseUpdateInfo(log types.Log) (*BindingContractUpdateInfo, error) {
	event := new(BindingContractUpdateInfo)
	if err := _BindingContract.contract.UnpackLog(event, "UpdateInfo", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
