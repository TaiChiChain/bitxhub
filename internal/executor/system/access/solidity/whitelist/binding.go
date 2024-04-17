// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package whitelist

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

// AuthInfo is an auto generated low-level Go binding around an user-defined struct.
type AuthInfo struct {
	Addr      common.Address
	Providers []common.Address
	Role      uint8
}

// ProviderInfo is an auto generated low-level Go binding around an user-defined struct.
type ProviderInfo struct {
	Addr common.Address
}

// BindingContractMetaData contains all meta data concerning the BindingContract contract.
var BindingContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"submitter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"Remove\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"submitter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"Submit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isAdd\",\"type\":\"bool\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"indexed\":false,\"internalType\":\"structProviderInfo[]\",\"name\":\"providers\",\"type\":\"tuple[]\"}],\"name\":\"UpdateProviders\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"queryAuthInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"providers\",\"type\":\"address[]\"},{\"internalType\":\"enumUserRole\",\"name\":\"role\",\"type\":\"uint8\"}],\"internalType\":\"structAuthInfo\",\"name\":\"authInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"queryProviderInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"internalType\":\"structProviderInfo\",\"name\":\"providerInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"remove\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"}],\"name\":\"submit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// QueryAuthInfo is a free data retrieval call binding the contract method 0x65408691.
//
// Solidity: function queryAuthInfo(address addr) view returns((address,address[],uint8) authInfo)
func (_BindingContract *BindingContractCaller) QueryAuthInfo(opts *bind.CallOpts, addr common.Address) (AuthInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "queryAuthInfo", addr)

	if err != nil {
		return *new(AuthInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(AuthInfo)).(*AuthInfo)

	return out0, err

}

// QueryAuthInfo is a free data retrieval call binding the contract method 0x65408691.
//
// Solidity: function queryAuthInfo(address addr) view returns((address,address[],uint8) authInfo)
func (_BindingContract *BindingContractSession) QueryAuthInfo(addr common.Address) (AuthInfo, error) {
	return _BindingContract.Contract.QueryAuthInfo(&_BindingContract.CallOpts, addr)
}

// QueryAuthInfo is a free data retrieval call binding the contract method 0x65408691.
//
// Solidity: function queryAuthInfo(address addr) view returns((address,address[],uint8) authInfo)
func (_BindingContract *BindingContractCallerSession) QueryAuthInfo(addr common.Address) (AuthInfo, error) {
	return _BindingContract.Contract.QueryAuthInfo(&_BindingContract.CallOpts, addr)
}

// QueryProviderInfo is a free data retrieval call binding the contract method 0x2d471513.
//
// Solidity: function queryProviderInfo(address addr) view returns((address) providerInfo)
func (_BindingContract *BindingContractCaller) QueryProviderInfo(opts *bind.CallOpts, addr common.Address) (ProviderInfo, error) {
	var out []interface{}
	err := _BindingContract.contract.Call(opts, &out, "queryProviderInfo", addr)

	if err != nil {
		return *new(ProviderInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ProviderInfo)).(*ProviderInfo)

	return out0, err

}

// QueryProviderInfo is a free data retrieval call binding the contract method 0x2d471513.
//
// Solidity: function queryProviderInfo(address addr) view returns((address) providerInfo)
func (_BindingContract *BindingContractSession) QueryProviderInfo(addr common.Address) (ProviderInfo, error) {
	return _BindingContract.Contract.QueryProviderInfo(&_BindingContract.CallOpts, addr)
}

// QueryProviderInfo is a free data retrieval call binding the contract method 0x2d471513.
//
// Solidity: function queryProviderInfo(address addr) view returns((address) providerInfo)
func (_BindingContract *BindingContractCallerSession) QueryProviderInfo(addr common.Address) (ProviderInfo, error) {
	return _BindingContract.Contract.QueryProviderInfo(&_BindingContract.CallOpts, addr)
}

// Remove is a paid mutator transaction binding the contract method 0x5e4ba17c.
//
// Solidity: function remove(address[] addresses) returns()
func (_BindingContract *BindingContractTransactor) Remove(opts *bind.TransactOpts, addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "remove", addresses)
}

