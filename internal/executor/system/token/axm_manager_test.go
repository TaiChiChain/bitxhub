package token

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var adminBalance *big.Int

func init() {
	axmBalance, _ := new(big.Int).SetString(repo.DefaultAXMBalance, 10)
	adminBalance = new(big.Int).Mul(axmBalance, big.NewInt(10).Exp(big.NewInt(10), big.NewInt(int64(repo.DefaultDecimals)), nil))
}
func TestInitAxmTokenManager(t *testing.T) {
	t.Parallel()
	t.Run("test init successfully", func(t *testing.T) {
		mockLg := newMockMinLedger(t)
		genesisConf := repo.DefaultGenesisConfig(false)
		conf, err := GenerateGenesisTokenConfig(genesisConf)
		require.Nil(t, err)
		err = InitAxmTokenManager(mockLg, conf)
		require.Nil(t, err)
	})

	t.Run("test transfer failed", func(t *testing.T) {
		mockLg := newMockMinLedger(t)
		genesisConf := repo.DefaultGenesisConfig(false)
		conf, err := GenerateGenesisTokenConfig(genesisConf)
		require.Nil(t, err)
		// insert totalSupply too small
		conf.TotalSupply = big.NewInt(1000000)
		err = InitAxmTokenManager(mockLg, conf)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), ErrInsufficientBalance.Error())
	})
}

func TestGetMeta(t *testing.T) {
	logger := log.NewWithModule("token")
	am := NewTokenManager(&common.SystemContractConfig{Logger: logger})
	am.account = newMockAccount(types.NewAddressByStr(common.TokenManagerContractAddr))
	require.Equal(t, "", am.Name())
	require.Equal(t, "", am.Symbol())
	require.Equal(t, uint8(0), am.Decimals())
	require.Equal(t, "0", am.TotalSupply().String())

	am = mockAxmManager(t)
	require.Equal(t, "Axiom", am.Name())
	require.Equal(t, "AXM", am.Symbol())
	require.Equal(t, uint8(18), am.Decimals())
	require.Equal(t, defaultTotalSupply, am.TotalSupply().String())
}

func TestAxmManager_BalanceOf(t *testing.T) {
	am := mockAxmManager(t)
	account1 := types.NewAddressByStr(admin1).ETHAddress()
	require.Equal(t, adminBalance.String(), am.BalanceOf(account1).String())

	s, err := types.GenerateSigner()
	require.Nil(t, err)
	randomAccount := types.NewAddressByStr(s.Addr.String()).ETHAddress()
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(randomAccount).String())
}

func TestAxmManager_Mint(t *testing.T) {
	am := mockAxmManager(t)
	contractAddr := types.NewAddressByStr(common.TokenManagerContractAddr).ETHAddress()
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(contractAddr).String())
	require.Equal(t, defaultTotalSupply, am.TotalSupply().String())

	// mint success
	addAmount := big.NewInt(1)
	require.Nil(t, am.Mint(addAmount))
	require.Equal(t, addAmount.String(), am.BalanceOf(contractAddr).String())
	oldTotalSupply, _ := new(big.Int).SetString(defaultTotalSupply, 10)
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
	contractAddr := types.NewAddressByStr(common.TokenManagerContractAddr).ETHAddress()
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(contractAddr).String())
	totalSupply, _ := new(big.Int).SetString(defaultTotalSupply, 10)
	require.Equal(t, totalSupply.String(), am.TotalSupply().String())

	err := am.Burn(big.NewInt(-1))
	require.NotNil(t, err, "burn value should be positive")
	require.Contains(t, err.Error(), ErrValue.Error())
	require.Equal(t, totalSupply.String(), am.TotalSupply().String())
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(contractAddr).String())

	err = am.Burn(big.NewInt(1))
	require.NotNil(t, err, "burn value should be less than balance")
	require.Contains(t, err.Error(), ErrInsufficientBalance.Error())
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(contractAddr).String())

	addAmount := big.NewInt(1)
	require.Nil(t, am.Mint(addAmount))
	require.Equal(t, addAmount.String(), am.BalanceOf(contractAddr).String())
	oldTotalSupply, _ := new(big.Int).SetString(defaultTotalSupply, 10)
	newTotalSupply := new(big.Int).Add(oldTotalSupply, addAmount)
	require.Equal(t, newTotalSupply.String(), am.TotalSupply().String())

	err = am.Burn(addAmount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0).String(), am.BalanceOf(contractAddr).String())
	require.Equal(t, oldTotalSupply.String(), am.TotalSupply().String())
}

