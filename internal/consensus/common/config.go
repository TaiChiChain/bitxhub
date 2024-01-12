package common

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/sirupsen/logrus"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

type Config struct {
	RepoRoot                                    string
	EVMConfig                                   repo.EVM
	Config                                      *repo.ConsensusConfig
	Logger                                      logrus.FieldLogger
	ConsensusType                               string
	ConsensusStorageType                        string
	PrivKey                                     *ecdsa.PrivateKey
	GenesisEpochInfo                            *rbft.EpochInfo
	Network                                     network.Network
	BlockSync                                   common.Sync
	TxPool                                      txpool.TxPool[types.Transaction, *types.Transaction]
	Applied                                     uint64
	Digest                                      string
	GenesisDigest                               string
	GetCurrentEpochInfoFromEpochMgrContractFunc func() (*rbft.EpochInfo, error)
	GetEpochInfoFromEpochMgrContractFunc        func(epoch uint64) (*rbft.EpochInfo, error)
	GetChainMetaFunc                            func() *types.ChainMeta
	GetBlockFunc                                func(height uint64) (*types.Block, error)
	GetAccountBalance                           func(address *types.Address) *big.Int
	GetAccountNonce                             func(address *types.Address) uint64
	EpochStore                                  storage.Storage
}

type Option func(*Config)

func WithConfig(repoRoot string, cfg *repo.ConsensusConfig) Option {
	return func(config *Config) {
		config.RepoRoot = repoRoot
		config.Config = cfg
	}
}

func WithTxPool(tp txpool.TxPool[types.Transaction, *types.Transaction]) Option {
	return func(config *Config) {
		config.TxPool = tp
	}
}

func WithGenesisEpochInfo(genesisEpochInfo *rbft.EpochInfo) Option {
	return func(config *Config) {
		config.GenesisEpochInfo = genesisEpochInfo
	}
}

func WithConsensusType(typ string) Option {
	return func(config *Config) {
		config.ConsensusType = typ
	}
}

func WithConsensusStorageType(consensusStorageType string) Option {
	return func(config *Config) {
		config.ConsensusStorageType = consensusStorageType
	}
}

func WithNetwork(net network.Network) Option {
	return func(config *Config) {
		config.Network = net
	}
}

func WithBlockSync(blockSync common.Sync) Option {
	return func(config *Config) {
		config.BlockSync = blockSync
	}
}

func WithPrivKey(privKey *ecdsa.PrivateKey) Option {
	return func(config *Config) {
		config.PrivKey = privKey
	}
}

func WithLogger(logger logrus.FieldLogger) Option {
	return func(config *Config) {
		config.Logger = logger
	}
}

func WithApplied(height uint64) Option {
	return func(config *Config) {
		config.Applied = height
	}
}

func WithDigest(digest string) Option {
	return func(config *Config) {
		config.Digest = digest
	}
}

func WithGenesisDigest(digest string) Option {
	return func(config *Config) {
		config.GenesisDigest = digest
	}
}

func WithGetChainMetaFunc(f func() *types.ChainMeta) Option {
	return func(config *Config) {
		config.GetChainMetaFunc = f
	}
}

func WithGetBlockFunc(f func(height uint64) (*types.Block, error)) Option {
	return func(config *Config) {
		config.GetBlockFunc = f
	}
}

func WithGetAccountBalanceFunc(f func(address *types.Address) *big.Int) Option {
	return func(config *Config) {
		config.GetAccountBalance = f
	}
}

func WithGetAccountNonceFunc(f func(address *types.Address) uint64) Option {
	return func(config *Config) {
		config.GetAccountNonce = f
	}
}

func WithGetEpochInfoFromEpochMgrContractFunc(f func(epoch uint64) (*rbft.EpochInfo, error)) Option {
	return func(config *Config) {
		config.GetEpochInfoFromEpochMgrContractFunc = f
	}
}

func WithGetCurrentEpochInfoFromEpochMgrContractFunc(f func() (*rbft.EpochInfo, error)) Option {
	return func(config *Config) {
		config.GetCurrentEpochInfoFromEpochMgrContractFunc = f
	}
}

func WithEVMConfig(c repo.EVM) Option {
	return func(config *Config) {
		config.EVMConfig = c
	}
}

func WithEpochStore(epochStore storage.Storage) Option {
	return func(config *Config) {
		config.EpochStore = epochStore
	}
}

func checkConfig(config *Config) error {
	if config.Logger == nil {
		return errors.New("logger is nil")
	}

	return nil
}

func GenerateConfig(opts ...Option) (*Config, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}

	if err := checkConfig(config); err != nil {
		return nil, fmt.Errorf("create consensus: %w", err)
	}

	return config, nil
}

type Logger struct {
	logrus.FieldLogger
}

// Trace implements rbft.Logger.
func (lg *Logger) Trace(name string, stage string, content any) {
	lg.Info(name, stage, content)
}

func (lg *Logger) Critical(v ...any) {
	lg.Fatal(v...)
}

func (lg *Logger) Criticalf(format string, v ...any) {
	lg.Fatalf(format, v...)
}

func (lg *Logger) Notice(v ...any) {
	lg.Info(v...)
}

func (lg *Logger) Noticef(format string, v ...any) {
	lg.Infof(format, v...)
}

func NeedChangeEpoch(height uint64, epochInfo *rbft.EpochInfo) bool {
	return height == (epochInfo.StartBlock + epochInfo.EpochPeriod - 1)
}

func GetQuorum(consensusType string, N int) uint64 {
	switch consensusType {
	case repo.ConsensusTypeRbft:
		f := (N - 1) / 3
		return uint64((N + f + 2) / 2)
	case repo.ConsensusTypeSolo, repo.ConsensusTypeSoloDev:
		fallthrough
	default:
		return 0
	}
}
