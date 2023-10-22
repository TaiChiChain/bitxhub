package access

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	vm "github.com/axiomesh/eth-kit/evm"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

func TestWhiteList_RunForSubmit(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	admins := []string{admin1, admin2}
	err = InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{ // case1 : submit success
			Caller: admin1,
			Data: generateRunData(t, SubmitMethod, &SubmitArgs{
				Addresses: []string{admin3},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateRunData(t, SubmitMethod, &SubmitArgs{
					Addresses: []string{admin3},
				})),
				Err: nil,
				ReturnData: generateReturnData(t, &SubmitArgs{
					Addresses: []string{admin3},
				}),
			},
			Err: nil,
		},
		{ // case2 : wrong method
			Caller: admin1,
			Data:   []byte{0, 1, 2, 3},
			Err:    ErrGetMethodName,
		},
		{ // case3 : query provider fail
			Caller: admin3,
			Data:   generateRunData(t, QueryWhiteListProviderMethod, &QueryWhiteListProviderArgs{WhiteListProviderAddr: admin1}),
			Err:    ErrQueryPermission,
		},
		{ // case4 : query provider success
			Caller: admin1,
			Data:   generateRunData(t, QueryWhiteListProviderMethod, &QueryWhiteListProviderArgs{WhiteListProviderAddr: admin1}),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, QueryWhiteListProviderMethod, &QueryWhiteListProviderArgs{WhiteListProviderAddr: admin1})),
				Err:        nil,
				ReturnData: generateQueryReturnData(t, QueryWhiteListProviderMethod, &WhiteListProvider{WhiteListProviderAddr: admin1}),
			},
			Err: nil,
		},
		{ // case4 : query auth info fail
			Caller: admin3,
			Data:   generateRunData(t, QueryAuthInfoMethod, &QueryAuthInfoArgs{User: admin3}),
			Err:    ErrQueryPermission,
		},
		{ // case5 : query auth info success
			Caller: admin1,
			Data:   generateRunData(t, QueryAuthInfoMethod, &QueryAuthInfoArgs{User: admin3}),
			Expected: vm.ExecutionResult{
				UsedGas: common.CalculateDynamicGas(generateRunData(t, QueryAuthInfoMethod, &QueryAuthInfoArgs{User: admin3})),
				Err:     nil,
				ReturnData: generateQueryReturnData(t, QueryAuthInfoMethod, &AuthInfo{
					User:      admin3,
					Providers: []string{admin1},
					Role:      BasicUser,
				}),
			},
			Err: nil,
		},
		{ // case6 : remove auth info success
			Caller: admin1,
			Data:   generateRunData(t, RemoveMethod, &RemoveArgs{Addresses: []string{admin3}}),
			Expected: vm.ExecutionResult{
				UsedGas:    common.CalculateDynamicGas(generateRunData(t, RemoveMethod, &RemoveArgs{Addresses: []string{admin3}})),
				Err:        nil,
				ReturnData: generateReturnData(t, &RemoveArgs{Addresses: []string{admin3}}),
			},
			Err: nil,
		},
	}

	for _, test := range testcases {
		cm.Reset(stateLedger)

		result, err := cm.Run(&vm.Message{
			From: types.NewAddressByStr(test.Caller).ETHAddress(),
			Data: test.Data,
		})
		assert.Equal(t, test.Err, err)

		if result != nil {
			assert.Equal(t, nil, result.Err)
			assert.Equal(t, test.Expected.UsedGas, result.UsedGas)
			assert.Equal(t, test.Expected.ReturnData, result.ReturnData)
		}
	}
}

func generateReturnData(t *testing.T, s any) []byte {
	marshal, err := json.Marshal(s)
	assert.Nil(t, err)
	return marshal
}

func generateQueryReturnData(t *testing.T, methodName string, s any) []byte {
	marshal, err := json.Marshal(s)
	assert.Nil(t, err)
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	b, err := cm.PackOutputArgs(methodName, marshal)
	assert.Nil(t, err)
	return b
}

func TestEstimateGas(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.WhiteListContractAddr).ETHAddress()

	data := hexutil.Bytes(generateRunData(t, SubmitMethod, &SubmitArgs{
		Addresses: []string{
			admin3,
		},
	}))
	gas, err := cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(data), gas)

	data = hexutil.Bytes(generateRunData(t, RemoveMethod, &RemoveArgs{Addresses: []string{admin1}}))
	gas, err = cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, common.CalculateDynamicGas(data), gas)

	// test error args
	data = hexutil.Bytes([]byte{0, 1, 2, 3})
	gas, err = cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), gas)
}

