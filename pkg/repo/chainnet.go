package repo

import (
	"time"

	rbft "github.com/axiomesh/axiom-bft"
)

const (
	AriesTestnetName = "aries"
)

var (
	TestNetConfigBuilderMap = map[string]func() *Config{
		AriesTestnetName: AriesConfig,
	}

	TestNetConsensusConfigBuilderMap = map[string]func() *ConsensusConfig{
		AriesTestnetName: AriesConsensusConfig,
	}

	TestNetGenesisConfigBuilderMap = map[string]func() *GenesisConfig{
		AriesTestnetName: AriesGenesisConfig,
	}
)

func AriesConfig() *Config {
	return &Config{
		Ulimit: 65535,
		Access: Access{
			EnableWhiteList: false,
		},
		Port: Port{
			JsonRpc:   8881,
			WebSocket: 9991,
			P2P:       4001,
			PProf:     53121,
			Monitor:   40011,
		},
		JsonRPC: JsonRPC{
			GasCap:     300000000,
			EVMTimeout: Duration(5 * time.Second),
			ReadLimiter: JLimiter{
				Interval: 50,
				Quantum:  500,
				Capacity: 10000,
				Enable:   true,
			},
			WriteLimiter: JLimiter{
				Interval: 50,
				Quantum:  500,
				Capacity: 10000,
				Enable:   true,
			},
			RejectTxsIfConsensusAbnormal: false,
		},
		P2P: P2P{
			Security:    P2PSecurityTLS,
			SendTimeout: Duration(5 * time.Second),
			ReadTimeout: Duration(5 * time.Second),
			Pipe: P2PPipe{
				ReceiveMsgCacheSize: 1024,
				BroadcastType:       P2PPipeBroadcastGossip,
				SimpleBroadcast: P2PPipeSimpleBroadcast{
					WorkerCacheSize:        1024,
					WorkerConcurrencyLimit: 20,
				},
				Gossipsub: P2PPipeGossipsub{
					SubBufferSize:          1024,
					PeerOutboundBufferSize: 1024,
					ValidateBufferSize:     1024,
					SeenMessagesTTL:        Duration(120 * time.Second),
					EnableMetrics:          false,
				},
				UnicastReadTimeout:       Duration(5 * time.Second),
				UnicastSendRetryNumber:   5,
				UnicastSendRetryBaseTime: Duration(100 * time.Millisecond),
				FindPeerTimeout:          Duration(10 * time.Second),
				ConnectTimeout:           Duration(1 * time.Second),
			},
		},
		Sync: Sync{
			RequesterRetryTimeout: Duration(5 * time.Second),
			TimeoutCountLimit:     uint64(10),
			ConcurrencyLimit:      1000,
		},
		Consensus: Consensus{
			Type:        ConsensusTypeRbft,
			StorageType: ConsensusStorageTypeMinifile,
		},
		Storage: Storage{
			KvType: KVStorageTypeLeveldb,
		},
		Ledger: Ledger{
			ChainLedgerCacheSize:           100,
			StateLedgerCacheMegabytesLimit: 128,
			StateLedgerAccountCacheSize:    1024,
		},
		Executor: Executor{
			Type:            ExecTypeNative,
			DisableRollback: false,
			EVM: EVM{
				DisableMaxCodeSizeLimit: false,
			},
		},
		PProf: PProf{
			Enable:   true,
			PType:    PprofTypeHTTP,
			Mode:     PprofModeMem,
			Duration: Duration(30 * time.Second),
		},
		Monitor: Monitor{
			Enable: true,
		},
		Log: Log{
			Level:            "info",
			Filename:         "axiom-ledger",
			ReportCaller:     false,
			EnableCompress:   false,
			EnableColor:      true,
			DisableTimestamp: false,
			MaxAge:           30,
			MaxSize:          128,
			RotationTime:     Duration(24 * time.Hour),
			Module: LogModule{
				P2P:        "info",
				Consensus:  "debug",
				Executor:   "info",
				Governance: "info",
				API:        "error",
				CoreAPI:    "info",
				Storage:    "info",
				Profile:    "info",
				Finance:    "error",
				BlockSync:  "info",
				APP:        "info",
				Access:     "info",
				TxPool:     "info",
				Epoch:      "info",
			},
		},
	}
}

