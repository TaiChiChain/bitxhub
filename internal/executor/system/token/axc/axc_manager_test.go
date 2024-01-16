package axc

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitAxmTokenManager(t *testing.T) {
	mockLg := newMockMinLedger(t)
	genesisConf := repo.DefaultGenesisConfig(false)
	conf, err := GenerateConfig(genesisConf)
	require.Nil(t, err)
	err = Init(mockLg, conf)
	require.Nil(t, err)
}

func TestGetMeta(t *testing.T) {
	logger := log.NewWithModule("token")
	am := New(&common.SystemContractConfig{Logger: logger})
	am.account = newMockAccount(types.NewAddressByStr(common.AXCContractAddr))
	require.Equal(t, "", am.Name())
	require.Equal(t, "", am.Symbol())
	require.Equal(t, uint8(0), am.Decimals())
	require.Equal(t, "0", am.TotalSupply().String())

	am = mockAxcManager(t)
	require.Equal(t, "Axiomesh Credit", am.Name())
	require.Equal(t, "axc", am.Symbol())
	require.Equal(t, uint8(18), am.Decimals())
	require.Equal(t, defaultTotalSupply, am.TotalSupply().String())
}

func TestAxmManager_BalanceOf(t *testing.T) {
	am := mockAxcManager(t)
	genesisConf := repo.DefaultGenesisConfig(false)
	totalSupply, _ := new(big.Int).SetString(genesisConf.Axc.TotalSupply, 10)
	for _, distribution := range genesisConf.Incentive.Distributions {
		percentage := big.NewInt(int64(distribution.Percentage * Decimals))
		expectedBalance := new(big.Int).Div(new(big.Int).Mul(totalSupply, percentage), big.NewInt(Decimals))
		_, actualBalanceBytes := am.account.GetState([]byte(getBalancesKey(types.NewAddressByStr(distribution.Addr).ETHAddress())))
		actualBalanceUnlock := new(big.Int).SetBytes(actualBalanceBytes)
		_, actualBalanceLockBytes := am.account.GetState([]byte(getLockedBalanceKey(types.NewAddressByStr(distribution.Addr).ETHAddress())))
		actualBalanceLock := new(big.Int).SetBytes(actualBalanceLockBytes)
		if !distribution.Locked {
			require.Equal(t, expectedBalance, actualBalanceUnlock)
		} else {
			require.Equal(t, expectedBalance, actualBalanceLock)
		}
	}
}

func TestAxmManager_Allowance(t *testing.T) {
	am := mockAxcManager(t)
	account1 := types.NewAddressByStr(admin1).ETHAddress()

	owner := types.NewAddressByStr(testMiningAddr).ETHAddress()
	spender := account1
	require.Equal(t, big.NewInt(0).String(), am.Allowance(owner, spender).String())
	allowanceKey := getAllowancesKey(owner, spender)
	amount := new(big.Int).SetUint64(100)
	am.account.SetState([]byte(allowanceKey), amount.Bytes())
	require.Equal(t, amount.String(), am.Allowance(owner, spender).String())
}

func TestAxmManager_Approve(t *testing.T) {
	t.Parallel()
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testMiningAddr).ETHAddress()

	t.Run("test approve value is negative", func(t *testing.T) {
		am.msgFrom = owner
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		err := am.Approve(account1, big.NewInt(-1))
		require.EqualError(t, err, ErrValue.Error())
		require.Contains(t, err.Error(), ErrValue.Error())
	})

	t.Run("test approve success", func(t *testing.T) {
		am.msgFrom = owner
		account1 := types.NewAddressByStr(admin1).ETHAddress()

		err := am.Approve(account1, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(am.msgFrom, account1))

		// approve increase
		err = am.Approve(account1, big.NewInt(2))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(2), am.Allowance(am.msgFrom, account1))

		// approve decrease
		err = am.Approve(account1, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(am.msgFrom, account1))
	})
}

func TestAxmManager_Transfer(t *testing.T) {
	t.Parallel()
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testMiningAddr).ETHAddress()

	t.Run("sender is nil", func(t *testing.T) {
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		am.msgFrom = ethcommon.Address{}
		err := am.Transfer(account1, big.NewInt(1))
		require.EqualError(t, err, ErrEmptyAccount.Error())
	})

	t.Run("sender has insufficient balance", func(t *testing.T) {
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()

		err := am.Transfer(owner, big.NewInt(1))
		require.EqualError(t, err, ErrInsufficientBalance.Error())
	})

	t.Run("transfer success", func(t *testing.T) {
		am.msgFrom = owner
		account1 := types.NewAddressByStr(admin1).ETHAddress()

		fromBalance := am.BalanceOf(am.msgFrom)
		toBalance := am.BalanceOf(account1)
		transferValue := big.NewInt(1)

		err := am.Transfer(account1, big.NewInt(1))
		require.Nil(t, err)

		require.Equal(t, fromBalance.Sub(fromBalance, transferValue), am.BalanceOf(am.msgFrom))
		require.Equal(t, toBalance.Add(toBalance, transferValue), am.BalanceOf(account1))
	})
}

