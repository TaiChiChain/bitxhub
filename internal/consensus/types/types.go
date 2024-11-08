package types

import (
	"encoding/json"
	"fmt"

	rbft "github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
)

const (
	LocalTxEvent = iota
	RemoteTxEvent
)

const (
	Batch     timer.TimeoutEvent = "Batch"
	NoTxBatch timer.TimeoutEvent = "NoTxBatch"
)

const (
	Dagbft = "dagbft"
	Rbft   = "rbft"
	Solo   = "solo"
)

var (
	ErrorPreCheck       = errors.New("precheck failed")
	ErrorAddTxPool      = errors.New("add txpool failed")
	ErrorConsensusStart = errors.New("consensus not start yet")
)

var DataSyncerPipeName = []string{
	"NULL_REQUEST",           // primary heartbeat
	"PRE_PREPARE",            // get batch
	"SIGNED_CHECKPOINT",      // get checkpoint
	"SYNC_STATE_RESPONSE",    // get quorum state
	"FETCH_MISSING_RESPONSE", // get missing txs in local pool
	"EPOCH_CHANGE_PROOF",     // get epoch change for state update
}

var DataSyncerRequestName = []string{
	"SYNC_STATE",            // get quorum state
	"FETCH_MISSING_REQUEST", // get missing txs in local pool
	"EPOCH_CHANGE_REQUEST",  // get epoch change for state update
}

// UncheckedTxEvent represents misc event sent by local modules
type UncheckedTxEvent struct {
	EventType int
	Event     any
}

type TxWithResp struct {
	Tx      *types.Transaction
	CheckCh chan *TxResp
	PoolCh  chan *TxResp
}

type TxResp struct {
	Status   bool
	ErrorMsg string
}

type CommitEvent struct {
	Block                  *types.Block
	RecvConsensusTime      int64
	StateUpdatedCheckpoint *Checkpoint
	CommitSequence         uint64
}

type Checkpoint struct {
	Epoch  uint64
	Height uint64
	Digest string
}

var QuorumCheckpointConstructor = map[string]func() types.QuorumCheckpoint{
	Dagbft: func() types.QuorumCheckpoint {
		return &DagbftQuorumCheckpoint{}
	},
	Rbft: func() types.QuorumCheckpoint {
		return &rbft.RbftQuorumCheckpoint{}
	},
}

type Attestation struct {
	Epoch         uint64 `json:"epoch"`
	ConsensusType string `json:"consensus_type"`
	Block         []byte `json:"block"`
	Proof         []byte `json:"proof"`
}

func (a *Attestation) GetBlock() (*types.Block, error) {
	block := &types.Block{}
	if err := block.Unmarshal(a.Block); err != nil {
		return nil, err
	}
	return block, nil
}

func (a *Attestation) GetProof() (types.QuorumCheckpoint, error) {
	switch a.ConsensusType {
	case repo.ConsensusTypeDagBft:
		cp := &DagbftQuorumCheckpoint{}
		if err := cp.Unmarshal(a.Proof); err != nil {
			return nil, err
		}
		return cp, nil
	case repo.ConsensusTypeRbft:
		cp := &rbft.RbftQuorumCheckpoint{}
		if err := cp.Unmarshal(a.Proof); err != nil {
			return nil, err
		}
		return cp, nil
	case repo.ConsensusTypeSolo:
		cp := &MockQuorumCheckpoint{}
		if err := cp.Unmarshal(a.Proof); err != nil {
			return nil, err
		}
		return cp, nil
	default:
		return nil, fmt.Errorf("unsupported consensus type: %s", a.ConsensusType)
	}
}

func (a *Attestation) GetEpoch() uint64 {
	return a.Epoch
}

type EpochChange struct {
	types.QuorumCheckpoint
}

type DagbftQuorumCheckpoint struct {
	*dagtypes.QuorumCheckpoint
}

func (q *DagbftQuorumCheckpoint) Epoch() uint64 {
	return q.QuorumCheckpoint.Epoch()
}

func (q *DagbftQuorumCheckpoint) NextEpoch() uint64 {
	return q.QuorumCheckpoint.NextEpoch()
}

func (q *DagbftQuorumCheckpoint) Marshal() ([]byte, error) {
	if q.QuorumCheckpoint == nil {
		return nil, errors.New("nil quorum checkpoint")
	}
	return q.QuorumCheckpoint.Marshal()
}

func (q *DagbftQuorumCheckpoint) Unmarshal(raw []byte) error {
	q.QuorumCheckpoint = &dagtypes.QuorumCheckpoint{}
	if err := q.QuorumCheckpoint.Unmarshal(raw); err != nil {
		return err
	}
	return nil
}

func (q *DagbftQuorumCheckpoint) GetHeight() uint64 {
	return q.QuorumCheckpoint.Height()
}

func (q *DagbftQuorumCheckpoint) GetStateDigest() string {
	return q.GetCheckpoint().GetExecuteState().GetStateRoot()
}

func (q *DagbftQuorumCheckpoint) GetSignatures() []types.Signature {
	signatures, ok := q.Signatures().(dagtypes.MultiSignature)
	if !ok {
		return nil
	}

	return lo.Map(signatures.Signatures, func(sig protocol.Signature, _ int) types.Signature {
		return types.Signature{Singer: uint64(sig.Signer), Signature: sig.Signature}
	})
}

func (q *DagbftQuorumCheckpoint) EndEpoch() bool {
	return q.QuorumCheckpoint.EndsEpoch()
}

type MockQuorumCheckpoint struct {
	BlockEpoch      uint64 `json:"block_epoch"`
	Height          uint64 `json:"height"`
	Digest          string `json:"digest"`
	NeedUpdateEpoch bool   `json:"need_update_epoch"`
}

func (q *MockQuorumCheckpoint) Epoch() uint64 {
	return q.BlockEpoch
}

func (q *MockQuorumCheckpoint) NextEpoch() uint64 {
	return q.BlockEpoch + 1
}

func (q *MockQuorumCheckpoint) Marshal() ([]byte, error) {
	return json.Marshal(q)
}

func (q *MockQuorumCheckpoint) Unmarshal(raw []byte) error {
	return json.Unmarshal(raw, q)
}

func (q *MockQuorumCheckpoint) GetHeight() uint64 {
	return q.Height
}

func (q *MockQuorumCheckpoint) GetStateDigest() string {
	return q.Digest
}

func (q *MockQuorumCheckpoint) GetSignatures() []byte {
	return nil
}

func (q *MockQuorumCheckpoint) EndEpoch() bool {
	return q.NeedUpdateEpoch
}
