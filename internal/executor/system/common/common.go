package common

import (
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// internal contract
	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000000001"

	// ProposalIDContractAddr is the contract to used to generate the proposal ID
	ProposalIDContractAddr = "0x0000000000000000000000000000000000001000"

	// system contract address range 0x1001-0xffff
	NodeManagerContractAddr    = "0x0000000000000000000000000000000000001001"
	CouncilManagerContractAddr = "0x0000000000000000000000000000000000001002"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr                = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr                = "0x0000000000000000000000000000000000001004"
	WhiteListProviderManagerContractAddr = "0x0000000000000000000000000000000000001005"
	NotFinishedProposalContractAddr      = "0x0000000000000000000000000000000000001006"
)

type SystemContractConfig struct {
	Logger logrus.FieldLogger
}

type SystemContractConstruct func(cfg *SystemContractConfig) SystemContract

type SystemContract interface {
	// Reset the state of the system contract
	Reset(uint64, ledger.StateLedger)

	// Run the system contract
	Run(*vm.Message) (*vm.ExecutionResult, error)

	// EstimateGas estimate the gas cost of the system contract
	EstimateGas(*types.CallArgs) (uint64, error)
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
