package token

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitAxmTokenManager(t *testing.T) {
	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)

	axcContract := AXCBuildConfig.Build(common.NewTestVMContext(lg.StateLedger, ethcommon.Address{}))
	err = axcContract.GenesisInit(rep.GenesisConfig)
	assert.Nil(t, err)

	oldTotalSupply, err := axcContract.TotalSupply()
	assert.Nil(t, err)

	oldBalance := axcContract.StateAccount.GetBalance()
	assert.Equal(t, "0", oldBalance.String())

	err = axcContract.Mint(big.NewInt(-1))
	assert.ErrorContains(t, err, "amount below zero")

	err = axcContract.Mint(big.NewInt(1))
	assert.Nil(t, err)

	totalSupply1, err := axcContract.TotalSupply()
	assert.Nil(t, err)
	assert.Equal(t, "1", totalSupply1.Sub(totalSupply1, oldTotalSupply).String())

	balance1 := axcContract.StateAccount.GetBalance()
	assert.Equal(t, "1", balance1.String())

	err = axcContract.Burn(big.NewInt(-1))
	assert.ErrorContains(t, err, "amount below zero")

	err = axcContract.Burn(big.NewInt(2))
	assert.ErrorContains(t, err, "Value exceeds balance")

	err = axcContract.Burn(big.NewInt(1))
	assert.Nil(t, err)

	totalSupply2, err := axcContract.TotalSupply()
	assert.Nil(t, err)
	assert.Equal(t, "0", totalSupply2.Sub(totalSupply2, oldTotalSupply).String())

	balance2 := axcContract.StateAccount.GetBalance()
	assert.Equal(t, "0", balance2.String())
}
