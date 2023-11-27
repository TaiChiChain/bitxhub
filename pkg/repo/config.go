package repo

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/fileutil"
)

type Duration time.Duration

func (d *Duration) MarshalText() (text []byte, err error) {
	return []byte(time.Duration(*d).String()), nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func StringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(Duration(5)) {
			return data, nil
		}

		d, err := time.ParseDuration(data.(string))
		if err != nil {
			return nil, err
		}
		return Duration(d), nil
	}
}

func (d *Duration) ToDuration() time.Duration {
	return time.Duration(*d)
}

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

type Config struct {
	Ulimit    uint64    `mapstructure:"ulimit" toml:"ulimit"`
	Port      Port      `mapstructure:"port" toml:"port"`
	JsonRPC   JsonRPC   `mapstructure:"jsonrpc" toml:"jsonrpc"`
	P2P       P2P       `mapstructure:"p2p" toml:"p2p"`
	Sync      Sync      `mapstructure:"sync" toml:"sync"`
	Consensus Consensus `mapstructure:"consensus" toml:"consensus"`
	Storage   Storage   `mapstructure:"storage" toml:"storage"`
	Ledger    Ledger    `mapstructure:"ledger" toml:"ledger"`
	Executor  Executor  `mapstructure:"executor" toml:"executor"`
	Genesis   Genesis   `mapstructure:"genesis" toml:"genesis"`
	PProf     PProf     `mapstructure:"pprof" toml:"pprof"`
	Monitor   Monitor   `mapstructure:"monitor" toml:"monitor"`
	Log       Log       `mapstructure:"log" toml:"log"`
	Access    Access    `mapstructure:"access" toml:"access"`
}

type Port struct {
	JsonRpc   int64 `mapstructure:"jsonrpc" toml:"jsonrpc"`
	WebSocket int64 `mapstructure:"websocket" toml:"websocket"`
	P2P       int64 `mapstructure:"p2p" toml:"p2p"`
	PProf     int64 `mapstructure:"pprof" toml:"pprof"`
	Monitor   int64 `mapstructure:"monitor" toml:"monitor"`
}

type JsonRPC struct {
	GasCap                       uint64   `mapstructure:"gas_cap" toml:"gas_cap"`
	EVMTimeout                   Duration `mapstructure:"evm_timeout" toml:"evm_timeout"`
	ReadLimiter                  JLimiter `mapstructure:"read_limiter" toml:"read_limiter"`
	WriteLimiter                 JLimiter `mapstructure:"write_limiter" toml:"write_limiter"`
	RejectTxsIfConsensusAbnormal bool     `mapstructure:"reject_txs_if_consensus_abnormal" toml:"reject_txs_if_consensus_abnormal"`
}

type P2PPipeGossipsub struct {
	DisableCustomMsgIDFn   bool     `mapstructure:"disable_custom_msg_id_fn" toml:"disable_custom_msg_id_fn"`
	SubBufferSize          int      `mapstructure:"sub_buffer_size" toml:"sub_buffer_size"`
	PeerOutboundBufferSize int      `mapstructure:"peer_outbound_buffer_size" toml:"peer_outbound_buffer_size"`
	ValidateBufferSize     int      `mapstructure:"validate_buffer_size" toml:"validate_buffer_size"`
	SeenMessagesTTL        Duration `mapstructure:"seen_messages_ttl" toml:"seen_messages_ttl"`
	EnableMetrics          bool     `mapstructure:"enable_metrics" toml:"enable_metrics"`
}

type P2PPipeSimpleBroadcast struct {
	WorkerCacheSize        int `mapstructure:"worker_cache_size" toml:"worker_cache_size"`
	WorkerConcurrencyLimit int `mapstructure:"worker_concurrency_limit" toml:"worker_concurrency_limit"`
}

type P2PPipe struct {
	ReceiveMsgCacheSize      int                    `mapstructure:"receive_msg_cache_size" toml:"receive_msg_cache_size"`
	BroadcastType            string                 `mapstructure:"broadcast_type" toml:"broadcast_type"`
	SimpleBroadcast          P2PPipeSimpleBroadcast `mapstructure:"simple_broadcast" toml:"simple_broadcast"`
	Gossipsub                P2PPipeGossipsub       `mapstructure:"gossipsub" toml:"gossipsub"`
	UnicastReadTimeout       Duration               `mapstructure:"unicast_read_timeout" toml:"unicast_read_timeout"`
	UnicastSendRetryNumber   int                    `mapstructure:"unicast_send_retry_number" toml:"unicast_send_retry_number"`
	UnicastSendRetryBaseTime Duration               `mapstructure:"unicast_send_retry_base_time" toml:"unicast_send_retry_base_time"`
	FindPeerTimeout          Duration               `mapstructure:"find_peer_timeout" toml:"find_peer_timeout"`
	ConnectTimeout           Duration               `mapstructure:"connect_timeout" toml:"connect_timeout"`
}

