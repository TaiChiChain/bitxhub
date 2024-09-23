package repo

import (
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/fileutil"
	"github.com/axiomesh/axiom-kit/types"
)

const (
	GenerateBatchByTime     = "fifo" // default
	GenerateBatchByGasPrice = "price_priority"
)

type ReceiveMsgLimiter struct {
	Enable bool  `mapstructure:"enable" toml:"enable"`
	Limit  int64 `mapstructure:"limit" toml:"limit"`
	Burst  int64 `mapstructure:"burst" toml:"burst"`
}

type ConsensusConfig struct {
	TimedGenBlock TimedGenBlock     `mapstructure:"timed_gen_block" toml:"timed_gen_block"`
	Limit         ReceiveMsgLimiter `mapstructure:"limit" toml:"limit"`
	TxPool        TxPool            `mapstructure:"tx_pool" toml:"tx_pool"`
	TxCache       TxCache           `mapstructure:"tx_cache" toml:"tx_cache"`
	Rbft          RBFT              `mapstructure:"rbft" toml:"rbft"`
	Solo          Solo              `mapstructure:"solo" toml:"solo"`
	Dagbft        Dagbft            `mapstructure:"dag_bft" toml:"dag_bft"`
}

type TimedGenBlock struct {
	NoTxBatchTimeout Duration `mapstructure:"no_tx_batch_timeout" toml:"no_tx_batch_timeout"`
}

type TxPool struct {
	PoolSize               uint64            `mapstructure:"pool_size" toml:"pool_size"`
	ToleranceTime          Duration          `mapstructure:"tolerance_time" toml:"tolerance_time"`
	ToleranceRemoveTime    Duration          `mapstructure:"tolerance_remove_time" toml:"tolerance_remove_time"`
	CleanEmptyAccountTime  Duration          `mapstructure:"clean_empty_account_time" toml:"clean_empty_account_time"`
	ToleranceNonceGap      uint64            `mapstructure:"tolerance_nonce_gap" toml:"tolerance_nonce_gap"`
	EnableLocalsPersist    bool              `mapstructure:"enable_locals_persist" toml:"enable_locals_persist"`
	RotateTxLocalsInterval Duration          `mapstructure:"rotate_tx_locals_interval" toml:"rotate_tx_locals_interval"`
	PriceLimit             *types.CoinNumber `mapstructure:"price_limit" toml:"price_limit"`
	PriceBump              uint64            `mapstructure:"price_bump" toml:"price_bump"`
	GenerateBatchType      string            `mapstructure:"generate_batch_type" toml:"generate_batch_type"`
}

type TxCache struct {
	SetSize    int      `mapstructure:"set_size" toml:"set_size"`
	SetTimeout Duration `mapstructure:"set_timeout" toml:"set_timeout"`
}

type RBFT struct {
	EnableMetrics             bool        `mapstructure:"enable_metrics" toml:"enable_metrics"`
	CommittedBlockCacheNumber uint64      `mapstructure:"committed_block_cache_number" toml:"committed_block_cache_number"`
	Timeout                   RBFTTimeout `mapstructure:"timeout" toml:"timeout"`
}

type RBFTTimeout struct {
	NullRequest      Duration `mapstructure:"null_request" toml:"null_request"`
	Request          Duration `mapstructure:"request" toml:"request"`
	ResendViewChange Duration `mapstructure:"resend_viewchange" toml:"resend_viewchange"`
	CleanViewChange  Duration `mapstructure:"clean_viewchange" toml:"clean_viewchange"`
	NewView          Duration `mapstructure:"new_view" toml:"new_view"`
	SyncState        Duration `mapstructure:"sync_state" toml:"sync_state"`
	SyncStateRestart Duration `mapstructure:"sync_state_restart" toml:"sync_state_restart"`
	FetchCheckpoint  Duration `mapstructure:"fetch_checkpoint" toml:"fetch_checkpoint"`
	FetchView        Duration `mapstructure:"fetch_view" toml:"fetch_view"`
	BatchTimeout     Duration `mapstructure:"batch_timeout" toml:"batch_timeout"`
}

