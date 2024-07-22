package dagbft

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics/prometheus"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/adaptor"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagconfig "github.com/bcds/go-hpc-dagbft/common/config"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/bcds/go-hpc-dagbft/protocol/observer"
	"github.com/samber/lo"
)

type Config struct {
	dagconfig.DAGConfigs
	// Logger is the logger used to record logger in DagBFT.
	Logger *common.Logger

	// MetricsProv is the metrics Provider used to generate metrics instance.
	MetricsProv observer.MetricProvider

	NetworkConfig *adaptor.NetworkConfig

	ChainState *chainstate.ChainState

	GetBlockHeaderFunc func(height uint64) (*types.BlockHeader, error)
	GetAccountBalance  func(address string) *big.Int
	GetAccountNonce    func(address *types.Address) uint64
}

// todo(lrx): common metrics for consensus
// todo(lrx): add primary and workers
func GenerateDagBftConfig(config *common.Config) (*Config, error) {
	defaultConfig := &Config{DAGConfigs: dagconfig.DefaultDAGConfigs}
	defaultConfig.Logger = &common.Logger{FieldLogger: config.Logger}
	var provider observer.MetricProvider
	provider = &prometheus.Provider{
		Name: "axiom_ledger",
	}
	provider = provider.SubProvider("consensus").SubProvider("dagbft")
	defaultConfig.MetricsProv = provider
	defaultConfig.GetBlockHeaderFunc = config.GetBlockHeaderFunc
	defaultConfig.GetAccountBalance = config.GetAccountBalance
	defaultConfig.GetAccountNonce = config.GetAccountNonce
	defaultConfig.ChainState = config.ChainState
	// todo(lrx): network config should be configurable
	defaultConfig.NetworkConfig = adaptor.DefaultNetworkConfig()
	defaultConfig.NetworkConfig.LocalPrimary = containers.Pack2(config.ChainState.SelfNodeInfo.Primary, config.ChainState.SelfNodeInfo.P2PID)

	defaultConfig.NetworkConfig.LocalWorkers = lo.SliceToMap(config.ChainState.SelfNodeInfo.Workers, func(n dagtypes.Host) (dagtypes.Host, adaptor.Pid) {
		return n, config.ChainState.SelfNodeInfo.P2PID
	})

	readConfig := config.Repo.ConsensusConfig.Dagbft

	defaultConfig.DAGConfigs = readDagbftConfig(readConfig)

	return defaultConfig, nil
}

func readDagbftConfig(readDagbftConfig repo.Dagbft) dagconfig.DAGConfigs {
	defaultConfig := dagconfig.DefaultDAGConfigs
	defaultConfig.AllowInconsistentExecuteResult = readDagbftConfig.AllowInconsistentExecuteResult
	defaultConfig.CheckpointSequencePeriod = readDagbftConfig.CheckpointSequencePeriod
	defaultConfig.MaxHeaderBatchesSize = readDagbftConfig.MaxHeaderBatchesSize
	defaultConfig.MinHeaderBatchesSize = readDagbftConfig.MinHeaderBatchesSize
	defaultConfig.MaxHeaderDelay = readDagbftConfig.MaxHeaderDelay.ToDuration()
	defaultConfig.MinHeaderDelay = readDagbftConfig.MinHeaderDelay.ToDuration()
	defaultConfig.HeaderResentDelay = readDagbftConfig.HeaderResentDelay.ToDuration()
	defaultConfig.GcRoundDepth = readDagbftConfig.GcRoundDepth
	defaultConfig.ExpiredRoundDepth = readDagbftConfig.ExpiredRoundDepth
	defaultConfig.SealingBatchLimit = readDagbftConfig.SealingBatchLimit
	defaultConfig.MaxBatchCount = readDagbftConfig.MaxBatchCount
	defaultConfig.MaxBatchSize = readDagbftConfig.MaxBatchSize
	defaultConfig.MaxBatchDelay = readDagbftConfig.MaxBatchDelay.ToDuration()
	defaultConfig.SyncRetryDelay = readDagbftConfig.SyncRetryDelay.ToDuration()
	defaultConfig.SyncRetryNodes = readDagbftConfig.SyncRetryNodes
	defaultConfig.HeartbeatsTimeout = readDagbftConfig.HeartbeatsTimeout.ToDuration()
	defaultConfig.FetchingOutputLimit = readDagbftConfig.FetchingOutputLimit
	defaultConfig.ExecutingOutputLimit = readDagbftConfig.ExecutingOutputLimit
	defaultConfig.RemoveBufferSize = readDagbftConfig.RemoveBufferSize
	defaultConfig.CheckpointHeightPeriod = readDagbftConfig.CheckpointHeightPeriod
	defaultConfig.MaxCheckpointWaitingDuration = readDagbftConfig.MaxCheckpointWaitingDuration.ToDuration()
	defaultConfig.LeaderReputationThreshold = readDagbftConfig.LeaderReputationThreshold
	defaultConfig.LeaderReputationSchedulePeriod = readDagbftConfig.LeaderReputationSchedulePeriod
	defaultConfig.NewCommittedCertificateTimeout = readDagbftConfig.NewCommittedCertificateTimeout.ToDuration()
	defaultConfig.NewCertifiedCheckpointTimeout = readDagbftConfig.NewCertifiedCheckpointTimeout.ToDuration()

	defaultConfig.FeatureConfigs.EnforceIncreasingTimestamp = readDagbftConfig.FeatureConfigs.EnforceIncreasingTimestamp
	defaultConfig.FeatureConfigs.EnforceIncreasingCommitRound = readDagbftConfig.FeatureConfigs.EnforceIncreasingCommitRound
	defaultConfig.FeatureConfigs.EnforceIncreasingStateHeight = readDagbftConfig.FeatureConfigs.EnforceIncreasingStateHeight
	defaultConfig.FeatureConfigs.EnableFastSyncRecovery = readDagbftConfig.FeatureConfigs.EnableFastSyncRecovery
	defaultConfig.FeatureConfigs.EnableLeaderReputation = readDagbftConfig.FeatureConfigs.EnableLeaderReputation
	defaultConfig.FeatureConfigs.AllowInconsistentExecuteResult = readDagbftConfig.FeatureConfigs.AllowInconsistentExecuteResult
	return defaultConfig
}
