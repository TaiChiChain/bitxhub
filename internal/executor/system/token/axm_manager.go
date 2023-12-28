package token

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

var _ IToken = (*AxmManager)(nil)

var (
	axmManagerABI *abi.ABI
)

func init() {
	amAbi, err := abi.JSON(strings.NewReader(axmManagerABIData))
	if err != nil {
		panic(err)
	}
	axmManagerABI = &amAbi
}

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

func (am *AxmManager) Reset(lastHeight uint64, stateLedger ledger.StateLedger) {
	am.account = stateLedger.GetOrCreateAccount(types.NewAddressByStr(common.TokenManagerContractAddr))
	am.stateLedger = stateLedger
}

func (am *AxmManager) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	am.resetMsg(msg.From)
	result := &vm.ExecutionResult{}
	ret, err := func() ([]byte, error) {
		args, method, err := common.ParseContractCallArgs(axmManagerABI, msg.Data, methodSig2ArgsReceiverConstructor)
		if err != nil {
			return nil, err
		}
		switch t := args.(type) {
		case *totalSupplyArgs:
			totalSupply := am.TotalSupply()
			return method.Outputs.Pack(totalSupply)
		case *balanceOfArgs:
			balance := am.BalanceOf(t.Address)
			return method.Outputs.Pack(balance)
		case *transferArgs:
			if err = am.Transfer(t.To, t.Value); err != nil {
				return nil, err
			}
			return method.Outputs.Pack(true)
		case *allowanceArgs:
			allowance := am.Allowance(t.Owner, t.Spender)
			return method.Outputs.Pack(allowance)
		case *approveArgs:
			if err = am.Approve(t.Spender, t.Value); err != nil {
				return nil, err
			}
			return method.Outputs.Pack(true)
		case *transferFromArgs:
			if err = am.TransferFrom(t.From, t.To, t.Value); err != nil {
				return nil, err
			}
			return method.Outputs.Pack(true)
		case *nameArgs:
			name := am.Name()
			return method.Outputs.Pack(name)
		case *symbolArgs:
			symbol := am.Symbol()
			return method.Outputs.Pack(symbol)
		case *decimalsArgs:
			decimals := am.Decimals()
			return method.Outputs.Pack(decimals)
		case *mintArgs:
			if err = am.Mint(t.Amount); err != nil {
				return nil, err
			}
			return method.Outputs.Pack(true)
		case *burnArgs:
			if err = am.Burn(t.Amount); err != nil {
				return nil, err
			}
			return method.Outputs.Pack()
		default:
			return nil, errors.Errorf("%v: not support method", vm.ErrExecutionReverted)
		}
	}()
	if err != nil {
		result.Err = vm.ErrExecutionReverted
		result.ReturnData = []byte(err.Error())
	} else {
		result.ReturnData = ret
	}
	result.UsedGas = common.CalculateDynamicGas(msg.Data)
	return result, nil
}

func (am *AxmManager) resetMsg(from ethcommon.Address) {
	am.msgFrom = from
}

func (am *AxmManager) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	if callArgs == nil {
		return 0, errors.New("callArgs is nil")
	}

	var data []byte
	if callArgs.Data != nil {
		data = *callArgs.Data
	}

	_, _, err := common.ParseContractCallArgs(axmManagerABI, data, methodSig2ArgsReceiverConstructor)
	if err != nil {
		return 0, errors.Errorf("%v: %v", vm.ErrExecutionReverted, err)
	}

	return common.CalculateDynamicGas(*callArgs.Data), nil
}
