package saccount

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_factory"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var SmartAccountFactoryBuildConfig = &common.SystemContractBuildConfig[*SmartAccountFactory]{
	Name:    "saccount_account_factory",
	Address: common.AccountFactoryContractAddr,
	AbiStr:  smart_account_factory.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *SmartAccountFactory {
		return &SmartAccountFactory{
			SystemContractBase: systemContractBase,
		}
	},
}

/**
 * A sample factory contract for SimpleAccount
 * A UserOperations "initCode" holds the address of the factory, and a method call (to createAccount, in this sample factory).
 * The factory's createAccount returns the target account address even if it is already installed.
 * This way, the entryPoint.getSenderAddress() can be called either before or after the account is created.
 */
type SmartAccountFactory struct {
	common.SystemContractBase
}

func (factory *SmartAccountFactory) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

func (factory *SmartAccountFactory) SetContext(ctx *common.VMContext) {
	factory.SystemContractBase.SetContext(ctx)
}

func (factory *SmartAccountFactory) CreateAccount(owner ethcommon.Address, salt *big.Int, guardian ethcommon.Address) (*SmartAccount, error) {
	addr := factory.GetAddress(owner, salt)

	sa := SmartAccountBuildConfig.BuildWithAddress(factory.CrossCallSystemContractContext(), addr)
	if err := sa.Initialize(owner, guardian); err != nil {
		return nil, err
	}

	return sa, nil
}

func (factory *SmartAccountFactory) GetAddress(owner ethcommon.Address, salt *big.Int) ethcommon.Address {
	var saltBytes [32]byte
	copy(saltBytes[:], salt.Bytes())
	return crypto.CreateAddress2(owner, saltBytes, factory.EthAddress.Bytes())
}
