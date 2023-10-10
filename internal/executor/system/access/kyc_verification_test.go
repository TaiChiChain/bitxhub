package access

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestKycVerification_RunForSubmit(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	// init kyc services
	admins := []string{admin1, admin2}
	err = InitKycServicesAndKycInfos(stateLedger, admins, admins)
	assert.Nil(t, err)

	// case 1: add a new kyc service
	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data: generateRunData(t, SubmitMethod, &SubmitArgs{
				KycInfos: []*KycInfo{
					{
						User:    *types.NewAddressByStr(admin3),
						KycAddr: *types.NewAddressByStr(admin1),
						KycFlag: Verified,
						Expires: -1,
					},
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: KycSubmitGas,
				Err:     nil,
				ReturnData: generateReturnData(t, &SubmitArgs{
					KycInfos: []*KycInfo{
						{
							User:    *types.NewAddressByStr(admin3),
							KycAddr: *types.NewAddressByStr(admin1),
							KycFlag: Verified,
							Expires: -1,
						},
					},
				}),
			},
			Err: nil,
		},
		{
			Caller: admin1,
			Data:   []byte{0, 1, 2, 3},
			Expected: vm.ExecutionResult{
				UsedGas:    KycSubmitGas,
				Err:        fmt.Errorf("ACCESS ERROR: getMethodName"),
				ReturnData: nil,
			},
			Err: fmt.Errorf("ACCESS ERROR: getMethodName"),
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

			expectedArgs := &SubmitArgs{}
			err = json.Unmarshal(test.Expected.ReturnData, expectedArgs)
			assert.Nil(t, err)

			actualArgs := &SubmitArgs{}
			err = json.Unmarshal(result.ReturnData, actualArgs)
			assert.Nil(t, err)
			assert.Equal(t, *expectedArgs, *actualArgs)

			for _, info := range expectedArgs.KycInfos {
				state, bytes := account.GetState([]byte(KycInfoKey + info.User.ETHAddress().String()))
				assert.True(t, state)
				dbInfo := &KycInfo{}
				err = json.Unmarshal(bytes, dbInfo)
				assert.Nil(t, err)
				assert.Equal(t, info, dbInfo)

				verify, err := Verify(stateLedger, &info.User)
				assert.Nil(t, err)
				assert.Equal(t, true, verify)
			}
		}
	}
}

func TestKycVerification_RunForRemove(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	admins := []string{admin1, admin2}
	err = InitKycServicesAndKycInfos(stateLedger, admins, admins)
	assert.Nil(t, err)

	address := types.NewAddressByStr(admin1).ETHAddress()
	cm.Reset(stateLedger)
	_, err = cm.Submit(&address, &SubmitArgs{KycInfos: []*KycInfo{
		{
			User:    *types.NewAddressByStr(admin3),
			KycAddr: *types.NewAddressByStr(admin1),
			KycFlag: Verified,
			Expires: -1,
		},
	}})
	assert.Nil(t, err)

	testcases := []struct {
		Caller   string
		Data     []byte
		Expected vm.ExecutionResult
		Err      error
	}{
		{
			Caller: admin1,
			Data: generateRunData(t, RemoveMethod, &RemoveArgs{
				Addresses: []*types.Address{
					types.NewAddressByStr(admin3),
				},
			}),
			Expected: vm.ExecutionResult{
				UsedGas: KycRemoveGas,
				Err:     nil,
				ReturnData: generateReturnData(t, &RemoveArgs{
					Addresses: []*types.Address{
						types.NewAddressByStr(admin3),
					},
				}),
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

			expectedArgs := &RemoveArgs{}
			err = json.Unmarshal(test.Expected.ReturnData, expectedArgs)
			assert.Nil(t, err)

			actualArgs := &RemoveArgs{}
			err = json.Unmarshal(result.ReturnData, actualArgs)
			assert.Nil(t, err)
			assert.Equal(t, *expectedArgs, *actualArgs)

			for _, addr := range expectedArgs.Addresses {
				state, bytes := account.GetState([]byte(KycInfoKey + addr.ETHAddress().String()))
				if state {
					dbInfo := &KycInfo{}
					err := json.Unmarshal(bytes, dbInfo)
					assert.Nil(t, err)
					assert.Equal(t, NotVerified, dbInfo.KycFlag)
				}
			}
		}
	}
}

func generateReturnData(t *testing.T, s any) []byte {
	marshal, err := json.Marshal(s)
	assert.Nil(t, err)
	return marshal
}

func TestEstimateGas(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.KycVerifyContractAddr).ETHAddress()

	data := hexutil.Bytes(generateRunData(t, SubmitMethod, &SubmitArgs{
		KycInfos: []*KycInfo{
			{
				User:    *types.NewAddressByStr(admin1),
				KycAddr: *types.NewAddressByStr(admin1),
				KycFlag: Verified,
				Expires: -1,
			},
		},
	}))
	gas, err := cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, KycSubmitGas, gas)

	data = hexutil.Bytes(generateRunData(t, RemoveMethod, &RemoveArgs{Addresses: []*types.Address{types.NewAddressByStr(admin1)}}))
	gas, err = cm.EstimateGas(&types.CallArgs{
		From: &from,
		To:   &to,
		Data: &data,
	})
	assert.Nil(t, err)
	assert.Equal(t, KycRemoveGas, gas)

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

