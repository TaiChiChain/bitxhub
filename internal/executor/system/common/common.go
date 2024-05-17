package common

import (
	"flag"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/intutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
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

type LogsCollectorStateLedger struct {
	ledger.StateLedger

	logs                     []*types.EvmLog
	disableRecordLogToLedger bool
}

func (l *LogsCollectorStateLedger) AddLog(log *types.EvmLog) {
	l.logs = append(l.logs, log)
	if !l.disableRecordLogToLedger {
		l.StateLedger.AddLog(log)
	}
}

type VMContext struct {
	StateLedger    *LogsCollectorStateLedger
	BlockNumber    uint64
	From           ethcommon.Address
	Value          *big.Int
	CallFromSystem bool
	CurrentEVM     *vm.EVM

	disableRecordLogToLedger bool
	TestLogs                 []any
}

func NewVMContext(stateLedger ledger.StateLedger, evm *vm.EVM, from ethcommon.Address, value *big.Int) *VMContext {
	return &VMContext{
		StateLedger: &LogsCollectorStateLedger{StateLedger: stateLedger},
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		From:        from,
		Value:       value,
		CurrentEVM:  evm,
	}
}

func NewVMContextByExecutor(stateLedger ledger.StateLedger) *VMContext {
	return &VMContext{
		StateLedger:    &LogsCollectorStateLedger{StateLedger: stateLedger},
		BlockNumber:    stateLedger.CurrentBlockHeight(),
		From:           ethcommon.Address{},
		CallFromSystem: true,
	}
}

func NewViewVMContext(stateLedger ledger.StateLedger) *VMContext {
	return &VMContext{
		StateLedger: &LogsCollectorStateLedger{StateLedger: stateLedger},
		BlockNumber: stateLedger.CurrentBlockHeight(),
		From:        ethcommon.Address{},
		CurrentEVM:  nil,
	}
}

func NewTestVMContext(stateLedger ledger.StateLedger, from ethcommon.Address) *VMContext {
	return &VMContext{
		StateLedger: &LogsCollectorStateLedger{StateLedger: stateLedger},
		BlockNumber: stateLedger.CurrentBlockHeight(),
		From:        from,
		CurrentEVM:  nil,
	}
}

func (s *VMContext) SetFrom(from ethcommon.Address) *VMContext {
	s.From = from
	return s
}

func (s *VMContext) GetLogs() []*types.EvmLog {
	return s.StateLedger.logs
}

func (s *VMContext) DisableRecordLogToLedger() *VMContext {
	s.disableRecordLogToLedger = true
	s.StateLedger.disableRecordLogToLedger = true
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
		BlockNumber:    s.Ctx.BlockNumber,
		From:           s.EthAddress,
		CallFromSystem: true,
		// TODO: set caller
		CurrentEVM:               s.Ctx.CurrentEVM,
		disableRecordLogToLedger: s.Ctx.disableRecordLogToLedger,
	}
}

func (s *SystemContractBase) EmitEvent(packer packer.Event) {
	log, err := packer.Pack(s.Abi)
	if err != nil {
		panic(errors.Wrap(err, "emit event error"))
	}
	s.Ctx.StateLedger.AddLog(log)
	if s.isTest() {
		s.Ctx.TestLogs = append(s.Ctx.TestLogs, packer)
	}
}

func (s *SystemContractBase) isTest() bool {
	if flag.Lookup("test.v") == nil {
		return false
	}
	return true
}

func (s *SystemContractBase) Revert(err packer.Error) error {
	return err.Pack(s.Abi)
}

// CrossCallEVMContract return call result, left over gas and error
func (s *SystemContractBase) CrossCallEVMContract(gas *big.Int, to ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	return s.CrossCallEVMContractWithValue(gas, big.NewInt(0), to, callData)
}

// CrossCallEVMContractWithValue callWithValue return call result, left over gas and error
// nolint
func (s *SystemContractBase) CrossCallEVMContractWithValue(gas, value *big.Int, to ethcommon.Address, callData []byte) (returnData []byte, gasLeft uint64, err error) {
	if gas == nil || gas.Sign() == 0 {
		gas = big.NewInt(MaxCallGasLimit)
	}

	val, err := intutil.BigIntToUint256(value)
	if err != nil {
		return nil, 0, err
	}

	result, gasLeft, err := s.Ctx.CurrentEVM.Call(vm.AccountRef(s.EthAddress), to, callData, gas.Uint64(), val)
	if errors.Is(err, vm.ErrExecutionReverted) {
		err = errors.Errorf("%s, reason: %x", err.Error(), result)
	}
	return result, gasLeft, err
}

func (s *SystemContractBase) Transfer(to ethcommon.Address, amount *big.Int) error {
	if amount.Sign() == 0 {
		return nil
	}
	balance := s.StateAccount.GetBalance()
	if balance.Cmp(amount) < 0 {
		return errors.New("balance is not enough")
	}
	s.StateAccount.SubBalance(amount)
	s.Ctx.StateLedger.AddBalance(types.NewAddress(to.Bytes()), amount)
	return nil
}

func IsZeroAddress(addr ethcommon.Address) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}

	return true
}

func newRevertError(abiErr abi.Error, args []any) error {
	selector := ethcommon.CopyBytes(abiErr.ID.Bytes()[:4])
	packed, err := abiErr.Inputs.Pack(args...)
	if err != nil {
		return err
	}

	return &packer.RevertError{
		Err:  vm.ErrExecutionReverted,
		Data: append(selector, packed...),
		Str:  fmt.Sprintf("%s, args: %v", abiErr.String(), args),
	}
}

func NewRevertError(name string, inputs abi.Arguments, args []any) error {
	abiErr := abi.NewError(name, inputs)
	return newRevertError(abiErr, args)
}

var revertSelector = crypto.Keccak256([]byte("Error(string)"))[:4]

func NewRevertStringError(data string) *packer.RevertError {
	packed, err := (abi.Arguments{{Type: StringType}}).Pack(data)
	if err != nil {
		panic(err)
	}
	return &packer.RevertError{
		Err:  vm.ErrExecutionReverted,
		Data: append(revertSelector, packed...),
		Str:  data,
	}
}
