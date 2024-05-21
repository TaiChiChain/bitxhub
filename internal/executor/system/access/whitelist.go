package access

import (
	"bytes"
	"sort"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access/solidity/whitelist"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access/solidity/whitelist_client"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

const (
	whitelistAuthInfosStorageKey     = "authInfos"
	whitelistProviderInfosStorageKey = "providerInfos"
)

var (
	ErrProviderPermission = errors.New("permission denied, you are not a provider")
)

var WhitelistBuildConfig = &common.SystemContractBuildConfig[*Whitelist]{
	Name:    "access_whitelist",
	Address: common.WhiteListContractAddr,
	AbiStr:  whitelist_client.BindingContractMetaData.ABI,
	Constructor: func(systemContractBase common.SystemContractBase) *Whitelist {
		return &Whitelist{
			SystemContractBase: systemContractBase,
		}
	},
}

var _ whitelist.Whitelist = (*Whitelist)(nil)

const (
	AuthUserRoleBasic uint8 = iota
	AuthUserRoleSuper
)

type Whitelist struct {
	common.SystemContractBase

	authInfos     *common.VMMap[ethcommon.Address, whitelist.AuthInfo]
	providerInfos *common.VMMap[ethcommon.Address, whitelist.ProviderInfo]
}

func (w *Whitelist) GenesisInit(genesis *repo.GenesisConfig) error {
	admins := lo.Map[*repo.CouncilMember, string](genesis.CouncilMembers, func(x *repo.CouncilMember, _ int) string {
		return x.Address
	})
	accounts := lo.Map(genesis.Accounts, func(x *repo.Account, _ int) string {
		return x.Address
	})
	initVerifiedUsers := lo.Union(admins, genesis.WhitelistProviders, accounts)
	sort.Strings(initVerifiedUsers)
	sort.Strings(genesis.WhitelistProviders)
	// init super user
	for _, addrStr := range initVerifiedUsers {
		addr := ethcommon.HexToAddress(addrStr)
		if err := w.authInfos.Put(addr, whitelist.AuthInfo{
			Addr:      addr,
			Providers: []ethcommon.Address{},
			Role:      AuthUserRoleSuper,
		}); err != nil {
			return err
		}
	}
	for _, addrStr := range genesis.WhitelistProviders {
		addr := ethcommon.HexToAddress(addrStr)
		if err := w.providerInfos.Put(addr, whitelist.ProviderInfo{
			Addr: addr,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (w *Whitelist) SetContext(ctx *common.VMContext) {
	w.SystemContractBase.SetContext(ctx)

	w.authInfos = common.NewVMMap[ethcommon.Address, whitelist.AuthInfo](w.StateAccount, whitelistAuthInfosStorageKey, func(key ethcommon.Address) string {
		return key.String()
	})
	w.providerInfos = common.NewVMMap[ethcommon.Address, whitelist.ProviderInfo](w.StateAccount, whitelistProviderInfosStorageKey, func(key ethcommon.Address) string {
		return key.String()
	})
}

func (w *Whitelist) Submit(addresses []ethcommon.Address) error {
	if err := w.checkProviderPermission(); err != nil {
		return err
	}
	for _, addr := range addresses {
		exist, authInfo, err := w.authInfos.Get(addr)
		if err != nil {
			return err
		}
		if exist {
			if authInfo.Role == AuthUserRoleSuper {
				return errors.New("access error: submit: try to modify super user")
			}
			if lo.ContainsBy(authInfo.Providers, func(item ethcommon.Address) bool {
				return bytes.Equal(item.Bytes(), w.Ctx.From.Bytes())
			}) {
				continue
			}
		} else {
			authInfo = whitelist.AuthInfo{
				Addr:      addr,
				Providers: []ethcommon.Address{},
				Role:      AuthUserRoleBasic,
			}
		}
		authInfo.Providers = append(authInfo.Providers, w.Ctx.From)
		if err := w.authInfos.Put(addr, authInfo); err != nil {
			return err
		}
	}

	w.EmitEvent(&whitelist.EventSubmit{
		Submitter: w.Ctx.From,
		Addresses: addresses,
	})
	return nil
}

func (w *Whitelist) Remove(addresses []ethcommon.Address) error {
	if err := w.checkProviderPermission(); err != nil {
		return err
	}

	for _, addr := range addresses {
		exist, authInfo, err := w.authInfos.Get(addr)
		if err != nil {
			return err
		}
		if exist {
			if authInfo.Role == AuthUserRoleSuper {
				return errors.Errorf("access error: remove: try to modify super user [%s]", addr)
			}
			if !lo.ContainsBy(authInfo.Providers, func(item ethcommon.Address) bool {
				return bytes.Equal(item.Bytes(), w.Ctx.From.Bytes())
			}) {
				return errors.Errorf("access error: remove: no permission")
			}
		} else {
			return errors.Errorf("access error: remove: try to remove [%s] auth info that does not exist", addr)
		}

		authInfo.Providers = lo.Reject(authInfo.Providers, func(addr ethcommon.Address, _ int) bool {
			return bytes.Equal(addr.Bytes(), w.Ctx.From.Bytes())
		})
		if len(authInfo.Providers) == 0 {
			if err := w.authInfos.Delete(addr); err != nil {
				return err
			}
		} else {
			if err := w.authInfos.Put(addr, authInfo); err != nil {
				return err
			}
		}
	}

	w.EmitEvent(&whitelist.EventRemove{
		Submitter: w.Ctx.From,
		Addresses: addresses,
	})
	return nil
}

func (w *Whitelist) QueryAuthInfo(addr ethcommon.Address) (whitelist.AuthInfo, error) {
	if !w.Ctx.CallFromSystem {
		exist, fromAuthInfo, err := w.authInfos.Get(w.Ctx.From)
		if err != nil {
			return whitelist.AuthInfo{}, err
		}
		if !exist || fromAuthInfo.Role != AuthUserRoleSuper {
			return whitelist.AuthInfo{}, errors.New("permission denied, you are not a super user")
		}
	}

	_, info, err := w.authInfos.Get(addr)
	if err != nil {
		return whitelist.AuthInfo{}, err
	}
	return info, nil
}

func (w *Whitelist) QueryProviderInfo(addr ethcommon.Address) (whitelist.ProviderInfo, error) {
	if !w.Ctx.CallFromSystem {
		exist, fromAuthInfo, err := w.authInfos.Get(w.Ctx.From)
		if err != nil {
			return whitelist.ProviderInfo{}, err
		}
		if !exist || fromAuthInfo.Role != AuthUserRoleSuper {
			return whitelist.ProviderInfo{}, errors.New("permission denied, you are not a super user")
		}
	}

	_, info, err := w.providerInfos.Get(addr)
	if err != nil {
		return whitelist.ProviderInfo{}, err
	}
	return info, nil
}

func (w *Whitelist) ExistProvider(addr ethcommon.Address) bool {
	return w.providerInfos.Has(addr)
}

func (w *Whitelist) Verify(user ethcommon.Address) error {
	exist := w.authInfos.Has(user)
	if !exist {
		return errors.New("permission denied, you are not in whitelist")
	}

	return nil
}

func (w *Whitelist) UpdateProviders(isAdd bool, providers []whitelist.ProviderInfo) error {
	for _, provider := range providers {
		if isAdd {
			if w.providerInfos.Has(provider.Addr) {
				return errors.Errorf("provider %s already exist", provider.Addr.String())
			}
			if err := w.providerInfos.Put(provider.Addr, provider); err != nil {
				return err
			}
		} else {
			if !w.providerInfos.Has(provider.Addr) {
				return errors.Errorf("provider %s not exist", provider.Addr.String())
			}
			if err := w.providerInfos.Delete(provider.Addr); err != nil {
				return err
			}
			// TODO: how to process users who have this provider?
		}
	}

	w.EmitEvent(&whitelist.EventUpdateProviders{
		IsAdd:     isAdd,
		Providers: providers,
	})
	return nil
}

func (w *Whitelist) checkProviderPermission() error {
	if !w.providerInfos.Has(w.Ctx.From) {
		return ErrProviderPermission
	}
	return nil
}