func TestKycVerification_CheckAndUpdateState(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	cm.CheckAndUpdateState(100, stateLedger)
}

func TestKycVerification_ParseErrorArgs(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	truedata, err := cm.gabi.Pack(SubmitMethod, []byte(""))
	assert.Nil(t, err)
	testcases := []struct {
		method   string
		data     []byte
		Expected string
	}{
		{
			method:   "Submit",
			data:     []byte{1},
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: msg data length is not improperly formatted: %q - Bytes: %+v", []byte{1}, []byte{1}).Error(),
		},
		{
			method:   "Submit",
			data:     []byte{1, 2, 3, 4},
			Expected: fmt.Errorf("abi: attempting to unmarshall an empty string while arguments are expected").Error(),
		},
		{
			method:   "Submit",
			data:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: improperly formatted output: %q - Bytes: %+v", []byte{5, 6, 7, 8}, []byte{5, 6, 7, 8}).Error(),
		},
		{
			method:   "test method",
			data:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36},
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: could not locate named method: %s", "test method").Error(),
		},
		{
			method:   "test method",
			data:     truedata,
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: could not locate named method: %s", "test method").Error(),
		},
	}

	for _, test := range testcases {
		args := &SubmitArgs{}
		err := cm.ParseArgs(&vm.Message{
			Data: test.data,
		}, test.method, args)
		assert.Equal(t, test.Expected, err.Error())
	}
}

