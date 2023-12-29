package token

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

const (
	nameMethod         = "name"
	symbolMethod       = "symbol"
	totalSupplyMethod  = "totalSupply"
	decimalsMethod     = "decimals"
	balanceOfMethod    = "balanceOf"
	transferMethod     = "transfer"
	approveMethod      = "approve"
	allowanceMethod    = "allowance"
	transferFromMethod = "transferFrom"
	mintMethod         = "mint"
	burnMethod         = "burn"
)

var AxmManagerMethod2Sig = map[string]string{
	totalSupplyMethod:  "totalSupply()",
	balanceOfMethod:    "balanceOf(address)",
	transferMethod:     "transfer(address,uint256)",
	allowanceMethod:    "allowance(address,address)",
	approveMethod:      "approve(address,uint256)",
	transferFromMethod: "transferFrom(address,address,uint256)",
	nameMethod:         "name()",
	symbolMethod:       "symbol()",
	decimalsMethod:     "decimals()",
	mintMethod:         "mint(uint256)",
	burnMethod:         "burn(uint256)",
}

var _ IToken = (*AxmManager)(nil)

type AxmManager struct {
	logger      logrus.FieldLogger
	account     ledger.IAccount
	msgFrom     ethcommon.Address
	stateLedger ledger.StateLedger
}

func InitAxmTokenManager(lg ledger.StateLedger, config Config) error {
	contractAccount := lg.GetOrCreateAccount(types.NewAddressByStr(common.TokenManagerContractAddr))

	contractAccount.SetState([]byte(NameKey), []byte(config.Name))
	contractAccount.SetState([]byte(SymbolKey), []byte(config.Symbol))
	contractAccount.SetState([]byte(DecimalsKey), []byte{config.Decimals})

	var err error

	// init balance for contract contractAccount
	if err = mint(contractAccount, config.TotalSupply); err != nil {
		return err
	}

	lo.ForEach(config.Admins, func(admin string, _ int) {
		adminAccount := lg.GetOrCreateAccount(types.NewAddressByStr(admin))
		if err = transfer(contractAccount, adminAccount, config.Balance); err != nil {
			return
		}
	})

	return err
}

func (am *AxmManager) Name() string {
	ok, name := am.account.GetState([]byte(NameKey))
	if !ok {
		return ""
	}
	return string(name)
}

func (am *AxmManager) Symbol() string {
	ok, symbol := am.account.GetState([]byte(SymbolKey))
	if !ok {
		return ""
	}
	return string(symbol)
}

func (am *AxmManager) Decimals() uint8 {
	ok, decimals := am.account.GetState([]byte(DecimalsKey))
	if !ok {
		return 0
	}
	return decimals[0]
}

func (am *AxmManager) TotalSupply() *big.Int {
	ok, totalSupply := am.account.GetState([]byte(TotalSupplyKey))
	if !ok {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(totalSupply)
}

func (am *AxmManager) BalanceOf(account ethcommon.Address) *big.Int {
	acc := am.stateLedger.GetAccount(types.NewAddressByStr(account.String()))
	if acc == nil {
		return big.NewInt(0)
	}
	return acc.GetBalance()
}

func (am *AxmManager) Mint(amount *big.Int) error {
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
	if acc.GetAddress().String() != common.TokenManagerContractAddr {
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

func (am *AxmManager) Burn(value *big.Int) error {
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

func (am *AxmManager) Allowance(owner, spender ethcommon.Address) *big.Int {
	return am.getAllowance(owner, spender)
}

func (am *AxmManager) getAllowance(owner, spender ethcommon.Address) *big.Int {
	ok, v := am.account.GetState([]byte(getAllowancesKey(owner, spender)))
	if !ok {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(v)
}

func (am *AxmManager) Approve(spender ethcommon.Address, value *big.Int) error {
	return am.approve(am.msgFrom, spender, value)
}

func (am *AxmManager) approve(owner, spender ethcommon.Address, value *big.Int) error {
	var err error
	if err = checkValue(value); err != nil {
		return err
	}

	am.account.SetState([]byte(getAllowancesKey(owner, spender)), value.Bytes())
	return nil
}

func (am *AxmManager) Transfer(recipient ethcommon.Address, value *big.Int) error {
	fromAcc := am.stateLedger.GetAccount(types.NewAddressByStr(am.msgFrom.String()))
	toAcc := am.stateLedger.GetAccount(types.NewAddressByStr(recipient.String()))
	return transfer(fromAcc, toAcc, value)
}

func (am *AxmManager) TransferFrom(sender, recipient ethcommon.Address, value *big.Int) error {
	// get allowance for <sender, msgFrom>
	allowance := am.getAllowance(sender, am.msgFrom)
	if allowance.Cmp(value) < 0 {
		return errors.New("not enough allowance")
	}

	fromAcc := am.stateLedger.GetAccount(types.NewAddressByStr(sender.String()))
	toAcc := am.stateLedger.GetAccount(types.NewAddressByStr(recipient.String()))
	if err := transfer(fromAcc, toAcc, value); err != nil {
		return err
	}

	return am.approve(sender, am.msgFrom, new(big.Int).Sub(allowance, value))
}

func transfer(sender, recipient ledger.IAccount, value *big.Int) error {
	if sender == nil || recipient == nil {
		return ErrEmptyAccount
	}
	senderBalance := sender.GetBalance()
	if senderBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	sender.SubBalance(value)

	recipient.AddBalance(value)
	return nil
}

func NewTokenManager(cfg *common.SystemContractConfig) *AxmManager {
	return &AxmManager{
		logger: cfg.Logger,
	}
}

func (am *AxmManager) SetContext(context *common.VMContext) {
	am.account = context.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.TokenManagerContractAddr))
	am.stateLedger = context.StateLedger
	am.msgFrom = *context.CurrentUser
}
