package common

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// SystemContractStartAddr is the start address of system contract
	// the address range is 0x1000-0xffff, start from 1000, avoid conflicts with precompiled contracts
	SystemContractStartAddr = "0x0000000000000000000000000000000000001000"

	// ProposalIDContractAddr is the contract to used to generate the proposal ID
	ProposalIDContractAddr = "0x0000000000000000000000000000000000001000"
	GovernanceContractAddr = "0x0000000000000000000000000000000000001001"

	// AXMContractAddr is the contract to used to manager native token info
	AXMContractAddr = "0x0000000000000000000000000000000000001002"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr           = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr           = "0x0000000000000000000000000000000000001004"
	NotFinishedProposalContractAddr = "0x0000000000000000000000000000000000001005"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001006"

	// AXCContractAddr is the system contract for axc
	AXCContractAddr = "0x0000000000000000000000000000000000001007"

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
	Reset(currentHeight uint64, stateLedger ledger.StateLedger)

	// View return a view system contract
	View() VirtualMachine

	// GetContractInstance return the contract instance by given address
	GetContractInstance(addr *types.Address) SystemContract

	// EnhancedInput return the enhanced input by given from and to address
	EnhancedInput(from *ethcommon.Address, to *ethcommon.Address, originalInput []byte) ([]byte, error)

	// DecodeEnhancedInput decode the enhanced input
	// return the original input
	// reset 'from' and 'to' address for nvm
	DecodeEnhancedInput(data []byte) ([]byte, error)
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
	gas, _ := core.IntrinsicGas(bytes, []ethtype.AccessTuple{}, false, true, true, true)
	return gas
}
