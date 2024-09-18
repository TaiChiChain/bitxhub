package types

import (
	rbft "github.com/axiomesh/axiom-bft/common/consensus"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/pkg/errors"

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

type EpochChange struct {
	types.QuorumCheckpoint
}

type DagbftQuorumCheckpoint struct {
	dagtypes.QuorumCheckpoint
}

func (q *DagbftQuorumCheckpoint) Epoch() uint64 {
	return q.QuorumCheckpoint.Epoch()
}

func (q *DagbftQuorumCheckpoint) NextEpoch() uint64 {
	return q.QuorumCheckpoint.NextEpoch()
}

func (q *DagbftQuorumCheckpoint) Marshal() ([]byte, error) {
	return q.QuorumCheckpoint.Marshal()
}

func (q *DagbftQuorumCheckpoint) Unmarshal(raw []byte) error {
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
