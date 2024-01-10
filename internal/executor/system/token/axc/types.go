package axc

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
	Receivers   []*Distribution
}

type Distribution struct {
	Name         string
	Addr         string
	TotalValue   *big.Int
	InitEmission *big.Int
}

var (
	ErrTotalSupply         = errors.New("total supply below zero")
	ErrValue               = errors.New("input Value below zero")
	ErrInsufficientBalance = errors.New("Value exceeds balance")
	ErrEmptyAccount        = errors.New("account is empty")
	ErrNoCommunityAccount  = errors.New("entity[Community] is required in incentive")
	ErrNotEnoughAllowance  = errors.New("not enough allowance")
)

const (
	TotalSupplyKey = "axcTotalSupplyKey"
	DecimalsKey    = "axcDecimalsKey"

	SymbolKey     = "axcSymbolKey"
	NameKey       = "axcNameKey"
	AllowancesKey = "axcAllowancesKey"

	// BalancesKey is a map stores axc balance, mapping(address => uint256)
	BalancesKey = "axcBalances"
	// LockedBalanceKey is a map stores axc locked balance, mapping(address => uint256)
	LockedBalanceKey = "axcLockedBalance"

	nameMethod         = "name"
	symbolMethod       = "symbol"
	totalSupplyMethod  = "totalSupply"
	decimalsMethod     = "decimals"
	balanceOfMethod    = "balanceOf"
	transferMethod     = "transfer"
	approveMethod      = "approve"
	allowanceMethod    = "allowance"
	transferFromMethod = "transferFrom"
	unlockMethod       = "unlock"
)

var Method2Sig = map[string]string{
	totalSupplyMethod:  "totalSupply()",
	balanceOfMethod:    "balanceOf(address)",
	transferMethod:     "transfer(address,uint256)",
	allowanceMethod:    "allowance(address,address)",
	approveMethod:      "approve(address,uint256)",
	transferFromMethod: "transferFrom(address,address,uint256)",
	nameMethod:         "name()",
	symbolMethod:       "symbol()",
	decimalsMethod:     "decimals()",
	unlockMethod:       "unlock(address,uint256)",
}

var Event2Sig = map[string]string{
	approveMethod:  "Approval(address,address,uint256)",
	transferMethod: "Transfer(address,address,uint256)",
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

func getBalancesKey(owner ethcommon.Address) string {
	return fmt.Sprintf("%s-%s", BalancesKey, owner.String())
}

func getLockedBalanceKey(owner ethcommon.Address) string {
	return fmt.Sprintf("%s-%s", LockedBalanceKey, owner.String())
}

func GenerateConfig(genesis *repo.GenesisConfig) (Config, error) {
	totalSupply, _ := new(big.Int).SetString(genesis.Axc.TotalSupply, 10)
	if totalSupply.Cmp(big.NewInt(0)) < 0 {
		return Config{}, ErrTotalSupply
	}
	receivers := make([]*Distribution, 0)
	lo.ForEach(genesis.Incentive.Distributions, func(entity *repo.Distribution, index int) {
		percentage := big.NewInt(int64(entity.Percentage * 100))
		totalValue := new(big.Int).Div(new(big.Int).Mul(percentage, totalSupply), big.NewInt(100))
		emission := big.NewInt(int64(entity.InitEmission * 100))
		emissionValue := new(big.Int).Div(new(big.Int).Mul(emission, totalValue), big.NewInt(100))
		receivers = append(receivers, &Distribution{
			Name:         entity.Name,
			Addr:         entity.Addr,
			TotalValue:   totalValue,
			InitEmission: emissionValue,
		})
	})
	tokenConfig := Config{
		Name:        genesis.Axc.Name,
		Symbol:      genesis.Axc.Symbol,
		Decimals:    genesis.Axc.Decimals,
		Receivers:   receivers,
		TotalSupply: totalSupply,
	}
	return tokenConfig, nil
}
