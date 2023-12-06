package token

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
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
	t.Parallel()
	t.Run("test decode allowance failed, invalid allowance format", func(t *testing.T) {
		am := mockAxmManager(t)
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		wrongAllowancesValue := "wrongValue"
		am.account.SetState([]byte(AllowancesKey), []byte(wrongAllowancesValue))
		require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
	})

	t.Run("test decode allowance failed, invalid amount format", func(t *testing.T) {
		am := mockAxmManager(t)
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		wrongAmountFormat := fmt.Sprintf("%s-%s-%s", admin1, admin2, "wrongAmount")
		am.account.SetState([]byte(AllowancesKey), []byte(wrongAmountFormat))
		require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
	})

	t.Run("test get nil allowance owner", func(t *testing.T) {
		am := mockAxmManager(t)
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
		contractToAdmin2 := fmt.Sprintf("%s-%s-%s", common.TokenManagerContractAddr, admin2, "100")
		am.account.SetState([]byte(AllowancesKey), []byte(contractToAdmin2))
		require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())
	})

	t.Run("test get right allowance", func(t *testing.T) {
		am := mockAxmManager(t)
		account1 := types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()
		require.Equal(t, big.NewInt(0).String(), am.Allowance(account1, account2).String())

		allowanceValue := big.NewInt(100)
		admin1ToAdmin2 := fmt.Sprintf("%s-%s-%s", admin1, admin2, allowanceValue.String())
		am.account.SetState([]byte(AllowancesKey), []byte(admin1ToAdmin2))
		require.Equal(t, allowanceValue.String(), am.Allowance(account1, account2).String())
	})
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

	t.Run("test get old allowance err", func(t *testing.T) {
		am := mockAxmManager(t)
		am.msgFrom = types.NewAddressByStr(admin1).ETHAddress()
		account2 := types.NewAddressByStr(admin2).ETHAddress()

		// prepare allowance for admin1-admin2
		wrongAllowancesValue := "wrongValue"
		am.account.SetState([]byte(AllowancesKey), []byte(wrongAllowancesValue))

		err := am.Approve(account2, big.NewInt(1))
		require.NotNil(t, err, "old allowance format err")
		require.Contains(t, err.Error(), "invalid allowance entry format")
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

func TestReset(t *testing.T) {
	am := mockAxmManager(t)
	initLg := newMockMinLedger(t)

	err := am.Mint(big.NewInt(1))
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), am.BalanceOf(am.account.GetAddress().ETHAddress()))

	am.Reset(1, initLg)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), am.BalanceOf(am.account.GetAddress().ETHAddress()))
}

func TestEstimateGas(t *testing.T) {
	t.Parallel()

	t.Run("input data err", func(t *testing.T) {
		am := mockAxmManager(t)
		_, err := am.EstimateGas(nil)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "callArgs is nil")

		wrongData := []byte("invalid data")
		callArgs := &types.CallArgs{Data: (*hexutil.Bytes)(&wrongData)}
		_, err = am.EstimateGas(callArgs)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), vm.ErrExecutionReverted.Error())
	})

	t.Run("estimateGas success", func(t *testing.T) {
		am := mockAxmManager(t)

		method := axmManagerABI.Methods["name"]

		inputs, err := method.Inputs.Pack()
		assert.Nil(t, err)
		inputs = append(method.ID, inputs...)

		callArgs := &types.CallArgs{Data: (*hexutil.Bytes)(&inputs)}
		gas, err := am.EstimateGas(callArgs)
		require.Nil(t, err)
		require.True(t, gas > 0)
	})
}

