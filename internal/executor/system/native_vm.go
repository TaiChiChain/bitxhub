package system

import (
	_ "embed"
	"fmt"
	"math/big"
	"strings"

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

var _ common.VirtualMachine = (*NativeVM)(nil)

// NativeVM handle abi decoding for parameters and abi encoding for return data
type NativeVM struct {
	logger logrus.FieldLogger

	// contract addr -> contract cfg
	contractBuildConfigMap map[string]*common.SystemContractStaticConfig

	// priority -> contract addr
	contractGenesisInitPriority []string
}

func New() *NativeVM {
	nvm := &NativeVM{
		logger:                 loggers.Logger(loggers.SystemContract),
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
	// convert all contract address to lower due to EIP-55
	contractAddr = strings.ToLower(contractAddr)
	contractBuildCfg := nvm.contractBuildConfigMap[contractAddr]
	if contractBuildCfg == nil {
		return nil, ErrNotExistSystemContract
	}
	contractInstance := contractBuildCfg.Build(common.NewVMContext(adaptor.StateLedger, statefulArgs.EVM, statefulArgs.From, statefulArgs.Value.ToBig(), statefulArgs.Output))
	defer common.Recovery(nvm.logger, execErr)

	contractAbi := contractBuildCfg.GetAbi()
	returnRes, returnErr := common.CallSystemContract(nvm.logger, contractInstance, contractAddr, contractAbi, data)
	if returnErr == common.ErrMethodNotFound {
		// if method not found in abi, return nil, nil, like solidity fallback
		return nil, nil
	}

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
		methodName, _ := common.ParseMethodName(contractAbi, data)
		return common.PackOutputArgs(contractAbi, methodName, returnRes...)
	}

	return nil, nil
}

// RequiredGas used in Inter-contract calls for EVM
func (nvm *NativeVM) RequiredGas(input []byte) uint64 {
	return RunSystemContractGas
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
