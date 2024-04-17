package adaptor

import (
	"github.com/pkg/errors"
)

func (a *RBFTAdaptor) Sign(msg []byte) ([]byte, error) {
	if a.config.Repo.Config.Consensus.UseBlsKey {
		return a.config.Repo.ConsensusKeystore.PrivateKey.Sign(msg)
	}
	return a.config.Repo.P2PKeystore.PrivateKey.Sign(msg)
}

func (a *RBFTAdaptor) Verify(nodeID uint64, signature []byte, msg []byte) error {
	nodeInfo, err := a.config.ChainState.GetNodeInfo(nodeID)
	if err != nil {
		return err
	}
	if a.config.Repo.Config.Consensus.UseBlsKey {
		if !nodeInfo.ConsensusPubKey.Verify(msg, signature) {
			return errors.New("invalid signature")
		}
	} else {
		if !nodeInfo.P2PPubKey.Verify(msg, signature) {
			return errors.New("invalid signature")
		}
	}
	return nil
}
