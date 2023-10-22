package access

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	vm "github.com/axiomesh/eth-kit/evm"
)

const (
	SubmitMethod                 = "submit"
	RemoveMethod                 = "remove"
	QueryAuthInfoMethod          = "queryAuthInfo"
	QueryWhiteListProviderMethod = "queryWhiteListProvider"
	AuthInfoKey                  = "authinfo"
	WhiteListProviderKey         = "providers"
)

var _ common.SystemContract = (*WhiteList)(nil)

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

// Global variable
var Providers []string

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

var method2Sig = map[string]string{
	SubmitMethod:                 "submit(bytes)",
	RemoveMethod:                 "remove(bytes)",
	QueryAuthInfoMethod:          "queryAuthInfo(bytes)",
	QueryWhiteListProviderMethod: "queryWhiteListProvider(bytes)",
}

type WhiteList struct {
	stateLedger ledger.StateLedger
	account     ledger.IAccount
	currentLog  *common.Log
	logger      logrus.FieldLogger
	method2Sig  map[string][]byte
	gabi        *abi.ABI
}

// NewWhiteList constructs a new WhiteList
func NewWhiteList(cfg *common.SystemContractConfig) *WhiteList {
	gabi, err := GetABI()
	if err != nil {
		panic(err)
	}

	return &WhiteList{
		logger:     loggers.Logger(loggers.Access),
		gabi:       gabi,
		method2Sig: initMethodSignature(),
	}
}

const jsondata = `
[
	{"type": "function", "name": "submit", "inputs": [{"name": "addresses", "type": "bytes"}]},
	{"type": "function", "name": "remove", "inputs": [{"name": "addresses", "type": "bytes"}]},
	{"type": "function", "name": "queryAuthInfo", "stateMutability": "view", "inputs": [{"name": "user", "type": "bytes"}], "outputs": [{"name": "authInfo", "type": "bytes"}]},
	{"type": "function", "name": "queryWhiteListProvider", "stateMutability": "view", "inputs": [{"name": "whiteListProviderAddr", "type": "bytes"}], "outputs": [{"name": "whiteListProvider", "type": "bytes"}]}
]
`

// GetABI get system contract abi
func GetABI() (*abi.ABI, error) {
	gabi, err := abi.JSON(strings.NewReader(jsondata))
	if err != nil {
		return nil, err
	}
	return &gabi, nil
}

func initMethodSignature() map[string][]byte {
	m2sig := make(map[string][]byte)
	for methodName, methodSig := range method2Sig {
		m2sig[methodName] = crypto.Keccak256([]byte(methodSig))
	}
	return m2sig
}

