package common

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtype "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	// ZeroAddress is a special address, no one has control
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	// SystemContractStartAddr is the start address of system contract
	// the address range is 0x1000-0xffff, start from 1000, avoid conflicts with precompiled contracts
	SystemContractStartAddr = "0x0000000000000000000000000000000000001000"

	GovernanceContractAddr = "0x0000000000000000000000000000000000001001"

	// AXCContractAddr is the contract to used to manager native token info
	AXCContractAddr                = "0x0000000000000000000000000000000000001002"
	StakingManagerContractAddr     = "0x0000000000000000000000000000000000001003"
	LiquidStakingTokenContractAddr = "0x0000000000000000000000000000000000001004"
	WhiteListContractAddr          = "0x0000000000000000000000000000000000001005"

	// EpochManagerContractAddr is the contract to used to manager chain epoch info
	EpochManagerContractAddr = "0x0000000000000000000000000000000000001006"
	NodeManagerContractAddr  = "0x0000000000000000000000000000000000001007"

	// Smart account contract

	// EntryPointContractAddr is the address of entry point system contract
	EntryPointContractAddr = "0x0000000000000000000000000000000000001008"

	// AccountFactoryContractAddr is the address of account factory system contract
	AccountFactoryContractAddr = "0x0000000000000000000000000000000000001009"

	// VerifyingPaymasterContractAddr is the address of verifying paymaster system contract
	VerifyingPaymasterContractAddr = "0x000000000000000000000000000000000000100a"

	// TokenPaymasterContractAddr is the address of token paymaster system contract
	TokenPaymasterContractAddr = "0x000000000000000000000000000000000000100b"

	// SystemContractEndAddr is the end address of system contract
	SystemContractEndAddr = "0x000000000000000000000000000000000000ffff"

	// empty contract bin code
	// Attention: this is runtime bin code
	EmptyContractBinCode = "608060405260043610601e575f3560e01c8063f2a75fe4146028575f80fd5b36602457005b5f80fd5b3480156032575f80fd5b00fea26469706673582212200b18a08a695e9b66c6e7f5c4186fd44f80415402a02823d06bc183192f130e1b64736f6c63430008180033"
)

const (
	MaxCallGasLimit = 500000
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

type VirtualMachine interface {
	vm.PrecompiledContract
}

type VMContext struct {
	StateLedger    ledger.StateLedger
	RecordLog      bool
	BlockNumber    uint64
	From           ethcommon.Address
	CallFromSystem bool
	CurrentEVM     *vm.EVM
}

func NewVMContext(stateLedger ledger.StateLedger, evm *vm.EVM, from ethcommon.Address, recordLog bool) *VMContext {
	return &VMContext{
		StateLedger: stateLedger,
		RecordLog:   recordLog,
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		From:        from,
		CurrentEVM:  evm,
	}
}

func NewVMContextByExecutor(stateLedger ledger.StateLedger) *VMContext {
	return &VMContext{
		StateLedger:    stateLedger,
		RecordLog:      false,
		BlockNumber:    stateLedger.CurrentBlockHeight(),
		From:           ethcommon.Address{},
		CallFromSystem: true,
	}
}

func NewViewVMContext(stateLedger ledger.StateLedger) *VMContext {
	return &VMContext{
		StateLedger: stateLedger,
		RecordLog:   false,
		BlockNumber: stateLedger.CurrentBlockHeight(),
		From:        ethcommon.Address{},
		CurrentEVM:  nil,
	}
}

func (s *VMContext) SetFrom(from ethcommon.Address) *VMContext {
	s.From = from
	return s
}

// SystemContract must be implemented by all system contract
type SystemContract interface {
	GenesisInit(genesis *repo.GenesisConfig) error

	SetContext(ctx *VMContext)
}

type SystemContractBuildConfig[T SystemContract] struct {
	Name        string
	Address     string
	AbiStr      string
	Constructor func(systemContractBase SystemContractBase) T

	address    *types.Address
	ethAddress ethcommon.Address
	abi        abi.ABI
	once       sync.Once
}

func (cfg *SystemContractBuildConfig[T]) init() {
	cfg.once.Do(func() {
		cfg.address = types.NewAddressByStr(cfg.Address)
		cfg.ethAddress = ethcommon.HexToAddress(cfg.Address)
		var err error
		cfg.abi, err = abi.JSON(strings.NewReader(cfg.AbiStr))
		if err != nil {
			panic(err)
		}
	})
}

func (cfg *SystemContractBuildConfig[T]) Build(ctx *VMContext) T {
	cfg.init()
	systemContract := cfg.Constructor(SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract).WithFields(logrus.Fields{"contract": cfg.Name, "address": cfg.Address}),
		EthAddress: cfg.ethAddress,
		Abi:        cfg.abi,
		Address:    cfg.address,
	})
	systemContract.SetContext(ctx)
	return systemContract
}

