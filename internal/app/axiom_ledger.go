package app

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"syscall"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/ethereum/go-ethereum/common/fdlimit"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/api/jsonrpc"
	"github.com/axiomesh/axiom-ledger/internal/block_sync"
	"github.com/axiomesh/axiom-ledger/internal/consensus"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/executor"
	devexecutor "github.com/axiomesh/axiom-ledger/internal/executor/dev"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/genesis"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/profile"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type AxiomLedger struct {
	Ctx           context.Context
	Cancel        context.CancelFunc
	Repo          *repo.Repo
	logger        logrus.FieldLogger
	ViewLedger    *ledger.Ledger
	BlockExecutor executor.Executor
	Consensus     consensus.Consensus
	Network       network.Network
	Sync          block_sync.Sync
	Monitor       *profile.Monitor
	Pprof         *profile.Pprof
	LoggerWrapper *loggers.LoggerWrapper
	Jsonrpc       *jsonrpc.ChainBrokerService
}

func NewAxiomLedger(rep *repo.Repo, ctx context.Context, cancel context.CancelFunc) (*AxiomLedger, error) {
	axm, err := NewAxiomLedgerWithoutConsensus(rep, ctx, cancel)
	if err != nil {
		return nil, fmt.Errorf("generate axiom-ledger without consensus failed: %w", err)
	}

	chainMeta := axm.ViewLedger.ChainLedger.GetChainMeta()

	if !rep.ReadonlyMode {
		axm.Consensus, err = consensus.New(
			rep.Config.Consensus.Type,
			common.WithConfig(rep.RepoRoot, rep.ConsensusConfig),
			common.WithSelfAccountAddress(rep.AccountAddress),
			common.WithGenesisEpochInfo(rep.Config.Genesis.EpochInfo.Clone()),
			common.WithConsensusType(rep.Config.Consensus.Type),
			common.WithPrivKey(rep.AccountKey),
			common.WithNetwork(axm.Network),
			common.WithLogger(loggers.Logger(loggers.Consensus)),
			common.WithApplied(chainMeta.Height),
			common.WithDigest(chainMeta.BlockHash.String()),
			common.WithGenesisDigest(axm.ViewLedger.ChainLedger.GetBlockHash(1).String()),
			common.WithGetChainMetaFunc(axm.ViewLedger.ChainLedger.GetChainMeta),
			common.WithGetBlockFunc(axm.ViewLedger.ChainLedger.GetBlock),
			common.WithGetAccountBalanceFunc(func(address *types.Address) *big.Int {
				return axm.ViewLedger.NewView().StateLedger.GetBalance(address)
			}),
			common.WithGetAccountNonceFunc(func(address *types.Address) uint64 {
				return axm.ViewLedger.NewView().StateLedger.GetNonce(address)
			}),
			common.WithGetEpochInfoFromEpochMgrContractFunc(func(epoch uint64) (*rbft.EpochInfo, error) {
				return base.GetEpochInfo(axm.ViewLedger.NewView().StateLedger, epoch)
			}),
			common.WithGetCurrentEpochInfoFromEpochMgrContractFunc(func() (*rbft.EpochInfo, error) {
				return base.GetCurrentEpochInfo(axm.ViewLedger.NewView().StateLedger)
			}),
			common.WithEVMConfig(axm.Repo.Config.Executor.EVM),
			common.WithBlockSync(axm.Sync),
		)
		if err != nil {
			return nil, fmt.Errorf("initialize consensus failed: %w", err)
		}
	}

	return axm, nil
}

func PrepareAxiomLedger(rep *repo.Repo) error {
	types.InitEIP155Signer(big.NewInt(int64(rep.Config.Genesis.ChainID)))

	if err := storagemgr.Initialize(rep.Config.Storage.KvType, rep.Config.Storage.KvCacheSize); err != nil {
		return fmt.Errorf("storagemgr initialize: %w", err)
	}
	if err := raiseUlimit(rep.Config.Ulimit); err != nil {
		return fmt.Errorf("raise ulimit: %w", err)
	}
	return nil
}