func (c *WhiteList) Reset(stateLedger ledger.StateLedger) {
	addr := types.NewAddressByStr(common.WhiteListContractAddr)
	c.account = stateLedger.GetOrCreateAccount(addr)
	c.stateLedger = stateLedger
	c.currentLog = &common.Log{
		Address: addr,
	}
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

func (c *WhiteList) EstimateGas(callArgs *types.CallArgs) (uint64, error) {
	_, err := c.getArgs(&vm.Message{Data: *callArgs.Data})
	if err != nil {
		return 0, err
	}
	dynamicGas := common.CalculateDynamicGas(*callArgs.Data)
	return dynamicGas, nil
}

func (c *WhiteList) CheckAndUpdateState(lastHeight uint64, stateLedger ledger.StateLedger) {}

func (c *WhiteList) Run(msg *vm.Message) (*vm.ExecutionResult, error) {
	defer c.SaveLog(c.stateLedger, c.currentLog)
	// parse method and arguments from msg payload
	args, err := c.getArgs(msg)
	if err != nil {
		return nil, err
	}
	var result *vm.ExecutionResult
	switch v := args.(type) {
	case *SubmitArgs:
		result, err = c.submit(&msg.From, v)
	case *RemoveArgs:
		result, err = c.remove(&msg.From, v)
	case *QueryAuthInfoArgs:
		result, err = c.queryAuthInfo(&msg.From, v)
	case *QueryWhiteListProviderArgs:
		result, err = c.queryWhiteListProvider(&msg.From, v)
	default:
		return nil, fmt.Errorf("unknown access args")
	}
	usedGas := common.CalculateDynamicGas(msg.Data)
	if result != nil {
		result.UsedGas = usedGas
	}
	return result, err
}

func (c *WhiteList) getArgs(msg *vm.Message) (any, error) {
	data := msg.Data
	if data == nil {
		return nil, vm.ErrExecutionReverted
	}
	method, err := c.getMethodName(data)
	if err != nil {
		return nil, err
	}
	switch method {
	case SubmitMethod:
		args := &BaseExtraArgs{}
		if err = c.ParseArgs(msg, SubmitMethod, args); err != nil {
			return nil, err
		}
		submitArgs := &SubmitArgs{}
		if err = json.Unmarshal(args.Extra, submitArgs); err != nil {
			return nil, err
		}
		return submitArgs, nil
	case RemoveMethod:
		args := &BaseExtraArgs{}
		if err = c.ParseArgs(msg, RemoveMethod, args); err != nil {
			return nil, err
		}
		removeArgs := &RemoveArgs{}
		if err = json.Unmarshal(args.Extra, removeArgs); err != nil {
			return nil, err
		}
		return removeArgs, nil
	case QueryAuthInfoMethod:
		args := &BaseExtraArgs{}
		if err = c.ParseArgs(msg, QueryAuthInfoMethod, args); err != nil {
			return nil, err
		}
		queryAuthInfoArgs := &QueryAuthInfoArgs{}
		if err = json.Unmarshal(args.Extra, queryAuthInfoArgs); err != nil {
			return nil, err
		}
		return queryAuthInfoArgs, nil
	case QueryWhiteListProviderMethod:
		args := &BaseExtraArgs{}
		if err = c.ParseArgs(msg, QueryWhiteListProviderMethod, args); err != nil {
			return nil, err
		}
		providerArgs := &QueryWhiteListProviderArgs{}
		if err = json.Unmarshal(args.Extra, providerArgs); err != nil {
			return nil, err
		}
		return providerArgs, nil
	default:
		return nil, fmt.Errorf("wrong method name")
	}
}

// ParseArgs parse the arguments to specified interface by method name
func (c *WhiteList) ParseArgs(msg *vm.Message, methodName string, ret any) error {
	if len(msg.Data) < 4 {
		c.logger.Debugf("access error: Parse_args: msg data length is not improperly formatted: %q - Bytes: %+v", msg.Data, msg.Data)
		return ErrParseArgs
	}
	// discard method id
	data := msg.Data[4:]
	var args abi.Arguments
	if method, ok := c.gabi.Methods[methodName]; ok {
		if len(data)%32 != 0 {
			c.logger.Debugf("access error: parse_args: improperly formatted output: %q - Bytes: %+v", data, data)
			return ErrParseArgs
		}
		args = method.Inputs
	}
	if args == nil {
		c.logger.Debugf("access error: parse_args: could not locate named method: %s", methodName)
		return ErrParseArgs
	}
	unpacked, err := args.Unpack(data)
	if err != nil {
		return err
	}
	return args.Copy(ret, unpacked)
}

func (c *WhiteList) getMethodName(data []byte) (string, error) {
	for methodName, methodSig := range c.method2Sig {
		id := methodSig[:4]
		if bytes.Equal(id, data[:4]) {
			return methodName, nil
		}
	}
	c.logger.Debugf("access error: get_method_name: get method id: %v", data[:4])
	return "", ErrGetMethodName
}

func (c *WhiteList) submit(from *ethcommon.Address, args *SubmitArgs) (*vm.ExecutionResult, error) {
	if from == nil {
		return nil, ErrUser
	}
	c.logger.Debugf("begin submit: msg sender is %s, args is %v", from.String(), args.Addresses)
	if err := CheckInServices(c.account, from.String()); err != nil {
		c.logger.Debugf("access error: submit: fail by checking providers")
		return nil, err
	}
	for _, address := range args.Addresses {
		// check user addr
		if addr := types.NewAddressByStr(address); addr.ETHAddress().String() != address {
			c.logger.Debugf("access error: info user addr is invalid")
			return nil, ErrCheckSubmitInfo
		}
		authInfo := c.getAuthInfo(address)
		if authInfo != nil {
			// check super role
			if authInfo.Role == SuperUser {
				c.logger.Debugf("access error: submit: try to modify super user")
				return nil, ErrCheckSubmitInfo
			}
			// check exist info
			if common.IsInSlice(from.String(), authInfo.Providers) {
				c.logger.Debugf("access error: auth information already exist")
				return nil, ErrCheckSubmitInfo
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
			return nil, err
		}
	}

	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	// record log
	c.RecordLog(c.currentLog, SubmitMethod, b)
	return &vm.ExecutionResult{
		ReturnData: b,
	}, nil
}

func (c *WhiteList) remove(from *ethcommon.Address, args *RemoveArgs) (*vm.ExecutionResult, error) {
	if from == nil {
		return nil, ErrUser
	}
	c.logger.Debugf("begin remove: msg sender is %s, args is %v", from.String(), args.Addresses)
	if err := CheckInServices(c.account, from.String()); err != nil {
		return nil, err
	}
	for _, addr := range args.Addresses {
		// check admin
		info := c.getAuthInfo(addr)
		if info != nil {
			if info.Role == SuperUser {
				c.logger.Debugf("access error: remove: try to modify super user [%s]", addr)
				return nil, ErrCheckRemoveInfo
			}
			if !common.IsInSlice(from.String(), info.Providers) {
				c.logger.Debugf("access error: remove: try to remove auth info that does not exist")
				return nil, ErrCheckRemoveInfo
			}
		} else {
			c.logger.Debugf("access error: remove: try to remove [%s] auth info that does not exist", addr)
			return nil, ErrCheckRemoveInfo
		}
		info.Providers = common.RemoveFirstMatchStrInSlice(info.Providers, from.String())
		if err := c.saveAuthInfo(info); err != nil {
			return nil, err
		}
	}
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	// record log
	c.RecordLog(c.currentLog, RemoveMethod, b)
	return &vm.ExecutionResult{
		ReturnData: b,
	}, nil
}

func (c *WhiteList) queryAuthInfo(addr *ethcommon.Address, args *QueryAuthInfoArgs) (*vm.ExecutionResult, error) {
	if addr == nil {
		return nil, ErrUser
	}

	userAddr := types.NewAddressByStr(args.User).ETHAddress().String()
	if userAddr != args.User {
		return nil, fmt.Errorf("user address is invalid")
	}

	authInfo := c.getAuthInfo(addr.String())
	if authInfo == nil || authInfo.Role != SuperUser {
		return nil, ErrQueryPermission
	}

	res := &vm.ExecutionResult{}
	state, b := c.account.GetState([]byte(AuthInfoKey + args.User))
	if state {
		outputArgs, err := c.PackOutputArgs(QueryAuthInfoMethod, b)
		if err != nil {
			return nil, err
		}
		res.ReturnData = outputArgs
		return res, nil
	}

	return nil, ErrNotFound
}

func (c *WhiteList) queryWhiteListProvider(addr *ethcommon.Address, args *QueryWhiteListProviderArgs) (*vm.ExecutionResult, error) {
	if addr == nil {
		return nil, ErrUser
	}
	if types.NewAddressByStr(args.WhiteListProviderAddr).ETHAddress().String() != args.WhiteListProviderAddr {
		return nil, fmt.Errorf("provider address is invalid")
	}
	authInfo := c.getAuthInfo(addr.String())
	if authInfo == nil || authInfo.Role != SuperUser {
		return nil, ErrQueryPermission
	}
	res := &vm.ExecutionResult{}
	state, b := c.account.GetState([]byte(WhiteListProviderKey))
	if !state {
		return res, nil
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
	res.ReturnData, err = c.PackOutputArgs(QueryWhiteListProviderMethod, b)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Verify(lg ledger.StateLedger, needApprove string) error {
	logger := loggers.Logger(loggers.Access)
	logger.Debugf("starting verification: address requiring approval is [%s]", needApprove)
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.WhiteListContractAddr))
	state, b := account.GetState([]byte(AuthInfoKey + needApprove))
	if !state {
		logger.Debugf("verify user addr fail by GetState, addr is %s", needApprove)
		return ErrVerify
	}
	info := &AuthInfo{}
	if err := json.Unmarshal(b, &info); err != nil {
		logger.Debugf("verify fail by json.Unmarshal, addr is %s", needApprove)
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
			return fmt.Errorf("access error: remove provider from an empty list")
		}
	default:
		return fmt.Errorf("access error: wrong submit type")
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

func (c *WhiteList) RecordLog(currentLog *common.Log, method string, data []byte) {
	currentLog.Topics = append(currentLog.Topics, types.NewHash(c.method2Sig[method]))
	currentLog.Data = data
	currentLog.Removed = false
}

// SaveLog save log
func (c *WhiteList) SaveLog(stateLedger ledger.StateLedger, currentLog *common.Log) {
	if currentLog.Data != nil {
		stateLedger.AddLog(&types.EvmLog{
			Address: currentLog.Address,
			Topics:  currentLog.Topics,
			Data:    currentLog.Data,
			Removed: currentLog.Removed,
		})
	}
}

// PackOutputArgs pack the output arguments by method name
func (c *WhiteList) PackOutputArgs(methodName string, outputArgs ...any) ([]byte, error) {
	var args abi.Arguments
	if method, ok := c.gabi.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("gabi: could not locate named method: %s", methodName)
	}

	return args.Pack(outputArgs...)
}

// UnpackOutputArgs unpack the output arguments by method name
func (c *WhiteList) UnpackOutputArgs(methodName string, packed []byte) ([]any, error) {
	var args abi.Arguments
	if method, ok := c.gabi.Methods[methodName]; ok {
		args = method.Outputs
	}

	if args == nil {
		return nil, fmt.Errorf("gabi: could not locate named method: %s", methodName)
	}

	return args.Unpack(packed)
}
