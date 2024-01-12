package common

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// system contract address range 0x1000-0xffff, start from 1000, avoid conflicts with precompiled contracts
	// SystemContractStartAddr is the start address of system contract
	SystemContractStartAddr = "0x0000000000000000000000000000000000001000"

	// ProposalIDContractAddr is the contract to used to generate the proposal ID
	ProposalIDContractAddr = "0x0000000000000000000000000000000000001000"
	GovernanceContractAddr = "0x0000000000000000000000000000000000001001"

	// TokenManagerContractAddr is the contract to used to manager token info
	TokenManagerContractAddr = "0x0000000000000000000000000000000000001002"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr           = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr           = "0x0000000000000000000000000000000000001004"
	NotFinishedProposalContractAddr = "0x0000000000000000000000000000000000001005"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001006"

	// SystemContractEndAddr is the end address of system contract
	SystemContractEndAddr = "0x000000000000000000000000000000000000ffff"
)

type SystemContractConfig struct {
	Logger logrus.FieldLogger
}

type VirtualMachine interface {
	vm.PrecompiledContract

	// IsSystemContract judge if is system contract
	IsSystemContract(addr *types.Address) bool

	// Reset the state of the system contract
	Reset(currentHeight uint64, stateLedger ledger.StateLedger, from ethcommon.Address, to *ethcommon.Address)

	// View return a view system contract
	View() VirtualMachine
}

type VMContext struct {
	StateLedger   ledger.StateLedger
	CurrentHeight uint64
	CurrentLogs   *[]Log
	CurrentUser   *ethcommon.Address
}

// SystemContract must be implemented by all system contract
type SystemContract interface {
	SetContext(*VMContext)
}

func IsInSlice[T ~uint8 | ~string](value T, slice []T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

func RemoveFirstMatchStrInSlice(slice []string, val string) []string {
	for i, v := range slice {
		if v == val {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

type Log struct {
	Address *types.Address
	Topics  []*types.Hash
	Data    []byte
	Removed bool
}

func CalculateDynamicGas(bytes []byte) uint64 {
	gas, _ := vm.IntrinsicGas(bytes, []ethtype.AccessTuple{}, false, true, true, true)
	return gas
}