func generateRunData(t *testing.T, method string, anyArgs any) []byte {
	marshal, err := json.Marshal(anyArgs)
	assert.Nil(t, err)

	gabi, err := GetABI()
	assert.Nil(t, err)
	data, err := gabi.Pack(method, marshal)
	assert.Nil(t, err)

	return data
}

func TestWhiteList_CheckAndUpdateState(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	cm.CheckAndUpdateState(100, stateLedger)
}

func TestWhiteList_ParseErrorArgs(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	truedata, err := cm.gabi.Pack(SubmitMethod, []byte(""))
	assert.Nil(t, err)
	testcases := []struct {
		method   string
		data     []byte
		Expected error
	}{
		{
			method:   SubmitMethod,
			data:     []byte{1},
			Expected: ErrParseArgs,
		},
		{
			method:   SubmitMethod,
			data:     []byte{1, 2, 3, 4},
			Expected: fmt.Errorf("abi: attempting to unmarshall an empty string while arguments are expected"),
		},
		{
			method:   SubmitMethod,
			data:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Expected: ErrParseArgs,
		},
		{
			method:   "test method",
			data:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36},
			Expected: ErrParseArgs,
		},
		{
			method:   "test method",
			data:     truedata,
			Expected: ErrParseArgs,
		},
	}

	for _, test := range testcases {
		args := &SubmitArgs{}
		err := cm.ParseArgs(&vm.Message{
			Data: test.data,
		}, test.method, args)
		assert.Equal(t, test.Expected, err)
	}
}

func TestGetArgsForSubmit(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	submitArgs := &SubmitArgs{Addresses: []string{
		admin3,
	}}
	marshal, err := json.Marshal(submitArgs)
	assert.Nil(t, err)
	data, err := cm.gabi.Pack(SubmitMethod, marshal)
	assert.Nil(t, err)

	arg, err := cm.getArgs(&vm.Message{
		Data: data,
	})
	assert.Nil(t, err)

	actualArgs, ok := arg.(*SubmitArgs)
	assert.True(t, ok)
	assert.Equal(t, submitArgs.Addresses[0], actualArgs.Addresses[0])
}

func TestGetArgsForRemove(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	removeArgs := &RemoveArgs{Addresses: []string{
		admin1,
		admin2,
	}}
	marshal, err := json.Marshal(removeArgs)
	assert.Nil(t, err)
	data, err := cm.gabi.Pack(RemoveMethod, marshal)
	assert.Nil(t, err)
	arg, err := cm.getArgs(&vm.Message{
		Data: data,
	})
	assert.Nil(t, err)

	actualArgs, ok := arg.(*RemoveArgs)
	assert.True(t, ok)

	assert.Equal(t, removeArgs.Addresses[0], actualArgs.Addresses[0])

}

func TestWhiteList_GetErrArgs(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	testjsondata := `
	[
		{"type": "function", "name": "this_method", "inputs": [{"name": "extra", "type": "bytes"}], "outputs": [{"name": "proposalId", "type": "uint64"}]}
	]
	`
	gabi, err := abi.JSON(strings.NewReader(testjsondata))
	cm.gabi = &gabi

	errMethod := "no_this_method"
	errMethodData, err := cm.gabi.Pack(errMethod, []byte(""))
	assert.NotNil(t, err)

	thisMethod := "this_method"
	thisMethodData, err := cm.gabi.Pack(thisMethod, []byte(""))
	assert.Nil(t, err)

	testcases := [][]byte{
		nil,
		{1, 2, 3, 4},
		errMethodData,
		thisMethodData,
	}

	for _, test := range testcases {
		_, err = cm.getArgs(&vm.Message{Data: test})
		assert.NotNil(t, err)
	}
}

func TestAddAndRemoveProviders(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	providers := []WhiteListProvider{
		{
			admin1,
		},
		{
			admin2,
		},
	}

	testcases := []struct {
		method ModifyType
		data   []WhiteListProvider
		err    error
	}{
		{
			method: AddWhiteListProvider,
			data:   providers,
			err:    nil,
		},
		{
			method: RemoveWhiteListProvider,
			data:   providers,
			err:    nil,
		},
		{
			method: RemoveWhiteListProvider,
			data:   providers,
			err:    fmt.Errorf("access error: remove provider from an empty list"),
		},
		{
			method: 3,
			data:   nil,
			err:    fmt.Errorf("access error: wrong submit type"),
		},
	}
	for _, test := range testcases {
		err := AddAndRemoveProviders(stateLedger, test.method, test.data)
		assert.Equal(t, test.err, err)
	}

}

func TestGetProviders(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	// case 1 empty list
	services, err := GetProviders(stateLedger)
	assert.Nil(t, err)
	assert.True(t, len(services) == 0)

	// case 2 not nil list
	listProviders := []WhiteListProvider{
		{
			admin1,
		},
		{
			admin2,
		},
	}
	err = SetProviders(stateLedger, listProviders)
	assert.Nil(t, err)
	services, err = GetProviders(stateLedger)
	assert.Nil(t, err)
	assert.True(t, len(services) == 2)
}

