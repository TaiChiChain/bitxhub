package genesis

import (
	"encoding/json"

	common2 "github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func initializeGenesisConfig(genesis *repo.GenesisConfig, lg ledger.StateLedger) error {
	account := lg.GetOrCreateAccount(types.NewAddressByStr(common.ZeroAddress))

	genesisCfg, err := json.Marshal(genesis)
	if err != nil {
		return err
	}
	account.SetState([]byte("genesis_cfg"), genesisCfg)
	return nil
}

// Initialize initialize block
func Initialize(genesis *repo.GenesisConfig, lg *ledger.Ledger) error {
	dummyRootHash := common2.Hash{}
	lg.StateLedger.PrepareBlock(types.NewHash(dummyRootHash[:]), 1)

	if err := initializeGenesisConfig(genesis, lg.StateLedger); err != nil {
		return err
	}

	err := system.InitGenesisData(genesis, lg.StateLedger)
	if err != nil {
		return err
	}
	lg.StateLedger.Finalise()

	stateRoot, err := lg.StateLedger.Commit()
	if err != nil {
		return err
	}

	block := &types.Block{
		Header: &types.BlockHeader{
			Number:          1,
			StateRoot:       stateRoot,
			TxRoot:          &types.Hash{},
			ReceiptRoot:     &types.Hash{},
			ParentHash:      &types.Hash{},
			Timestamp:       genesis.Timestamp,
			Epoch:           genesis.EpochInfo.Epoch,
			Bloom:           new(types.Bloom),
			GasPrice:        int64(genesis.EpochInfo.FinanceParams.StartGasPrice),
			ProposerAccount: common.ZeroAddress,
		},
		Transactions: []*types.Transaction{},
	}
	blockData := &ledger.BlockData{
		Block: block,
	}
	block.Header.CalculateHash()
	lg.PersistBlockData(blockData)

	return nil
}

// GetGenesisConfig retrieves the genesis configuration from the given ledger.
func GetGenesisConfig(lg *ledger.Ledger) (*repo.GenesisConfig, error) {
	account := lg.StateLedger.GetAccount(types.NewAddressByStr(common.ZeroAddress))
	if account == nil {
		return nil, nil
	}

	state, bytes := account.GetState([]byte("genesis_cfg"))
	if !state {
		return nil, nil
	}

	genesis := &repo.GenesisConfig{}
	err := json.Unmarshal(bytes, genesis)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