func TestGetArgsForSubmit(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	submitArgs := &SubmitArgs{KycInfos: []*KycInfo{
		{
			User:    *types.NewAddressByStr(admin3),
			KycAddr: *types.NewAddressByStr(admin1),
			KycFlag: 1,
			Expires: -1,
		},
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
	assert.Equal(t, submitArgs.KycInfos[0].User.String(), actualArgs.KycInfos[0].User.String())
}

func TestGetArgsForRemove(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	removeArgs := &RemoveArgs{Addresses: []*types.Address{
		types.NewAddressByStr(admin1),
		types.NewAddressByStr(admin2),
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

	assert.Equal(t, removeArgs.Addresses[0].String(), actualArgs.Addresses[0].String())

}

func TestKycVerification_GetErrArgs(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
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

func TestAddAndRemoveKycService(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	kycservices := []*KycService{
		{
			*types.NewAddressByStr(admin1),
		},
		{
			*types.NewAddressByStr(admin2),
		},
	}

	testcases := []struct {
		method ModifyType
		data   []*KycService
		err    error
	}{
		{
			method: AddKycService,
			data:   kycservices,
			err:    nil,
		},
		{
			method: RemoveKycService,
			data:   kycservices,
			err:    nil,
		},
		{
			method: RemoveKycService,
			data:   kycservices,
			err:    fmt.Errorf("ACCESS ERROR: remove kyc services from an empty list"),
		},
	}
	for _, test := range testcases {
		err := AddAndRemoveKycService(stateLedger, test.method, test.data)
		assert.Equal(t, test.err, err)
	}

}

func TestGetKycServices(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	// case 1 empty list
	services, err := GetKycServices(stateLedger)
	assert.Nil(t, err)
	assert.True(t, len(services) == 0)

	// case 2 not nil list
	kycservices := []*KycService{
		{
			*types.NewAddressByStr(admin1),
		},
		{
			*types.NewAddressByStr(admin2),
		},
	}
	err = SetKycService(stateLedger, kycservices)
	assert.Nil(t, err)
	services, err = GetKycServices(stateLedger)
	assert.Nil(t, err)
	assert.True(t, len(services) == 2)
}

func TestSetKycService(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	kycservices := []*KycService{
		{
			*types.NewAddressByStr(admin1),
		},
		{
			*types.NewAddressByStr(admin2),
		},
	}
	err = SetKycService(stateLedger, kycservices)
	assert.Nil(t, err)
}

func TestKycVerification_SaveLog(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	cm := NewKycVerification(&common.SystemContractConfig{
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

func TestKycVerification_checkErrSubmitInfo(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	testcases := []struct {
		from    *types.Address
		kycinfo *KycInfo
		err     error
	}{
		{
			from: types.NewAddressByStr(admin1),
			kycinfo: &KycInfo{
				User:    *types.NewAddressByStr(admin1),
				KycAddr: *types.NewAddressByStr(admin2),
				KycFlag: 1,
				Expires: -1,
			},
			err: ErrCheckSubmitInfo,
		},
		{
			from:    types.NewAddressByStr(admin1),
			kycinfo: nil,
			err:     ErrCheckSubmitInfo,
		},
		{
			from: types.NewAddressByStr(admin1),
			kycinfo: &KycInfo{
				User:    *types.NewAddressByStr(admin1),
				KycAddr: *types.NewAddressByStr(admin1),
				KycFlag: 1,
				Expires: 1,
			},
			err: ErrCheckSubmitInfo,
		},
		{
			from: types.NewAddressByStr(admin1),
			kycinfo: &KycInfo{
				User:    *types.NewAddressByStr(admin1),
				KycAddr: *types.NewAddressByStr(admin1),
				KycFlag: 3,
				Expires: -1,
			},
			err: nil,
		},
	}

	for _, test := range testcases {
		address := test.from.ETHAddress()
		err := cm.checkSubmitInfo(&address, test.kycinfo)
		assert.Equal(t, test.err, err)
	}

}

func TestVerify(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	// test fail to get state
	testcase := struct {
		needApprove *types.Address
		expected    bool
		err         error
	}{
		needApprove: types.NewAddressByStr(admin2),
		expected:    false,
		err:         fmt.Errorf("ACCESS ERROR: Verify: fail by GetState"),
	}
	verify, err := Verify(stateLedger, testcase.needApprove)
	assert.Equal(t, testcase.err, err)
	assert.Equal(t, testcase.expected, verify)

	// test others
	admins := []string{admin1}
	InitKycServicesAndKycInfos(stateLedger, admins, admins)
	testcases := []struct {
		needApprove *types.Address
		submitAddr  *types.Address
		submitArgs  *SubmitArgs
		expected    bool
		err         error
	}{
		{
			needApprove: types.NewAddressByStr(admin2),
			submitAddr:  types.NewAddressByStr(admin1),
			submitArgs: &SubmitArgs{KycInfos: []*KycInfo{
				{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin1),
					KycFlag: 1,
					Expires: -1,
				},
			}},
			expected: true,
			err:      nil,
		},
		{
			needApprove: types.NewAddressByStr(admin2),
			submitAddr:  types.NewAddressByStr(admin1),
			submitArgs: &SubmitArgs{KycInfos: []*KycInfo{
				{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin1),
					KycFlag: 1,
					Expires: time.Now().Unix() + 1000,
				},
			}},
			expected: true,
			err:      nil,
		},
		{
			needApprove: types.NewAddressByStr(admin2),
			submitAddr:  types.NewAddressByStr(admin1),
			submitArgs: &SubmitArgs{KycInfos: []*KycInfo{
				{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin1),
					KycFlag: 0,
					Expires: -1,
				},
			}},
			expected: false,
			err:      fmt.Errorf("ACCESS ERROR: Verify: fail by checking kyc info"),
		},
		{
			needApprove: types.NewAddressByStr(admin2),
			submitAddr:  types.NewAddressByStr(admin1),
			submitArgs: &SubmitArgs{KycInfos: []*KycInfo{
				{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin1),
					KycFlag: 2,
					Expires: time.Now().Unix() + 10000,
				},
			}},
			expected: false,
			err:      fmt.Errorf("ACCESS ERROR: Verify: fail by checking kyc info"),
		},
	}

	for _, test := range testcases {
		submit := test.submitAddr.ETHAddress()
		cm.Reset(stateLedger)
		_, err := cm.Submit(&submit, test.submitArgs)
		assert.Nil(t, err)
		verify, err := Verify(stateLedger, test.needApprove)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, verify)
	}
}

func TestKycVerification_ErrorSubmit(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	cm.Reset(stateLedger)
	admins := []string{admin1}
	InitKycServicesAndKycInfos(stateLedger, admins, admins)

	testcases := []struct {
		from ethcommon.Address
		args *SubmitArgs
		err  error
	}{
		{
			from: types.NewAddressByStr(admin2).ETHAddress(),
			args: &SubmitArgs{
				KycInfos: []*KycInfo{{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin2),
					KycFlag: 0,
					Expires: 0,
				}},
			},
			err: fmt.Errorf("ACCESS ERROR: Submit: fail by checking kyc services"),
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &SubmitArgs{
				KycInfos: []*KycInfo{{
					User:    *types.NewAddressByStr(admin2),
					KycAddr: *types.NewAddressByStr(admin2),
					KycFlag: 0,
					Expires: 0,
				}},
			},
			err: ErrCheckSubmitInfo,
		},
	}

	for _, test := range testcases {
		_, err := cm.Submit(&test.from, test.args)
		assert.Equal(t, test.err, err)
	}
}

func TestKycVerification_ErrorRemove(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	cm.Reset(stateLedger)
	admins := []string{admin1}
	InitKycServicesAndKycInfos(stateLedger, admins, admins)

	testcases := []struct {
		from ethcommon.Address
		args *RemoveArgs
		err  error
	}{
		{
			from: types.NewAddressByStr(admin2).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []*types.Address{types.NewAddressByStr(admin1)},
			},
			err: fmt.Errorf("ACCESS ERROR: Remove: check kyc service fail"),
		},
		{
			from: types.NewAddressByStr(admin1).ETHAddress(),
			args: &RemoveArgs{
				Addresses: []*types.Address{types.NewAddressByStr(admin3)},
			},
			err: nil,
		},
	}

	for _, test := range testcases {
		_, err := cm.Remove(&test.from, test.args)
		assert.Equal(t, test.err, err)
	}
}

func TestCheckInServices(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	state := CheckInServices(account, admin1)
	assert.Equal(t, false, state)

	err = InitKycServicesAndKycInfos(stateLedger, nil, nil)
	assert.Nil(t, err)
	state = CheckInServices(account, admin1)
	assert.Equal(t, false, state)
}

func TestInitKycServicesAndKycInfos(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "kyc_verification"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.KycVerifyContractAddr), ledger.NewChanger())

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	richAccounts := []string{
		admin1,
	}
	adminAccounts := []string{
		admin1,
		admin2,
		admin3,
	}
	kycServices := []string{
		admin4,
	}
	err = InitKycServicesAndKycInfos(stateLedger, append(richAccounts, append(adminAccounts, kycServices...)...), kycServices)
	assert.Nil(t, err)
}

func TestKycVerification_getErrorArgs(t *testing.T) {
	cm := NewKycVerification(&common.SystemContractConfig{
		Logger: logrus.New(),
	})

	testcases := []struct {
		method   string
		data     *vm.Message
		Expected string
	}{
		{
			method: "Submit",
			data: &vm.Message{
				Data: []byte{83, 44, 81, 214, 0},
			},
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: improperly formatted output: %q - Bytes: %+v", []byte{0}, []byte{0}).Error(),
		},
		{
			method: "Remove",
			data: &vm.Message{
				Data: []byte{42, 198, 250, 166, 0},
			},
			Expected: fmt.Errorf("ACCESS ERROR: ParseArgs: improperly formatted output: %q - Bytes: %+v", []byte{0}, []byte{0}).Error(),
		},
	}

	for _, test := range testcases {
		getArgs, err := cm.getArgs(test.data)
		assert.Nil(t, getArgs)
		assert.Equal(t, test.Expected, err.Error())
	}
}