func TestSetProviders(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	providers := []WhiteListProvider{
		{
			admin1,
		},
		{
			admin2,
		},
	}
	err = SetProviders(stateLedger, providers)
	assert.Nil(t, err)
}

func TestWhiteList_SaveLog(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	assert.Nil(t, cm.currentLog)
	cm.Reset(stateLedger)
	assert.NotNil(t, cm.currentLog)
	cm.currentLog.Removed = false
	cm.currentLog.Data = []byte{0, 1, 2}
	cm.currentLog.Address = types.NewAddressByStr(admin1)
	cm.currentLog.Topics = []*types.Hash{
		types.NewHash([]byte{0, 1, 2}),
	}
	cm.SaveLog(stateLedger, cm.currentLog)
}

func TestVerify(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	// test fail to get state
	testcase := struct {
		needApproveAddr string
		expected        bool
		err             error
	}{
		needApproveAddr: admin2,
		expected:        false,
		// Verify: fail by GetState
		err: ErrVerify,
	}
	err = Verify(stateLedger, testcase.needApproveAddr)
	assert.Equal(t, testcase.err, err)

	// test others
	admins := []string{admin1}
	InitProvidersAndWhiteList(stateLedger, admins, admins)
	testcases := []struct {
		needApprove string
		authInfo    AuthInfo
		expected    error
	}{
		{
			needApprove: admin2,
			authInfo: AuthInfo{
				User:      admin2,
				Providers: []string{admin1},
				Role:      BasicUser,
			},
			expected: nil,
		},
		{
			needApprove: admin2,
			authInfo: AuthInfo{
				User:      admin2,
				Providers: []string{},
				Role:      BasicUser,
			},
			expected: ErrVerify,
		},
		{
			needApprove: admin2,
			authInfo: AuthInfo{
				User:      admin2,
				Providers: []string{},
				Role:      SuperUser,
			},
			// Verify: fail by checking info
			expected: nil,
		},
		{
			needApprove: admin2,
			authInfo: AuthInfo{
				User:      admin2,
				Providers: []string{admin1},
				Role:      2,
			},
			// Verify: fail by checking info
			expected: ErrVerify,
		},
		{
			needApprove: admin2,
			authInfo: AuthInfo{
				User:      admin2,
				Providers: []string{admin3},
				Role:      0,
			},
			expected: ErrVerify,
		},
	}

	for _, test := range testcases {
		cm.Reset(stateLedger)
		cm.saveAuthInfo(&test.authInfo)
		err := Verify(stateLedger, test.needApprove)
		assert.Equal(t, test.expected, err)
	}
}

func TestWhiteList_ErrorSubmit(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, expected := ledger.NewAccountCache()
	assert.Nil(t, expected)
	repoRoot := t.TempDir()
	ld, expected := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, expected)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	cm.Reset(stateLedger)
	admins := []string{admin1}
	InitProvidersAndWhiteList(stateLedger, admins, admins)

	testcases := []struct {
		from     ethcommon.Address
		args     *SubmitArgs
		expected error
	}{
		{
			from: types.NewAddressByStr(admin2).ETHAddress(),
			args: &SubmitArgs{
				Addresses: []string{admin2},
			},
			expected: ErrCheckWhiteListProvider,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &SubmitArgs{
				Addresses: []string{admin2},
			},
			expected: nil,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &SubmitArgs{
				Addresses: []string{"admin3"},
			},
			expected: ErrCheckSubmitInfo,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &SubmitArgs{
				Addresses: []string{admin1},
			},
			expected: ErrCheckSubmitInfo,
		},
	}

	for _, test := range testcases {
		_, expected := cm.submit(&test.from, test.args)
		assert.Equal(t, test.expected, expected)
	}

	// test auth info already existed
	testcases = []struct {
		from     ethcommon.Address
		args     *SubmitArgs
		expected error
	}{
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &SubmitArgs{
				Addresses: []string{admin4},
			},
			expected: ErrCheckSubmitInfo,
		},
	}
	b, _ := json.Marshal(&AuthInfo{
		User:      admin4,
		Providers: []string{admin1},
		Role:      0,
	})
	account.SetState([]byte(AuthInfoKey+admin4), b)
	_, err := cm.submit(&testcases[0].from, testcases[0].args)
	assert.Equal(t, testcases[0].expected, err)

	// test nil submit
	testcase := struct {
		from     *ethcommon.Address
		args     *SubmitArgs
		expected error
	}{
		from: nil,
		args: &SubmitArgs{
			Addresses: []string{},
		},
		expected: ErrUser,
	}
	_, err = cm.submit(testcase.from, testcase.args)
	assert.Equal(t, testcase.expected, err)
}

