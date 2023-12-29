package token

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Config struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	Balance     *big.Int
	Admins      []string
}

var (
	ErrTotalSupply         = errors.New("total supply below zero")
	ErrValue               = errors.New("input Value below zero")
	ErrInsufficientBalance = errors.New("Value exceeds balance")
	ErrEmptyAccount        = errors.New("account is empty")
	ErrContractAccount     = errors.New("account is not AXM IToken contract account")
)

const (
	TotalSupplyKey = "totalSupplyKey"
	DecimalsKey    = "decimalsKey"

	SymbolKey     = "symbolKey"
	NameKey       = "nameKey"
	AllowancesKey = "allowancesKey"
)

type totalSupplyArgs struct{}

type balanceOfArgs struct {
	Address ethcommon.Address
}

type transferArgs struct {
	To    ethcommon.Address
	Value *big.Int
}

type allowanceArgs struct {
	Owner   ethcommon.Address
	Spender ethcommon.Address
}

type approveArgs struct {
	Spender ethcommon.Address
	Value   *big.Int
}

type transferFromArgs struct {
	From  ethcommon.Address
	To    ethcommon.Address
	Value *big.Int
}

type nameArgs struct{}

type symbolArgs struct{}

type decimalsArgs struct{}

type mintArgs struct {
	Amount *big.Int
}

type burnArgs struct {
	Amount *big.Int
}

func (am *AxmManager) checkBeforeMint(account ethcommon.Address, value *big.Int) error {
	// todo: implement it
	return nil
}

func (am *AxmManager) checkBeforeBurn(account ethcommon.Address, value *big.Int) error {
	// Todo implement me
	return nil
}

func checkValue(value *big.Int) error {
	if value.Sign() < 0 {
		return ErrValue
	}
	return nil
}

func getAllowancesKey(owner, spender ethcommon.Address) string {
	return fmt.Sprintf("%s-%s-%s", AllowancesKey, owner.String(), spender.String())
}

func GenerateGenesisTokenConfig(genesis *repo.GenesisConfig) (Config, error) {
	balance, _ := new(big.Int).SetString(genesis.Balance, 10)
	totalSupply, _ := new(big.Int).SetString(genesis.Token.TotalSupply, 10)
	// calculate totalSupply - (balance * len(admins))
	contractBalance := new(big.Int).Set(totalSupply)
	transf := new(big.Int).Mul(balance, big.NewInt(int64(len(genesis.Admins))))
	if contractBalance.Cmp(transf) < 0 {
		return Config{}, ErrTotalSupply
	}
	adminAddrs := make([]string, 0)
	lo.ForEach(genesis.Admins, func(admin *repo.Admin, index int) {
		adminAddrs = append(adminAddrs, admin.Address)
	})
	lo.ForEach(genesis.Accounts, func(addr string, index int) {
		adminAddrs = append(adminAddrs, addr)
	})
	tokenConfig := Config{
		Name:        genesis.Token.Name,
		Symbol:      genesis.Token.Symbol,
		Decimals:    genesis.Token.Decimals,
		Admins:      adminAddrs,
		TotalSupply: contractBalance,
		Balance:     balance,
	}
	return tokenConfig, nil
}