// Remove is a paid mutator transaction binding the contract method 0x5e4ba17c.
//
// Solidity: function remove(address[] addresses) returns()
func (_BindingContract *BindingContractSession) Remove(addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Remove(&_BindingContract.TransactOpts, addresses)
}

// Remove is a paid mutator transaction binding the contract method 0x5e4ba17c.
//
// Solidity: function remove(address[] addresses) returns()
func (_BindingContract *BindingContractTransactorSession) Remove(addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Remove(&_BindingContract.TransactOpts, addresses)
}

// Submit is a paid mutator transaction binding the contract method 0xfde4af2d.
//
// Solidity: function submit(address[] addresses) returns()
func (_BindingContract *BindingContractTransactor) Submit(opts *bind.TransactOpts, addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.contract.Transact(opts, "submit", addresses)
}

// Submit is a paid mutator transaction binding the contract method 0xfde4af2d.
//
// Solidity: function submit(address[] addresses) returns()
func (_BindingContract *BindingContractSession) Submit(addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Submit(&_BindingContract.TransactOpts, addresses)
}

// Submit is a paid mutator transaction binding the contract method 0xfde4af2d.
//
// Solidity: function submit(address[] addresses) returns()
func (_BindingContract *BindingContractTransactorSession) Submit(addresses []common.Address) (*types.Transaction, error) {
	return _BindingContract.Contract.Submit(&_BindingContract.TransactOpts, addresses)
}

