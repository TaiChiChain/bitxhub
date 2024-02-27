package system

import (
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axm"
	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var systemContractAddrs = []string{
	common.GovernanceContractAddr,
	common.EpochManagerContractAddr,
	common.WhiteListContractAddr,
	common.AXMContractAddr,
	common.AXCContractAddr,
}

var notSystemContractAddrs = []string{
	"0x1000000000000000000000000000000000000000",
	"0x0340000000000000000000000000000000000000",
	"0x0200000000000000000000000000000000000000",
	"0xffddd00000000000000000000000000000000000",
}

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

func TestContract_IsSystemContract(t *testing.T) {
	nvm := New()
	for _, addr := range systemContractAddrs {
		ok := nvm.IsSystemContract(types.NewAddressByStr(addr))
		assert.True(t, ok)
	}

	for _, addr := range notSystemContractAddrs {
		ok := nvm.IsSystemContract(types.NewAddressByStr(addr))
		assert.False(t, ok)
	}

	// test nil address
	ok := nvm.IsSystemContract(nil)
	assert.False(t, ok)

	// test empty address
	ok = nvm.IsSystemContract(&types.Address{})
	assert.False(t, ok)
}

func TestContractInitGenesisData(t *testing.T) {
	t.Parallel()
	t.Run("init genesis data success", func(t *testing.T) {
		mockCtl := gomock.NewController(t)
		chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
		stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
		mockLedger := &ledger.Ledger{
			ChainLedger: chainLedger,
			StateLedger: stateLedger,
		}

		genesis := repo.DefaultGenesisConfig(false)

		account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))
		axmAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXMContractAddr))
		axcAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXCContractAddr))
		stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
			if address.String() == common.AXMContractAddr {
				return axmAccount
			}
			if address.String() == common.AXCContractAddr {
				return axcAccount
			}
			return account
		}).AnyTimes()

		stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

		err := InitGenesisData(genesis, mockLedger.StateLedger)
		assert.Nil(t, err)
	})

	t.Run("init epoch contract failed", func(t *testing.T) {
		mockCtl := gomock.NewController(t)
		chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
		stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
		mockLedger := &ledger.Ledger{
			ChainLedger: chainLedger,
			StateLedger: stateLedger,
		}

		genesis := repo.DefaultGenesisConfig(false)
		// set wrong address for validatorSet
		genesis.EpochInfo.ValidatorSet[0].AccountAddress = "wrong address"

		account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))
		axmAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXMContractAddr))
		axcAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXCContractAddr))
		stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
			if address.String() == common.AXMContractAddr {
				return axmAccount
			}
			if address.String() == common.AXCContractAddr {
				return axcAccount
			}
			return account
		}).AnyTimes()

		stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

		err := InitGenesisData(genesis, mockLedger.StateLedger)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid account address")
	})

	t.Run("init token contract failed", func(t *testing.T) {
		mockCtl := gomock.NewController(t)
		chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
		stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
		mockLedger := &ledger.Ledger{
			ChainLedger: chainLedger,
			StateLedger: stateLedger,
		}

		genesis := repo.DefaultGenesisConfig(false)
		// set wrong balance for token manager
		genesis.Accounts[0].Balance = "wrong balance"

		account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))
		axmAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXMContractAddr))
		axcAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXCContractAddr))
		stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
			if address.String() == common.AXMContractAddr {
				return axmAccount
			}
			if address.String() == common.AXCContractAddr {
				return axcAccount
			}
			return account
		}).AnyTimes()

		stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

		err := InitGenesisData(genesis, mockLedger.StateLedger)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid balance")

		genesis = repo.DefaultGenesisConfig(false)
		// decrease total supply
		genesis.Axm.TotalSupply = "10"
		err = InitGenesisData(genesis, mockLedger.StateLedger)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), axm.ErrTotalSupply.Error())
	})

}

func TestWhiteListContractInitGenesisData(t *testing.T) {
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	genesis := repo.DefaultGenesisConfig(false)

	//WhiteListContractAddr
	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.WhiteListContractAddr))
	axmAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXMContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
		if address.String() == common.AXMContractAddr {
			return axmAccount
		}
		return account
	}).AnyTimes()

	axcAccount := ledger.NewMockAccount(2, types.NewAddressByStr(common.AXCContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
		if address.String() == common.AXCContractAddr {
			return axcAccount
		}
		return account
	}).AnyTimes()

	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	err := InitGenesisData(genesis, mockLedger.StateLedger)
	assert.Nil(t, err)
}

