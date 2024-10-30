package adaptor

import (
	"context"

	kittypes "github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	"github.com/bcds/go-hpc-dagbft/common/config"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/events"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/bcds/go-hpc-dagbft/protocol/layer"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Ready struct {
	Txs                    []*kittypes.Transaction
	Height                 uint64
	RecvConsensusTimestamp int64
	Timestamp              int64
	ProposerNodeID         uint64
	ExecutedCh             chan<- *events.ExecutedEvent
	CommitState            *types.CommitState
	EpochChangedCh         chan<- struct{}
}

type DagBFTAdaptor struct {
	narwhalConfig config.Configs
	ledgerConfig  *LedgerConfig
	Chain         *BlockChain

	logger logrus.FieldLogger

	ctx context.Context

	crashCh chan bool
}

func NewAdaptor(precheck precheck.PreCheck, cnf *common.Config, networkConfig *NetworkConfig, readyC chan *Ready, dagConfig config.DAGConfigs, ctx context.Context, closeCh chan bool) (*DagBFTAdaptor, error) {
	host, _ := networkConfig.LocalPrimary.Unpack()
	narwhalConfig := config.Configs{
		PrimaryNode: host,
		WorkerNodes: lo.MapValues(networkConfig.LocalWorkers, func(value Pid, key types.Host) bool { return true }),
		DAGConfigs:  dagConfig,
	}

	bc, err := newBlockchain(precheck, cnf, narwhalConfig, readyC, closeCh)
	if err != nil {
		return nil, err
	}

	d := &DagBFTAdaptor{
		narwhalConfig: config.Configs{
			PrimaryNode: networkConfig.LocalPrimary.First,
			WorkerNodes: lo.MapValues(networkConfig.LocalWorkers, func(value Pid, key types.Host) bool { return true }),
			DAGConfigs:  dagConfig,
		},
		ledgerConfig: &LedgerConfig{
			useBls:             cnf.Repo.Config.Consensus.UseBlsKey,
			ChainState:         cnf.ChainState,
			GetBlockHeaderFunc: cnf.GetBlockHeaderFunc,
		},
		Chain:  bc,
		logger: cnf.Logger,

		ctx: ctx,
	}
	return d, nil
}

func (d *DagBFTAdaptor) Start() {
	go d.Chain.listenChainEvent()
}

func (d *DagBFTAdaptor) GetLedger() layer.Ledger {
	return d.Chain
}

func (d *DagBFTAdaptor) GetCryptoVerifier() protocol.ValidatorVerifier {
	return d.Chain.crypto.GetValidatorVerifier()
}
