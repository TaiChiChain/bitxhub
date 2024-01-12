package common

import (
	"context"

	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Peers               []string
	StartEpochChangeNum uint64
	SnapPersistedEpoch  uint64
	LatestPersistEpoch  uint64
}

type Option func(*Config)

// ============= snap sync config ================

func WithPeers(peers []string) Option {
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

type ModeConfig struct {
	Logger  logrus.FieldLogger
	Network network.Network
	Ctx     context.Context
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

func WithContext(ctx context.Context) ModeOption {
	return func(config *ModeConfig) {
		config.Ctx = ctx
	}
}