func TestAxmManager_Allowance(t *testing.T) {
	am := mockAxmManager(t)
	account1 := types.NewAddressByStr(admin1).ETHAddress()
	account2 := types.NewAddressByStr(admin2).ETHAddress()

	owner := ethcommon.HexToAddress(common.TokenManagerContractAddr)
	spender := account2
	require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
	contractToAdmin2Key := getAllowancesKey(owner, spender)
	amount := new(big.Int).SetUint64(100)
	am.account.SetState([]byte(contractToAdmin2Key), amount.Bytes())
	require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
	require.Equal(t, amount.String(), am.Allowance(owner, spender).String())

}

func TestAxmManager_Approve(t *testing.T) {
	t.Parallel()

	t.Run("test approve value is negative", func(t *testing.T) {
		am := mockAxmManager(t)
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		err := am.Approve(account2, big.NewInt(-1))
		require.NotNil(t, err, "approve value should be positive")
		require.Contains(t, err.Error(), ErrValue.Error())
	})

	t.Run("test approve success", func(t *testing.T) {
		am := mockAxmManager(t)
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		err := am.Approve(account2, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(am.msgFrom, account2))

		// approve again
		err = am.Approve(account2, big.NewInt(2))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(2), am.Allowance(am.msgFrom, account2), "approve value should be set not increased")
	})
}

func TestAxmManager_Transfer(t *testing.T) {
	t.Parallel()

	t.Run("sender is nil", func(t *testing.T) {
		am := mockAxmManager(t)
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		err := am.Transfer(account2, big.NewInt(1))
		require.NotNil(t, err, "sender is nil")
		require.Contains(t, err.Error(), ErrEmptyAccount.Error())
	})

	t.Run("sender has insufficient balance", func(t *testing.T) {
		am := mockAxmManager(t)

		am.account.SetBalance(big.NewInt(0))

		am.msgFrom = types.NewAddressByStr(am.account.GetAddress().String()).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		err := am.Transfer(account2, big.NewInt(1))
		require.NotNil(t, err, "sender has insufficient balance")
		require.Contains(t, err.Error(), ErrInsufficientBalance.Error())
	})

	t.Run("transfer success", func(t *testing.T) {
		am := mockAxmManager(t)
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		fromBalance := am.BalanceOf(am.msgFrom)
		toBalance := am.BalanceOf(account2)
		transferValue := big.NewInt(1)

		err := am.Transfer(account2, big.NewInt(1))
		require.Nil(t, err)

		require.Equal(t, fromBalance.Sub(fromBalance, transferValue), am.BalanceOf(am.msgFrom))
		require.Equal(t, toBalance.Add(toBalance, transferValue), am.BalanceOf(account2))
	})
}

func TestAxmManager_TransferFrom(t *testing.T) {
	t.Parallel()

	t.Run("transfer from success", func(t *testing.T) {
		am := mockAxmManager(t)
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		am.msgFrom = account1
		err := am.Approve(account2, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(account1, account2))

		am.msgFrom = account2
		err = am.TransferFrom(account1, am.account.GetAddress().ETHAddress(), big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(0), am.Allowance(account1, account2))
	})

	t.Run("sender have not enough allowance for recipient", func(t *testing.T) {
		am := mockAxmManager(t)
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		err := am.TransferFrom(am.msgFrom, account2, big.NewInt(1))
		require.NotNil(t, err, "sender have not enough allowance for recipient")
		require.Contains(t, err.Error(), "not enough allowance")
	})

	t.Run("sender have not enough balance for recipient", func(t *testing.T) {
		am := mockAxmManager(t)
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		am.msgFrom = types.NewAddressByStr(common.TokenManagerContractAddr).ETHAddress()
		err := am.Approve(account2, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(am.msgFrom, account2))

		// reset balance to 0 to ensure sender have not enough balance for recipient
		am.account.SetBalance(big.NewInt(0))

		am.msgFrom = account2
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		err = am.TransferFrom(am.account.GetAddress().ETHAddress(), account1, big.NewInt(1))
		require.NotNil(t, err, "sender have not enough balance for recipient")
		require.Contains(t, err.Error(), ErrInsufficientBalance.Error())
	})
}
