package chainstate

import (
	"math/big"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func NewMockChainState(genesisConfig *repo.GenesisConfig, epochMap map[uint64]*types.EpochInfo) *ChainState {
	nodeInfoMap := map[uint64]*ExpandedNodeInfo{}
	for i, nodeInfo := range genesisConfig.Nodes {
		nodeID := uint64(i + 1)

		consensusPubKey := &crypto.Bls12381PublicKey{}
		if err := consensusPubKey.Unmarshal(hexutil.Decode(nodeInfo.ConsensusPubKey)); err != nil {
			panic(errors.Wrap(err, "failed to unmarshal consensus public key"))
		}

		p2pPubKey := &crypto.Ed25519PublicKey{}
		if err := p2pPubKey.Unmarshal(hexutil.Decode(nodeInfo.P2PPubKey)); err != nil {
			panic(errors.Wrap(err, "failed to unmarshal p2p public key"))
		}
		p2pID, err := repo.P2PPubKeyToID(p2pPubKey)
		if err != nil {
			panic(errors.Wrap(err, "failed to calculate p2p id from p2p public key"))
		}

		nodeInfoMap[nodeID] = &ExpandedNodeInfo{
			NodeInfo: types.NodeInfo{
				ID:              nodeID,
				ConsensusPubKey: nodeInfo.ConsensusPubKey,
				P2PPubKey:       nodeInfo.P2PPubKey,
				P2PID:           p2pID,
				OperatorAddress: nodeInfo.OperatorAddress,
				MetaData: types.NodeMetaData{
					Name:       nodeInfo.MetaData.Name,
					Desc:       nodeInfo.MetaData.Desc,
					ImageURL:   nodeInfo.MetaData.ImageURL,
					WebsiteURL: nodeInfo.MetaData.WebsiteURL,
				},
				Status: types.NodeStatusActive,
			},
			P2PPubKey:       p2pPubKey,
			ConsensusPubKey: consensusPubKey,
		}
	}
	if epochMap == nil {
		epochMap = map[uint64]*types.EpochInfo{}
	}
	if _, ok := epochMap[genesisConfig.EpochInfo.Epoch]; !ok {
		epochMap[genesisConfig.EpochInfo.Epoch] = genesisConfig.EpochInfo
	}
	return &ChainState{
		nodeInfoCacheLock:     sync.RWMutex{},
		p2pID2NodeIDCacheLock: sync.RWMutex{},
		epochInfoCacheLock:    sync.RWMutex{},
		getNodeInfoFn: func(u uint64) (*types.NodeInfo, error) {
			return nil, errors.New("node not found")
		},
		getNodeIDByP2PIDFn: func(p2pID string) (uint64, error) {
			return 0, errors.New("node not found")
		},
		getEpochInfoFn: func(epoch uint64) (*types.EpochInfo, error) {
			return nil, errors.New("epoch not found")
		},
		nodeInfoCache: nodeInfoMap,
		p2pID2NodeIDCache: lo.MapEntries(nodeInfoMap, func(key uint64, value *ExpandedNodeInfo) (string, uint64) {
			return value.NodeInfo.P2PID, key
		}),
		epochInfoCache: epochMap,
		selfRegistered: true,
		EpochInfo:      genesisConfig.EpochInfo,
		ChainMeta: &types.ChainMeta{
			Height:    0,
			GasPrice:  new(big.Int),
			BlockHash: &types.Hash{},
		},
		ValidatorSet: lo.Map(genesisConfig.Nodes, func(_ repo.GenesisNodeInfo, index int) ValidatorInfo {
			return ValidatorInfo{ID: uint64(index + 1), ConsensusVotingPower: 1000}
		}),
		SelfNodeInfo: nodeInfoMap[1],
		IsValidator:  true,
	}
}
