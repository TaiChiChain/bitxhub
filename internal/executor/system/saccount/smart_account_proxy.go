package saccount

import (
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_proxy"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/saccount/solidity/smart_account_proxy_client"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

var AccountProxyBuildConfig = &common.SystemContractBuildConfig[*SmartAccountProxy]{
	Name:    "saccount_account_proxy",
	Address: common.AccountProxyContractAddr,
	AbiStr:  smart_account_proxy_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *SmartAccountProxy {
		return &SmartAccountProxy{}
	},
}

var _ common.SystemContract = (*SmartAccountProxy)(nil)
var _ smart_account_proxy.SmartAccountProxy = (*SmartAccountProxy)(nil)

type SmartAccountProxy struct {
	common.SystemContractBase
}

// GenesisInit implements common.SystemContract.
func (proxy *SmartAccountProxy) GenesisInit(genesis *repo.GenesisConfig) error {
	return nil
}

// SetContext implements common.SystemContract.
func (proxy *SmartAccountProxy) SetContext(ctx *common.VMContext) {
	proxy.SystemContractBase.SetContext(ctx)
}

// GetGuardian implements smart_account_proxy.SmartAccountProxy.
func (proxy *SmartAccountProxy) GetGuardian(account ethcommon.Address) (ethcommon.Address, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(proxy.CrossCallSystemContractContext(), account)

	return sa.GetGuardian()
}

// GetOwner implements smart_account_proxy.SmartAccountProxy.
func (proxy *SmartAccountProxy) GetOwner(account ethcommon.Address) (ethcommon.Address, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(proxy.CrossCallSystemContractContext(), account)

	return sa.GetOwner()
}

// GetPasskeys implements smart_account_proxy.SmartAccountProxy.
func (proxy *SmartAccountProxy) GetPasskeys(account ethcommon.Address) ([]smart_account_proxy.PassKey, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(proxy.CrossCallSystemContractContext(), account)

	return sa.GetPasskeys()
}

// GetSessions implements smart_account_proxy.SmartAccountProxy.
func (proxy *SmartAccountProxy) GetSessions(account ethcommon.Address) ([]smart_account_proxy.SessionKey, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(proxy.CrossCallSystemContractContext(), account)

	return sa.GetSessions()
}

// GetStatus implements smart_account_proxy.SmartAccountProxy.
func (proxy *SmartAccountProxy) GetStatus(account ethcommon.Address) (uint64, error) {
	sa := SmartAccountBuildConfig.BuildWithAddress(proxy.CrossCallSystemContractContext(), account)

	return sa.GetStatus()
}
