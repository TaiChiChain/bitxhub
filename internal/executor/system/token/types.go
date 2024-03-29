package token

import (
	"fmt"
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	Accounts    []*InitialAccount
}

type InitialAccount struct {
	Address *types.Address
	Balance *big.Int
}

var (
	ErrTotalSupply         = errors.New("total supply below zero")
	ErrValue               = errors.New("input Value below zero")
	ErrInsufficientBalance = errors.New("Value exceeds balance")
	ErrEmptyAccount        = errors.New("account is empty")
	ErrContractAccount     = errors.New("account is not Token IToken contract account")
)

const (
	TotalSupplyKey = "totalSupplyKey"
	DecimalsKey    = "decimalsKey"

	SymbolKey     = "symbolKey"
	NameKey       = "nameKey"
	AllowancesKey = "allowancesKey"

	totalSupplyMethod = "totalSupply"
	mintMethod        = "mint"
	burnMethod        = "burn"
)

func (am *Manager) checkBeforeMint(account ethcommon.Address, value *big.Int) error {
	// todo: implement it
	return nil
}

func (am *Manager) checkBeforeBurn(account ethcommon.Address, value *big.Int) error {
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

func GenerateConfig(genesis *repo.GenesisConfig) (Config, error) {
	var err error
	totalSupply, ok := new(big.Int).SetString(genesis.Axc.TotalSupply, 10)
	if !ok || totalSupply.Sign() < 0 {
		err = fmt.Errorf("invalid balance: %s", totalSupply)
	}
	initialAccounts := lo.Map(genesis.Accounts, func(account *repo.Account, index int) *InitialAccount {
		balance, ok := new(big.Int).SetString(account.Balance, 10)
		if !ok || balance.Sign() < 0 {
			err = fmt.Errorf("invalid balance: %s", totalSupply)
			return nil
		}
		return &InitialAccount{
			Address: types.NewAddressByStr(account.Address),
			Balance: balance,
		}
	})

	if err != nil {
		return Config{}, err
	}

	tokenConfig := Config{
		Name:        genesis.Axc.Name,
		Symbol:      genesis.Axc.Symbol,
		Decimals:    genesis.Axc.Decimals,
		Accounts:    initialAccounts,
		TotalSupply: totalSupply,
	}
	return tokenConfig, nil
}
