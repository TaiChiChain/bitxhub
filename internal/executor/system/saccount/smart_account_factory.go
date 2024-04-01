package saccount

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/interfaces"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

/**
 * A sample factory contract for SimpleAccount
 * A UserOperations "initCode" holds the address of the factory, and a method call (to createAccount, in this sample factory).
 * The factory's createAccount returns the target account address even if it is already installed.
 * This way, the entryPoint.getSenderAddress() can be called either before or after the account is created.
 */
type SmartAccountFactory struct {
	entryPoint  interfaces.IEntryPoint
	factoryAddr *types.Address
	logger      logrus.FieldLogger

	account     ledger.IAccount
	currentUser *ethcommon.Address
	currentLogs *[]common.Log
	stateLedger ledger.StateLedger
	currentEVM  *vm.EVM
}

func NewSmartAccountFactory(cfg *common.SystemContractConfig, entryPoint interfaces.IEntryPoint) *SmartAccountFactory {
	return &SmartAccountFactory{
		entryPoint:  entryPoint,
		factoryAddr: types.NewAddressByStr(common.AccountFactoryContractAddr),
		logger:      cfg.Logger,
	}
}

func (factory *SmartAccountFactory) SetContext(context *common.VMContext) {
	factory.currentUser = context.CurrentUser
	factory.currentLogs = context.CurrentLogs
	factory.stateLedger = context.StateLedger
	factory.currentEVM = context.CurrentEVM

	factory.account = factory.stateLedger.GetOrCreateAccount(factory.factoryAddr)
}

func (factory *SmartAccountFactory) CreateAccount(owner ethcommon.Address, salt *big.Int, guardian ethcommon.Address) *SmartAccount {
	addr := factory.GetAddress(owner, salt)

	sa := NewSmartAccount(factory.logger, factory.entryPoint)
	currentUser := factory.factoryAddr.ETHAddress()
	sa.SetContext(&common.VMContext{
		CurrentUser: &currentUser,
		StateLedger: factory.stateLedger,
		CurrentLogs: factory.currentLogs,
		CurrentEVM:  factory.currentEVM,
	})
	// initialize account or load account data
	sa.InitializeOrLoad(addr, owner, guardian, nil)

	return sa
}

func (factory *SmartAccountFactory) GetAddress(owner ethcommon.Address, salt *big.Int) ethcommon.Address {
	var saltBytes [32]byte
	copy(saltBytes[:], salt.Bytes())
	return crypto.CreateAddress2(owner, saltBytes, factory.factoryAddr.Bytes())
}
