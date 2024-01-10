package axc

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

var _ token.ICredit = (*Manager)(nil)

var (
	communityAddr ethcommon.Address
)

type byter interface {
	Bytes() []byte
}

type Manager struct {
	logger      logrus.FieldLogger
	account     ledger.IAccount
	msgFrom     ethcommon.Address
	stateLedger ledger.StateLedger
	logs        []common.Log
}

func Init(lg ledger.StateLedger, config Config) error {
	contractAccount := lg.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))

	contractAccount.SetState([]byte(NameKey), []byte(config.Name))
	contractAccount.SetState([]byte(SymbolKey), []byte(config.Symbol))
	contractAccount.SetState([]byte(DecimalsKey), []byte{config.Decimals})
	contractAccount.SetState([]byte(TotalSupplyKey), config.TotalSupply.Bytes())

	lo.ForEach(config.Receivers, func(entity *Distribution, _ int) {
		addr := types.NewAddressByStr(entity.Addr)
		contractAccount.SetState([]byte(getLockedBalanceKey(addr.ETHAddress())), entity.InitEmission.Bytes())
		if entity.Name == "Community" {
			contractAccount.SetState([]byte(getBalancesKey(addr.ETHAddress())), entity.TotalValue.Bytes())
			communityAddr = addr.ETHAddress()
		}
	})

	if communityAddr == (ethcommon.Address{}) {
		return ErrNoCommunityAccount
	}

	return nil
}

func (ac *Manager) Name() string {
	ok, name := ac.account.GetState([]byte(NameKey))
	if !ok {
		return ""
	}
	return string(name)
}

func (ac *Manager) Symbol() string {
	ok, symbol := ac.account.GetState([]byte(SymbolKey))
	if !ok {
		return ""
	}
	return string(symbol)
}

func (ac *Manager) Decimals() uint8 {
	ok, decimals := ac.account.GetState([]byte(DecimalsKey))
	if !ok {
		return 0
	}
	return decimals[0]
}

func (ac *Manager) TotalSupply() *big.Int {
	ok, totalSupply := ac.account.GetState([]byte(TotalSupplyKey))
	if !ok {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(totalSupply)
}

func (ac *Manager) BalanceOf(account ethcommon.Address) *big.Int {
	if ok, balanceBytes := ac.account.GetState([]byte(getBalancesKey(account))); ok {
		return new(big.Int).SetBytes(balanceBytes)
	}
	return big.NewInt(0)
}

func (ac *Manager) Allowance(owner, spender ethcommon.Address) *big.Int {
	return ac.getAllowance(owner, spender)
}

func (ac *Manager) getAllowance(owner, spender ethcommon.Address) *big.Int {
	if ok, v := ac.account.GetState([]byte(getAllowancesKey(owner, spender))); ok {
		return new(big.Int).SetBytes(v)
	}
	return big.NewInt(0)
}

func (ac *Manager) Approve(spender ethcommon.Address, value *big.Int) error {
	return ac.approve(ac.msgFrom, spender, value)
}

func (ac *Manager) approve(owner, spender ethcommon.Address, value *big.Int) error {
	var err error
	if err = checkValue(value); err != nil {
		return err
	}

	ac.recordLog(approveMethod, []byter{owner, spender}, []byter{value})
	ac.account.SetState([]byte(getAllowancesKey(owner, spender)), value.Bytes())
	return nil
}

func (ac *Manager) Transfer(recipient ethcommon.Address, value *big.Int) error {
	from := types.NewAddressByStr(ac.msgFrom.String()).ETHAddress()
	return ac.transfer(from, recipient, value, getBalancesKey)
}

func (ac *Manager) TransferFrom(sender, recipient ethcommon.Address, value *big.Int) error {
	// get allowance for <sender, msgFrom>
	allowance := ac.getAllowance(sender, ac.msgFrom)
	if allowance.Cmp(value) < 0 {
		return ErrNotEnoughAllowance
	}

	if err := ac.transfer(sender, recipient, value, getBalancesKey); err != nil {
		return err
	}

	return ac.approve(sender, ac.msgFrom, new(big.Int).Sub(allowance, value))
}

func (ac *Manager) TransferLocked(recipient ethcommon.Address, value *big.Int) error {
	from := types.NewAddressByStr(ac.msgFrom.String()).ETHAddress()
	return ac.transfer(from, recipient, value, getLockedBalanceKey)
}

func (ac *Manager) transfer(sender, recipient ethcommon.Address, value *big.Int, getKeyFunc func(ethcommon.Address) string) error {
	if sender == (ethcommon.Address{}) || recipient == (ethcommon.Address{}) {
		return ErrEmptyAccount
	}
	ok, senderBalanceBytes := ac.account.GetState([]byte(getKeyFunc(sender)))
	if !ok {
		return ErrInsufficientBalance
	}
	senderBalance := new(big.Int).SetBytes(senderBalanceBytes)
	if senderBalance.Cmp(value) < 0 {
		return ErrInsufficientBalance
	}

	_, receiverBalanceBytes := ac.account.GetState([]byte(getKeyFunc(recipient)))
	receiverBalance := new(big.Int).SetBytes(receiverBalanceBytes)

	newSenderBalance := new(big.Int).Sub(senderBalance, value)
	newReceiverBalance := new(big.Int).Add(receiverBalance, value)

	ac.account.SetState([]byte(getKeyFunc(sender)), newSenderBalance.Bytes())
	ac.account.SetState([]byte(getKeyFunc(recipient)), newReceiverBalance.Bytes())

	ac.recordLog("transfer", []byter{sender, recipient}, []byter{value})

	return nil
}

func (ac *Manager) Unlock(_ ethcommon.Address, _ *big.Int) error {
	// todo: implement it
	return nil
}

// recordLog record execution log for governance
func (ac *Manager) recordLog(method string, topics []byter, data []byter) {
	methodSig := Event2Sig[method]

	sigHash := sha3.NewLegacyKeccak256()
	sigHash.Write([]byte(methodSig))
	signature := sigHash.Sum(nil)[:4]
	currentLog := common.Log{
		Address: types.NewAddressByStr(common.AXCContractAddr),
	}
	currentLog.Topics = append(currentLog.Topics, types.NewHash(signature))
	for _, topic := range topics {
		currentLog.Topics = append(currentLog.Topics, types.NewHash(topic.Bytes()))
	}
	var currentData []byte
	for _, d := range data {
		currentData = append(currentData, types.NewHash(d.Bytes()).Bytes()...)
	}
	currentLog.Data = currentData

	ac.logs = append(ac.logs, currentLog)
}

func (ac *Manager) Logs() []common.Log {
	return ac.logs
}

func New(cfg *common.SystemContractConfig) *Manager {
	return &Manager{
		logger: cfg.Logger,
	}
}

func (ac *Manager) SetContext(context *common.VMContext) {
	ac.account = context.StateLedger.GetOrCreateAccount(types.NewAddressByStr(common.AXCContractAddr))
	ac.stateLedger = context.StateLedger
	ac.msgFrom = *context.CurrentUser
	ac.logs = *context.CurrentLogs
}
