package repo

import "github.com/axiomesh/axiom-kit/types"

const (
	AppName = "AxiomLedger"

	// CfgFileName is the default config name
	CfgFileName = "config.toml"

	consensusCfgFileName = "consensus.toml"

	genesisCfgFileName = "genesis.toml"

	// defaultRepoRoot is the path to the default config dir location.
	defaultRepoRoot = "~/.axiom-ledger"

	// rootPathEnvVar is the environment variable used to change the path root.
	rootPathEnvVar = "AXIOM_LEDGER_PATH"

	P2PKeystoreFileName = "p2p-keystore.json"

	p2pKeystoreIDKey = "p2p_id"

	ConsensusKeystoreFileName = "consensus-keystore.json"

	DefaultKeystorePassword = "2023@axiomesh"

	pidFileName = "running.pid"

	LogsDirName = "logs"
)

const (
	ConsensusTypeSolo    = "solo"
	ConsensusTypeRbft    = "rbft"
	ConsensusTypeSoloDev = "solo_dev"

	ConsensusStorageTypeMinifile = "minifile"
	ConsensusStorageTypeRosedb   = "rosedb"

	KVStorageTypeLeveldb = "leveldb"
	KVStorageTypePebble  = "pebble"
	KVStorageCacheSize   = 16
	KVStorageSync        = true

	P2PSecurityTLS   = "tls"
	P2PSecurityNoise = "noise"

	PprofModeMem     = "mem"
	PprofModeCpu     = "cpu"
	PprofTypeHTTP    = "http"
	PprofTypeRuntime = "runtime"

	ExecTypeNative = "native"
	ExecTypeDev    = "dev"

	// txSlotSize is used to calculate how many data slots a single transaction
	// takes up based on its size. The slots are used as DoS protection, ensuring
	// that validating a new transaction remains a constant operation (in reality
	// O(maxslots), where max slots are 4 currently).
	txSlotSize = 32 * 1024

	// DefaultTxMaxSize is the maximum size a single transaction can have. This field has
	// non-trivial consequences: larger transactions are significantly harder and
	// more expensive to propagate; larger transactions also take more resources
	// to validate whether they fit into the pool or not.
	DefaultTxMaxSize = 4 * txSlotSize // 128KB
)

var (
	DefaultAXCBalance     = types.CoinNumberByAxc(1000000000) // 1 billion AXC
	DefaultAccountBalance = types.CoinNumberByAxc(10000000)   // 10 million AXC

	DefaultStartGasPrice = types.CoinNumberByGmol(5000)
	DefaultMaxGasPrice   = types.CoinNumberByGmol(10000)
	DefaultMinGasPrice   = types.CoinNumberByGmol(1000)
)
