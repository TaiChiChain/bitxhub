package adaptor

import "github.com/axiomesh/axiom-kit/types"

func (a *RBFTAdaptor) GetCurrentEpochInfo() (*types.EpochInfo, error) {
	currentEpochInfo := a.config.ChainState.GetCurrentEpochInfo()
	return &currentEpochInfo, nil
}

func (a *RBFTAdaptor) GetEpochInfo(epoch uint64) (*types.EpochInfo, error) {
	return a.config.ChainState.GetEpochInfo(epoch)
}

func (a *RBFTAdaptor) StoreEpochState(_ uint64, value types.QuorumCheckpoint) error {
	return a.epochManager.StoreEpochState(value)
}

func (a *RBFTAdaptor) ReadEpochState(epoch uint64) (types.QuorumCheckpoint, error) {
	return a.epochManager.ReadEpochState(epoch)
}
