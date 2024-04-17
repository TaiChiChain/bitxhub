package access

import (
	"encoding/json"
	"errors"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
)

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

func TestWhiteList_Submit(t *testing.T) {
	whitelist := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	admins := []string{admin1, admin2}
	err := InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)

	testcases := []struct {
		Caller string
		Data   []byte
		Err    error
		HasErr bool
	}{
		{ // case1 : submit success
			Caller: admin1,
			Data: generateRunData(t, &SubmitArgs{
				Addresses: []string{admin3},
			}),
			HasErr: false,
		},
		{ // case2 : wrong user
			Caller: "",
			Data: generateRunData(t, &SubmitArgs{
				Addresses: []string{admin3},
			}),
			Err: ErrUser,
		},
		{
			Caller: admin1,
			Data:   []byte("error data"),
			HasErr: true,
		},
		{
			Caller: admin3,
			Data: generateRunData(t, &SubmitArgs{
				Addresses: []string{admin3},
			}),
			Err: ErrProviderPermission,
		},
		{
			Caller: admin1,
			Data: generateRunData(t, &SubmitArgs{
				Addresses: []string{"0x111"},
			}),
			Err: ErrCheckSubmitInfo,
		},
		{
			Caller: admin1,
			Data: generateRunData(t, &SubmitArgs{
				Addresses: []string{admin2},
			}),
			Err: ErrCheckSubmitInfo,
		},
	}

	for _, test := range testcases {
		from := types.NewAddressByStr(test.Caller).ETHAddress()
		user := &from
		if test.Caller == "" {
			user = nil
		}
		logs := make([]common.Log, 0)
		whitelist.SetContext(&common.VMContext{
			BlockNumber: 1,
			StateLedger: stateLedger,
			From:        user,
			CurrentLogs: &logs,
		})

		err := whitelist.Submit(test.Data)
		if test.Err != nil {
			assert.Equal(t, test.Err, err)
		} else {
			if test.HasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func generateRunData(t *testing.T, anyArgs any) []byte {
	marshal, err := json.Marshal(anyArgs)
	assert.Nil(t, err)

	return marshal
}

func TestAddAndRemoveProviders(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	providers := []WhitelistProviderInfo{
		{
			Addr: admin1,
		},
		{
			Addr: admin2,
		},
	}

	testcases := []struct {
		method ModifyType
		data   []WhitelistProviderInfo
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
			err:    errors.New("access error: remove provider from an empty list"),
		},
		{
			method: 3,
			data:   nil,
			err:    errors.New("access error: wrong submit type"),
		},
	}
	for _, test := range testcases {
		err := UpdateProviders(stateLedger, test.method, test.data)
		assert.Equal(t, test.err, err)
	}
}

func TestGetProviders(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	// case 1 empty list
	services, err := GetProviders(stateLedger)
	assert.Nil(t, err)
	assert.True(t, len(services) == 0)

	// case 2 not nil list
	listProviders := []WhitelistProviderInfo{
		{
			Addr: admin1,
		},
		{
			Addr: admin2,
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

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	providers := []WhitelistProviderInfo{
		{
			Addr: admin1,
		},
		{
			Addr: admin2,
		},
	}
	err := SetProviders(stateLedger, providers)
	assert.Nil(t, err)
}

func TestVerify(t *testing.T) {
	whitelist := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))
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
	err := Verify(stateLedger, testcase.needApproveAddr)
	assert.Equal(t, testcase.err, err)

	// test others
	admins := []string{admin1}
	err = InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)
	testcases := []struct {
		needApprove string
		authInfo    WhitelistAuthInfo
		expected    error
	}{
		{
			needApprove: admin2,
			authInfo: WhitelistAuthInfo{
				User:      admin2,
				Providers: []string{admin1},
				Role:      BasicUser,
			},
			expected: nil,
		},
		{
			needApprove: admin2,
			authInfo: WhitelistAuthInfo{
				User:      admin2,
				Providers: []string{},
				Role:      BasicUser,
			},
			expected: ErrVerify,
		},
		{
			needApprove: admin2,
			authInfo: WhitelistAuthInfo{
				User:      admin2,
				Providers: []string{},
				Role:      SuperUser,
			},
			// Verify: fail by checking info
			expected: nil,
		},
		{
			needApprove: admin2,
			authInfo: WhitelistAuthInfo{
				User:      admin2,
				Providers: []string{admin1},
				Role:      2,
			},
			// Verify: fail by checking info
			expected: ErrVerify,
		},
		{
			needApprove: admin2,
			authInfo: WhitelistAuthInfo{
				User:      admin2,
				Providers: []string{admin3},
				Role:      0,
			},
			expected: ErrVerify,
		},
	}

	for _, test := range testcases {
		whitelist.SetContext(&common.VMContext{
			BlockNumber: 1,
			StateLedger: stateLedger,
		})
		err := whitelist.saveAuthInfo(&test.authInfo)
		assert.Nil(t, err)
		err = Verify(stateLedger, test.needApprove)
		assert.Equal(t, test.expected, err)
	}
}

