package common

import (
	"github.com/axiomesh/axiom-kit/types"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/network"
)

type Config struct {
	Peers               []*Node
	StartEpochChangeNum uint64
	SnapPersistedEpoch  uint64
	LatestPersistEpoch  uint64
	EpochChangeSendCh   chan<- containers.Tuple[types.QuorumCheckpoint, error, bool]
}

type Option func(*Config)

// ============= snap sync config ================

func WithPeers(peers []*Node) Option {
	return func(config *Config) {
		config.Peers = peers
	}
}

func WithStartEpochChangeNum(num uint64) Option {
	return func(config *Config) {
		config.StartEpochChangeNum = num
	}
}

func WithLatestPersistEpoch(latestPersistEpoch uint64) Option {
	return func(config *Config) {
		config.LatestPersistEpoch = latestPersistEpoch
	}
}

func WithSnapCurrentEpoch(epoch uint64) Option {
	return func(config *Config) {
		config.SnapPersistedEpoch = epoch
	}
}

func WithEpochChangeSendCh(ch chan<- containers.Tuple[types.QuorumCheckpoint, error, bool]) Option {
	return func(config *Config) {
		config.EpochChangeSendCh = ch
	}
}

type ModeConfig struct {
	Logger        logrus.FieldLogger
	Network       network.Network
	ConsensusType string
}

type ModeOption func(config *ModeConfig)

func WithLogger(logger logrus.FieldLogger) ModeOption {
	return func(config *ModeConfig) {
		config.Logger = logger
	}
}

func WithNetwork(network network.Network) ModeOption {
	return func(config *ModeConfig) {
		config.Network = network
	}
}

func WithConsensusType(consensusType string) ModeOption {
	return func(config *ModeConfig) {
		config.ConsensusType = consensusType
	}
}
