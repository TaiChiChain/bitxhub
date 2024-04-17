package adaptor

import (
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
)

func (a *RBFTAdaptor) GetNodeInfo(nodeID uint64) (*types.NodeInfo, error) {
	nodeInfo, err := a.config.ChainState.GetNodeInfo(nodeID)
	if err != nil {
		return nil, err
	}
	return &nodeInfo.NodeInfo, nil
}

func (a *RBFTAdaptor) GetNodeIDByP2PID(p2pID string) (uint64, error) {
	return a.config.ChainState.GetNodeIDByP2PID(p2pID)
}

func (a *RBFTAdaptor) GetValidatorSet() (map[uint64]int64, error) {
	return lo.SliceToMap(a.config.ChainState.ValidatorSet, func(item chainstate.ValidatorInfo) (uint64, int64) {
		return item.ID, item.ConsensusVotingPower
	}), nil
}