func (cfg *SystemContractBuildConfig[T]) BuildWithAddress(ctx *VMContext, addr ethcommon.Address) T {
	cfg.init()
	systemContract := cfg.Constructor(SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract).WithFields(logrus.Fields{"contract": cfg.Name, "address": cfg.Address}),
		EthAddress: addr,
		Abi:        cfg.abi,
		Address:    types.NewAddress(addr.Bytes()),
	})
	systemContract.SetContext(ctx)
	return systemContract
}

func (cfg *SystemContractBuildConfig[T]) StaticConfig() *SystemContractStaticConfig {
	return &SystemContractStaticConfig{
		Name:    cfg.Name,
		Address: cfg.Address,
		AbiStr:  cfg.AbiStr,
		Constructor: func(systemContractBase SystemContractBase) SystemContract {
			return cfg.Constructor(systemContractBase)
		},
		address:    cfg.address,
		ethAddress: cfg.ethAddress,
		abi:        cfg.abi,
		once:       sync.Once{},
	}
}

type SystemContractStaticConfig struct {
	Name        string
	Address     string
	AbiStr      string
	Constructor func(systemContractBase SystemContractBase) SystemContract

	address    *types.Address
	ethAddress ethcommon.Address
	abi        abi.ABI
	once       sync.Once
}

func (cfg *SystemContractStaticConfig) init() {
	cfg.once.Do(func() {
		cfg.address = types.NewAddressByStr(cfg.Address)
		cfg.ethAddress = ethcommon.HexToAddress(cfg.Address)
		var err error
		cfg.abi, err = abi.JSON(strings.NewReader(cfg.AbiStr))
		if err != nil {
			panic(err)
		}
	})
}

func (cfg *SystemContractStaticConfig) Build(ctx *VMContext) SystemContract {
	cfg.init()
	systemContract := cfg.Constructor(SystemContractBase{
		Logger:     loggers.Logger(loggers.SystemContract).WithFields(logrus.Fields{"contract": cfg.Name, "address": cfg.Address}),
		EthAddress: cfg.ethAddress,
		Abi:        cfg.abi,
		Address:    cfg.address,
	})
	systemContract.SetContext(ctx)
	return systemContract
}

func (cfg *SystemContractStaticConfig) GetAbi() abi.ABI {
	cfg.init()
	return cfg.abi
}

func (cfg *SystemContractStaticConfig) GetEthAddress() ethcommon.Address {
	cfg.init()
	return cfg.ethAddress
}

func (cfg *SystemContractStaticConfig) GetAddress() *types.Address {
	cfg.init()
	return cfg.address
}

type SystemContractBase struct {
	Ctx          *VMContext
	Logger       logrus.FieldLogger
	Address      *types.Address
	EthAddress   ethcommon.Address
	Abi          abi.ABI
	StateAccount ledger.IAccount
}

func (s *SystemContractBase) SetContext(ctx *VMContext) {
	s.Ctx = ctx
	s.StateAccount = ctx.StateLedger.GetOrCreateAccount(s.Address)
}

func (s *SystemContractBase) CrossCallSystemContractContext() *VMContext {
	return &VMContext{
		StateLedger:    s.Ctx.StateLedger,
		RecordLog:      s.Ctx.RecordLog,
		BlockNumber:    s.Ctx.BlockNumber,
		From:           s.EthAddress,
		CallFromSystem: true,
		CurrentEVM:     s.Ctx.CurrentEVM,
	}
}

