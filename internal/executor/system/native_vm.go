package system

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/core"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axc"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axm"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/core/vm"
)

var (
	ErrNotExistSystemContract         = errors.New("not exist this system contract")
	ErrNotExistMethodName             = errors.New("not exist method name of this system contract")
	ErrNotExistSystemContractABI      = errors.New("not exist this system contract abi")
	ErrNotImplementFuncSystemContract = errors.New("not implement the function for this system contract")
	ErrInvalidStateDB                 = errors.New("invalid statedb")
)

//go:embed sol/Governance.abi
var governanceABI string

//go:embed sol/WhiteList.abi
var whiteListABI string

//go:embed sol/EpochManager.abi
var epochManagerABI string

//go:embed sol/AxmManager.abi
var axmManagerABI string

//go:embed sol/AxcManager.abi
var axcManagerABI string

var _ common.VirtualMachine = (*NativeVM)(nil)

// NativeVM handle abi decoding for parameters and abi encoding for return data
type NativeVM struct {
	logger logrus.FieldLogger

	// contract address mapping to method signature
	contract2MethodSig map[string]map[string][]byte
	// contract address mapping to contract abi
	contract2ABI map[string]abi.ABI
	// contract address mapping to contact instance
	contracts map[string]struct{}
}

func New() common.VirtualMachine {
	nvm := &NativeVM{
		logger:             loggers.Logger(loggers.SystemContract),
		contract2MethodSig: make(map[string]map[string][]byte),
		contract2ABI:       make(map[string]abi.ABI),
		contracts:          make(map[string]struct{}),
	}

	// deploy all system contract
	nvm.Deploy(common.GovernanceContractAddr, governanceABI, governance.GovernanceMethod2Sig)
	nvm.Deploy(common.EpochManagerContractAddr, epochManagerABI, base.EpochManagerMethod2Sig)
	nvm.Deploy(common.WhiteListContractAddr, whiteListABI, access.WhiteListMethod2Sig)
	nvm.Deploy(common.AXMContractAddr, axmManagerABI, axm.Method2Sig)
	nvm.Deploy(common.AXCContractAddr, axcManagerABI, axc.Method2Sig)

	return nvm
}

func (nvm *NativeVM) View() common.VirtualMachine {
	return &NativeVM{
		logger:             nvm.logger,
		contract2MethodSig: nvm.contract2MethodSig,
		contract2ABI:       nvm.contract2ABI,
		contracts:          nvm.contracts,
	}
}

func (nvm *NativeVM) Deploy(addr string, abiFile string, method2Sig map[string]string) {
	// check system contract range
	if addr < common.SystemContractStartAddr || addr > common.SystemContractEndAddr {
		panic(fmt.Sprintf("this system contract %s is out of range", addr))
	}

	if _, ok := nvm.contracts[addr]; ok {
		panic("deploy system contract repeated")
	}
	nvm.contracts[addr] = struct{}{}

	contractABI, err := abi.JSON(strings.NewReader(abiFile))
	if err != nil {
		panic(err)
	}
	nvm.contract2ABI[addr] = contractABI

	m2sig := make(map[string][]byte)
	for methodName, methodSig := range method2Sig {
		m2sig[methodName] = crypto.Keccak256([]byte(methodSig))
	}
	nvm.contract2MethodSig[addr] = m2sig

	nvm.setEVMPrecompiled(addr)
}

