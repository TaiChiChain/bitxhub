package genesis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitialize(t *testing.T) {
	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)

	genesisConfig := repo.DefaultGenesisConfig()
	err = Initialize(genesisConfig, lg)
	assert.Nil(t, err)
}

func TestGetGenesisConfig(t *testing.T) {
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}
	// test get account is nil
	stateLedger.EXPECT().GetAccount(gomock.Any()).Return(nil).Times(1)
	accountGenesisConfig, err := GetGenesisConfig(mockLedger)
	assert.Nil(t, err)
	// test get state is nil
	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.ZeroAddress))
	stateLedger.EXPECT().GetAccount(gomock.Any()).Return(account).AnyTimes()
	accountGenesisConfig, err = GetGenesisConfig(mockLedger)
	assert.Nil(t, err)
	// test get genesis config success
	genesisConfig := repo.DefaultGenesisConfig()
	genesisCfg, err := json.Marshal(genesisConfig)
	assert.Nil(t, err)
	account.SetState(genesisConfigKey, genesisCfg)
	accountGenesisConfig, err = GetGenesisConfig(mockLedger)
	assert.Nil(t, err)
	assert.NotNil(t, accountGenesisConfig.ChainID)

	// test get genesis config fail
	account.SetState(genesisConfigKey, []byte{})
	accountGenesisConfig, err = GetGenesisConfig(mockLedger)
	assert.NotNil(t, err)
}
