package token

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-ledger/pkg/repo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
)

const (
	admin1             = "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013"
	admin2             = "0x79a1215469FaB6f9c63c1816b45183AD3624bE34"
	defaultTotalSupply = "24000000000000000000000000000"
)

const (
	nameMethod         = "name"
	symbolMethod       = "symbol"
	totalSupplyMethod  = "totalSupply"
	decimalsMethod     = "decimals"
	balanceOfMethod    = "balanceOf"
	transferMethod     = "transfer"
	approveMethod      = "approve"
	allowanceMethod    = "allowance"
	transferFromMethod = "transferFrom"
	mintMethod         = "mint"
	burnMethod         = "burn"
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
	//contractAccount := newMockAccount(types.NewAddressByStr(common.TokenManagerContractAddr))
	//contractAccount.
	//mockLg.accountDb[common.TokenManagerContractAddr] = contractAccount
	return mockLg
}

func mockAxmManager(t *testing.T) *AxmManager {
	logger := log.NewWithModule("token")

	mockLg := newMockMinLedger(t)
	genesisConf := repo.DefaultGenesisConfig(false)
	conf, err := GenerateGenesisTokenConfig(genesisConf)
	require.Nil(t, err)
	err = InitAxmTokenManager(mockLg, conf)
	require.Nil(t, err)
	contractAccount := mockLg.GetOrCreateAccount(types.NewAddressByStr(common.TokenManagerContractAddr))

	am := &AxmManager{logger: logger, account: contractAccount, stateLedger: mockLg}
	return am
}

func generateRunData(t *testing.T, method string, args ...any) []byte {
	var (
		inputs []byte
		err    error
	)
	m := axmManagerABI.Methods[method]
	if len(args) == 0 || len(m.Inputs) == 0 {
		inputs, err = m.Inputs.Pack()
	} else {
		inputs, err = m.Inputs.Pack(args...)
	}
	require.Nil(t, err)
	inputs = append(m.ID, inputs...)

	return inputs
}

func wrapperReturnData(t *testing.T, method string, data ...any) []byte {
	var wd []byte
	var err error
	m := axmManagerABI.Methods[method]
	if len(data) == 0 || len(m.Outputs) == 0 {
		wd, err = m.Outputs.Pack()
	} else {
		wd, err = m.Outputs.Pack(data...)
	}
	require.Nil(t, err)
	return wd
}
