package access

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access/solidity/whitelist"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

var (
	admin0 = ethcommon.HexToAddress("0x1200000000000000000000000000000000000000")
	admin1 = ethcommon.HexToAddress("0x1210000000000000000000000000000000000000")
	admin2 = ethcommon.HexToAddress("0x1220000000000000000000000000000000000000")
	admin3 = ethcommon.HexToAddress("0x1230000000000000000000000000000000000000")
	admin4 = ethcommon.HexToAddress("0x1240000000000000000000000000000000000000")
)

func TestWhitelist_AuthInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.WhitelistProviders = []string{admin0.String(), admin1.String(), admin2.String()}
	whitelistContract := WhitelistBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(whitelistContract)

	testNVM.Call(whitelistContract, admin1, func() {
		admin1AuthInfo, err := whitelistContract.QueryAuthInfo(admin1)
		assert.Nil(t, err)
		assert.Equal(t, admin1, admin1AuthInfo.Addr)
		assert.Equal(t, AuthUserRoleSuper, admin1AuthInfo.Role)
		assert.Empty(t, admin1AuthInfo.Providers)
	})

	testNVM.Call(whitelistContract, ethcommon.Address{}, func() {
		_, err := whitelistContract.QueryAuthInfo(admin1)
		assert.ErrorContains(t, err, "permission denied")
	})

	testNVM.Call(whitelistContract, admin1, func() {
		zeroInfo, err := whitelistContract.QueryAuthInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, whitelist.AuthInfo{}, zeroInfo)
	})

	testNVM.RunSingleTX(whitelistContract, ethcommon.Address{}, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin3})
		assert.ErrorContains(t, err, ErrProviderPermission.Error())
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin4, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin3})
		assert.ErrorContains(t, err, ErrProviderPermission.Error())
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin3})
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin2, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin3})
		assert.Nil(t, err)
		return err
	})

	// resubmit
	testNVM.RunSingleTX(whitelistContract, admin2, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin3})
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(whitelistContract, admin1, func() {
		admin3AuthInfo, err := whitelistContract.QueryAuthInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, admin3, admin3AuthInfo.Addr)
		assert.Equal(t, AuthUserRoleBasic, admin3AuthInfo.Role)
		assert.Equal(t, 2, len(admin3AuthInfo.Providers))
		assert.Equal(t, admin1, admin3AuthInfo.Providers[0])
	})

	testNVM.Call(whitelistContract, admin3, func() {
		err := whitelistContract.Verify(admin3)
		assert.Nil(t, err)
	})

	testNVM.Call(whitelistContract, admin3, func() {
		_, err := whitelistContract.QueryAuthInfo(admin1)
		assert.ErrorContains(t, err, "permission denied")
	})

	testNVM.RunSingleTX(whitelistContract, ethcommon.Address{}, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin4})
		assert.ErrorContains(t, err, ErrProviderPermission.Error())
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin3, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin4})
		assert.ErrorContains(t, err, ErrProviderPermission.Error())
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin4})
		assert.ErrorContains(t, err, "not exist")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin2})
		assert.ErrorContains(t, err, "try to modify super user")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin0, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin3})
		assert.ErrorContains(t, err, "no permission")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin3})
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(whitelistContract, admin1, func() {
		admin3AuthInfo, err := whitelistContract.QueryAuthInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, admin3, admin3AuthInfo.Addr)
		assert.Equal(t, AuthUserRoleBasic, admin3AuthInfo.Role)
		assert.Equal(t, 1, len(admin3AuthInfo.Providers))
		assert.Equal(t, admin2, admin3AuthInfo.Providers[0])
	})

	testNVM.Call(whitelistContract, admin3, func() {
		err := whitelistContract.Verify(admin3)
		assert.Nil(t, err)
	})

	testNVM.Call(whitelistContract, admin3, func() {
		_, err := whitelistContract.QueryProviderInfo(admin1)
		assert.ErrorContains(t, err, "permission denied")
	})

	testNVM.RunSingleTX(whitelistContract, admin2, func() error {
		err := whitelistContract.Remove([]ethcommon.Address{admin3})
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(whitelistContract, admin2, func() {
		zeroInfo, err := whitelistContract.QueryAuthInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, whitelist.AuthInfo{}, zeroInfo)
	})

	testNVM.Call(whitelistContract, admin3, func() {
		err := whitelistContract.Verify(admin3)
		assert.ErrorContains(t, err, "permission denied")
	})
}

func TestWhitelist_ProviderInfo(t *testing.T) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.WhitelistProviders = []string{admin0.String(), admin1.String(), admin2.String()}
	whitelistContract := WhitelistBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(whitelistContract)

	testNVM.Call(whitelistContract, admin1, func() {
		zeroInfo, err := whitelistContract.QueryProviderInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, whitelist.ProviderInfo{}, zeroInfo)
	})

	testNVM.RunSingleTX(whitelistContract, ethcommon.Address{}, func() error {
		_, err := whitelistContract.QueryProviderInfo(admin3)
		assert.ErrorContains(t, err, "permission denied")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin4, func() error {
		_, err := whitelistContract.QueryProviderInfo(admin3)
		assert.ErrorContains(t, err, "permission denied")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		assert.True(t, whitelistContract.ExistProvider(admin1))
		assert.False(t, whitelistContract.ExistProvider(admin3))
		return nil
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		admin1ProviderInfo, err := whitelistContract.QueryProviderInfo(admin1)
		assert.Nil(t, err)
		assert.Equal(t, admin1, admin1ProviderInfo.Addr)
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.UpdateProviders(false, []whitelist.ProviderInfo{{Addr: admin4}})
		assert.ErrorContains(t, err, "not exist")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.UpdateProviders(true, []whitelist.ProviderInfo{{Addr: admin1}})
		assert.ErrorContains(t, err, "already exist")
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.UpdateProviders(true, []whitelist.ProviderInfo{{Addr: admin3}})
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin3, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{admin4})
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		err := whitelistContract.UpdateProviders(false, []whitelist.ProviderInfo{{Addr: admin3}})
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin3, func() error {
		err := whitelistContract.Submit([]ethcommon.Address{ethcommon.HexToAddress("0x1240000000000000000000000000000000000001")})
		assert.ErrorContains(t, err, ErrProviderPermission.Error())
		return err
	})

	testNVM.RunSingleTX(whitelistContract, admin1, func() error {
		zeroInfo, err := whitelistContract.QueryProviderInfo(admin3)
		assert.Nil(t, err)
		assert.Equal(t, whitelist.ProviderInfo{}, zeroInfo)
		return err
	})
}
