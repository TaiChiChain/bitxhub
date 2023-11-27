package adaptor

import rbfttypes "github.com/axiomesh/axiom-bft/types"

func (a *RBFTAdaptor) GetBlockMeta(num uint64) (*rbfttypes.BlockMeta, error) {
	block, err := a.config.GetBlockFunc(num)
	if err != nil {
		return nil, err
	}
	return &rbfttypes.BlockMeta{
		ProcessorNodeID: block.BlockHeader.ProposerNodeID,
		BlockNum:        block.BlockHeader.Number,
	}, nil
}
