package repo

var (
	TestNetConfigBuilderMap = map[string]func() *Config{}

	TestNetConsensusConfigBuilderMap = map[string]func() *ConsensusConfig{}

	TestNetGenesisConfigBuilderMap = map[string]func() *GenesisConfig{}
)