func AriesConsensusConfig() *ConsensusConfig {
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
				NullRequest:      Duration(9 * time.Second),
				Request:          Duration(6 * time.Second),
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

func AriesGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		ChainID: 23411,
		Balance: "1000000000000000000000000000",
		Admins: []*Admin{
			{
				Address: "0xecFE18Dc453CCdF96f1b9b58ccb4db3c6115A1D0",
				Weight:  1,
				Name:    "S2luZw==",
			},
			{
				Address: "0x13f30647b99Edeb8CF3725eCd1Eaf545D9283335",
				Weight:  1,
				Name:    "UmVk",
			},
			{
				Address: "0x6cdB717de826334faD8FB0ce0547Bac0230ba5a4",
				Weight:  1,
				Name:    "QXBwbGU=",
			},
			{
				Address: "0xAc7DD5009788f2CB14db8dCd6728d94Cbd4d705e",
				Weight:  1,
				Name:    "Q2F0",
			},
		},
		Accounts: []string{},
		EpochInfo: &rbft.EpochInfo{
			Version:     1,
			Epoch:       1,
			EpochPeriod: 100000000,
			StartBlock:  1,
			P2PBootstrapNodeAddresses: []string{
				"/ip4/127.0.0.1/tcp/4001/p2p/16Uiu2HAm9VjBKpMJyzXUzLCd4wWigkPD9HHUEmg628pPxGhkyoVg",
				"/ip4/127.0.0.1/tcp/4002/p2p/16Uiu2HAmQ2EnGWAeRLNB8ZfiPAQxwofWbCpA7sfQymTGbbe64z4G",
				"/ip4/127.0.0.1/tcp/4003/p2p/16Uiu2HAkwWBiECscWVK3mp3xTUpGdx5qkBs91RbhT2psQAZHkx5i",
				"/ip4/127.0.0.1/tcp/4004/p2p/16Uiu2HAm3ikUE3LjJeatMMgDuV2cAG9da8ZJJFLA8nBy6qcN1MMg",
			},
			ConsensusParams: rbft.ConsensusParams{
				ValidatorElectionType:         rbft.ValidatorElectionTypeWRF,
				ProposerElectionType:          rbft.ProposerElectionTypeAbnormalRotation,
				CheckpointPeriod:              10,
				HighWatermarkCheckpointPeriod: 4,
				MaxValidatorNum:               20,
				BlockMaxTxNum:                 500,
				EnableTimedGenEmptyBlock:      false,
				NotActiveWeight:               1,
				AbnormalNodeExcludeView:       1,
				AgainProposeIntervalBlockInValidatorsNumPercentage: 30,
				ContinuousNullRequestToleranceNumber:               3,
				ReBroadcastToleranceNumber:                         2,
			},
			CandidateSet: []rbft.NodeInfo{},
			ValidatorSet: []rbft.NodeInfo{
				{
					ID:                   1,
					AccountAddress:       "0xecFE18Dc453CCdF96f1b9b58ccb4db3c6115A1D0",
					P2PNodeID:            "16Uiu2HAm9VjBKpMJyzXUzLCd4wWigkPD9HHUEmg628pPxGhkyoVg",
					ConsensusVotingPower: 1000,
				},
				{
					ID:                   2,
					AccountAddress:       "0x13f30647b99Edeb8CF3725eCd1Eaf545D9283335",
					P2PNodeID:            "16Uiu2HAmQ2EnGWAeRLNB8ZfiPAQxwofWbCpA7sfQymTGbbe64z4G",
					ConsensusVotingPower: 1000,
				},
				{
					ID:                   3,
					AccountAddress:       "0x6cdB717de826334faD8FB0ce0547Bac0230ba5a4",
					P2PNodeID:            "16Uiu2HAkwWBiECscWVK3mp3xTUpGdx5qkBs91RbhT2psQAZHkx5i",
					ConsensusVotingPower: 1000,
				},
				{
					ID:                   4,
					AccountAddress:       "0xAc7DD5009788f2CB14db8dCd6728d94Cbd4d705e",
					P2PNodeID:            "16Uiu2HAm3ikUE3LjJeatMMgDuV2cAG9da8ZJJFLA8nBy6qcN1MMg",
					ConsensusVotingPower: 1000,
				},
			},
			FinanceParams: rbft.FinanceParams{
				GasLimit:               0x5f5e100,
				StartGasPriceAvailable: true,
				StartGasPrice:          5000000000000,
				MaxGasPrice:            10000000000000,
				MinGasPrice:            1000000000000,
				GasChangeRateValue:     1250,
				GasChangeRateDecimals:  4,
			},
			MiscParams: rbft.MiscParams{
				TxMaxSize: DefaultTxMaxSize,
			},
		},
	}
}
