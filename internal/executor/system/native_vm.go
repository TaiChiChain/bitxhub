package system

import (
	"bytes"
	_ "embed"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/packer"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	ErrNotExistSystemContract         = errors.New("not exist this system contract")
	ErrNotExistMethodName             = errors.New("not exist method name of this system contract")
	ErrNotImplementFuncSystemContract = errors.New("not implement the function for this system contract")
	ErrInvalidStateDB                 = errors.New("invalid statedb")
)

const (
	RunSystemContractGas = 50000
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

var _ common.VirtualMachine = (*NativeVM)(nil)

// NativeVM handle abi decoding for parameters and abi encoding for return data
type NativeVM struct {
	logger logrus.FieldLogger

	// contract addr -> method id -> method name
	contractMethodID2Name map[string]map[[4]byte]string

	// contract addr -> contract cfg
	contractBuildConfigMap map[string]*common.SystemContractStaticConfig

	// priority -> contract addr
	contractGenesisInitPriority []string
}

func New() *NativeVM {
	nvm := &NativeVM{
		logger:                 loggers.Logger(loggers.SystemContract),
		contractMethodID2Name:  make(map[string]map[[4]byte]string),
		contractBuildConfigMap: make(map[string]*common.SystemContractStaticConfig),
	}

	// deploy all system contract
	nvm.Deploy(token.AXCBuildConfig.StaticConfig())

	nvm.Deploy(framework.EpochManagerBuildConfig.StaticConfig())
	nvm.Deploy(framework.NodeManagerBuildConfig.StaticConfig())
	nvm.Deploy(framework.LiquidStakingTokenBuildConfig.StaticConfig())
	nvm.Deploy(framework.StakingManagerBuildConfig.StaticConfig())

	nvm.Deploy(access.WhitelistBuildConfig.StaticConfig())
	nvm.Deploy(governance.BuildConfig.StaticConfig())

	nvm.Deploy(saccount.EntryPointBuildConfig.StaticConfig())
	nvm.Deploy(saccount.VerifyingPaymasterBuildConfig.StaticConfig())
	nvm.Deploy(saccount.TokenPaymasterBuildConfig.StaticConfig())
	nvm.Deploy(saccount.SmartAccountFactoryBuildConfig.StaticConfig())

	return nvm
}

func (nvm *NativeVM) View() common.VirtualMachine {
	return &NativeVM{
		logger:                 nvm.logger,
		contractMethodID2Name:  nvm.contractMethodID2Name,
		contractBuildConfigMap: nvm.contractBuildConfigMap,
	}
}

func (nvm *NativeVM) Deploy(cfg *common.SystemContractStaticConfig) {
	// check system contract range
	if cfg.Address < common.SystemContractStartAddr || cfg.Address > common.SystemContractEndAddr {
		panic(fmt.Sprintf("this system contract %s is out of range", cfg.Address))
	}

	if _, ok := nvm.contractBuildConfigMap[cfg.Address]; ok {
		panic(fmt.Sprintf("this system contract %s has been deployed", cfg.Address))
	}
	nvm.contractBuildConfigMap[cfg.Address] = cfg

	id2Name := make(map[[4]byte]string)
	for methodName, method := range cfg.GetAbi().Methods {
		id2Name[[4]byte(method.ID)] = methodName
	}
	nvm.contractMethodID2Name[cfg.Address] = id2Name
	nvm.contractGenesisInitPriority = append(nvm.contractGenesisInitPriority, cfg.Address)
	nvm.setEVMPrecompiled(cfg.Address)
}

func (nvm *NativeVM) Run(data []byte, statefulArgs *vm.StatefulArgs) (execResult []byte, execErr error) {
	adaptor, ok := statefulArgs.EVM.StateDB.(*ledger.EvmStateDBAdaptor)
	if !ok {
		return nil, ErrInvalidStateDB
	}

	if statefulArgs.To == nil {
		return nil, ErrNotExistSystemContract
	}
	contractAddr := statefulArgs.To.Hex()
	contractBuildCfg := nvm.contractBuildConfigMap[contractAddr]
	if contractBuildCfg == nil {
		return nil, ErrNotExistSystemContract
	}
	contractInstance := contractBuildCfg.Build(common.NewVMContext(adaptor.StateLedger, statefulArgs.EVM, statefulArgs.From, statefulArgs.Value.ToBig(), statefulArgs.Output))
	defer func() {
		if err := recover(); err != nil {
			nvm.logger.Error(err)
			execErr = fmt.Errorf("%s", err)

			// get panic stack info
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, false)]

			lines := bytes.Split(stack, []byte("\n"))
			var errMsg []byte
			isRecord := false
			errLineNum := 0
			for _, line := range lines {
				if bytes.Contains(line, []byte("panic")) {
					isRecord = true
				}
				if isRecord {
					errMsg = append(errMsg, line...)
					errMsg = append(errMsg, []byte("\n")...)
					errLineNum++
				}
				if errLineNum > 20 {
					break
				}
			}
			nvm.logger.Errorf("panic stack info: %s", errMsg)
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
	args, err := nvm.parseArgs(contractBuildCfg, data, methodName)
	if err != nil {
		return nil, err
	}
	var inputs []reflect.Value
	for _, arg := range args {
		inputs = append(inputs, reflect.ValueOf(arg))
	}
	// convert input args, if input args contains slice struct which is defined in abi, convert to dest slice struct which is defined in golang
	inputs = convertInputArgs(method, inputs)

	// maybe panic when inputs mismatch, but we recover
	results := method.Call(inputs)

	var returnRes []any
	var returnErr error
	for i, result := range results {
		if checkIsError(result.Type()) {
			if i != len(results)-1 {
				panic(fmt.Sprintf("contract[%s] call method[%s] return error: %s, but not the last return value", contractAddr, methodName, returnErr))
			}
			if !result.IsNil() {
				returnErr = result.Interface().(error)
			}
		} else {
			returnRes = append(returnRes, result.Interface())
		}
	}

	nvm.logger.Debugf("Contract addr: %s, method name: %s, return result: %+v, return error: %s", contractAddr, methodName, returnRes, returnErr)

	if returnErr != nil {
		// if err is execution reverted, get reason
		var err *packer.RevertError
		if errors.As(returnErr, &err) {
			return err.Data, err.Err
		}
		revertErr := common.NewRevertStringError(returnErr.Error())
		return revertErr.Data, revertErr.Err
	}

	if returnRes != nil {
		return nvm.packOutputArgs(contractBuildCfg, methodName, returnRes...)
	}
	return nil, nil
}

// RequiredGas used in Inter-contract calls for EVM
func (nvm *NativeVM) RequiredGas(input []byte) uint64 {
	return RunSystemContractGas
}

// getMethodName quickly returns the name of a method of specified contract.
// This is a quick way to get the name of a method.
// The method name is the first 4 bytes of the keccak256 hash of the method signature.
// If the method name is not found, the empty string is returned.
func (nvm *NativeVM) getMethodName(contractAddr string, data []byte) (string, error) {
	// data is empty, maybe transfer to system contract
	if len(data) == 0 {
		return "", nil
	}

	if len(data) < 4 {
		nvm.logger.Errorf("system contract abi: data length is less than 4, %x", data)
		return "", ErrNotExistMethodName
	}

	methodID2Name, ok := nvm.contractMethodID2Name[contractAddr]
	if !ok {
		return "", ErrNotExistSystemContract
	}

	methodName, ok := methodID2Name[[4]byte(data[:4])]
	if !ok {
		nvm.logger.Errorf("system contract abi: could not locate method name, %x", data)
		return "", ErrNotExistMethodName
	}

	return methodName, nil
}

// parseArgs parse the arguments to specified interface by method name
func (nvm *NativeVM) parseArgs(contractBuildCfg *common.SystemContractStaticConfig, data []byte, methodName string) ([]any, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("msg data length is not improperly formatted: %q - Bytes: %+v", data, data)
	}

	// dinvmard method id
	msgData := data[4:]

	contractABI := contractBuildCfg.GetAbi()

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

// packOutputArgs pack the output arguments by method name
func (nvm *NativeVM) packOutputArgs(contractBuildCfg *common.SystemContractStaticConfig, methodName string, outputArgs ...any) ([]byte, error) {
	contractABI := contractBuildCfg.GetAbi()

	var args abi.Arguments
	if method, ok := contractABI.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("system contract abi: could not locate named method: %s", methodName)
	}

	return args.Pack(outputArgs...)
}