type Solo struct {
	BatchTimeout Duration `mapstructure:"batch_timeout" toml:"batch_timeout"`
}

type Dagbft struct {
	BatchTimeout      Duration `mapstructure:"batch_timeout" toml:"batch_timeout"`
	ReportBatchResult bool     `mapstructure:"report_batch_result" toml:"report_batch_result"`
	// The maximum number of batch digests included in a header.
	MaxHeaderBatchesSize int `toml:"max_header_batches_size" mapstructure:"max_header_batches_size"`
	// The threshold number of batch digests included in a header which is reached `MinHeaderDelay`.
	MinHeaderBatchesSize int `toml:"min_header_batches_size" mapstructure:"min_header_batches_size"`
	// The maximum delay that the primary should wait between generating two headers, even if
	// other conditions are not satisfied besides having enough parent stakes (e.g. insufficient batch digests).
	MaxHeaderDelay Duration `toml:"max_header_delay" mapstructure:"max_header_delay"`
	// When the delay from last header reaches `MinHeaderDelay`, a new header can be proposed
	// even if batches have not reached `MaxHeaderBatchesSize`.
	MinHeaderDelay Duration `toml:"min_header_delay" mapstructure:"min_header_delay"`
	// The maximum delay that the primary should wait between generating two headers, even if
	// other conditions are not satisfied besides having enough parent stakes.
	HeaderResentDelay Duration `toml:"header_resent_delay" mapstructure:"header_resent_delay"`
	// The depth of the garbage collection (Denominated in number of rounds).
	GcRoundDepth uint64 `toml:"gc_round_depth" mapstructure:"gc_round_depth"`
	// The depth of the batch expiration (Denominated in number of rounds).
	ExpiredRoundDepth uint64 `toml:"expired_round_depth" mapstructure:"expired_round_depth"`
	// Limit max sealing batches in the worker.
	SealingBatchLimit int `toml:"sealing_batch_limit" mapstructure:"sealing_batch_limit"`
	// The preferred batch count. The workers seal a batch of transactions when it reaches this count.
	MaxBatchCount int `toml:"max_batch_count" mapstructure:"max_batch_count"`
	// The preferred batch size. The workers seal a batch of transactions when it reaches this size.
	// Denominated in bytes.
	MaxBatchSize int `toml:"max_batch_size" mapstructure:"max_batch_size"`
	// The delay after which the workers seal a batch of transactions, even if `MaxBatchCount`
	// and `MaxBatchSize` is not reached.
	MaxBatchDelay Duration `toml:"max_batch_delay" mapstructure:"max_batch_delay"`
	// The delay after which the synchronizer retries to send sync batches.
	SyncRetryDelay Duration `toml:"sync_retry_delay" mapstructure:"sync_retry_delay"`
	// Determine with how many nodes to sync when re-trying to send sync-request. These nodes
	// are picked at random from the committee.
	SyncRetryNodes int `toml:"sync_retry_nodes" mapstructure:"sync_retry_nodes"`
	// Worker sends to a watermark synchronize request for each HeartbeatsTimeout.
	HeartbeatsTimeout Duration `toml:"heartbeats_timeout" mapstructure:"heartbeats_timeout"`

	// FetchingOutputLimit defines the output fetching task limits without waiting to finish.
	FetchingOutputLimit int `toml:"fetching_output_limit" mapstructure:"fetching_output_limit"`
	// ExecutingOutputLimit defines the output executing task limits without waiting to finish.
	ExecutingOutputLimit int `toml:"executing_output_limit" mapstructure:"executing_output_limit"`
	// RemoveBufferSize keeps a buffer to prevent removing of recent data for stores.
	RemoveBufferSize int `toml:"remove_buffer_size" mapstructure:"remove_buffer_size"`

	// CheckpointHeightPeriod defines the period for generating checkpoints based on state height.
	CheckpointHeightPeriod uint64 `toml:"checkpoint_height_period" mapstructure:"checkpoint_height_period"`
	// CheckpointSequencePeriod defines the period for generating checkpoints based on commit sequence.
	CheckpointSequencePeriod uint64 `toml:"checkpoint_sequence_period" mapstructure:"checkpoint_sequence_period"`
	// MaxCheckpointWaitingDuration defines the max duration allowed for waiting execution result when receiving checkpoint.
	MaxCheckpointWaitingDuration Duration `toml:"max_checkpoint_waiting_duration" mapstructure:"max_checkpoint_waiting_duration"`

	// LeaderReputationThreshold dictates the threshold (percentage of stake) that is used to calculate the "bad" nodes to be
	// swapped when creating the consensus schedule. The values should be of the range [0 - 33%]. Anything
	// above 33% (f) will not be allowed.
	LeaderReputationThreshold uint64 `toml:"leader_reputation_threshold" mapstructure:"leader_reputation_threshold"`
	// LeaderReputationSchedulePeriod dictates the window of reputation schedule changing.
	LeaderReputationSchedulePeriod uint64 `toml:"leader_reputation_schedule_period" mapstructure:"leader_reputation_schedule_period"`

	// NewCommittedCertificateTimeout defines the timeout of no new certificate committed.
	NewCommittedCertificateTimeout Duration `toml:"new_committed_certificate_timeout" mapstructure:"new_committed_certificate_timeout"`
	// NewCertifiedCheckpointTimeout defines the timeout of no new checkpoint certified.
	NewCertifiedCheckpointTimeout Duration `toml:"new_certified_checkpoint_timeout" mapstructure:"new_certified_checkpoint_timeout"`

	FeatureConfigs `toml:"feature_configs" mapstructure:"feature_configs"`
}

