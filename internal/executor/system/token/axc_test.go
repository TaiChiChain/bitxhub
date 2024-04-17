package token

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitAxmTokenManager(t *testing.T) {
	t.Parallel()
	t.Run("test init successfully", func(t *testing.T) {
		mockLg := newMockMinLedger(t)
		genesisConf := repo.DefaultGenesisConfig()
		conf, err := GenerateConfig(genesisConf)
		require.Nil(t, err)
		err = Init(mockLg, conf)
		require.Nil(t, err)
	})
}

func TestGetMeta(t *testing.T) {
	logger := log.NewWithModule("token")
	am := New(&common.SystemContractConfig{Logger: logger})
	am.account = newMockAccount(types.NewAddressByStr(common.AXCContractAddr))
	require.Equal(t, "0", am.TotalSupply().String())

	am = mockAxmManager(t)
	require.Equal(t, repo.DefaultAXCBalance, am.TotalSupply().String())
}

func TestAxmManager_Mint(t *testing.T) {
	am := mockAxmManager(t)
	contractAddr := types.NewAddressByStr(common.AXCContractAddr)
	require.Equal(t, big.NewInt(0).String(), am.stateLedger.GetBalance(contractAddr).String())
	require.Equal(t, repo.DefaultAXCBalance, am.TotalSupply().String())

	// mint success
	addAmount := big.NewInt(1)
	require.Nil(t, am.Mint(addAmount))
	require.Equal(t, addAmount.String(), am.stateLedger.GetBalance(contractAddr).String())
	oldTotalSupply, _ := new(big.Int).SetString(repo.DefaultAXCBalance, 10)
	newTotalSupply := new(big.Int).Add(oldTotalSupply, addAmount)
	require.Equal(t, newTotalSupply.String(), am.TotalSupply().String())

	// wrong mint value
	err := am.Mint(big.NewInt(-1))
	require.NotNil(t, err, "mint value should be positive")
	require.Contains(t, err.Error(), ErrValue.Error())
	require.Equal(t, newTotalSupply.String(), am.TotalSupply().String())

	// wrong contract account
	am.account = newMockAccount(types.NewAddressByStr(admin1))
	err = am.Mint(big.NewInt(1))
	require.NotNil(t, err, "only token contract account can mint")
	require.Contains(t, err.Error(), ErrContractAccount.Error())
}

func TestAxmManager_Burn(t *testing.T) {
	am := mockAxmManager(t)
	contractAddr := types.NewAddressByStr(common.AXCContractAddr)
	require.Equal(t, big.NewInt(0).String(), am.stateLedger.GetBalance(contractAddr).String())
	totalSupply, _ := new(big.Int).SetString(repo.DefaultAXCBalance, 10)
	require.Equal(t, totalSupply.String(), am.TotalSupply().String())

	err := am.Burn(big.NewInt(-1))
	require.NotNil(t, err, "burn value should be positive")
	require.Contains(t, err.Error(), ErrValue.Error())
	require.Equal(t, totalSupply.String(), am.TotalSupply().String())
	require.Equal(t, big.NewInt(0).String(), am.stateLedger.GetBalance(contractAddr).String())

	err = am.Burn(big.NewInt(1))
	require.NotNil(t, err, "burn value should be less than balance")
	require.Contains(t, err.Error(), ErrInsufficientBalance.Error())
	require.Equal(t, big.NewInt(0).String(), am.stateLedger.GetBalance(contractAddr).String())

	addAmount := big.NewInt(1)
	require.Nil(t, am.Mint(addAmount))
	require.Equal(t, addAmount.String(), am.stateLedger.GetBalance(contractAddr).String())
	oldTotalSupply, _ := new(big.Int).SetString(repo.DefaultAXCBalance, 10)
	newTotalSupply := new(big.Int).Add(oldTotalSupply, addAmount)
	require.Equal(t, newTotalSupply.String(), am.TotalSupply().String())

	err = am.Burn(addAmount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0).String(), am.stateLedger.GetBalance(contractAddr).String())
	require.Equal(t, oldTotalSupply.String(), am.TotalSupply().String())
}
