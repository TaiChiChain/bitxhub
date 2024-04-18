// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package liquid_staking_token

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

type LiquidStakingToken interface {

	// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
	//
	// Solidity: function approve(address _approved, uint256 _tokenId) payable returns()
	Approve(_approved common.Address, _tokenId *big.Int) error

	// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
	//
	// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId) payable returns()
	SafeTransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) error

	// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
	//
	// Solidity: function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes data) payable returns()
	SafeTransferFrom0(_from common.Address, _to common.Address, _tokenId *big.Int, data []byte) error

	// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
	//
	// Solidity: function setApprovalForAll(address _operator, bool _approved) returns()
	SetApprovalForAll(_operator common.Address, _approved bool) error

	// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
	//
	// Solidity: function transferFrom(address _from, address _to, uint256 _tokenId) payable returns()
	TransferFrom(_from common.Address, _to common.Address, _tokenId *big.Int) error

	// GetLockedCoin is a free data retrieval call binding the contract method 0x4d4b209c.
	//
	// Solidity: function GetLockedCoin(uint256 _tokenId) view returns(uint256)
	GetLockedCoin(_tokenId *big.Int) (*big.Int, error)

	// GetLockedReward is a free data retrieval call binding the contract method 0xadfd0c07.
	//
	// Solidity: function GetLockedReward(uint256 _tokenId) view returns(uint256)
	GetLockedReward(_tokenId *big.Int) (*big.Int, error)

	// GetTotalCoin is a free data retrieval call binding the contract method 0xc5f50eb9.
	//
	// Solidity: function GetTotalCoin(uint256 _tokenId) view returns(uint256)
	GetTotalCoin(_tokenId *big.Int) (*big.Int, error)

	// GetUnlockedCoin is a free data retrieval call binding the contract method 0x262cd446.
	//
	// Solidity: function GetUnlockedCoin(uint256 _tokenId) view returns(uint256)
	GetUnlockedCoin(_tokenId *big.Int) (*big.Int, error)

	// GetUnlockingCoin is a free data retrieval call binding the contract method 0x273d9adb.
	//
	// Solidity: function GetUnlockingCoin(uint256 _tokenId) view returns(uint256)
	GetUnlockingCoin(_tokenId *big.Int) (*big.Int, error)

	// Info is a free data retrieval call binding the contract method 0xe6f250dc.
	//
	// Solidity: function Info(uint256 _tokenId) view returns((uint64,uint256,uint256,uint64,(uint256,uint64)[]) info)
	Info(_tokenId *big.Int) (LiquidStakingTokenInfo, error)

	// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
	//
	// Solidity: function balanceOf(address _owner) view returns(uint256)
	BalanceOf(_owner common.Address) (*big.Int, error)

	// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
	//
	// Solidity: function getApproved(uint256 _tokenId) view returns(address)
	GetApproved(_tokenId *big.Int) (common.Address, error)

	// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
	//
	// Solidity: function isApprovedForAll(address _owner, address _operator) view returns(bool)
	IsApprovedForAll(_owner common.Address, _operator common.Address) (bool, error)

	// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
	//
	// Solidity: function ownerOf(uint256 _tokenId) view returns(address)
	OwnerOf(_tokenId *big.Int) (common.Address, error)
}

// LiquidStakingTokenApproval represents a Approval event raised by the LiquidStakingToken contract.
type EventApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
}

func (_event *EventApproval) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Approval}"])
}

// LiquidStakingTokenApprovalForAll represents a ApprovalForAll event raised by the LiquidStakingToken contract.
type EventApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
}

func (_event *EventApprovalForAll) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["ApprovalForAll}"])
}

// LiquidStakingTokenTransfer represents a Transfer event raised by the LiquidStakingToken contract.
type EventTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
}

func (_event *EventTransfer) Pack(abi abi.ABI) (log *types.EvmLog, err error) {
	return packer.PackEvent(_event, abi.Events["Transfer}"])
}