type FeatureConfigs struct {
	// EnforceIncreasingTimestamp enforce promise the timestamp of the header is increasing.
	EnforceIncreasingTimestamp bool `toml:"enforce_increasing_timestamp" mapstructure:"enforce_increasing_timestamp"`
	// EnforceIncreasingCommitRound enforce promise the commit round is increasing.
	EnforceIncreasingCommitRound bool `toml:"enforce_increasing_commit_round" mapstructure:"enforce_increasing_commit_round"`
	// EnforceIncreasingStateHeight enforce promise the state height is increasing.
	EnforceIncreasingStateHeight bool `toml:"enforce_increasing_state_height" mapstructure:"enforce_increasing_state_height"`
	// EnableFastSyncRecovery enable the backward nodes syncing state fast.
	EnableFastSyncRecovery bool `toml:"enable_fast_sync_recovery" mapstructure:"enable_fast_sync_recovery"`
	// EnableLeaderReputation enable choosing leader with reputation calculation.
	EnableLeaderReputation bool `toml:"enable_leader_reputation" mapstructure:"enable_leader_reputation"`
	// AllowInconsistentExecuteResult allow inconsistent execution result and try to recommit.
	AllowInconsistentExecuteResult bool `toml:"allow_inconsistent_execute_result" mapstructure:"allow_inconsistent_execute_result"`
}

func DefaultConsensusConfig() *ConsensusConfig {
	if testNetConsensusConfigBuilder, ok := TestNetConsensusConfigBuilderMap[BuildNet]; ok {
		return testNetConsensusConfigBuilder()
	}
	return defaultConsensusConfig()
}