func TestRun(t *testing.T) {
	am := mockAxmManager(t)
	initLedger := newMockMinLedger(t)

	generateRunData(t, balanceOfMethod, am.account.GetAddress().ETHAddress())

	mockLg := newMockMinLedger(t)
	genesisConf := repo.DefaultGenesisConfig(false)
	conf, err := GenerateGenesisTokenConfig(genesisConf)
	require.Nil(t, err)
	err = InitAxmTokenManager(mockLg, conf)
	require.Nil(t, err)
	initLedger.accountDb = am.stateLedger.(*mockLedger).accountDb

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
		NotReset bool
	}{
		{ // case1 : get name success
			Caller: admin1,
			Data:   generateRunData(t, nameMethod),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, nameMethod)),
				Err:        nil,
				ReturnData: wrapperReturnData(t, nameMethod, am.Name()),
			},
			Err: nil,
		},
		{ // case2: get symbol success
			Caller: admin1,
			Data:   generateRunData(t, symbolMethod),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, symbolMethod)),
				Err:        nil,
				ReturnData: wrapperReturnData(t, symbolMethod, am.Symbol()),
			},
			Err: nil,
		},
		{ // case3: get decimals success
			Caller: admin1,
			Data:   generateRunData(t, decimalsMethod),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, decimalsMethod)),
				Err:        nil,
				ReturnData: wrapperReturnData(t, decimalsMethod, am.Decimals()),
			},
			Err: nil,
		},
		{ // case4: get totalSupply success
			Caller: admin1,
			Data:   generateRunData(t, totalSupplyMethod),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, totalSupplyMethod)),
				Err:        nil,
				ReturnData: wrapperReturnData(t, totalSupplyMethod, am.TotalSupply()),
			},
			Err: nil,
		},
		{ // case5: get balanceOf success
			Caller: admin1,
			Data:   generateRunData(t, balanceOfMethod, am.account.GetAddress().ETHAddress()),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, balanceOfMethod, am.account.GetAddress().ETHAddress())),
				Err:        nil,
				ReturnData: wrapperReturnData(t, balanceOfMethod, am.BalanceOf(am.account.GetAddress().ETHAddress())),
			},
			Err: nil,
		},
		{ // case6: transfer success
			Caller: admin1,
			Data:   generateRunData(t, transferMethod, am.account.GetAddress().ETHAddress(), big.NewInt(10)),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, transferMethod, am.account.GetAddress().ETHAddress(), big.NewInt(10))),
				Err:        nil,
				ReturnData: wrapperReturnData(t, transferMethod, true),
			},
			Err:      nil,
			NotReset: true,
		},
		{ // case7: approve success
			Caller: admin1,
			Data:   generateRunData(t, approveMethod, am.account.GetAddress().ETHAddress(), big.NewInt(10)),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, approveMethod, am.account.GetAddress().ETHAddress(), big.NewInt(10))),
				Err:        nil,
				ReturnData: wrapperReturnData(t, approveMethod, true),
			},
			Err:      nil,
			NotReset: true,
		},
		{ // case8: allowance success
			Caller: admin1,
			Data:   generateRunData(t, allowanceMethod, am.account.GetAddress().ETHAddress(), am.account.GetAddress().ETHAddress()),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, allowanceMethod, am.account.GetAddress().ETHAddress(), am.account.GetAddress().ETHAddress())),
				Err:        nil,
				ReturnData: wrapperReturnData(t, allowanceMethod, am.Allowance(am.account.GetAddress().ETHAddress(), am.account.GetAddress().ETHAddress())),
			},
			Err:      nil,
			NotReset: true,
		},
		{ // case8: transferFrom success
			Caller: common.TokenManagerContractAddr,
			Data:   generateRunData(t, transferFromMethod, types.NewAddressByStr(admin1).ETHAddress(), types.NewAddressByStr(admin2).ETHAddress(), big.NewInt(10)),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, transferFromMethod, types.NewAddressByStr(admin1).ETHAddress(), types.NewAddressByStr(admin2).ETHAddress(), big.NewInt(10))),
				Err:        nil,
				ReturnData: wrapperReturnData(t, transferFromMethod, true),
			},
			Err: nil,
		},
		{ // case9: mint success
			Caller: common.TokenManagerContractAddr,
			Data:   generateRunData(t, mintMethod, big.NewInt(10)),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, mintMethod, big.NewInt(10))),
				Err:        nil,
				ReturnData: wrapperReturnData(t, mintMethod, true),
			},
			Err: nil,
		},
		{ // case10: burn success
			Caller: common.TokenManagerContractAddr,
			Data:   generateRunData(t, burnMethod, big.NewInt(10)),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, burnMethod, big.NewInt(10))),
				Err:        nil,
				ReturnData: wrapperReturnData(t, burnMethod),
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		result, err := am.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		require.Equal(t, test.Err, err)

		if result != nil {
			require.Equal(t, nil, result.Err)
			require.Equal(t, test.Expected.UsedGas, result.UsedGas)
			require.Equal(t, test.Expected.ReturnData, result.ReturnData)
		}

		if !test.NotReset {
			am.Reset(1, initLedger)
		}
	}
}