func (nvm *NativeVM) Run(data []byte, statefulArgs *vm.StatefulArgs) (execResult []byte, execErr error) {
	adaptor, ok := statefulArgs.StateDB.(*ledger.EvmStateDBAdaptor)
	if !ok {
		return nil, ErrInvalidStateDB
	}

	if statefulArgs.To == nil {
		return nil, ErrNotExistSystemContract
	}
	contractAddr := statefulArgs.To.Hex()
	contractInstance := nvm.GetContractInstance(types.NewAddressByStr(contractAddr))
	if contractInstance == nil {
		return nil, ErrNotExistSystemContract
	}
	currentLogs := make([]common.Log, 0)
	vmContext := &common.VMContext{
		// set context first
		StateLedger:   adaptor.StateLedger,
		CurrentHeight: statefulArgs.Height.Uint64(),
		CurrentLogs:   &currentLogs,
		CurrentUser:   &statefulArgs.From,
	}
	contractInstance.SetContext(vmContext)

	defer nvm.saveLogs(vmContext.StateLedger, vmContext.CurrentLogs)
	defer func() {
		if err := recover(); err != nil {
			nvm.logger.Error(err)
			execErr = fmt.Errorf("%s", err)
		}
	}()

	methodName, err := nvm.getMethodName(contractAddr, data)
	if err != nil {
		return nil, err
	}
	// method name may be proposed, but we implement Propose
	// capitalize the first letter of a function
	funcName := methodName
	if len(methodName) >= 2 {
		funcName = fmt.Sprintf("%s%s", strings.ToUpper(methodName[:1]), methodName[1:])
	}
	nvm.logger.Debugf("run system contract method name: %s\n", funcName)
	method := reflect.ValueOf(contractInstance).MethodByName(funcName)
	if !method.IsValid() {
		return nil, ErrNotImplementFuncSystemContract
	}
	args, err := nvm.parseArgs(contractAddr, data, methodName)
	if err != nil {
		return nil, err
	}
	var inputs []reflect.Value
	for _, arg := range args {
		inputs = append(inputs, reflect.ValueOf(arg))
	}
	// maybe panic when inputs mismatch, but we recover
	results := method.Call(inputs)

	var returnRes []any
	var returnErr error
	for _, result := range results {
		// basic type(such as bool, number, string, can't call isNil)
		if result.CanInt() || result.CanFloat() || result.CanUint() || result.Kind() == reflect.Bool || result.Kind() == reflect.String {
			returnRes = append(returnRes, result.Interface())
			continue
		}

		if result.IsNil() {
			continue
		}
		if err, ok := result.Interface().(error); ok {
			returnErr = err
			break
		}
		returnRes = append(returnRes, result.Interface())
	}

	nvm.logger.Debugf("Contract addr: %s, method name: %s, return result: %+v, return error: %s", contractAddr, methodName, returnRes, returnErr)

	if returnErr != nil {
		return nil, returnErr
	}

	if returnRes != nil {
		return nvm.PackOutputArgs(contractAddr, methodName, returnRes...)
	}
	return nil, nil
}

// RequiredGas used in Inter-contract calls for EVM
func (nvm *NativeVM) RequiredGas(input []byte) uint64 {
	return common.CalculateDynamicGas(input)
}

// getMethodName quickly returns the name of a method of specified contract.
// This is a quick way to get the name of a method.
// The method name is the first 4 bytes of the keccak256 hash of the method signature.
// If the method name is not found, the empty string is returned.
func (nvm *NativeVM) getMethodName(contractAddr string, data []byte) (string, error) {
	if len(data) < 4 {
		return "", ErrNotExistMethodName
	}

	method2Sig, ok := nvm.contract2MethodSig[contractAddr]
	if !ok {
		return "", ErrNotExistSystemContract
	}

	for methodName, methodSig := range method2Sig {
		id := methodSig[:4]
		if bytes.Equal(id, data[:4]) {
			return methodName, nil
		}
	}

	return "", ErrNotExistMethodName
}

// parseArgs parse the arguments to specified interface by method name
func (nvm *NativeVM) parseArgs(contractAddr string, data []byte, methodName string) ([]any, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("msg data length is not improperly formatted: %q - Bytes: %+v", data, data)
	}

	// dinvmard method id
	msgData := data[4:]

	contractABI, ok := nvm.contract2ABI[contractAddr]
	if !ok {
		return nil, ErrNotExistSystemContractABI
	}

	var args abi.Arguments
	if method, ok := contractABI.Methods[methodName]; ok {
		if len(msgData)%32 != 0 {
			return nil, fmt.Errorf("system contract abi: improperly formatted output: %q - Bytes: %+v", msgData, msgData)
		}
		args = method.Inputs
	}

	if args == nil {
		return nil, fmt.Errorf("system contract abi: could not locate named method: %s", methodName)
	}

	unpacked, err := args.Unpack(msgData)
	if err != nil {
		return nil, err
	}
	return unpacked, nil
}

// PackOutputArgs pack the output arguments by method name
func (nvm *NativeVM) PackOutputArgs(contractAddr, methodName string, outputArgs ...any) ([]byte, error) {
	contractABI, ok := nvm.contract2ABI[contractAddr]
	if !ok {
		return nil, ErrNotExistSystemContractABI
	}

	var args abi.Arguments
	if method, ok := contractABI.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("system contract abi: could not locate named method: %s", methodName)
	}

	return args.Pack(outputArgs...)
}

