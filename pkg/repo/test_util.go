package repo

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/types"
)

var (
	consensusKeys = []string{
		"0x099383c2b41a282936fe9e656467b2ad6ecafd38753eefa080b5a699e3276372",
		"0x5d21b741bd16e05c3a883b09613d36ad152f1586393121d247bdcfef908cce8f",
		"0x42cc8e862b51a1c21a240bb2ae6f2dbad59668d86fe3c45b2e4710eebd2a63fd",
		"0x6e327c2d5a284b89f9c312a02b2714a90b38e721256f9a157f03ec15c1a386a6",
	}
	p2pKeys = []string{
		"0xce374993d8867572a043e443355400ff4628662486d0d6ef9d76bc3c8b2aa8a8",
		"0x43dd946ade57013fd4e7d0f11d84b94e2fda4336829f154ae345be94b0b63616",
		"0x875e5ef34c34e49d35ff5a0f8a53003d8848fc6edd423582c00edc609a1e3239",
		"0xf0aac0c25791d0bd1b96b2ec3c9c25539045cf6cc5cc9ad0f3cb64453d1f38c0",
	}
)

func MockRepo(t testing.TB) *Repo {
	return MockRepoWithNodeID(t, 1)
}

func MockRepoWithNodeID(t testing.TB, nodeID uint64) *Repo {
	repoRoot := t.TempDir()
	consensusKeystore, err := GenerateConsensusKeystore(repoRoot, consensusKeys[nodeID-1], DefaultKeystorePassword)
	require.Nil(t, err)
	p2pKeystore, err := GenerateP2PKeystore(repoRoot, p2pKeys[nodeID-1], DefaultKeystorePassword)
	require.Nil(t, err)
	rep := &Repo{
		RepoRoot:          repoRoot,
		Config:            defaultConfig(),
		ConsensusConfig:   defaultConsensusConfig(),
		GenesisConfig:     defaultGenesisConfig(),
		ConsensusKeystore: consensusKeystore,
		P2PKeystore:       p2pKeystore,
		StartArgs:         &StartArgs{},
	}
	rep.GenesisConfig.Nodes = []GenesisNodeInfo{
		{
			ConsensusPubKey: "0xac9bb2675ab6b60b1c6d3ed60e95bdabb16517525458d8d50fa1065014184823556b0bd97922fab8c688788006e8b1030cd506d19101522e203769348ea10d21780e5c26a5c03c0cfcb8de23c7cf16d4d384140613bb953d446a26488fbaf6e0",
			P2PPubKey:       "0xd4c0ac1567bcb2c855bb1692c09ab2a2e2c84c45376592674530ce95f1fda351",
			OperatorAddress: "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013",
			MetaData: GenesisNodeMetaData{
				Name: "node1",
				Desc: "node1",
			},
			IsDataSyncer:   false,
			StakeNumber:    types.CoinNumberByAxc(1000),
			CommissionRate: 0,
		},
		{
			ConsensusPubKey: "0xa41eb8e086872835b17e323dadd569d98caa8645b694c5a3095a1a0790a4390cb0db7a79af5411328bc17c9fb213d7f407a471c929e8aa2fe33e9e1472adb000c86990dd81078906d2cccd831a7fa4a0772a094e02d58db361162e95ac5e29fa",
			P2PPubKey:       "0xeec45cda21da07acb89d3e7db8ce78933773b4b1daf567c2efddc6d5fd687001",
			OperatorAddress: "0x79a1215469FaB6f9c63c1816b45183AD3624bE34",
			MetaData: GenesisNodeMetaData{
				Name: "node2",
				Desc: "node2",
			},
			IsDataSyncer:   false,
			StakeNumber:    types.CoinNumberByAxc(1000),
			CommissionRate: 0,
		},
		{
			ConsensusPubKey: "0xa8c8b2635518df0212e92e3056ffbc3388cdcaf175227c914cdf713419bb25d3ba73e9f4981330aa20b3016ee668c19f0d3408226aeca43261abf01bd17b3cb5992ada1d0b3bb90b28930eed40d95b3e0a72bf6df5a30feb3330a9e7561eb82b",
			P2PPubKey:       "0xb20879ca4baa02370d8a5d033be54e49df5163c0c8257cb0d49b38385aa14930",
			OperatorAddress: "0x97c8B516D19edBf575D72a172Af7F418BE498C37",
			MetaData: GenesisNodeMetaData{
				Name: "node3",
				Desc: "node3",
			},
			IsDataSyncer:   false,
			StakeNumber:    types.CoinNumberByAxc(1000),
			CommissionRate: 0,
		},
		{
			ConsensusPubKey: "0xa1a0d4cea22621b1b61adf2e050b4f191b6504344271229eb613045a832e92b3ee152d3f32ef7e79408e190c80709b0006ddc3f10cae1da52a98fb27c0d5eaa29b906c416b3b60e8ec09619fd67a3a4fa7b468f3de96acd120fe1c39ff579ea1",
			P2PPubKey:       "0x98cdfe5dd41fe8d336d45572778ced304f8571e21c822de148b960d34ccca256",
			OperatorAddress: "0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8",
			MetaData: GenesisNodeMetaData{
				Name: "node4",
				Desc: "node4",
			},
			IsDataSyncer:   false,
			StakeNumber:    types.CoinNumberByAxc(1000),
			CommissionRate: 0,
		},
	}

	rep.GenesisConfig.Accounts = lo.Map(rep.GenesisConfig.Nodes, func(item GenesisNodeInfo, index int) *Account {
		return &Account{
			Address: item.OperatorAddress,
			Balance: types.CoinNumberByAxc(10000000),
		}
	})
	return rep
}
