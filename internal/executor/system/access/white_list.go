package access

import (
	"encoding/json"
	"errors"
	"sort"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
)

const (
	AuthInfoKey          = "authinfo"
	WhiteListProviderKey = "providers"
)

var (
	ErrCheckSubmitInfo        = errors.New("submit args check fail")
	ErrCheckRemoveInfo        = errors.New("remove args check fail")
	ErrUser                   = errors.New("user is invalid")
	ErrCheckWhiteListProvider = errors.New("white list provider check fail")
	ErrParseArgs              = errors.New("parse args fail")
	ErrGetMethodName          = errors.New("get method name fail")
	ErrVerify                 = errors.New("access error")
	ErrQueryPermission        = errors.New("insufficient query permissions")
	ErrNotFound               = errors.New("not found")
)

const (
	SubmitMethod                 = "submit"
	RemoveMethod                 = "remove"
	QueryAuthInfoMethod          = "queryAuthInfo"
	QueryWhiteListProviderMethod = "queryWhiteListProvider"
)

type Role uint8

type ModifyType uint8

const (
	AddWhiteListProvider    ModifyType = 4
	RemoveWhiteListProvider ModifyType = 5
)

const (
	BasicUser Role = iota
	SuperUser
)

type AuthInfo struct {
	User      string
	Providers []string
	Role      Role
}

// Global providers
// TODO: refactor
var Providers []string

type BaseExtraArgs struct {
	Extra []byte
}

type SubmitArgs struct {
	Addresses []string
}

type RemoveArgs struct {
	Addresses []string
}

type QueryAuthInfoArgs struct {
	User string
}

type QueryWhiteListProviderArgs struct {
	WhiteListProviderAddr string
}

type WhiteListProvider struct {
	WhiteListProviderAddr string
}

type WhiteListProviderArgs struct {
	Providers []WhiteListProvider
}

type WhiteList struct {
	stateLedger ledger.StateLedger
	account     ledger.IAccount
	currentLogs *[]common.Log
	currentUser *ethcommon.Address
	logger      logrus.FieldLogger
}

// NewWhiteList constructs a new WhiteList
func NewWhiteList(cfg *common.SystemContractConfig) *WhiteList {
	return &WhiteList{
		logger: loggers.Logger(loggers.Access),
	}
}

func (c *WhiteList) SetContext(context *common.VMContext) {
	addr := types.NewAddressByStr(common.WhiteListContractAddr)
	c.account = context.StateLedger.GetOrCreateAccount(addr)
	c.stateLedger = context.StateLedger
	c.currentLogs = context.CurrentLogs
	c.currentUser = context.CurrentUser

	// TODO: lazy load
	state, b := c.account.GetState([]byte(WhiteListProviderKey))
	if state {
		var services []WhiteListProvider
		if err := json.Unmarshal(b, &services); err != nil {
			c.logger.Debugf("system contract reset error: unmarshal services fail!")
			services = []WhiteListProvider{}
		}
		Providers = []string{}
		for _, item := range services {
			Providers = append(Providers, item.WhiteListProviderAddr)
		}
	}
}

func (c *WhiteList) Submit(data []byte) error {
	from := c.currentUser
	if from == nil {
		return ErrUser
	}

	var args SubmitArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return err
	}

	c.logger.Debugf("begin submit: msg sender is %s, args is %v", from.String(), args.Addresses)
	if err := CheckInServices(c.account, from.String()); err != nil {
		c.logger.Debugf("access error: submit: fail by checking providers")
		return err
	}
	for _, address := range args.Addresses {
		// check user addr
		if addr := types.NewAddressByStr(address); addr.ETHAddress().String() != address {
			c.logger.Debugf("access error: info user addr is invalid")
			return ErrCheckSubmitInfo
		}
		authInfo := c.getAuthInfo(address)
		if authInfo != nil {
			// check super role
			if authInfo.Role == SuperUser {
				c.logger.Debugf("access error: submit: try to modify super user")
				return ErrCheckSubmitInfo
			}
			// check exist info
			if common.IsInSlice(from.String(), authInfo.Providers) {
				c.logger.Debugf("access error: auth information already exist")
				return ErrCheckSubmitInfo
			}
		} else {
			authInfo = &AuthInfo{
				User:      address,
				Providers: []string{},
				Role:      BasicUser,
			}
		}
		// fill auth info with providers
		authInfo.Providers = append(authInfo.Providers, from.String())
		if err := c.saveAuthInfo(authInfo); err != nil {
			return err
		}
	}

	b, err := json.Marshal(args)
	if err != nil {
		return err
	}
	// record log
	c.RecordLog(SubmitMethod, b)
	return nil
}

