package token

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type Manager struct {
	logger      logrus.FieldLogger
	account     ledger.IAccount
	msgFrom     ethcommon.Address
	stateLedger ledger.StateLedger
}

func Init(lg ledger.StateLedger, config Config) error {
	contractAccount := lg.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))

	contractAccount.SetState([]byte(NameKey), []byte(config.Name))
	contractAccount.SetState([]byte(SymbolKey), []byte(config.Symbol))
	contractAccount.SetState([]byte(DecimalsKey), []byte{config.Decimals})
	contractAccount.SetState([]byte(TotalSupplyKey), config.TotalSupply.Bytes())

	// init balance for all accounts
	for _, account := range config.Accounts {
		supplyAccount := lg.GetOrCreateAccount(account.Address)
		supplyAccount.AddBalance(account.Balance)
	}
	return nil
}

func (am *Manager) TotalSupply() *big.Int {
	ok, totalSupply := am.account.GetState([]byte(TotalSupplyKey))
	if !ok {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(totalSupply)
}

func (am *Manager) Mint(amount *big.Int) error {
	if err := am.checkBeforeMint(am.msgFrom, amount); err != nil {
		return err
	}
	return mint(am.account, amount)
}

func mint(contractAccount ledger.IAccount, value *big.Int) error {
	if err := checkValue(value); err != nil {
		return err
	}

	contractAccount.AddBalance(value)

	return changeTotalSupply(contractAccount, value, true)
}

func changeTotalSupply(acc ledger.IAccount, amount *big.Int, increase bool) error {
	if acc.GetAddress().String() != common.AXCContractAddr {
		return ErrContractAccount
	}

	totalSupply := big.NewInt(0)
	ok, v := acc.GetState([]byte(TotalSupplyKey))
	if ok {
		totalSupply = new(big.Int).SetBytes(v)
	}

	if increase {
		totalSupply.Add(totalSupply, amount)
	} else {
		totalSupply.Sub(totalSupply, amount)
		if totalSupply.Cmp(big.NewInt(0)) < 0 {
			return ErrTotalSupply
		}
	}

	acc.SetState([]byte(TotalSupplyKey), totalSupply.Bytes())

	return nil
}

func (am *Manager) Burn(value *big.Int) error {
	// todo: check role(only council contract account)
	if err := am.checkBeforeBurn(am.msgFrom, value); err != nil {
		return err
	}
	return burn(am.account, value)
}

func burn(contractAccount ledger.IAccount, value *big.Int) error {
	// 1. check Value arg
	if err := checkValue(value); err != nil {
		return err
	}

	// 2. check if enough balance
	oldBalance := contractAccount.GetBalance()
	if oldBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	if err := changeTotalSupply(contractAccount, value, false); err != nil {
		return err
	}
	contractAccount.SubBalance(value)

	return nil
}

func New(cfg *common.SystemContractConfig) *Manager {
	return &Manager{
		logger: cfg.Logger,
	}
}

func (am *Manager) SetContext(context *common.VMContext) {
	am.account = context.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	am.stateLedger = context.StateLedger
	am.msgFrom = *context.CurrentUser
}