// BindingContractRemoveIterator is returned from FilterRemove and is used to iterate over the raw logs and unpacked data for Remove events raised by the BindingContract contract.
type BindingContractRemoveIterator struct {
	Event *BindingContractRemove // Event containing the contract specifics and raw log

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
func (it *BindingContractRemoveIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractRemove)
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
		it.Event = new(BindingContractRemove)
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
func (it *BindingContractRemoveIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractRemoveIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractRemove represents a Remove event raised by the BindingContract contract.
type BindingContractRemove struct {
	Submitter common.Address
	Addresses []common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRemove is a free log retrieval operation binding the contract event 0xf500adc86f9a3a87ee0f8f212871b20160cc3b42ac0c746f4e546b7cb2aacc7a.
//
// Solidity: event Remove(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) FilterRemove(opts *bind.FilterOpts, submitter []common.Address) (*BindingContractRemoveIterator, error) {

	var submitterRule []interface{}
	for _, submitterItem := range submitter {
		submitterRule = append(submitterRule, submitterItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Remove", submitterRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractRemoveIterator{contract: _BindingContract.contract, event: "Remove", logs: logs, sub: sub}, nil
}

// WatchRemove is a free log subscription operation binding the contract event 0xf500adc86f9a3a87ee0f8f212871b20160cc3b42ac0c746f4e546b7cb2aacc7a.
//
// Solidity: event Remove(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) WatchRemove(opts *bind.WatchOpts, sink chan<- *BindingContractRemove, submitter []common.Address) (event.Subscription, error) {

	var submitterRule []interface{}
	for _, submitterItem := range submitter {
		submitterRule = append(submitterRule, submitterItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Remove", submitterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractRemove)
				if err := _BindingContract.contract.UnpackLog(event, "Remove", log); err != nil {
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

// ParseRemove is a log parse operation binding the contract event 0xf500adc86f9a3a87ee0f8f212871b20160cc3b42ac0c746f4e546b7cb2aacc7a.
//
// Solidity: event Remove(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) ParseRemove(log types.Log) (*BindingContractRemove, error) {
	event := new(BindingContractRemove)
	if err := _BindingContract.contract.UnpackLog(event, "Remove", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractSubmitIterator is returned from FilterSubmit and is used to iterate over the raw logs and unpacked data for Submit events raised by the BindingContract contract.
type BindingContractSubmitIterator struct {
	Event *BindingContractSubmit // Event containing the contract specifics and raw log

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
func (it *BindingContractSubmitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractSubmit)
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
		it.Event = new(BindingContractSubmit)
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
func (it *BindingContractSubmitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractSubmitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractSubmit represents a Submit event raised by the BindingContract contract.
type BindingContractSubmit struct {
	Submitter common.Address
	Addresses []common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSubmit is a free log retrieval operation binding the contract event 0x4b4a5053d153cd51836bbbbf5e9903ae3e020799bfeb018dc14ea86796fcf128.
//
// Solidity: event Submit(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) FilterSubmit(opts *bind.FilterOpts, submitter []common.Address) (*BindingContractSubmitIterator, error) {

	var submitterRule []interface{}
	for _, submitterItem := range submitter {
		submitterRule = append(submitterRule, submitterItem)
	}

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "Submit", submitterRule)
	if err != nil {
		return nil, err
	}
	return &BindingContractSubmitIterator{contract: _BindingContract.contract, event: "Submit", logs: logs, sub: sub}, nil
}

// WatchSubmit is a free log subscription operation binding the contract event 0x4b4a5053d153cd51836bbbbf5e9903ae3e020799bfeb018dc14ea86796fcf128.
//
// Solidity: event Submit(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) WatchSubmit(opts *bind.WatchOpts, sink chan<- *BindingContractSubmit, submitter []common.Address) (event.Subscription, error) {

	var submitterRule []interface{}
	for _, submitterItem := range submitter {
		submitterRule = append(submitterRule, submitterItem)
	}

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "Submit", submitterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractSubmit)
				if err := _BindingContract.contract.UnpackLog(event, "Submit", log); err != nil {
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

// ParseSubmit is a log parse operation binding the contract event 0x4b4a5053d153cd51836bbbbf5e9903ae3e020799bfeb018dc14ea86796fcf128.
//
// Solidity: event Submit(address indexed submitter, address[] addresses)
func (_BindingContract *BindingContractFilterer) ParseSubmit(log types.Log) (*BindingContractSubmit, error) {
	event := new(BindingContractSubmit)
	if err := _BindingContract.contract.UnpackLog(event, "Submit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BindingContractUpdateProvidersIterator is returned from FilterUpdateProviders and is used to iterate over the raw logs and unpacked data for UpdateProviders events raised by the BindingContract contract.
type BindingContractUpdateProvidersIterator struct {
	Event *BindingContractUpdateProviders // Event containing the contract specifics and raw log

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
func (it *BindingContractUpdateProvidersIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BindingContractUpdateProviders)
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
		it.Event = new(BindingContractUpdateProviders)
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
func (it *BindingContractUpdateProvidersIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BindingContractUpdateProvidersIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BindingContractUpdateProviders represents a UpdateProviders event raised by the BindingContract contract.
type BindingContractUpdateProviders struct {
	IsAdd     bool
	Providers []ProviderInfo
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUpdateProviders is a free log retrieval operation binding the contract event 0x98821a905e777564b618021b0af7b68de9661458f0fead22a0e965bc9eea9e2f.
//
// Solidity: event UpdateProviders(bool isAdd, (address)[] providers)
func (_BindingContract *BindingContractFilterer) FilterUpdateProviders(opts *bind.FilterOpts) (*BindingContractUpdateProvidersIterator, error) {

	logs, sub, err := _BindingContract.contract.FilterLogs(opts, "UpdateProviders")
	if err != nil {
		return nil, err
	}
	return &BindingContractUpdateProvidersIterator{contract: _BindingContract.contract, event: "UpdateProviders", logs: logs, sub: sub}, nil
}

// WatchUpdateProviders is a free log subscription operation binding the contract event 0x98821a905e777564b618021b0af7b68de9661458f0fead22a0e965bc9eea9e2f.
//
// Solidity: event UpdateProviders(bool isAdd, (address)[] providers)
func (_BindingContract *BindingContractFilterer) WatchUpdateProviders(opts *bind.WatchOpts, sink chan<- *BindingContractUpdateProviders) (event.Subscription, error) {

	logs, sub, err := _BindingContract.contract.WatchLogs(opts, "UpdateProviders")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BindingContractUpdateProviders)
				if err := _BindingContract.contract.UnpackLog(event, "UpdateProviders", log); err != nil {
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

// ParseUpdateProviders is a log parse operation binding the contract event 0x98821a905e777564b618021b0af7b68de9661458f0fead22a0e965bc9eea9e2f.
//
// Solidity: event UpdateProviders(bool isAdd, (address)[] providers)
func (_BindingContract *BindingContractFilterer) ParseUpdateProviders(log types.Log) (*BindingContractUpdateProviders, error) {
	event := new(BindingContractUpdateProviders)
	if err := _BindingContract.contract.UnpackLog(event, "UpdateProviders", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
