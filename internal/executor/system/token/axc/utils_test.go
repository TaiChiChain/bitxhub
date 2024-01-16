package axc

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
)

const (
	testMiningAddr     = "0xE7aEe2a87E7d5129Bd2fBf12fAfF7534Ed146666"
	testCommunityAddr  = "0xf16F8B02df2Dd7c4043C41F3f1EBB17f15358888"
	admin1             = "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"
	defaultTotalSupply = "2000000000000000000000000000"
)

type mockLedger struct {
	*mock_ledger.MockStateLedger
	accountDb map[string]ledger.IAccount
}

type mockAccount struct {
	*ledger.SimpleAccount
	balance *big.Int
	stateDb map[string][]byte
}

func newMockAccount(addr *types.Address) *mockAccount {
	return &mockAccount{
		SimpleAccount: ledger.NewMockAccount(1, addr),
		balance:       big.NewInt(0),
		stateDb:       make(map[string][]byte),
	}
}

func (ma *mockAccount) GetBalance() *big.Int {
	if ma == nil {
		return big.NewInt(0)
	}
	return ma.balance
}

func (ma *mockAccount) SetBalance(b *big.Int) {
	ma.balance = b
}

func (ma *mockAccount) AddBalance(b *big.Int) {
	ma.balance.Add(ma.balance, b)
}

func (ma *mockAccount) SubBalance(b *big.Int) {
	ma.balance.Sub(ma.balance, b)
}

func (ma *mockAccount) SetState(key []byte, value []byte) {
	if ma.stateDb == nil {
		ma.stateDb = make(map[string][]byte)
	}
	ma.stateDb[string(key)] = value
}

func (ma *mockAccount) GetState(key []byte) (bool, []byte) {
	if ma.stateDb == nil {
		return false, nil
	}
	v, ok := ma.stateDb[string(key)]
	return ok, v
}

func newMockMinLedger(t *testing.T) *mockLedger {
	mockLg := &mockLedger{
		accountDb: make(map[string]ledger.IAccount),
	}
	ctrl := gomock.NewController(t)
	mockLg.MockStateLedger = mock_ledger.NewMockStateLedger(ctrl)

	mockLg.EXPECT().GetOrCreateAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
		if mockLg.accountDb[address.String()] == nil {
			mockLg.accountDb[address.String()] = newMockAccount(address)
		}
		return mockLg.accountDb[address.String()]
	}).AnyTimes()

	mockLg.EXPECT().GetState(gomock.Any(), gomock.Any()).DoAndReturn(func(address *types.Address, bytes []byte) (bool, []byte) {
		if mockLg.accountDb[address.String()] == nil {
			return false, nil
		}
		return mockLg.accountDb[address.String()].GetState(bytes)
	}).AnyTimes()

	mockLg.EXPECT().GetAccount(gomock.Any()).DoAndReturn(func(address *types.Address) ledger.IAccount {
		return mockLg.accountDb[address.String()]
	}).AnyTimes()

	mockLg.EXPECT().GetBalance(gomock.Any()).DoAndReturn(func(address *types.Address) *big.Int {
		if mockLg.accountDb[address.String()] == nil {
			return big.NewInt(0)
		}
		return mockLg.accountDb[address.String()].GetBalance()
	}).AnyTimes()
	//
	//contractAccount := newMockAccount(types.NewAddressByStr(common.AXMManagerContractAddr))
	//contractAccount.
	//mockLg.accountDb[common.AXMManagerContractAddr] = contractAccount
	return mockLg
}

func mockAxcManager(t *testing.T) *Manager {
	logger := log.NewWithModule("axc")

	mockLg := newMockMinLedger(t)
	genesisConf := repo.DefaultGenesisConfig(false)
	conf, err := GenerateConfig(genesisConf)
	require.Nil(t, err)
	err = Init(mockLg, conf)
	require.Nil(t, err)

	randUser := types.NewAddressByStr(admin1).ETHAddress()
	logs := make([]common.Log, 0)
	am := &Manager{logger: logger}
	am.SetContext(&common.VMContext{
		StateLedger: mockLg,
		CurrentUser: &randUser,
		CurrentLogs: &logs,
	})
	return am
}
