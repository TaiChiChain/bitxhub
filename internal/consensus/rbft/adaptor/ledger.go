package adaptor

import rbfttypes "github.com/axiomesh/axiom-bft/types"

func (a *RBFTAdaptor) GetBlockMeta(num uint64) (*rbfttypes.BlockMeta, error) {
	blockHeader, err := a.config.GetBlockHeaderFunc(num)
	if err != nil {
		return nil, err
	}
	return &rbfttypes.BlockMeta{
		ProcessorNodeID: blockHeader.ProposerNodeID,
		BlockNum:        blockHeader.Number,
	}, nil
}
