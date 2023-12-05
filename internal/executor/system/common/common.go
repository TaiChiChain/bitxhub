package common

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// system contract address range 0x1000-0xffff, start from 1000, avoid conflicts with precompiled contracts

	// ProposalIDContractAddr is the contract to used to generate the proposal ID
	ProposalIDContractAddr = "0x0000000000000000000000000000000000001000"

	NodeManagerContractAddr    = "0x0000000000000000000000000000000000001001"
	CouncilManagerContractAddr = "0x0000000000000000000000000000000000001002"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr                = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr                = "0x0000000000000000000000000000000000001004"
	WhiteListProviderManagerContractAddr = "0x0000000000000000000000000000000000001005"
	NotFinishedProposalContractAddr      = "0x0000000000000000000000000000000000001006"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001007"

	GasManagerContractAddr = "0x0000000000000000000000000000000000001008"

	// TokenManagerContractAddr is the contract to used to manager token info
	TokenManagerContractAddr = "0x0000000000000000000000000000000000001009"
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

func ParseContractCallArgs(contractAbi *abi.ABI, data []byte, methodSig2ArgsReceiverConstructor map[string]func() any) (any, *abi.Method, error) {
	if len(data) < 4 {
		return nil, nil, errors.New("gabi: data is invalid")
	}

	method, err := contractAbi.MethodById(data[:4])
	if err != nil {
		return nil, nil, errors.Errorf("gabi: not found method: %v", err)
	}
	argsReceiverConstructor, ok := methodSig2ArgsReceiverConstructor[method.Sig]
	if !ok {
		return nil, nil, errors.Errorf("gabi: not support method: %v", method.Name)
	}
	args := argsReceiverConstructor()
	unpacked, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, nil, errors.Errorf("gabi: decode method[%s] args failed: %v", method.Name, err)
	}
	if err = method.Inputs.Copy(args, unpacked); err != nil {
		return nil, nil, errors.Errorf("gabi: decode method[%s] args failed: %v", method.Name, err)
	}
	return args, method, nil
}
