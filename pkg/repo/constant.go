package repo

const (
	AppName = "AxiomLedger"

	// CfgFileName is the default config name
	CfgFileName = "config.toml"

	consensusCfgFileName = "consensus.toml"

	// defaultRepoRoot is the path to the default config dir location.
	defaultRepoRoot = "~/.axiom-ledger"

	// rootPathEnvVar is the environment variable used to change the path root.
	rootPathEnvVar = "AXIOM_LEDGER_PATH"

	p2pKeyFileName = "p2p.key"

	AccountKeyFileName = "account.key"

	pidFileName = "running.pid"

	LogsDirName = "logs"
)

const (
	ConsensusTypeSolo    = "solo"
	ConsensusTypeRbft    = "rbft"
	ConsensusTypeSoloDev = "solo_dev"

	KVStorageTypeLeveldb = "leveldb"
	KVStorageTypePebble  = "pebble"
	KVStorageCacheSize   = 16

	P2PSecurityTLS   = "tls"
	P2PSecurityNoise = "noise"

	PprofModeMem     = "mem"
	PprofModeCpu     = "cpu"
	PprofTypeHTTP    = "http"
	PprofTypeRuntime = "runtime"

	P2PPipeBroadcastSimple = "simple"
	P2PPipeBroadcastGossip = "gossip"

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

	// DefaultGasPremiumRate is the default gas price premium rate,
	// For example: if current gas price is 10wei, the default gas price premium rate is 0.2.
	// Then the gasPrice interface will return 10+10 * 0.2 = 12 wei
	DefaultGasPremiumRate = 0.2
)

var (
	DefaultNodeNames = []string{
		"S2luZw==", // base64 encode King
		"UmVk",     // base64 encode Red
		"QXBwbGU=", // base64 encode Apple
		"Q2F0",     // base64 encode Cat
	}

	DefaultNodeKeys = []string{
		"b6477143e17f889263044f6cf463dc37177ac4526c4c39a7a344198457024a2f",
		"05c3708d30c2c72c4b36314a41f30073ab18ea226cf8c6b9f566720bfe2e8631",
		"85a94dd51403590d4f149f9230b6f5de3a08e58899dcaf0f77768efb1825e854",
		"72efcf4bb0e8a300d3e47e6a10f630bcd540de933f01ed5380897fc5e10dc95d",

		// candidates
		"06bf783a69c860a2ab33fe2f99fed38d14bbdba7ef2295bbcb5a073e6c8847ec",
		"bfee1d369f1a98070f85b3b5b3508aaf071440fcdf7bdcb9c725fea835f17433",
		"508d3fd4ec16aff6443cc58bf3df44e55d5d384b1e56529bf52b0c25e8fcf8f7",
		"ffa932acb7c1099de1029070e7def812f8b2c9433adfb8a90b3cb132233a7690",
	}

	DefaultNodeAddrs = []string{
		"0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013",
		"0x79a1215469FaB6f9c63c1816b45183AD3624bE34",
		"0x97c8B516D19edBf575D72a172Af7F418BE498C37",
		"0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8",

		// candidates
		"0xd0091F6D0b39B9E9D2E9051fA46d13B63b8C7B18",
		"0xFd19030f51719D5601Bb079e5c5Be1eD07E01de2",
		"0xE4b988C0BEa762B8809a0E4D14F3ac3f922B41B3",
		"0x5FC85d64dE2125986b1581b4805a43Bfb3af5E52",
	}

	defaultNodeIDs = []string{
		"16Uiu2HAmJ38LwfY6pfgDWNvk3ypjcpEMSePNTE6Ma2NCLqjbZJSF",
		"16Uiu2HAmRypzJbdbUNYsCV2VVgv9UryYS5d7wejTJXT73mNLJ8AK",
		"16Uiu2HAmTwEET536QC9MZmYFp1NUshjRuaq5YSH1sLjW65WasvRk",
		"16Uiu2HAmQBFTnRr84M3xNhi3EcWmgZnnBsDgewk4sNtpA3smBsHJ",

		// candidates
		"16Uiu2HAm2HeK145KTfLaURhcoxBUMZ1PfhVnLRfnmE8qncvXWoZj",
		"16Uiu2HAm2CVtLveAtroaN7pcR8U2saBKjwYqRAikMSwxqdoYMxtv",
		"16Uiu2HAmQv3m5SSyYAoafKmYbTbGmXBaS4DXHXR9wxWKQ9xLzC3n",
		"16Uiu2HAkx1o5fzWLdAobanvE6vqbf1XSbDSgCnid3AoqDGQYFVxo",
	}
)