type P2P struct {
	BootstrapNodeAddresses []string `mapstructure:"bootstrap_node_addresses" toml:"bootstrap_node_addresses"`
	Security               string   `mapstructure:"security" toml:"security"`
	SendTimeout            Duration `mapstructure:"send_timeout" toml:"send_timeout"`
	ReadTimeout            Duration `mapstructure:"read_timeout" toml:"read_timeout"`
	Pipe                   P2PPipe  `mapstructure:"pipe" toml:"pipe"`
}

type Monitor struct {
	Enable          bool `mapstructure:"enable" toml:"enable"`
	EnableExpensive bool `mapstructure:"enable_expensive" toml:"enable_expensive"`
}

type PProf struct {
	Enable   bool     `mapstructure:"enable" toml:"enbale"`
	PType    string   `mapstructure:"ptype" toml:"ptype"`
	Mode     string   `mapstructure:"mode" toml:"mode"`
	Duration Duration `mapstructure:"duration" toml:"duration"`
}

type JLimiter struct {
	Interval Duration `mapstructure:"interval" toml:"interval"`
	Quantum  int64    `mapstructure:"quantum" toml:"quantum"`
	Capacity int64    `mapstructure:"capacity" toml:"capacity"`
	Enable   bool     `mapstructure:"enable" toml:"enable"`
}

type Log struct {
	Level            string `mapstructure:"level" toml:"level"`
	Filename         string `mapstructure:"filename" toml:"filename"`
	ReportCaller     bool   `mapstructure:"report_caller" toml:"report_caller"`
	EnableCompress   bool   `mapstructure:"enable_compress" toml:"enable_compress"`
	EnableColor      bool   `mapstructure:"enable_color" toml:"enable_color"`
	DisableTimestamp bool   `mapstructure:"disable_timestamp" toml:"disable_timestamp"`

	// unit: day
	MaxAge uint `mapstructure:"max_age" toml:"max_age"`

	// unit: MB
	MaxSize uint `mapstructure:"max_size" toml:"max_size"`

	RotationTime Duration  `mapstructure:"rotation_time" toml:"rotation_time"`
	Module       LogModule `mapstructure:"module" toml:"module"`
}

type LogModule struct {
	P2P        string `mapstructure:"p2p" toml:"p2p"`
	Consensus  string `mapstructure:"consensus" toml:"consensus"`
	Executor   string `mapstructure:"executor" toml:"executor"`
	Governance string `mapstructure:"governance" toml:"governance"`
	API        string `mapstructure:"api" toml:"api"`
	APP        string `mapstructure:"app" toml:"app"`
	CoreAPI    string `mapstructure:"coreapi" toml:"coreapi"`
	Storage    string `mapstructure:"storage" toml:"storage"`
	Profile    string `mapstructure:"profile" toml:"profile"`
	Finance    string `mapstructure:"finance" toml:"finance"`
	TxPool     string `mapstructure:"txpool" toml:"txpool"`
	Access     string `mapstructure:"access" toml:"access"`
	BlockSync  string `mapstructure:"blocksync" toml:"blocksync"`
	Epoch      string `mapstructure:"epoch" toml:"epoch"`
}

type Genesis struct {
	ChainID                uint64          `mapstructure:"chainid" toml:"chainid"`
	Balance                string          `mapstructure:"balance" toml:"balance"`
	Admins                 []*Admin        `mapstructure:"admins" toml:"admins"`
	InitWhiteListProviders []string        `mapstructure:"init_white_list_providers" toml:"init_white_list_providers"`
	Accounts               []string        `mapstructure:"accounts" toml:"accounts"`
	EpochInfo              *rbft.EpochInfo `mapstructure:"epoch_info" toml:"epoch_info"`
	NodeNames              []*NodeName     `mapstructure:"node_names" toml:"node_names"`
}

type Access struct {
	EnableWhiteList bool `mapstructure:"enable_white_list" toml:"enable_white_list"`
}

type Admin struct {
	Address string `mapstructure:"address" toml:"address"`
	Weight  uint64 `mapstructure:"weight" toml:"weight"`
	Name    string `mapstructure:"name" toml:"name"`
}

type NodeName struct {
	ID   uint64 `mapstructure:"id" toml:"id"`
	Name string `mapstructure:"name" toml:"name"`
}

