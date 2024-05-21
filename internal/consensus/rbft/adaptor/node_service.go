package adaptor

import (
	"github.com/samber/lo"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
)

func (a *RBFTAdaptor) GetNodeInfo(nodeID uint64) (*rbft.NodeInfo, error) {
	nodeInfo, err := a.config.ChainState.GetNodeInfo(nodeID)
	if err != nil {
		return nil, err
	}
	return &rbft.NodeInfo{
		ID:        nodeInfo.ID,
		P2PNodeID: nodeInfo.P2PID,
	}, nil
}

func (a *RBFTAdaptor) GetNodeIDByP2PID(p2pID string) (uint64, error) {
	return a.config.ChainState.GetNodeIDByP2PID(p2pID)
}

func (a *RBFTAdaptor) GetValidatorSet() (map[uint64]int64, error) {
	return lo.SliceToMap(a.config.ChainState.ValidatorSet, func(item chainstate.ValidatorInfo) (uint64, int64) {
		return item.ID, item.ConsensusVotingPower
	}), nil
}