func (s *SystemContractBase) EmitEvent(eventName string, args ...any) {
	log, err := packEvent(s.Abi, eventName, args...)
	if err != nil {
		panic(errors.Wrap(err, "emit event error"))
	}

	s.Logger.Debugf("Emit event %s: %s", eventName, hex.EncodeToString(log.Data))
	if s.Ctx.RecordLog {
		s.Ctx.StateLedger.AddLog(log)
	}
}

func (s *SystemContractBase) RevertCustomError(name string, args []any) error {
	abiErr := s.Abi.Errors[name]

	return newRevertError(abiErr, args)
}

// CrossCallEVMContract return call result, left over gas and error
func (s *SystemContractBase) CrossCallEVMContract(gas *big.Int, to ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	return s.CrossCallEVMContractWithValue(gas, big.NewInt(0), to, callData)
}

// callWithValue return call result, left over gas and error
// nolint
func (s *SystemContractBase) CrossCallEVMContractWithValue(gas, value *big.Int, to ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	if gas == nil || gas.Sign() == 0 {
		gas = big.NewInt(MaxCallGasLimit)
	}

	result, gasLeft, err := s.Ctx.CurrentEVM.Call(vm.AccountRef(s.EthAddress), to, callData, gas.Uint64(), value)
	if errors.Is(err, vm.ErrExecutionReverted) {
		err = errors.Errorf("%s, reason: %x", err.Error(), result)
	}
	return result, gasLeft, err
}

func packEvent(contractAbi abi.ABI, eventName string, args ...any) (*types.EvmLog, error) {
	event, ok := contractAbi.Events[eventName]
	if !ok {
		return nil, errors.Errorf("event %s not defined", eventName)
	}
	// references: https://medium.com/mycrypto/understanding-event-logs-on-the-ethereum-blockchain-f4ae7ba50378
	var noIndexedArgs []any
	topicArgs := [][]any{
		{event.ID},
	}
	for i, input := range event.Inputs {
		if !input.Indexed {
			noIndexedArgs = append(noIndexedArgs, args[i])
		} else {
			topicArgs = append(topicArgs, []any{args[i]})
		}
	}

	topics, err := abi.MakeTopics(topicArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "event %s make topics error", eventName)
	}

	packedData, err := event.Inputs.NonIndexed().Pack(noIndexedArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "event %s pack args error", eventName)
	}

	return &types.EvmLog{
		Topics: lo.Map(topics, func(t []ethcommon.Hash, i int) *types.Hash {
			return types.NewHash(t[0].Bytes())
		}),
		Data:    packedData,
		Removed: false,
	}, nil
}

func CalculateDynamicGas(bytes []byte) uint64 {
	gas, _ := core.IntrinsicGas(bytes, []ethtype.AccessTuple{}, false, true, true, true)
	return gas
}

func IsZeroAddress(addr ethcommon.Address) bool {
	for _, b := range addr {
		if b != 0 {
			return true
		}
	}

	return true
}

// Bool2Bytes Used for record evm log
func Bool2Bytes(b bool) []byte {
	if b {
		return []byte{1}
	}

	return []byte{0}
}

var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

type RevertError struct {
	Err error

	// Data is encoded reverted reason, or result
	Data []byte

	// reverted result
	Str string
}

func NewRevertStringError(data string) *RevertError {
	packed, err := (abi.Arguments{{Type: StringType}}).Pack(data)
	if err != nil {
		panic(errors.Wrap(err, "pack revert string error"))
	}
	return &RevertError{
		Err:  vm.ErrExecutionReverted,
		Data: append(revertSelector, packed...),
		Str:  data,
	}
}

func newRevertError(abiErr abi.Error, args []any) error {
	selector := ethcommon.CopyBytes(abiErr.ID.Bytes()[:4])
	packed, err := abiErr.Inputs.Pack(args...)
	if err != nil {
		return err
	}

	return &RevertError{
		Err:  vm.ErrExecutionReverted,
		Data: append(selector, packed...),
		Str:  fmt.Sprintf("%s, args: %v", abiErr.String(), args),
	}
}

func NewRevertError(name string, inputs abi.Arguments, args []any) error {
	abiErr := abi.NewError(name, inputs)
	return newRevertError(abiErr, args)
}

func (e *RevertError) Error() string {
	return fmt.Sprintf("%s errdata %s", e.Err.Error(), e.Str)
}
