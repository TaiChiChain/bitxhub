package token

import (
	"math/big"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/solidity/axc"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	TotalSupplyKey = "totalSupplyKey"
)

var AXCBuildConfig = &common.SystemContractBuildConfig[*AXC]{
	Name:    "token_axc",
	Address: common.AXCContractAddr,
	AbiStr:  axc.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *AXC {
		return &AXC{
			SystemContractBase: systemContractBase,
		}
	},
}

type AXC struct {
	common.SystemContractBase

	totalSupply *common.VMSlot[*big.Int]
}

func (axc *AXC) GenesisInit(genesis *repo.GenesisConfig) error {
	totalSupply := genesis.Axc.TotalSupply.ToBigInt()
	for _, account := range genesis.Accounts {
		balance := account.Balance.ToBigInt()
		supplyAccount := axc.Ctx.StateLedger.GetOrCreateAccount(types.NewAddressByStr(account.Address))
		supplyAccount.AddBalance(balance)
	}
	if err := axc.totalSupply.Put(totalSupply); err != nil {
		return err
	}
	return nil
}

func (axc *AXC) SetContext(context *common.VMContext) {
	axc.SystemContractBase.SetContext(context)

	axc.totalSupply = common.NewVMSlot[*big.Int](axc.StateAccount, TotalSupplyKey)
}

func (axc *AXC) TotalSupply() (*big.Int, error) {
	return axc.totalSupply.MustGet()
}

func (axc *AXC) Mint(amount *big.Int) error {
	if amount.Sign() < 0 {
		return errors.New("mint amount below zero")
	}

	axc.StateAccount.AddBalance(amount)

	totalSupply, err := axc.totalSupply.MustGet()
	if err != nil {
		return err
	}
	totalSupply = totalSupply.Add(totalSupply, amount)
	if err := axc.totalSupply.Put(totalSupply); err != nil {
		return err
	}
	return nil
}

func (axc *AXC) Burn(amount *big.Int) error {
	if amount.Sign() < 0 {
		return errors.New("burn amount below zero")
	}

	// 2. check if enough balance
	oldBalance := axc.StateAccount.GetBalance()
	if oldBalance.Cmp(amount) < 0 {
		return errors.New("Value exceeds balance")
	}

	totalSupply, err := axc.totalSupply.MustGet()
	if err != nil {
		return err
	}

	if totalSupply.Cmp(amount) < 0 {
		return errors.New("Value exceeds balance")
	}

	totalSupply = totalSupply.Sub(totalSupply, amount)
	if err := axc.totalSupply.Put(totalSupply); err != nil {
		return err
	}

	axc.StateAccount.SubBalance(amount)

	return nil
}