func TestAxmManager_TransferFrom(t *testing.T) {
	t.Parallel()
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testMiningAddr).ETHAddress()

	t.Run("transfer from success", func(t *testing.T) {
		account1 := types.NewAddressByStr(admin1).ETHAddress()

		am.msgFrom = owner
		err := am.Approve(account1, big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(1), am.Allowance(owner, account1))

		am.msgFrom = account1
		err = am.TransferFrom(owner, am.account.GetAddress().ETHAddress(), big.NewInt(1))
		require.Nil(t, err)
		require.Equal(t, big.NewInt(0), am.Allowance(owner, account1))
	})

	t.Run("sender have not enough allowance for recipient", func(t *testing.T) {
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		err := am.TransferFrom(owner, am.account.GetAddress().ETHAddress(), big.NewInt(1))
		require.EqualError(t, err, ErrNotEnoughAllowance.Error())
	})

	t.Run("sender have not enough balance for recipient", func(t *testing.T) {
		account1 := types.NewAddressByStr(admin1).ETHAddress()

		am.msgFrom = owner
		amount := new(big.Int).Add(am.BalanceOf(owner), big.NewInt(1))
		err := am.Approve(account1, amount)
		require.Nil(t, err)
		require.Equal(t, amount, am.Allowance(am.msgFrom, account1))

		am.msgFrom = account1
		err = am.TransferFrom(owner, am.account.GetAddress().ETHAddress(), amount)
		require.EqualError(t, err, ErrInsufficientBalance.Error())
	})
}

func TestAxmManager_TransferLocked(t *testing.T) {
	t.Parallel()
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testCommunityAddr).ETHAddress()

	t.Run("sender is nil", func(t *testing.T) {
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		am.msgFrom = ethcommon.Address{}
		err := am.TransferLocked(account1, big.NewInt(1))
		require.EqualError(t, err, ErrEmptyAccount.Error())
	})

	t.Run("sender has insufficient balance", func(t *testing.T) {
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()

		err := am.TransferLocked(owner, big.NewInt(1))
		require.EqualError(t, err, ErrInsufficientBalance.Error())
	})

	t.Run("transferLocked success", func(t *testing.T) {
		am.msgFrom = owner
		account1 := types.NewAddressByStr(admin1).ETHAddress()

		fromBeforeBalance := am.BalanceOfLocked(owner)
		require.True(t, fromBeforeBalance.Cmp(big.NewInt(0)) == 1)
		toBeforeBalance := am.BalanceOfLocked(account1)
		require.True(t, toBeforeBalance.Cmp(big.NewInt(0)) == 0)
		transferValue := big.NewInt(1)

		err := am.TransferLocked(account1, transferValue)
		require.Nil(t, err)

		fromAfterBalance := am.BalanceOfLocked(owner)
		toAfterBalance := am.BalanceOfLocked(account1)
		require.Equal(t, transferValue, new(big.Int).Sub(fromBeforeBalance, fromAfterBalance))
		require.Equal(t, transferValue, new(big.Int).Sub(toAfterBalance, toBeforeBalance))
	})
}

func TestManager_Unlock(t *testing.T) {
	t.Parallel()
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testCommunityAddr).ETHAddress()
	err := am.Unlock(owner, nil)
	require.EqualError(t, ErrNotImplemented, err.Error())
}

func TestLogs(t *testing.T) {
	am := mockAxcManager(t)
	owner := types.NewAddressByStr(testMiningAddr).ETHAddress()
	am.msgFrom = owner
	am.logs = make([]common.Log, 0)
	account1 := types.NewAddressByStr(admin1).ETHAddress()
	contract := types.NewAddressByStr(common.AXCContractAddr).ETHAddress()

	err := am.Approve(account1, big.NewInt(1))
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), am.Allowance(am.msgFrom, account1))

	expectedApproveSig := "0x000000000000000000000000000000000000000000000000000000008c5be1e5"
	require.Equal(t, expectedApproveSig, am.logs[0].Topics[0].String())

	am.msgFrom = account1
	err = am.TransferFrom(owner, contract, big.NewInt(1))
	require.Nil(t, err)

	expectedTransferSig := "0x00000000000000000000000000000000000000000000000000000000ddf252ad"
	require.Equal(t, expectedTransferSig, am.logs[1].Topics[0].String())
}