func (c *WhiteList) Remove(data []byte) error {
	from := c.currentUser
	if from == nil {
		return ErrUser
	}
	var args RemoveArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return err
	}

	c.logger.Debugf("begin remove: msg sender is %s, args is %v", from.String(), args.Addresses)
	if err := CheckInServices(c.account, from.String()); err != nil {
		return err
	}
	for _, addr := range args.Addresses {
		// check admin
		info := c.getAuthInfo(addr)
		if info != nil {
			if info.Role == SuperUser {
				c.logger.Debugf("access error: remove: try to modify super user [%s]", addr)
				return ErrCheckRemoveInfo
			}
			if !common.IsInSlice(from.String(), info.Providers) {
				c.logger.Debugf("access error: remove: try to remove auth info that does not exist")
				return ErrCheckRemoveInfo
			}
		} else {
			c.logger.Debugf("access error: remove: try to remove [%s] auth info that does not exist", addr)
			return ErrCheckRemoveInfo
		}
		info.Providers = common.RemoveFirstMatchStrInSlice(info.Providers, from.String())
		if err := c.saveAuthInfo(info); err != nil {
			return err
		}
	}
	b, err := json.Marshal(args)
	if err != nil {
		return err
	}
	// record log
	c.RecordLog(RemoveMethod, b)
	return nil
}

func (c *WhiteList) QueryAuthInfo(data []byte) ([]byte, error) {
	addr := c.currentUser
	if addr == nil {
		return nil, ErrUser
	}
	var args QueryAuthInfoArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, err
	}

	userAddr := types.NewAddressByStr(args.User).ETHAddress().String()
	if userAddr != args.User {
		return nil, errors.New("user address is invalid")
	}

	authInfo := c.getAuthInfo(addr.String())
	if authInfo == nil || authInfo.Role != SuperUser {
		return nil, ErrQueryPermission
	}

	state, b := c.account.GetState([]byte(AuthInfoKey + args.User))
	if state {
		return b, nil
	}

	return nil, ErrNotFound
}

func (c *WhiteList) QueryWhiteListProvider(data []byte) ([]byte, error) {
	addr := c.currentUser
	if addr == nil {
		return nil, ErrUser
	}
	var args QueryWhiteListProviderArgs
	if err := json.Unmarshal(data, &args); err != nil {
		return nil, err
	}
	if types.NewAddressByStr(args.WhiteListProviderAddr).ETHAddress().String() != args.WhiteListProviderAddr {
		return nil, errors.New("provider address is invalid")
	}
	authInfo := c.getAuthInfo(addr.String())
	if authInfo == nil || authInfo.Role != SuperUser {
		return nil, ErrQueryPermission
	}
	state, b := c.account.GetState([]byte(WhiteListProviderKey))
	if !state {
		return nil, ErrNotFound
	}
	var whiteListProviders []WhiteListProvider
	if err := json.Unmarshal(b, &whiteListProviders); err != nil {
		return nil, err
	}
	associate := lo.Associate[WhiteListProvider, string, *WhiteListProvider](whiteListProviders, func(k WhiteListProvider) (string, *WhiteListProvider) {
		return k.WhiteListProviderAddr, &k
	})
	provider := associate[args.WhiteListProviderAddr]
	if provider == nil {
		return nil, ErrNotFound
	}
	b, err := json.Marshal(provider)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Verify(lg ledger.StateLedger, needApprove string) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.WhiteListContractAddr))
	state, b := account.GetState([]byte(AuthInfoKey + needApprove))
	if !state {
		return ErrVerify
	}
	info := &AuthInfo{}
	if err := json.Unmarshal(b, &info); err != nil {
		return ErrVerify
	}
	if info.Role == SuperUser {
		return nil
	}
	role := info.Role
	if role != BasicUser && role != SuperUser {
		return ErrVerify
	}
	providers := info.Providers
	for _, addr := range providers {
		if common.IsInSlice(addr, Providers) {
			return nil
		}
	}
	return ErrVerify
}

func (c *WhiteList) saveAuthInfo(info *AuthInfo) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	c.account.SetState([]byte(AuthInfoKey+info.User), b)
	return nil
}

func (c *WhiteList) getAuthInfo(addr string) *AuthInfo {
	state, i := c.account.GetState([]byte(AuthInfoKey + addr))
	if state {
		res := &AuthInfo{}
		if err := json.Unmarshal(i, res); err == nil {
			return res
		}
	}
	return nil
}

