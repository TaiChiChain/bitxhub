package epochmgr

import (
	"encoding/binary"
	"fmt"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
)

const (
	EpochStatePrefix = "epoch_q_chkpt."
	EpochIndexKey    = "epoch_latest_idx"
	ProofIndexKey    = "proof_latest_idx"
)

type EpochManager struct {
	consensusType string
	store         kv.Storage
	genesisHeader *types.BlockHeader
}

func NewEpochManager(consensusType string, store kv.Storage, genesisHeader *types.BlockHeader) *EpochManager {
	return &EpochManager{
		consensusType: consensusType,
		store:         store,
		genesisHeader: genesisHeader,
	}
}

func (e *EpochManager) StoreEpochState(c types.QuorumCheckpoint) error {
	newEpoch := c.NextEpoch()
	raw, err := c.Marshal()
	if err != nil {
		return fmt.Errorf("Persist epoch %d quorum chkpt failed with marshal err: %s ", newEpoch, err)
	}

	e.store.Put(e.genEpochKey(newEpoch), raw)

	// update latest epoch index
	if e.GetLatestEpoch() < c.NextEpoch() {
		data := make([]byte, 8)
		binary.BigEndian.PutUint64(data, c.NextEpoch())
		e.store.Put(e.genLatestEpochIndexKey(), data)
	}
	return nil
}

func (e *EpochManager) ReadEpochState(epoch uint64) (types.QuorumCheckpoint, error) {
	if epoch == e.genesisHeader.Epoch {
		return e.genGenesisEpochCheckpoint()
	}
	key := e.genEpochKey(epoch)
	b := e.store.Get(key)
	if b == nil {
		return nil, fmt.Errorf("[EpochService] epoch state not found at epoch %d", epoch)
	}
	ckpt := consensustypes.QuorumCheckpointConstructor[e.consensusType]()
	if ckpt == nil {
		return nil, fmt.Errorf("[EpochService] epoch state constructor not found at consensus type %s", e.consensusType)
	}
	if err := ckpt.Unmarshal(b); err != nil {
		return nil, fmt.Errorf("[EpochService] epoch state unmarshal error at epoch %d", epoch)
	}
	return ckpt, nil
}

func (e *EpochManager) GetLatestEpoch() uint64 {
	data := e.store.Get(e.genLatestEpochIndexKey())
	if data == nil {
		return 0
	}
	return binary.BigEndian.Uint64(data)
}

func (e *EpochManager) StoreLatestProof(proof []byte) error {
	e.store.Put(e.genProofKey(), proof)
	return nil
}

func (e *EpochManager) GetLatestProof() ([]byte, error) {
	data := e.store.Get(e.genProofKey())
	if data == nil {
		return nil, fmt.Errorf("[EpochService] latest proof not found")
	}

	return data, nil
}

func (e *EpochManager) genEpochKey(epoch uint64) []byte {
	return []byte("epoch." + fmt.Sprintf("%s%d", EpochStatePrefix, epoch))
}

func (e *EpochManager) genProofKey() []byte {
	return []byte("proof." + ProofIndexKey)
}

func (e *EpochManager) genLatestEpochIndexKey() []byte {
	return []byte("epoch." + EpochIndexKey)
}

func (e *EpochManager) GenGenesisEpochCheckpoint() (types.QuorumCheckpoint, error) {
	return e.genGenesisEpochCheckpoint()
}

func (e *EpochManager) genGenesisEpochCheckpoint() (types.QuorumCheckpoint, error) {
	switch e.consensusType {
	case repo.ConsensusTypeDagBft:
		return &consensustypes.DagbftQuorumCheckpoint{
			QuorumCheckpoint: &dagtypes.QuorumCheckpoint{
				QuorumCheckpoint: protos.QuorumCheckpoint{
					Checkpoint: &protos.Checkpoint{
						Epoch: e.genesisHeader.Epoch,
						ExecuteState: &protos.ExecuteState{
							Height:    e.genesisHeader.Number,
							StateRoot: e.genesisHeader.Hash().String(),
						},
					},
				},
			},
		}, nil

	case repo.ConsensusTypeRbft:
		return &consensus.RbftQuorumCheckpoint{
			QuorumCheckpoint: &consensus.QuorumCheckpoint{
				Checkpoint: &consensus.Checkpoint{
					Epoch: e.genesisHeader.Epoch,
					ExecuteState: &consensus.Checkpoint_ExecuteState{
						Height: e.genesisHeader.Number,
						Digest: e.genesisHeader.Hash().String(),
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported consensus type %s", e.consensusType)
	}
}
