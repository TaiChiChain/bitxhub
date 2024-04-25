// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package staking_manager_client

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

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"liquidStakingTokenID\",\"type\":\"uint256\"}],\"name\":\"AddStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"liquidStakingTokenID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unlock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"liquidStakingTokenID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"addStake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"liquidStakingTokenID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"unlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"poolID\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"liquidStakingTokenID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// AddStake is a paid mutator transaction binding the contract method 0xad899a39.
//
// Solidity: function addStake(uint64 poolID, address owner, uint256 amount) payable returns()
func (_BindingContract *BindingContractTransactor) AddStake(opts *bind.TransactOpts, poolID uint64, owner common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "addStake", poolID, owner, amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xad899a39.
//
// Solidity: function addStake(uint64 poolID, address owner, uint256 amount) payable returns()
func (_BindingContract *BindingContractSession) AddStake(poolID uint64, owner common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.AddStake(&_BindingContract.TransactOpts, poolID, owner, amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xad899a39.
//
// Solidity: function addStake(uint64 poolID, address owner, uint256 amount) payable returns()
func (_BindingContract *BindingContractTransactorSession) AddStake(poolID uint64, owner common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.AddStake(&_BindingContract.TransactOpts, poolID, owner, amount)
}

// Unlock is a paid mutator transaction binding the contract method 0x7a94f25b.
//
// Solidity: function unlock(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractTransactor) Unlock(opts *bind.TransactOpts, poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "unlock", poolID, owner, liquidStakingTokenID, amount)
}

// Unlock is a paid mutator transaction binding the contract method 0x7a94f25b.
//
// Solidity: function unlock(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractSession) Unlock(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Unlock(&_BindingContract.TransactOpts, poolID, owner, liquidStakingTokenID, amount)
}

// Unlock is a paid mutator transaction binding the contract method 0x7a94f25b.
//
// Solidity: function unlock(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractTransactorSession) Unlock(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Unlock(&_BindingContract.TransactOpts, poolID, owner, liquidStakingTokenID, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xcdfc4800.
//
// Solidity: function withdraw(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractTransactor) Withdraw(opts *bind.TransactOpts, poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "withdraw", poolID, owner, liquidStakingTokenID, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xcdfc4800.
//
// Solidity: function withdraw(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractSession) Withdraw(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Withdraw(&_BindingContract.TransactOpts, poolID, owner, liquidStakingTokenID, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xcdfc4800.
//
// Solidity: function withdraw(uint64 poolID, address owner, uint256 liquidStakingTokenID, uint256 amount) returns()
func (_BindingContract *BindingContractTransactorSession) Withdraw(poolID uint64, owner common.Address, liquidStakingTokenID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BindingContract.Contract.Withdraw(&_BindingContract.TransactOpts, poolID, owner, liquidStakingTokenID, amount)
}

// BindingContractAddStakeIterator is returned from FilterAddStake and is used to iterate over the raw logs and unpacked data for AddStake events raised by the BindingContract contract.
type BindingContractAddStakeIterator struct {
	Event *BindingContractAddStake // Event containing the contract specifics and raw log

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
func (it *BindingContractAddStakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractAddStake)
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
		it.Event = new(BindingContractAddStake)
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
func (it *BindingContractAddStakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractAddStakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractAddStake represents a AddStake event raised by the BindingContract contract.
type BindingContractAddStake struct {
	PoolID               uint64
	Owner                common.Address
	Amount               *big.Int
	LiquidStakingTokenID *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterAddStake is a free log retrieval operation binding the contract event 0x3301397a2de959044a175731cb4b6ec2b2759b40fddd83e761657d93c598c073.
//
// Solidity: event AddStake(uint64 indexed poolID, address indexed owner, uint256 amount, uint256 liquidStakingTokenID)
func (_BindingContract *BindingContractFilterer) FilterAddStake(opts *bind.FilterOpts, poolID []uint64, owner []common.Address) (*BindingContractAddStakeIterator, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "AddStake", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractAddStakeIterator{contract: _BindingContract.contract, event: "AddStake", logs: logs, sub: sub}, nil
}

// WatchAddStake is a free log subscription operation binding the contract event 0x3301397a2de959044a175731cb4b6ec2b2759b40fddd83e761657d93c598c073.
//
// Solidity: event AddStake(uint64 indexed poolID, address indexed owner, uint256 amount, uint256 liquidStakingTokenID)
func (_BindingContract *BindingContractFilterer) WatchAddStake(opts *bind.WatchOpts, sink chan<- *BindingContractAddStake, poolID []uint64, owner []common.Address) (event.Subscription, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "AddStake", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractAddStake)
				if err := _BindingContract.contract.UnpackLog(event, "AddStake", log); err != nil {
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

// ParseAddStake is a log parse operation binding the contract event 0x3301397a2de959044a175731cb4b6ec2b2759b40fddd83e761657d93c598c073.
//
// Solidity: event AddStake(uint64 indexed poolID, address indexed owner, uint256 amount, uint256 liquidStakingTokenID)
func (_BindingContract *BindingContractFilterer) ParseAddStake(log types.Log) (*BindingContractAddStake, error) {
	event := new(BindingContractAddStake)
	if err := _BindingContract.contract.UnpackLog(event, "AddStake", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractUnlockIterator is returned from FilterUnlock and is used to iterate over the raw logs and unpacked data for Unlock events raised by the BindingContract contract.
type BindingContractUnlockIterator struct {
	Event *BindingContractUnlock // Event containing the contract specifics and raw log

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
func (it *BindingContractUnlockIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUnlock)
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
		it.Event = new(BindingContractUnlock)
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
func (it *BindingContractUnlockIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUnlockIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUnlock represents a Unlock event raised by the BindingContract contract.
type BindingContractUnlock struct {
	PoolID               uint64
	Owner                common.Address
	LiquidStakingTokenID *big.Int
	Amount               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterUnlock is a free log retrieval operation binding the contract event 0x9e001325f29422fd56979ba9277ab9ef57ef9b425517a2efa6979dcbabe20006.
//
// Solidity: event Unlock(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) FilterUnlock(opts *bind.FilterOpts, poolID []uint64, owner []common.Address) (*BindingContractUnlockIterator, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Unlock", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractUnlockIterator{contract: _BindingContract.contract, event: "Unlock", logs: logs, sub: sub}, nil
}

// WatchUnlock is a free log subscription operation binding the contract event 0x9e001325f29422fd56979ba9277ab9ef57ef9b425517a2efa6979dcbabe20006.
//
// Solidity: event Unlock(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) WatchUnlock(opts *bind.WatchOpts, sink chan<- *BindingContractUnlock, poolID []uint64, owner []common.Address) (event.Subscription, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Unlock", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUnlock)
				if err := _BindingContract.contract.UnpackLog(event, "Unlock", log); err != nil {
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

// ParseUnlock is a log parse operation binding the contract event 0x9e001325f29422fd56979ba9277ab9ef57ef9b425517a2efa6979dcbabe20006.
//
// Solidity: event Unlock(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) ParseUnlock(log types.Log) (*BindingContractUnlock, error) {
	event := new(BindingContractUnlock)
	if err := _BindingContract.contract.UnpackLog(event, "Unlock", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the BindingContract contract.
type BindingContractWithdrawIterator struct {
	Event *BindingContractWithdraw // Event containing the contract specifics and raw log

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
func (it *BindingContractWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractWithdraw)
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
		it.Event = new(BindingContractWithdraw)
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
func (it *BindingContractWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractWithdraw represents a Withdraw event raised by the BindingContract contract.
type BindingContractWithdraw struct {
	PoolID               uint64
	Owner                common.Address
	LiquidStakingTokenID *big.Int
	Amount               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x1d7b95cb608751bc6718009a76582b8205568cfa1f0dd76b226139f147622f69.
//
// Solidity: event Withdraw(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) FilterWithdraw(opts *bind.FilterOpts, poolID []uint64, owner []common.Address) (*BindingContractWithdrawIterator, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Withdraw", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractWithdrawIterator{contract: _BindingContract.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x1d7b95cb608751bc6718009a76582b8205568cfa1f0dd76b226139f147622f69.
//
// Solidity: event Withdraw(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *BindingContractWithdraw, poolID []uint64, owner []common.Address) (event.Subscription, error) {

	var poolIDRule []interface{}
	for _, poolIDItem := range poolID {
		poolIDRule = append(poolIDRule, poolIDItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Withdraw", poolIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractWithdraw)
				if err := _BindingContract.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x1d7b95cb608751bc6718009a76582b8205568cfa1f0dd76b226139f147622f69.
//
// Solidity: event Withdraw(uint64 indexed poolID, address indexed owner, uint256 liquidStakingTokenID, uint256 amount)
func (_BindingContract *BindingContractFilterer) ParseWithdraw(log types.Log) (*BindingContractWithdraw, error) {
	event := new(BindingContractWithdraw)
	if err := _BindingContract.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