type Sync struct {
	RequesterRetryTimeout Duration `mapstructure:"requester_retry_timeout" toml:"requester_retry_timeout"`
	TimeoutCountLimit     uint64   `mapstructure:"timeout_count_limit" toml:"timeout_count_limit"`
	ConcurrencyLimit      uint64   `mapstructure:"concurrency_limit" toml:"concurrency_limit"`
}

type Consensus struct {
	Type        string `mapstructure:"type" toml:"type"`
	StorageType string `mapstructure:"storage_type" toml:"storage_type"`
}

type Storage struct {
	KvType      string `mapstructure:"kv_type" toml:"kv_type"`
	KvCacheSize int    `mapstructure:"kv_cache_size" toml:"kv_cache_size"`
	Sync        bool   `mapstructure:"sync" toml:"sync"`
}

type Ledger struct {
	ChainLedgerCacheSize           int `mapstructure:"chain_ledger_cache_size" toml:"chain_ledger_cache_size"`
	StateLedgerCacheMegabytesLimit int `mapstructure:"state_ledger_cache_megabytes_limit" toml:"state_ledger_cache_megabytes_limit"`
	StateLedgerAccountCacheSize    int `mapstructure:"state_ledger_account_cache_size" toml:"state_ledger_account_cache_size"`
}

type EVM struct {
	DisableMaxCodeSizeLimit bool `mapstructure:"disable_max_code_size_limit" toml:"disable_max_code_size_limit"`
}

type Executor struct {
	Type            string `mapstructure:"type" toml:"type"`
	DisableRollback bool   `mapstructure:"disable_rollback" toml:"disable_rollback"`
	EVM             EVM    `mapstructure:"evm" toml:"evm"`
}

var SupportMultiNode = make(map[string]bool)
var registrationMutex sync.Mutex

func Register(consensusType string, isSupported bool) {
	registrationMutex.Lock()
	defer registrationMutex.Unlock()
	SupportMultiNode[consensusType] = isSupported
}