func TestWhiteList_ErrorRemove(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	cm.Reset(stateLedger)
	admins := []string{admin1}
	InitProvidersAndWhiteList(stateLedger, admins, admins)
	// before case: auth info not exist
	testcases := []struct {
		from ethcommon.Address
		args *RemoveArgs
		err  error
	}{
		{
			from: types.NewAddressByStr(admin2).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin1},
			},
			err: ErrCheckWhiteListProvider,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin3},
			},
			err: ErrCheckRemoveInfo,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin1},
			},
			err: ErrCheckRemoveInfo,
		},
	}
	for _, test := range testcases {
		_, err := cm.remove(&test.from, test.args)
		assert.Equal(t, test.err, err)
	}

	authInfo := &AuthInfo{
		User:      admin3,
		Providers: []string{admin2},
		Role:      BasicUser,
	}
	cm.saveAuthInfo(authInfo)
	testcases = []struct {
		from ethcommon.Address
		args *RemoveArgs
		err  error
	}{
		{
			from: types.NewAddressByStr(admin2).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin3},
			},
			err: ErrCheckWhiteListProvider,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin3},
			},
			err: ErrCheckRemoveInfo,
		},
	}
	for _, test := range testcases {
		_, err := cm.remove(&test.from, test.args)
		assert.Equal(t, test.err, err)
	}

	// others: test err user
	_, err = cm.remove(nil, testcases[0].args)
	assert.Equal(t, ErrUser, err)

}

func TestCheckInServices(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	state := CheckInServices(account, admin1)
	assert.Equal(t, ErrCheckWhiteListProvider, state)

	err = InitProvidersAndWhiteList(stateLedger, nil, []string{admin1})
	assert.Nil(t, err)
	state = CheckInServices(account, admin1)
	assert.Equal(t, true, true)
}

func TestInitProvidersAndWhiteList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	richAccounts := []string{
		admin1,
	}
	adminAccounts := []string{
		admin1,
		admin2,
		admin3,
	}
	services := []string{
		admin4,
	}
	err = InitProvidersAndWhiteList(stateLedger, append(richAccounts, append(adminAccounts, services...)...), services)
	assert.Nil(t, err)
}

func TestWhiteList_getErrorArgs(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	testcases := []struct {
		method   string
		data     *vm.Message
		Expected string
	}{
		{
			method: SubmitMethod,
			data: &vm.Message{
				Data: []byte{239, 127, 167, 27, 0},
			},
			Expected: ErrParseArgs.Error(),
		},
		{
			method: RemoveMethod,
			data: &vm.Message{
				Data: []byte{88, 237, 239, 76, 0},
			},
			Expected: ErrParseArgs.Error(),
		},
		{
			method: QueryWhiteListProviderMethod,
			data: &vm.Message{
				Data: []byte{154, 100, 149, 61, 0},
			},
			Expected: ErrParseArgs.Error(),
		},
		{
			method: QueryAuthInfoMethod,
			data: &vm.Message{
				Data: []byte{254, 117, 73, 238, 0},
			},
			Expected: ErrParseArgs.Error(),
		},
	}

	for _, test := range testcases {
		getArgs, err := cm.getArgs(test.data)
		assert.Nil(t, getArgs)
		assert.Equal(t, test.Expected, err.Error())
	}
}

func TestWhiteList_Reset(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "white_list"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.WhiteListContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	account.SetState([]byte(WhiteListProviderKey), []byte{1, 2, 3})
	cm.Reset(stateLedger)
}

func TestWhiteList_PackOutputArgs(t *testing.T) {
	cm := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	testcases := []struct {
		Method string
		Inputs []any
		IsErr  bool
	}{
		{
			Method: SubmitMethod,
			Inputs: []any{[]byte("{}")},
			IsErr:  true,
		},
		{
			Method: SubmitMethod,
			Inputs: []any{"test str"},
			IsErr:  true,
		},
		{
			Method: SubmitMethod,
			Inputs: []any{uint64(999)},
			IsErr:  true,
		},
		{
			Method: QueryAuthInfoMethod,
			Inputs: []any{[]byte("bytes")},
			IsErr:  false,
		},
	}
	for _, testcase := range testcases {
		packed, err := cm.PackOutputArgs(testcase.Method, testcase.Inputs...)
		if testcase.IsErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)

			ret, err := cm.UnpackOutputArgs(testcase.Method, packed)
			assert.Nil(t, err)
			assert.Equal(t, testcase.Inputs, ret)
		}
	}
}
