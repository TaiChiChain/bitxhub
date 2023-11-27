package repo

import (
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/fileutil"
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
}

type TimedGenBlock struct {
	NoTxBatchTimeout Duration `mapstructure:"no_tx_batch_timeout" toml:"no_tx_batch_timeout"`
}

type TxPool struct {
	PoolSize              uint64   `mapstructure:"pool_size" toml:"pool_size"`
	ToleranceTime         Duration `mapstructure:"tolerance_time" toml:"tolerance_time"`
	ToleranceRemoveTime   Duration `mapstructure:"tolerance_remove_time" toml:"tolerance_remove_time"`
	CleanEmptyAccountTime Duration `mapstructure:"clean_empty_account_time" toml:"clean_empty_account_time"`
	ToleranceNonceGap     uint64   `mapstructure:"tolerance_nonce_gap" toml:"tolerance_nonce_gap"`
}

type TxCache struct {
	SetSize    int      `mapstructure:"set_size" toml:"set_size"`
	SetTimeout Duration `mapstructure:"set_timeout" toml:"set_timeout"`
}

type RBFT struct {
	EnableMultiPipes          bool        `mapstructure:"enable_multi_pipes" toml:"enable_multi_pipes"`
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
	BatchTimeout     Duration `mapstructure:"batch_timeout" toml:"batch_timeout"`
	CheckpointPeriod uint64   `mapstructure:"checkpoint_period" toml:"checkpoint_period"`
}

func DefaultConsensusConfig() *ConsensusConfig {
	if testNetConsensusConfigBuilder, ok := TestNetConsensusConfigBuilderMap[BuildNet]; ok {
		return testNetConsensusConfigBuilder()
	}

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
			PoolSize:              50000,
			ToleranceTime:         Duration(5 * time.Minute),
			ToleranceRemoveTime:   Duration(15 * time.Minute),
			CleanEmptyAccountTime: Duration(10 * time.Minute),
			ToleranceNonceGap:     1000,
		},
		TxCache: TxCache{
			SetSize:    50,
			SetTimeout: Duration(100 * time.Millisecond),
		},
		Rbft: RBFT{
			EnableMultiPipes:          false,
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
			BatchTimeout:     Duration(500 * time.Millisecond),
			CheckpointPeriod: 10,
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
			if err := readConfigFromFile(cfgPath, cfg); err != nil {
				return nil, err
			}
		}
		return cfg, nil
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load network config")
	}
	return cfg, nil
}