// UnpackOutputArgs unpack the output arguments by method name
func (nvm *NativeVM) UnpackOutputArgs(contractAddr, methodName string, packed []byte) ([]any, error) {
	contractABI, ok := nvm.contract2ABI[contractAddr]
	if !ok {
		return nil, ErrNotExistSystemContractABI
	}

	var args abi.Arguments
	if method, ok := contractABI.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("system contract abi: could not locate named method: %s", methodName)
	}

	return args.Unpack(packed)
}

// saveLogs save all logs during the system execution
func (nvm *NativeVM) saveLogs(l ledger.StateLedger, currentLogs *[]common.Log) {
	nvm.logger.Debugf("logs: %+v", currentLogs)

	for _, currentLog := range *currentLogs {
		l.AddLog(&types.EvmLog{
			Address: currentLog.Address,
			Topics:  currentLog.Topics,
			Data:    currentLog.Data,
			Removed: currentLog.Removed,
		})
	}
}

func (nvm *NativeVM) setEVMPrecompiled(addr string) {
	// set system contracts into vm.precompiled
	vm.PrecompiledAddressesByzantium = append(vm.PrecompiledAddressesByzantium, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesBerlin = append(vm.PrecompiledAddressesBerlin, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesHomestead = append(vm.PrecompiledAddressesHomestead, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesIstanbul = append(vm.PrecompiledAddressesIstanbul, ethcommon.HexToAddress(addr))

	vm.PrecompiledContractsBerlin[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsByzantium[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsHomestead[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsIstanbul[ethcommon.HexToAddress(addr)] = nvm
}

func (nvm *NativeVM) GetContractInstance(addr *types.Address) common.SystemContract {
	cfg := &common.SystemContractConfig{
		Logger: nvm.logger,
	}
	switch addr.String() {
	case common.GovernanceContractAddr:
		return governance.NewGov(cfg)
	case common.EpochManagerContractAddr:
		return base.NewEpochManager(cfg)
	case common.WhiteListContractAddr:
		return access.NewWhiteList(cfg)
	case common.AXMContractAddr:
		return axm.New(cfg)
	case common.AXCContractAddr:
		return axc.New(cfg)
	}
	return nil
}

func RunAxiomNativeVM(nvm common.VirtualMachine, height uint64, l ledger.StateLedger, data []byte, from ethcommon.Address, to *ethcommon.Address) *core.ExecutionResult {
	usedGas := nvm.RequiredGas(data)
	returnData, err := nvm.Run(data, &vm.StatefulArgs{
		StateDB: &ledger.EvmStateDBAdaptor{StateLedger: l},
		Height:  big.NewInt(int64(height)),
		From:    from,
		To:      to,
	})
	return &core.ExecutionResult{
		UsedGas:    usedGas,
		Err:        err,
		ReturnData: returnData,
	}
}

func InitGenesisData(genesis *repo.GenesisConfig, lg ledger.StateLedger) error {
	if err := base.InitEpochInfo(lg, genesis.EpochInfo.Clone()); err != nil {
		return err
	}
	if err := governance.InitCouncilMembers(lg, genesis.Admins); err != nil {
		return err
	}
	if err := governance.InitNodeMembers(lg, genesis.NodeNames, genesis.EpochInfo); err != nil {
		return err
	}
	if err := governance.InitGasParam(lg, genesis.EpochInfo); err != nil {
		return err
	}

	axmConfig, err := axm.GenerateConfig(genesis)
	if err != nil {
		return err
	}
	if err = axm.Init(lg, axmConfig); err != nil {
		return err
	}

	axcConfig, err := axc.GenerateConfig(genesis)
	if err != nil {
		return err
	}
	if err = axc.Init(lg, axcConfig); err != nil {
		return err
	}

	admins := lo.Map[*repo.Admin, string](genesis.Admins, func(x *repo.Admin, _ int) string {
		return x.Address
	})
	totalLength := len(admins) + len(genesis.InitWhiteListProviders) + len(genesis.Accounts)
	combined := make([]string, 0, totalLength)
	combined = append(combined, admins...)
	combined = append(combined, genesis.InitWhiteListProviders...)
	accountAddrs := lo.Map(genesis.Accounts, func(ac *repo.Account, _ int) string {
		return ac.Address
	})
	combined = append(combined, accountAddrs...)
	if err = access.InitProvidersAndWhiteList(lg, combined, genesis.InitWhiteListProviders); err != nil {
		return err
	}
	return nil
}