// unpackOutputArgs unpack the output arguments by method name
func (nvm *NativeVM) unpackOutputArgs(contractBuildCfg *common.SystemContractStaticConfig, methodName string, packed []byte) ([]any, error) {
	contractABI := contractBuildCfg.GetAbi()
	var args abi.Arguments
	if method, ok := contractABI.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("system contract abi: could not locate named method: %s", methodName)
	}

	return args.Unpack(packed)
}

func (nvm *NativeVM) setEVMPrecompiled(addr string) {
	// set system contractConstructor into vm.precompiled
	vm.PrecompiledAddressesByzantium = append(vm.PrecompiledAddressesByzantium, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesBerlin = append(vm.PrecompiledAddressesBerlin, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesHomestead = append(vm.PrecompiledAddressesHomestead, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesIstanbul = append(vm.PrecompiledAddressesIstanbul, ethcommon.HexToAddress(addr))
	vm.PrecompiledAddressesCancun = append(vm.PrecompiledAddressesCancun, ethcommon.HexToAddress(addr))

	vm.PrecompiledContractsBerlin[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsByzantium[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsHomestead[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsIstanbul[ethcommon.HexToAddress(addr)] = nvm
	vm.PrecompiledContractsCancun[ethcommon.HexToAddress(addr)] = nvm
}

func (nvm *NativeVM) GenesisInit(genesis *repo.GenesisConfig, lg ledger.StateLedger) error {
	nvm.initSystemContractCode(lg)

	for _, contractAddr := range nvm.contractGenesisInitPriority {
		contractBuildCfg := nvm.contractBuildConfigMap[contractAddr]
		if err := contractBuildCfg.Build(common.NewVMContextByExecutor(lg).DisableRecordLogToLedger()).GenesisInit(genesis); err != nil {
			return err
		}
	}

	return nil
}

// InitSystemContractCode init system contract code for compatible with ethereum abstract account
func (nvm *NativeVM) initSystemContractCode(lg ledger.StateLedger) {
	contractAddr := types.NewAddressByStr(common.SystemContractStartAddr)
	contractAddrBig := contractAddr.ETHAddress().Big()
	endContractAddrBig := types.NewAddressByStr(common.SystemContractEndAddr).ETHAddress().Big()
	for contractAddrBig.Cmp(endContractAddrBig) <= 0 {
		lg.SetCode(contractAddr, ethcommon.Hex2Bytes(common.EmptyContractBinCode))

		contractAddrBig.Add(contractAddrBig, big.NewInt(1))
		contractAddr = types.NewAddress(contractAddrBig.Bytes())
	}
}

func convertInputArgs(method reflect.Value, inputArgs []reflect.Value) []reflect.Value {
	outputArgs := make([]reflect.Value, len(inputArgs))
	copy(outputArgs, inputArgs)

	rt := method.Type()
	for i := 0; i < rt.NumIn(); i++ {
		argType := rt.In(i)
		if argType.Kind() == reflect.Slice && inputArgs[i].Type().Kind() == reflect.Slice {
			if argType.Elem().Kind() == reflect.Struct && inputArgs[i].Type().Elem().Kind() == reflect.Struct {
				slice := reflect.MakeSlice(argType, inputArgs[i].Len(), inputArgs[i].Len())
				for j := 0; j < inputArgs[i].Len(); j++ {
					v := reflect.New(argType.Elem()).Elem()
					for k := 0; k < v.NumField(); k++ {
						field := inputArgs[i].Index(j).Field(k)
						v.Field(k).Set(field)
					}
					slice.Index(j).Set(v)
				}

				outputArgs[i] = slice
			}
		}
	}
	return outputArgs
}

func checkIsError(tpe reflect.Type) bool {
	return tpe.AssignableTo(errorType)
}