func TestContractRun(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.GovernanceContractAddr).ETHAddress()
	data := hexutil.Bytes(generateVoteData(t, 1, governance.Pass))
	getIDData := hexutil.Bytes(generateGetLatestProposalIDData(t))

	nvm := New()

	testcases := []struct {
		Message       *core.Message
		IsExpectedErr bool
	}{
		{
			Message: &core.Message{
				From: from,
				To:   nil,
				Data: data,
			},
			IsExpectedErr: true,
		},
		{
			Message: &core.Message{
				From: from,
				To:   &to,
				Data: data,
			},
			IsExpectedErr: true,
		},
		{
			Message: &core.Message{
				From: from,
				To:   &to,
				Data: hexutil.Bytes("error data"),
			},
			IsExpectedErr: true,
		},
		{
			Message: &core.Message{
				From: from,
				To:   &to,
				Data: crypto.Keccak256([]byte("vote(uint64,uint8)")),
			},
			IsExpectedErr: true,
		},
		{
			Message: &core.Message{
				From: from,
				To:   &to,
				Data: getIDData,
			},
			IsExpectedErr: false,
		},
		{
			Message: &core.Message{
				From: from,
				To:   &from,
				Data: data,
			},
			IsExpectedErr: true,
		},
	}
	for _, testcase := range testcases {
		nvm.Reset(2, stateLedger)
		input, err := nvm.EnhancedInput(&testcase.Message.From, testcase.Message.To, testcase.Message.Data)
		assert.Nil(t, err)
		_, err = nvm.Run(input)
		if testcase.IsExpectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func generateVoteData(t *testing.T, proposalID uint64, voteResult governance.VoteResult) []byte {
	gabi, err := abi.JSON(strings.NewReader(governanceABI))
	assert.Nil(t, err)

	data, err := gabi.Pack(governance.VoteMethod, proposalID, voteResult)
	assert.Nil(t, err)

	return data
}

func generateGetLatestProposalIDData(t *testing.T) []byte {
	gabi, err := abi.JSON(strings.NewReader(governanceABI))
	assert.Nil(t, err)

	data, err := gabi.Pack(governance.GetLatestProposalIDMethod)
	assert.Nil(t, err)

	return data
}

func TestNativeVM_RequiredGas(t *testing.T) {
	nvm := New()
	input, err := nvm.EnhancedInput(nil, nil, []byte{1})
	assert.Nil(t, err)
	gas := nvm.RequiredGas(input)
	assert.Equal(t, uint64(21740), gas)
}

func TestNativeVM_EnhancedInput(t *testing.T) {
	nvm := New()
	data := hexutil.Bytes(generateVoteData(t, 1, governance.Pass))
	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.GovernanceContractAddr).ETHAddress()
	testCases := []struct {
		from          *ethcommon.Address
		to            *ethcommon.Address
		originalInput []byte
		err           error
	}{
		{
			from:          nil,
			to:            nil,
			originalInput: nil,
			err:           nil,
		},
		{
			from:          &from,
			to:            &to,
			originalInput: data,
			err:           nil,
		},
	}

	for _, testCase := range testCases {
		_, err := nvm.EnhancedInput(testCase.from, testCase.to, testCase.originalInput)
		assert.Equal(t, testCase.err, err)
	}
}

func TestNativeVM_DecodeEnhancedInput(t *testing.T) {
	nvm := New()
	originalData := generateVoteData(t, 1, governance.Pass)
	input, err := nvm.EnhancedInput(nil, nil, originalData)
	assert.Nil(t, err)
	testCases := []struct {
		originalData []byte
		input        []byte
		nilError     bool
	}{
		{
			originalData: nil,
			input:        nil,
			nilError:     false,
		},
		{
			originalData: originalData,
			input:        input,
			nilError:     true,
		},
	}

	for _, testCase := range testCases {
		data, err := nvm.DecodeEnhancedInput(testCase.input)
		assert.True(t, err == nil == testCase.nilError)
		if testCase.nilError {
			assert.Equal(t, originalData, data)
		}
	}
}

func TestRunAxiomNativeVM(t *testing.T) {
	nvm := New()
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()

	from := types.NewAddressByStr(admin1).ETHAddress()
	to := types.NewAddressByStr(common.GovernanceContractAddr).ETHAddress()
	data := hexutil.Bytes(generateVoteData(t, 1, governance.Pass))
	input, err := nvm.EnhancedInput(&from, &to, data)
	assert.Nil(t, err)
	nativeVM := RunAxiomNativeVM(nvm, 2, stateLedger, input)
	assert.NotNil(t, nativeVM.Err)
}