func TestWhiteList_Remove(t *testing.T) {
	whitelist := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	admins := []string{admin1}
	err := InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)

	// add a user
	user1 := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	whitelist.SetContext(&common.VMContext{
		StateLedger: stateLedger,
		BlockNumber: 1,
		From:        &user1,
		CurrentLogs: &logs,
	})
	err = whitelist.Submit(generateRunData(t, SubmitArgs{
		Addresses: []string{admin2},
	}))
	assert.Nil(t, err)

	// before case: auth info not exist
	testcases := []struct {
		from   ethcommon.Address
		args   *RemoveArgs
		err    error
		hasErr bool
	}{
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin2},
			},
			hasErr: false,
		},
		{
			from: types.NewAddressByStr(admin3).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{admin1},
			},
			err: ErrProviderPermission,
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
		{
			from: ethcommon.Address{},
			args: &RemoveArgs{
				Addresses: []string{admin1},
			},
			err: ErrUser,
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []string{"0x1222"},
			},
			hasErr: true,
		},
	}
	for _, test := range testcases {
		user := &test.from
		if (test.from == ethcommon.Address{}) {
			user = nil
		}
		logs := make([]common.Log, 0)
		whitelist.SetContext(&common.VMContext{
			BlockNumber: 1,
			StateLedger: stateLedger,
			From:        user,
			CurrentLogs: &logs,
		})
		err := whitelist.Remove(generateRunData(t, test.args))
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else {
			if test.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestWhiteList_QueryAuthInfo(t *testing.T) {
	whitelist := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	admins := []string{admin1}
	err := InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)

	user1 := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	whitelist.SetContext(&common.VMContext{
		StateLedger: stateLedger,
		BlockNumber: 1,
		From:        &user1,
		CurrentLogs: &logs,
	})

	err = whitelist.Submit(generateRunData(t, SubmitArgs{
		Addresses: []string{admin2},
	}))
	assert.Nil(t, err)

	testcases := []struct {
		from        string
		arg         *QueryAuthInfoArgs
		expectedErr error
		hasErr      bool
	}{
		{
			from: admin1,
			arg: &QueryAuthInfoArgs{
				User: admin1,
			},
			hasErr: false,
		},
		{
			from: admin1,
			arg: &QueryAuthInfoArgs{
				User: admin2,
			},
			hasErr: false,
		},
		{
			from: admin1,
			arg: &QueryAuthInfoArgs{
				User: admin3,
			},
			expectedErr: ErrNotFound,
		},
		{
			from: admin1,
			arg: &QueryAuthInfoArgs{
				User: "qwty",
			},
			hasErr: true,
		},
		{
			from: admin3,
			arg: &QueryAuthInfoArgs{
				User: admin1,
			},
			expectedErr: ErrQueryPermission,
		},
		{
			from: "",
			arg: &QueryAuthInfoArgs{
				User: admin1,
			},
			expectedErr: ErrUser,
		},
		{
			from:   admin1,
			arg:    nil,
			hasErr: true,
		},
	}

	for _, testcase := range testcases {
		user := types.NewAddressByStr(testcase.from).ETHAddress()
		from := &user
		if testcase.from == "" {
			from = nil
		}
		whitelist.SetContext(&common.VMContext{
			StateLedger: stateLedger,
			BlockNumber: 1,
			From:        from,
			CurrentLogs: &logs,
		})

		data, err := json.Marshal(testcase.arg)
		if err != nil {
			data = nil
		}

		_, err = whitelist.QueryAuthInfo(data)
		if testcase.expectedErr != nil {
			assert.Equal(t, testcase.expectedErr, err)
		} else {
			if testcase.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestWhiteList_QueryWhiteListProvider(t *testing.T) {
	whitelist := NewWhiteList(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	admins := []string{admin1}
	err := InitProvidersAndWhiteList(stateLedger, admins, admins)
	assert.Nil(t, err)

	testcases := []struct {
		from        string
		arg         *QueryWhiteListProviderArgs
		expectedErr error
		hasErr      bool
	}{
		{
			from: admin1,
			arg: &QueryWhiteListProviderArgs{
				WhiteListProviderAddr: admin1,
			},
			hasErr: false,
		},
		{
			from: admin1,
			arg: &QueryWhiteListProviderArgs{
				WhiteListProviderAddr: admin2,
			},
			expectedErr: ErrNotFound,
		},
		{
			from: admin1,
			arg: &QueryWhiteListProviderArgs{
				WhiteListProviderAddr: "tyu",
			},
			hasErr: true,
		},
		{
			from: admin3,
			arg: &QueryWhiteListProviderArgs{
				WhiteListProviderAddr: admin1,
			},
			expectedErr: ErrQueryPermission,
		},
		{
			from: "",
			arg: &QueryWhiteListProviderArgs{
				WhiteListProviderAddr: admin1,
			},
			expectedErr: ErrUser,
		},
		{
			from:   admin1,
			arg:    nil,
			hasErr: true,
		},
	}

	for _, testcase := range testcases {
		user := types.NewAddressByStr(testcase.from).ETHAddress()
		from := &user
		if testcase.from == "" {
			from = nil
		}
		logs := make([]common.Log, 0)
		whitelist.SetContext(&common.VMContext{
			StateLedger: stateLedger,
			BlockNumber: 1,
			From:        from,
			CurrentLogs: &logs,
		})

		data, err := json.Marshal(testcase.arg)
		if err != nil {
			data = nil
		}

		_, err = whitelist.QueryWhiteListProvider(data)
		if testcase.expectedErr != nil {
			assert.Equal(t, testcase.expectedErr, err)
		} else {
			if testcase.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}

func TestCheckInServices(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	state := CheckInServices(account, admin1)
	assert.Equal(t, ErrProviderPermission, state)

	err := InitProvidersAndWhiteList(stateLedger, nil, []string{admin1})
	assert.Nil(t, err)
	state = CheckInServices(account, admin1)
	assert.Nil(t, state)
}

func TestInitProvidersAndWhiteList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.WhiteListContractAddr))

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
	err := InitProvidersAndWhiteList(stateLedger, append(richAccounts, append(adminAccounts, services...)...), services)
	assert.Nil(t, err)
}