func NewAxiomLedgerWithoutConsensus(rep *repo.Repo, ctx context.Context, cancel context.CancelFunc) (*AxiomLedger, error) {
	if err := PrepareAxiomLedger(rep); err != nil {
		return nil, err
	}

	logger := loggers.Logger(loggers.App)

	// 0. load ledger
	rwLdg, err := ledger.NewLedger(rep)
	if err != nil {
		return nil, fmt.Errorf("create RW ledger: %w", err)
	}

	if rwLdg.ChainLedger.GetChainMeta().Height == 0 {
		if err := genesis.Initialize(&rep.Config.Genesis, rwLdg); err != nil {
			return nil, err
		}
		logger.WithFields(logrus.Fields{
			"genesis block hash": rwLdg.ChainLedger.GetChainMeta().BlockHash,
		}).Info("Initialize genesis")
	}

	var txExec executor.Executor
	if rep.Config.Executor.Type == repo.ExecTypeDev {
		txExec, err = devexecutor.New(loggers.Logger(loggers.Executor))
	} else {
		txExec, err = executor.New(rep, rwLdg)
	}
	if err != nil {
		return nil, fmt.Errorf("create BlockExecutor: %w", err)
	}

	var net network.Network
	if !rep.ReadonlyMode {
		net, err = network.New(rep, loggers.Logger(loggers.P2P), rwLdg.NewView())
		if err != nil {
			return nil, fmt.Errorf("create peer manager: %w", err)
		}
	}

	vl := rwLdg.NewView()
	sync, err := block_sync.NewBlockSync(loggers.Logger(loggers.BlockSync), vl.ChainLedger.GetBlock, net, rep.Config.Sync)
	if err != nil {
		return nil, fmt.Errorf("create block sync: %w", err)
	}

	axm := &AxiomLedger{
		Ctx:           ctx,
		Cancel:        cancel,
		Repo:          rep,
		logger:        logger,
		ViewLedger:    vl,
		BlockExecutor: txExec,
		Network:       net,
		Sync:          sync,
	}
	// read current epoch info from ledger
	axm.Repo.EpochInfo, err = base.GetCurrentEpochInfo(axm.ViewLedger.StateLedger)
	if err != nil {
		return nil, err
	}
	return axm, nil
}

func (axm *AxiomLedger) Start() error {
	if repo.SupportMultiNode[axm.Repo.Config.Consensus.Type] && !axm.Repo.ReadonlyMode {
		if err := axm.Network.Start(); err != nil {
			return fmt.Errorf("peer manager start: %w", err)
		}

		if err := axm.Sync.Start(); err != nil {
			return fmt.Errorf("block sync start: %w", err)
		}
	}

	if !axm.Repo.ReadonlyMode {
		if err := axm.Consensus.Start(); err != nil {
			return fmt.Errorf("consensus start: %w", err)
		}
	}

	if err := axm.BlockExecutor.Start(); err != nil {
		return fmt.Errorf("block executor start: %w", err)
	}

	axm.start()

	axm.printLogo()

	return nil
}

func (axm *AxiomLedger) Stop() error {
	if axm.Repo.Config.Consensus.Type != repo.ConsensusTypeSolo && !axm.Repo.ReadonlyMode {
		if err := axm.Network.Stop(); err != nil {
			return fmt.Errorf("network stop: %w", err)
		}
	}
	if !axm.Repo.ReadonlyMode {
		axm.Consensus.Stop()
	}
	if err := axm.BlockExecutor.Stop(); err != nil {
		return fmt.Errorf("block executor stop: %w", err)
	}
	axm.Cancel()

	axm.logger.Infof("%s stopped", repo.AppName)

	return nil
}

func (axm *AxiomLedger) printLogo() {
	if !axm.Repo.ReadonlyMode {
		for {
			time.Sleep(100 * time.Millisecond)
			err := axm.Consensus.Ready()
			if err == nil {
				break
			}
		}
	}

	axm.logger.WithFields(logrus.Fields{
		"consensus_type": axm.Repo.Config.Consensus.Type,
	}).Info("Consensus is ready")
	fig := figure.NewFigure(repo.AppName, "slant", true)
	axm.logger.WithField(log.OnlyWriteMsgWithoutFormatterField, nil).Infof(`
=========================================================================================
%s
=========================================================================================
`, fig.String())
}

func raiseUlimit(limitNew uint64) error {
	_, err := fdlimit.Raise(limitNew)
	if err != nil {
		return fmt.Errorf("set limit failed: %w", err)
	}

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return fmt.Errorf("getrlimit error: %w", err)
	}

	if limit.Cur != limitNew && limit.Cur != limit.Max {
		return errors.New("failed to raise ulimit")
	}

	return nil
}
