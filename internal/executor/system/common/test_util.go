package common

import (
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type TestNVM struct {
	t           testing.TB
	Rep         *repo.Repo
	Ledger      *ledger.Ledger
	StateLedger ledger.StateLedger
}

func NewTestNVM(t testing.TB) *TestNVM {
	rep := repo.MockRepo(t)
	lg, err := ledger.NewMemory(rep)
	assert.Nil(t, err)
	return &TestNVM{
		t:           t,
		Rep:         rep,
		Ledger:      lg,
		StateLedger: lg.StateLedger,
	}
}

func (nvm *TestNVM) GenesisInit(contracts ...SystemContract) {
	for _, contract := range contracts {
		contract.SetContext(NewVMContextByExecutor(nvm.StateLedger).DisableRecordLogToLedger())
		err := contract.GenesisInit(nvm.Rep.GenesisConfig)
		assert.Nil(nvm.t, err)
	}
}

type TestNVMRunOption func(ctx *VMContext)

func TestNVMRunOptionCallFromSystem() TestNVMRunOption {
	return func(ctx *VMContext) {
		ctx.CallFromSystem = true
	}
}

func (nvm *TestNVM) RunSingleTX(contract SystemContract, from ethcommon.Address, executor func() error, opts ...TestNVMRunOption) {
	snapshot := nvm.Ledger.StateLedger.Snapshot()
	stateLedger := &LogsCollectorStateLedger{StateLedger: nvm.Ledger.StateLedger, disableRecordLogToLedger: true}
	ctx := &VMContext{
		StateLedger:    stateLedger,
		BlockNumber:    1,
		From:           from,
		CallFromSystem: false,
		CurrentEVM:     nvm.NewEVM(from, stateLedger, nil, nil),
	}
	for _, opt := range opts {
		opt(ctx)
	}
	contract.SetContext(ctx)
	if err := executor(); err != nil {
		nvm.Ledger.StateLedger.RevertToSnapshot(snapshot)
		return
	}
	nvm.Ledger.StateLedger.Finalise()
}

func (nvm *TestNVM) Call(contract SystemContract, from ethcommon.Address, executor func()) {
	snapshot := nvm.Ledger.StateLedger.Snapshot()
	stateLedger := &LogsCollectorStateLedger{StateLedger: nvm.Ledger.StateLedger, disableRecordLogToLedger: true}
	contract.SetContext(&VMContext{
		StateLedger:              stateLedger,
		BlockNumber:              1,
		From:                     from,
		CallFromSystem:           false,
		CurrentEVM:               nvm.NewEVM(from, stateLedger, nil, nil),
		disableRecordLogToLedger: true,
	})
	executor()
	nvm.Ledger.StateLedger.RevertToSnapshot(snapshot)
}

func (nvm *TestNVM) NewEVM(caller ethcommon.Address, stateLedger *LogsCollectorStateLedger, blkCtxSetter func(ctx *vm.BlockContext), txCtxSetter func(ctx *vm.TxContext)) *vm.EVM {
	coinbase := "0xed17543171C1459714cdC6519b58fFcC29A3C3c9"
	blkCtx := &vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash: func(u uint64) ethcommon.Hash {
			header, err := nvm.Ledger.ChainLedger.GetBlockHeader(u)
			if err != nil {
				return ethcommon.Hash{}
			}
			return header.Hash().ETHHash()
		},
		Coinbase:    ethcommon.HexToAddress(coinbase),
		BlockNumber: new(big.Int).SetUint64(1),
		Time:        uint64(time.Now().Unix()),
		Difficulty:  big.NewInt(0x2000),
		BaseFee:     big.NewInt(0),
		GasLimit:    0x2fefd8,
		Random:      &ethcommon.Hash{},
	}
	if blkCtxSetter != nil {
		blkCtxSetter(blkCtx)
	}

	txCtx := &vm.TxContext{
		Origin:   caller,
		GasPrice: big.NewInt(100000000000),
	}
	if txCtxSetter != nil {
		txCtxSetter(txCtx)
	}

	shanghaiTime := uint64(0)
	cancunTime := uint64(0)
	pragueTime := uint64(0)

	return vm.NewEVM(*blkCtx, *txCtx, &ledger.EvmStateDBAdaptor{
		StateLedger: stateLedger,
	}, &params.ChainConfig{
		ChainID:                 big.NewInt(1356),
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		ShanghaiTime:            &shanghaiTime,
		CancunTime:              &cancunTime,
		PragueTime:              &pragueTime,
	}, vm.Config{})
}
