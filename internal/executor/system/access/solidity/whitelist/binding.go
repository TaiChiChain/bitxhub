// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package whitelist

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

type Whitelist interface {

	// Remove is a paid mutator transaction binding the contract method 0x5e4ba17c.
	//
	// Solidity: function remove(address[] addresses) returns()
	Remove(addresses []common.Address) error

	// Submit is a paid mutator transaction binding the contract method 0xfde4af2d.
	//
	// Solidity: function submit(address[] addresses) returns()
	Submit(addresses []common.Address) error

	// QueryAuthInfo is a free data retrieval call binding the contract method 0x65408691.
	//
	// Solidity: function queryAuthInfo(address addr) view returns((address,address[],uint8) authInfo)
	QueryAuthInfo(addr common.Address) (AuthInfo, error)

	// QueryProviderInfo is a free data retrieval call binding the contract method 0x2d471513.
	//
	// Solidity: function queryProviderInfo(address addr) view returns((address) providerInfo)
	QueryProviderInfo(addr common.Address) (ProviderInfo, error)
}

// WhitelistRemove represents a Remove event raised by the Whitelist contract.
type EventRemove struct {
	Submitter common.Address
	Addresses []common.Address
}

func (_event *EventRemove) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Remove}"])
}

// WhitelistSubmit represents a Submit event raised by the Whitelist contract.
type EventSubmit struct {
	Submitter common.Address
	Addresses []common.Address
}

func (_event *EventSubmit) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Submit}"])
}

// WhitelistUpdateProviders represents a UpdateProviders event raised by the Whitelist contract.
type EventUpdateProviders struct {
	IsAdd     bool
	Providers []ProviderInfo
}

func (_event *EventUpdateProviders) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["UpdateProviders}"])
}
