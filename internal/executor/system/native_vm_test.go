package system

import (
	"math/big"
	"strings"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/governance"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axc"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/token/axm"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	admin1 = "0x1210000000000000000000000000000000000000"
	admin2 = "0x1220000000000000000000000000000000000000"
	admin3 = "0x1230000000000000000000000000000000000000"
	admin4 = "0x1240000000000000000000000000000000000000"
)

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
		args := &vm.StatefulArgs{
			StateDB: &ledger.EvmStateDBAdaptor{
				StateLedger: stateLedger,
			},
			Height: big.NewInt(2),
			From:   testcase.Message.From,
			To:     testcase.Message.To,
		}
		_, err := nvm.Run(testcase.Message.Data, args)
		if testcase.IsExpectedErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestNativeVM_GetContractInstance(t *testing.T) {
	nvm := New()
	cfg := &common.SystemContractConfig{}

	testCases := []struct {
		ContractAddr *types.Address
		expectType   interface{}
	}{
		{
			ContractAddr: types.NewAddressByStr(common.GovernanceContractAddr),
			expectType:   governance.NewGov(cfg),
		},
		{
			ContractAddr: types.NewAddressByStr(common.AXCContractAddr),
			expectType:   axc.New(cfg),
		},
		{
			ContractAddr: types.NewAddressByStr(common.AXMContractAddr),
			expectType:   axm.New(cfg),
		},
		{
			ContractAddr: types.NewAddressByStr(common.WhiteListContractAddr),
			expectType:   access.NewWhiteList(cfg),
		},
		{
			ContractAddr: types.NewAddressByStr(common.EpochManagerContractAddr),
			expectType:   base.NewEpochManager(cfg),
		},
	}

	for _, testCase := range testCases {
		contractInstance := nvm.GetContractInstance(testCase.ContractAddr)
		// 使用 assert.IsType 来判断类型是否一致
		assert.IsType(t, testCase.expectType, contractInstance)
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
	gas := nvm.RequiredGas([]byte{1})
	assert.Equal(t, uint64(21016), gas)
}
