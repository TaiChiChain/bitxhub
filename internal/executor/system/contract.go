package system

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

var (
	ErrNotExistSystemContract         = errors.New("not exist this system contract")
	ErrNotExistMethodName             = errors.New("not exist method name of this system contract")
	ErrNotExistSystemContractABI      = errors.New("not exist this system contract abi")
	ErrNotDeploySystemContract        = errors.New("not deploy this system contract")
	ErrNotImplementFuncSystemContract = errors.New("not implement the function for this system contract")
)

// SystemContractImpl handle abi decoding for parameters and abi encoding for return data
type SystemContractImpl struct {
	logger        logrus.FieldLogger
	stateLedger   ledger.StateLedger
	currentLogs   []common.Log
	currentHeight uint64

	// contract address mapping to method signature
	contract2MethodSig map[string]map[string][]byte
	// contract address mapping to contract abi
	contract2ABI map[string]abi.ABI
	// contract address mapping to contact instance
	contract2Instance map[string]common.InnerSystemContract
}

func New() common.SystemContract {
	sc := &SystemContractImpl{
		logger:             loggers.Logger(loggers.SystemContract),
		contract2MethodSig: common.Contract2MethodSig,
		contract2ABI:       common.Contract2ABI,
		contract2Instance:  make(map[string]common.InnerSystemContract),
	}

	cfg := &common.SystemContractConfig{
		Logger: sc.logger,
	}

	// deploy all system contract
	sc.Deploy(common.GovernanceContractAddr, governance.NewGov(cfg))
	sc.Deploy(common.EpochManagerContractAddr, base.NewEpochManager(cfg))
	sc.Deploy(common.WhiteListContractAddr, access.NewWhiteList(cfg))

	return sc
}

func (sc *SystemContractImpl) View() common.SystemContract {
	return &SystemContractImpl{
		logger:             sc.logger,
		contract2MethodSig: sc.contract2MethodSig,
		contract2ABI:       sc.contract2ABI,
		contract2Instance:  sc.contract2Instance,
	}
}

func (sc *SystemContractImpl) Deploy(addr string, instance common.InnerSystemContract) {
	if _, ok := sc.contract2Instance[addr]; ok {
		panic("deploy system contract repeated")
	}
	sc.contract2Instance[addr] = instance
}

func (sc *SystemContractImpl) Reset(currentHeight uint64, stateLedger ledger.StateLedger) {
	sc.stateLedger = stateLedger
	sc.currentHeight = currentHeight
	sc.currentLogs = make([]common.Log, 0)
}

func (sc *SystemContractImpl) Run(msg *vm.Message) (executionResult *vm.ExecutionResult, execErr error) {
	defer sc.saveLogs()
	defer func() {
		if err := recover(); err != nil {
			sc.logger.Error(err)
			execErr = fmt.Errorf("%s", err)
		}
	}()

	// get args and method, call the contract method
	contractAddr := msg.To.Hex()
	methodName, err := sc.getMethodName(contractAddr, msg.Data)
	if err != nil {
		return nil, err
	}
	contractInstance, ok := sc.contract2Instance[contractAddr]
	if !ok {
		return nil, ErrNotDeploySystemContract
	}

	// set context first
	contractInstance.SetContext(&common.SystemContractContext{
		StateLedger:   sc.stateLedger,
		CurrentHeight: sc.currentHeight,
		CurrentLogs:   &sc.currentLogs,
		CurrentUser:   &msg.From,
	})

	// method name may be propose, but we implement Propose
	// capitalize the first letter of a function
	funcName := methodName
	if len(methodName) >= 2 {
		funcName = fmt.Sprintf("%s%s", strings.ToUpper(methodName[:1]), methodName[1:])
	}
	sc.logger.Debugf("run system contract method name: %s\n", funcName)
	method := reflect.ValueOf(contractInstance).MethodByName(funcName)
	if !method.IsValid() {
		return nil, ErrNotImplementFuncSystemContract
	}
	args, err := sc.parseArgs(contractAddr, msg.Data, methodName)
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

	// calculate gas
	// wrap the result and return
	executionResult = &vm.ExecutionResult{
		UsedGas: common.CalculateDynamicGas(msg.Data),
	}

	sc.logger.Debugf("Contract addr: %s, method name: %s, return result: %+v, return error: %s", contractAddr, methodName, returnRes, returnErr)

	if returnErr != nil {
		executionResult.Err = returnErr
	}

	if returnRes != nil {
		returnData, err := sc.PackOutputArgs(contractAddr, methodName, returnRes...)
		if err != nil {
			sc.logger.Errorf("Pack return data error: %s", err)
			return executionResult, err
		}
		executionResult.ReturnData = returnData
	}

	return executionResult, nil
}

