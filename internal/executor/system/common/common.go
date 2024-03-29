package common

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
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

	// AXCContractAddr is the contract to used to manager native token info
	AXCContractAddr = "0x0000000000000000000000000000000000001002"

	// Addr2NameContractAddr for unique name mapping to address
	Addr2NameContractAddr           = "0x0000000000000000000000000000000000001003"
	WhiteListContractAddr           = "0x0000000000000000000000000000000000001004"
	NotFinishedProposalContractAddr = "0x0000000000000000000000000000000000001005"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001006"

	NodeManagerContractAddr = "0x0000000000000000000000000000000000001007"

	// Smart account contract
	// EntryPointContractAddr is the address of entry point system contract
	EntryPointContractAddr = "0x0000000000000000000000000000000000001008"

	// AccountFactoryContractAddr is the address of account factory system contract
	AccountFactoryContractAddr = "0x0000000000000000000000000000000000001009"

	// VerifyingPaymasterContractAddr is the address of verifying paymaster system contract
	VerifyingPaymasterContractAddr = "0x000000000000000000000000000000000000100a"

	// TokenPaymasterContractAddr is the address of token paymaster system contract
	TokenPaymasterContractAddr = "0x000000000000000000000000000000000000100b"

	StakingManagerContractAddr     = "0x0000000000000000000000000000000000001012"
	LiquidStakingTokenContractAddr = "0x0000000000000000000000000000000000001013"

	// SystemContractEndAddr is the end address of system contract
	SystemContractEndAddr = "0x000000000000000000000000000000000000ffff"

	// empty contract bin code
	// Attention: this is runtime bin code
	EmptyContractBinCode = "608060405260043610601e575f3560e01c8063f2a75fe4146028575f80fd5b36602457005b5f80fd5b3480156032575f80fd5b00fea26469706673582212200b18a08a695e9b66c6e7f5c4186fd44f80415402a02823d06bc183192f130e1b64736f6c63430008180033"
)

var (
	BoolType, _         = abi.NewType("bool", "", nil)
	BigIntType, _       = abi.NewType("uint256", "", nil)
	UInt64Type, _       = abi.NewType("uint64", "", nil)
	UInt48Type, _       = abi.NewType("uint48", "", nil)
	StringType, _       = abi.NewType("string", "", nil)
	AddressType, _      = abi.NewType("address", "", nil)
	BytesType, _        = abi.NewType("bytes", "", nil)
	Bytes32Type, _      = abi.NewType("bytes32", "", nil)
	AddressSliceType, _ = abi.NewType("address[]", "", nil)
	BytesSliceType, _   = abi.NewType("bytes[]", "", nil)
)

type SystemContractConfig struct {
	Logger logrus.FieldLogger
}

type VirtualMachine interface {
	vm.PrecompiledContract

	// View return a view system contract
	View() VirtualMachine

	// GetContractInstance return the contract instance by given address
	GetContractInstance(addr *types.Address) SystemContract
}

type VMContext struct {
	StateLedger   ledger.StateLedger
	CurrentHeight uint64
	CurrentLogs   *[]Log
	CurrentUser   *ethcommon.Address
	CurrentEVM    *vm.EVM
	ABI           *abi.ABI
}

// SystemContract must be implemented by all system contract
type SystemContract interface {
	SetContext(*VMContext)
}

func EmitEvent(selfAddr *types.Address, abi abi.ABI, event string, args ...[]byte) Log {
	currentLog := Log{
		Address: selfAddr,
	}
	ev := abi.Events[event]
	currentLog.Topics = append(currentLog.Topics, types.NewHash(ev.ID.Bytes()))
	for i := 0; i < len(ev.Inputs); i++ {
		if ev.Inputs[i].Indexed {
			currentLog.Topics = append(currentLog.Topics, types.NewHash(args[i]))
		} else {
			currentLog.Data = append(currentLog.Data, ethcommon.LeftPadBytes(args[i], 32)...)
		}
	}
	return currentLog
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

// Used for record evm log
func Bool2Bytes(b bool) []byte {
	if b {
		return []byte{1}
	}

	return []byte{0}
}

var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

type RevertError struct {
	err error

	// data is encoded reverted reason, or result
	data []byte

	// reverted result
	str string
}

func NewRevertStringError(data string) error {
	packed, err := (abi.Arguments{{Type: StringType}}).Pack(data)
	if err != nil {
		return err
	}
	return &RevertError{
		err:  vm.ErrExecutionReverted,
		data: append(revertSelector, packed...),
		str:  data,
	}
}

func NewRevertError(name string, inputs abi.Arguments, args []any) error {
	abiErr := abi.NewError(name, inputs)
	selector := ethcommon.CopyBytes(abiErr.ID.Bytes()[:4])
	packed, err := inputs.Pack(args...)
	if err != nil {
		return err
	}

	return &RevertError{
		err:  vm.ErrExecutionReverted,
		data: append(selector, packed...),
		str:  fmt.Sprintf("%s, args: %v", abiErr.String(), args),
	}
}

func (e *RevertError) Error() string {
	return fmt.Sprintf("%s errdata %s", e.err.Error(), e.str)
}

func (e *RevertError) GetError() error {
	return e.err
}

func (e *RevertError) Data() []byte {
	return e.data
}