func CheckInServices(account ledger.IAccount, addr string) error {
	isExist, data := account.GetState([]byte(WhiteListProviderKey))
	if !isExist {
		return ErrCheckWhiteListProvider
	}
	var Services []WhiteListProvider
	if err := json.Unmarshal(data, &Services); err != nil {
		return ErrCheckWhiteListProvider
	}
	if len(Services) == 0 {
		return ErrCheckWhiteListProvider
	}
	if !common.IsInSlice[string](addr, lo.Map[WhiteListProvider, string](Services, func(item WhiteListProvider, index int) string {
		return item.WhiteListProviderAddr
	})) {
		return ErrCheckWhiteListProvider
	}
	return nil
}

func AddAndRemoveProviders(lg ledger.StateLedger, modifyType ModifyType, inputServices []WhiteListProvider) error {
	existServices, err := GetProviders(lg)
	if err != nil {
		return err
	}
	switch modifyType {
	case AddWhiteListProvider:
		existServices = append(existServices, inputServices...)
		addrToServiceMap := lo.Associate(existServices, func(service WhiteListProvider) (string, WhiteListProvider) {
			return service.WhiteListProviderAddr, service
		})
		existServices = lo.MapToSlice(addrToServiceMap, func(key string, value WhiteListProvider) WhiteListProvider {
			return value
		})
		return SetProviders(lg, existServices)
	case RemoveWhiteListProvider:
		if len(existServices) > 0 {
			addrToServiceMap := lo.Associate(existServices, func(service WhiteListProvider) (string, WhiteListProvider) {
				return service.WhiteListProviderAddr, service
			})
			filteredMembers := lo.Reject(inputServices, func(service WhiteListProvider, _ int) bool {
				_, exists := addrToServiceMap[service.WhiteListProviderAddr]
				return exists
			})
			existServices = filteredMembers
			return SetProviders(lg, existServices)
		} else {
			return errors.New("access error: remove provider from an empty list")
		}
	default:
		return errors.New("access error: wrong submit type")
	}
}

func SetProviders(lg ledger.StateLedger, services []WhiteListProvider) error {
	// Sort list based on the Providers
	sort.Slice(services, func(i, j int) bool {
		return services[i].WhiteListProviderAddr < services[j].WhiteListProviderAddr
	})
	cb, err := json.Marshal(services)
	if err != nil {
		return err
	}
	lg.GetOrCreateAccount(types.NewAddressByStr(common.WhiteListContractAddr)).SetState([]byte(WhiteListProviderKey), cb)
	return nil
}

func GetProviders(lg ledger.StateLedger) ([]WhiteListProvider, error) {
	success, data := lg.GetOrCreateAccount(types.NewAddressByStr(common.WhiteListContractAddr)).GetState([]byte(WhiteListProviderKey))
	var services []WhiteListProvider
	if success {
		if err := json.Unmarshal(data, &services); err != nil {
			return nil, err
		}
		return services, nil
	}
	return services, nil
}

func InitProvidersAndWhiteList(lg ledger.StateLedger, initVerifiedUsers []string, initProviders []string) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.WhiteListContractAddr))
	uniqueMap := make(map[string]struct{})
	for _, str := range initVerifiedUsers {
		uniqueMap[str] = struct{}{}
	}
	for _, str := range initProviders {
		uniqueMap[str] = struct{}{}
	}
	allAddresses := make([]string, len(uniqueMap))
	i := 0
	for key := range uniqueMap {
		allAddresses[i] = key
		i++
	}
	sort.Strings(allAddresses)
	// init super user
	for _, addrStr := range allAddresses {
		info := &AuthInfo{
			User:      addrStr,
			Providers: []string{},
			Role:      SuperUser,
		}
		b, err := json.Marshal(info)
		if err != nil {
			return err
		}
		account.SetState([]byte(AuthInfoKey+addrStr), b)
	}
	var whiteListProviders []WhiteListProvider
	for _, addrStr := range initProviders {
		service := WhiteListProvider{
			WhiteListProviderAddr: addrStr,
		}
		whiteListProviders = append(whiteListProviders, service)
	}
	// Sort list based on the Providers
	sort.Slice(whiteListProviders, func(i, j int) bool {
		return whiteListProviders[i].WhiteListProviderAddr < whiteListProviders[j].WhiteListProviderAddr
	})
	marshal, err := json.Marshal(whiteListProviders)
	if err != nil {
		return err
	}
	account.SetState([]byte(WhiteListProviderKey), marshal)
	loggers.Logger(loggers.Access).Debugf("finish init providers and white list")
	return nil
}

func (c *WhiteList) RecordLog(method string, data []byte) {
	currentLog := common.Log{
		Address: types.NewAddressByStr(common.WhiteListContractAddr),
	}
	currentLog.Topics = append(currentLog.Topics, types.NewHashByStr(method))
	currentLog.Data = data
	currentLog.Removed = false

	*c.currentLogs = append(*c.currentLogs, currentLog)
}