func (c *Config) Bytes() ([]byte, error) {
	ret, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func GenesisNodeNameInfo(epochEnable bool) []*NodeName {
	var nodes []*NodeName
	var sliceLength int
	if epochEnable {
		sliceLength = len(DefaultNodeAddrs)
	} else {
		sliceLength = 4
	}
	nodes = lo.Map(DefaultNodeAddrs[:sliceLength], func(item string, idx int) *NodeName {
		return &NodeName{
			Name: DefaultNodeNames[idx],
			ID:   uint64(idx + 1),
		}
	})
	return nodes
}

func GenesisEpochInfo(epochEnable bool) *rbft.EpochInfo {
	var candidateSet, validatorSet []rbft.NodeInfo
	if epochEnable {
		candidateSet = lo.Map(DefaultNodeAddrs[4:], func(item string, idx int) rbft.NodeInfo {
			idx += 4
			return rbft.NodeInfo{
				ID:                   uint64(idx + 1),
				AccountAddress:       DefaultNodeAddrs[idx],
				P2PNodeID:            defaultNodeIDs[idx],
				ConsensusVotingPower: int64(len(DefaultNodeAddrs)-idx) * 1000,
			}
		})
		validatorSet = lo.Map(DefaultNodeAddrs[0:4], func(item string, idx int) rbft.NodeInfo {
			return rbft.NodeInfo{
				ID:                   uint64(idx + 1),
				AccountAddress:       DefaultNodeAddrs[idx],
				P2PNodeID:            defaultNodeIDs[idx],
				ConsensusVotingPower: int64(len(DefaultNodeAddrs)-idx) * 1000,
			}
		})
	} else {
		validatorSet = lo.Map(DefaultNodeAddrs[0:4], func(item string, idx int) rbft.NodeInfo {
			return rbft.NodeInfo{
				ID:                   uint64(idx + 1),
				AccountAddress:       DefaultNodeAddrs[idx],
				P2PNodeID:            defaultNodeIDs[idx],
				ConsensusVotingPower: 1000,
			}
		})
	}

	return &rbft.EpochInfo{
		Version:     1,
		Epoch:       1,
		EpochPeriod: 100,
		StartBlock:  1,
		P2PBootstrapNodeAddresses: lo.Map(defaultNodeIDs[0:4], func(item string, idx int) string {
			return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", 4001+idx, item)
		}),
		ConsensusParams: rbft.ConsensusParams{
			ProposerElectionType:                 rbft.ProposerElectionTypeWRF,
			ValidatorElectionType:                rbft.ValidatorElectionTypeWRF,
			CheckpointPeriod:                     1,
			HighWatermarkCheckpointPeriod:        10,
			MaxValidatorNum:                      4,
			BlockMaxTxNum:                        500,
			EnableTimedGenEmptyBlock:             false,
			NotActiveWeight:                      1,
			AbnormalNodeExcludeView:              10,
			AgainProposeIntervalBlock:            0,
			ContinuousNullRequestToleranceNumber: 3,
			ReBroadcastToleranceNumber:           2,
		},
		CandidateSet: candidateSet,
		ValidatorSet: validatorSet,
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
	}
}

func DefaultConfig(epochEnable bool) *Config {
	if testNetConfigBuilder, ok := TestNetConfigBuilderMap[BuildNet]; ok {
		return testNetConfigBuilder()
	}
	return &Config{
		Ulimit: 65535,
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
				Enable:   false,
			},
			WriteLimiter: JLimiter{
				Interval: 50,
				Quantum:  500,
				Capacity: 10000,
				Enable:   false,
			},
			RejectTxsIfConsensusAbnormal: false,
		},
		P2P: P2P{
			Security:    P2PSecurityTLS,
			SendTimeout: Duration(5 * time.Second),
			ReadTimeout: Duration(5 * time.Second),
			Pipe: P2PPipe{
				ReceiveMsgCacheSize: 10240,
				BroadcastType:       P2PPipeBroadcastGossip,
				SimpleBroadcast: P2PPipeSimpleBroadcast{
					WorkerCacheSize:        1024,
					WorkerConcurrencyLimit: 20,
				},
				Gossipsub: P2PPipeGossipsub{
					SubBufferSize:          10240,
					PeerOutboundBufferSize: 10240,
					ValidateBufferSize:     10240,
					SeenMessagesTTL:        Duration(120 * time.Second),
					EnableMetrics:          true,
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
			KvType:      KVStorageTypePebble,
			KvCacheSize: 128,
			Sync:        true,
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
		Genesis: Genesis{
			ChainID: 1356,
			Balance: "1000000000000000000000000000",
			Admins: lo.Map(DefaultNodeAddrs[0:4], func(item string, idx int) *Admin {
				return &Admin{
					Address: item,
					Weight:  1,
					Name:    DefaultAdminNames[idx],
				}
			}),
			NodeNames:              GenesisNodeNameInfo(epochEnable),
			InitWhiteListProviders: DefaultNodeAddrs,
			Accounts: []string{
				"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
				"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
				"0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
				"0x90F79bf6EB2c4f870365E785982E1f101E93b906",
				"0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65",
				"0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc",
				"0x976EA74026E726554dB657fA54763abd0C3a0aa9",
				"0x14dC79964da2C08b23698B3D3cc7Ca32193d9955",
				"0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f",
				"0xa0Ee7A142d267C1f36714E4a8F75612F20a79720",
				"0xBcd4042DE499D14e55001CcbB24a551F3b954096",
				"0x71bE63f3384f5fb98995898A86B02Fb2426c5788",
				"0xFABB0ac9d68B0B445fB7357272Ff202C5651694a",
				"0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec",
				"0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097",
				"0xcd3B766CCDd6AE721141F452C550Ca635964ce71",
				"0x2546BcD3c84621e976D8185a91A922aE77ECEc30",
				"0xbDA5747bFD65F08deb54cb465eB87D40e51B197E",
				"0xdD2FD4581271e230360230F9337D5c0430Bf44C0",
				"0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
			},
			EpochInfo: GenesisEpochInfo(epochEnable),
		},
		PProf: PProf{
			Enable:   true,
			PType:    PprofTypeHTTP,
			Mode:     PprofModeMem,
			Duration: Duration(30 * time.Second),
		},
		Monitor: Monitor{
			Enable:          true,
			EnableExpensive: true,
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
				API:        "info",
				CoreAPI:    "info",
				Storage:    "info",
				Profile:    "info",
				Finance:    "error",
				BlockSync:  "info",
				APP:        "info",
				Access:     "info",
				TxPool:     "debug",
				Epoch:      "info",
			},
		},
		Access: Access{
			EnableWhiteList: false,
		},
	}
}

func LoadConfig(repoRoot string) (*Config, error) {
	cfg, err := func() (*Config, error) {
		cfg := DefaultConfig(false)
		cfgPath := path.Join(repoRoot, CfgFileName)
		existConfig := fileutil.Exist(cfgPath)
		if !existConfig {
			err := os.MkdirAll(repoRoot, 0755)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build default config")
			}

			if err := writeConfigWithEnv(cfgPath, cfg); err != nil {
				return nil, errors.Wrap(err, "failed to build default config")
			}
		} else {
			if err := CheckWritable(repoRoot); err != nil {
				return nil, err
			}
			if err := readConfigFromFile(cfgPath, cfg); err != nil {
				return nil, err
			}
		}

		return cfg, nil
	}()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}
	return cfg, nil
}