func defaultConsensusConfig() *ConsensusConfig {
	// nolint
	return &ConsensusConfig{
		TimedGenBlock: TimedGenBlock{
			NoTxBatchTimeout: Duration(2 * time.Second),
		},
		Limit: ReceiveMsgLimiter{
			Enable: false,
			Limit:  10000,
			Burst:  10000,
		},
		TxPool: TxPool{
			PoolSize:               50000,
			ToleranceTime:          Duration(5 * time.Minute),
			ToleranceRemoveTime:    Duration(15 * time.Minute),
			CleanEmptyAccountTime:  Duration(10 * time.Minute),
			RotateTxLocalsInterval: Duration(1 * time.Hour),
			ToleranceNonceGap:      1000,
			EnableLocalsPersist:    true,
			PriceLimit:             GetDefaultMinGasPrice(),
			PriceBump:              10,
			GenerateBatchType:      GenerateBatchByTime,
		},
		TxCache: TxCache{
			SetSize:    50,
			SetTimeout: Duration(100 * time.Millisecond),
		},
		Rbft: RBFT{
			EnableMetrics:             true,
			CommittedBlockCacheNumber: 10,
			Timeout: RBFTTimeout{
				NullRequest:      Duration(3 * time.Second),
				Request:          Duration(2 * time.Second),
				ResendViewChange: Duration(10 * time.Second),
				CleanViewChange:  Duration(60 * time.Second),
				NewView:          Duration(8 * time.Second),
				SyncState:        Duration(1 * time.Second),
				SyncStateRestart: Duration(10 * time.Minute),
				FetchCheckpoint:  Duration(5 * time.Second),
				FetchView:        Duration(1 * time.Second),
				BatchTimeout:     Duration(500 * time.Millisecond),
			},
		},
		Solo: Solo{
			BatchTimeout: Duration(500 * time.Millisecond),
		},
		Dagbft: Dagbft{
			BatchTimeout:         Duration(50 * time.Millisecond),
			ReportBatchResult:    true,
			MinHeaderBatchesSize: 10,
			MaxHeaderBatchesSize: 100,
			MaxHeaderDelay:       Duration(time.Millisecond * 200),
			MinHeaderDelay:       Duration(time.Millisecond * 50),
			HeaderResentDelay:    Duration(time.Minute * 1),
			GcRoundDepth:         50,
			ExpiredRoundDepth:    100,

			SealingBatchLimit: 1000,
			MaxBatchCount:     10,
			MaxBatchSize:      100 * DefaultTxMaxSize,
			MaxBatchDelay:     Duration(time.Millisecond * 50),
			SyncRetryDelay:    Duration(time.Second * 10),
			SyncRetryNodes:    3,
			HeartbeatsTimeout: Duration(time.Second * 1),

			FetchingOutputLimit:  100,
			ExecutingOutputLimit: 100,
			RemoveBufferSize:     5,

			CheckpointHeightPeriod:       1,
			CheckpointSequencePeriod:     100,
			MaxCheckpointWaitingDuration: Duration(time.Second * 10),

			LeaderReputationThreshold:      33,
			LeaderReputationSchedulePeriod: 300,

			NewCommittedCertificateTimeout: Duration(time.Second * 30),
			NewCertifiedCheckpointTimeout:  Duration(time.Second * 60),

			FeatureConfigs: FeatureConfigs{
				EnforceIncreasingTimestamp:     false,
				EnforceIncreasingCommitRound:   false,
				EnforceIncreasingStateHeight:   false,
				EnableFastSyncRecovery:         true,
				EnableLeaderReputation:         true,
				AllowInconsistentExecuteResult: true,
			},
		},
	}
}

func LoadConsensusConfig(repoRoot string) (*ConsensusConfig, error) {
	cfg, err := func() (*ConsensusConfig, error) {
		cfg := DefaultConsensusConfig()
		cfgPath := path.Join(repoRoot, consensusCfgFileName)
		existConfig := fileutil.Exist(cfgPath)
		if !existConfig {
			if err := writeConfigWithEnv(cfgPath, cfg); err != nil {
				return nil, errors.Wrap(err, "failed to build default consensus config")
			}
		} else {
			if err := ReadConfigFromFile(cfgPath, cfg); err != nil {
				return nil, err
			}
		}
		return cfg, nil
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load consensus config")
	}
	return cfg, nil
}
