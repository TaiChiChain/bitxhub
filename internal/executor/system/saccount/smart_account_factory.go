package saccount

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_factory"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_factory_client"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	accountFactoryKey = "accountFactory"
)

var SmartAccountFactoryBuildConfig = &common.SystemContractBuildConfig[*SmartAccountFactory]{
	Name:    "saccount_account_factory",
	Address: common.AccountFactoryContractAddr,
	AbiStr:  smart_account_factory_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *SmartAccountFactory {
		return &SmartAccountFactory{
			Ownable: Ownable{SystemContractBase: systemContractBase},
		}
	},
}

var _ common.SystemContract = (*SmartAccountFactory)(nil)
var _ smart_account_factory.SmartAccountFactory = (*SmartAccountFactory)(nil)

/**
 * A sample factory contract for SimpleAccount
 * A UserOperations "initCode" holds the address of the factory, and a method call (to createAccount, in this sample factory).
 * The factory's createAccount returns the target account address even if it is already installed.
 * This way, the entryPoint.getSenderAddress() can be called either before or after the account is created.
 */
type SmartAccountFactory struct {
	Ownable

	// accountFactory is the address of the factory contract which implements function GetAddress
	accountFactory *common.VMSlot[ethcommon.Address]
}

func (factory *SmartAccountFactory) GenesisInit(genesis *repo.GenesisConfig) error {
	factory.Init(ethcommon.HexToAddress(genesis.SmartAccountAdmin))
	return nil
}

func (factory *SmartAccountFactory) SetContext(ctx *common.VMContext) {
	factory.Ownable.SetContext(ctx)

	factory.accountFactory = common.NewVMSlot[ethcommon.Address](factory.StateAccount, accountFactoryKey)
}

func (factory *SmartAccountFactory) CreateAccount2(owner ethcommon.Address, salt *big.Int, guardian ethcommon.Address) (ethcommon.Address, error) {
	var addr ethcommon.Address
	factoryAddr, err := factory.GetAccountFactory()
	if err != nil {
		addr, _ = factory.GetAddress(owner, salt)
	} else {
		// call the factory contract to get address
		packed, err := factory.Abi.Pack("getAddress", owner, salt)
		if err != nil {
			return ethcommon.Address{}, err
		}
		returnData, _, err := factory.CrossCallEVMContract(big.NewInt(MaxCallGasLimit), factoryAddr, packed)
		if err != nil {
			return ethcommon.Address{}, err
		}
		addr = ethcommon.BytesToAddress(returnData)
	}

	sa := SmartAccountBuildConfig.BuildWithAddress(factory.CrossCallSystemContractContext(), addr)
	if err := sa.Initialize(owner, guardian); err != nil {
		return ethcommon.Address{}, err
	}

	return sa.EthAddress, nil
}

func (factory *SmartAccountFactory) CreateAccount(owner ethcommon.Address, salt *big.Int) (ethcommon.Address, error) {
	return factory.CreateAccount2(owner, salt, ethcommon.Address{})
}

func (factory *SmartAccountFactory) GetAddress(owner ethcommon.Address, salt *big.Int) (ethcommon.Address, error) {
	var saltBytes [32]byte
	copy(saltBytes[:], ethcommon.LeftPadBytes(salt.Bytes(), 32))
	return crypto.CreateAddress2(owner, saltBytes, factory.EthAddress.Bytes()), nil
}

func (factory *SmartAccountFactory) GetAccountFactory() (ethcommon.Address, error) {
	return factory.accountFactory.MustGet()
}

func (factory *SmartAccountFactory) SetAccountFactory(factoryAddr ethcommon.Address) error {
	return factory.accountFactory.Put(factoryAddr)
}
