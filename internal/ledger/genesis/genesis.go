package genesis

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func initializeGenesisConfig(genesis *repo.Genesis, lg ledger.StateLedger) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.ZeroAddress))

	genesisCfg, err := json.Marshal(genesis)
	if err != nil {
		return err
	}
	account.SetState([]byte("genesis_cfg"), genesisCfg)
	return nil
}

// Initialize initialize block
func Initialize(genesis *repo.Genesis, lg *ledger.Ledger) error {
	lg.StateLedger.PrepareBlock(nil, 1)

	if err := initializeGenesisConfig(genesis, lg.StateLedger); err != nil {
		return err
	}

	balance, _ := new(big.Int).SetString(genesis.Balance, 10)
	for _, addr := range genesis.Accounts {
		lg.StateLedger.SetBalance(types.NewAddressByStr(addr), balance)
	}
	err := system.InitGenesisData(genesis, lg.StateLedger)
	if err != nil {
		return err
	}

	accounts, stateRoot := lg.StateLedger.FlushDirtyData()

	block := &types.Block{
		BlockHeader: &types.BlockHeader{
			Number:          1,
			StateRoot:       stateRoot,
			TxRoot:          &types.Hash{},
			ReceiptRoot:     &types.Hash{},
			ParentHash:      &types.Hash{},
			Timestamp:       time.Now().Unix(),
			GasPrice:        int64(genesis.GasPrice),
			Epoch:           genesis.EpochInfo.Epoch,
			Bloom:           new(types.Bloom),
			ProposerAccount: common.ZeroAddress,
		},
		Transactions: []*types.Transaction{},
	}
	block.BlockHash = block.Hash()
	blockData := &ledger.BlockData{
		Block:    block,
		Accounts: accounts,
	}

	lg.PersistBlockData(blockData)

	return nil
}