func (sc *SystemContractImpl) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	contractAddr := callArgs.To.Hex()
	methodName, err := sc.getMethodName(contractAddr, *callArgs.Data)
	if err != nil {
		return 0, err
	}

	_, err = sc.parseArgs(contractAddr, *callArgs.Data, methodName)
	if err != nil {
		return 0, err
	}
	return common.CalculateDynamicGas(*callArgs.Data), nil
}

// getMethodName quickly returns the name of a method of specified contract.
// This is a quick way to get the name of a method.
// The method name is the first 4 bytes of the keccak256 hash of the method signature.
// If the method name is not found, the empty string is returned.
func (sc *SystemContractImpl) getMethodName(contractAddr string, data []byte) (string, error) {
	method2Sig, ok := sc.contract2MethodSig[contractAddr]
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
func (sc *SystemContractImpl) parseArgs(contractAddr string, data []byte, methodName string) ([]any, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("msg data length is not improperly formatted: %q - Bytes: %+v", data, data)
	}

	// discard method id
	msgData := data[4:]

	contractABI, ok := sc.contract2ABI[contractAddr]
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
func (sc *SystemContractImpl) PackOutputArgs(contractAddr, methodName string, outputArgs ...any) ([]byte, error) {
	contractABI, ok := sc.contract2ABI[contractAddr]
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
func (sc *SystemContractImpl) UnpackOutputArgs(contractAddr, methodName string, packed []byte) ([]any, error) {
	contractABI, ok := sc.contract2ABI[contractAddr]
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
func (sc *SystemContractImpl) saveLogs() {
	sc.logger.Debugf("logs: %+v", sc.currentLogs)

	for _, currentLog := range sc.currentLogs {
		sc.stateLedger.AddLog(&types.EvmLog{
			Address: currentLog.Address,
			Topics:  currentLog.Topics,
			Data:    currentLog.Data,
			Removed: currentLog.Removed,
		})
	}
}

// IsSystemContract judge if it is system contract
// return true if system contract, false if not
func (sc *SystemContractImpl) IsSystemContract(addr *types.Address) bool {
	if addr == nil {
		return false
	}

	if _, ok := sc.contract2Instance[addr.String()]; ok {
		return true
	}
	return false
}

func InitGenesisData(genesis *repo.GenesisConfig, lg ledger.StateLedger) error {
	if err := base.InitEpochInfo(lg, genesis.EpochInfo.Clone()); err != nil {
		return err
	}
	if err := governance.InitCouncilMembers(lg, genesis.Admins, genesis.Balance); err != nil {
		return err
	}
	if err := governance.InitNodeMembers(lg, genesis.NodeNames, genesis.EpochInfo); err != nil {
		return err
	}
	if err := governance.InitGasParam(lg, genesis.EpochInfo); err != nil {
		return err
	}

	admins := lo.Map[*repo.Admin, string](genesis.Admins, func(x *repo.Admin, _ int) string {
		return x.Address
	})
	totalLength := len(admins) + len(genesis.InitWhiteListProviders) + len(genesis.Accounts)
	combined := make([]string, 0, totalLength)
	combined = append(combined, admins...)
	combined = append(combined, genesis.InitWhiteListProviders...)
	combined = append(combined, genesis.Accounts...)
	if err := access.InitProvidersAndWhiteList(lg, combined, genesis.InitWhiteListProviders); err != nil {
		return err
	}
	return nil
}
