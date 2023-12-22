package common

import (
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
)

const (
	SyncBlockRequestPipe  = "sync_block_pipe_v1_request"
	SyncBlockResponsePipe = "sync_block_pipe_v1_response"

	SyncChainDataRequestPipe  = "sync_chain_data_pipe_v1_request"
	SyncChainDataResponsePipe = "sync_chain_data_pipe_v1_response"

	MaxRetryCount = 5
)

type SyncMsgType int

const (
	SyncMsgType_InvalidBlock SyncMsgType = iota
	SyncMsgType_TimeoutBlock
	SyncMsgType_ErrorMsg
)

type SyncMode int

const (
	SyncModeFull SyncMode = iota
	SyncModeSnapshot
)

var SyncModeMap = map[SyncMode]string{
	SyncModeFull:     "full",
	SyncModeSnapshot: "snapshot",
}

type SyncParams struct {
	Peers            []string
	LatestBlockHash  string
	Quorum           uint64
	CurHeight        uint64
	TargetHeight     uint64
	QuorumCheckpoint *consensus.SignedCheckpoint
	EpochChanges     []*consensus.EpochChange
}

type InvalidMsg struct {
	NodeID string
	Height uint64
	ErrMsg error
	Typ    SyncMsgType
}

type WrapperStateResp struct {
	PeerID string
	Hash   string
	Resp   *pb.SyncStateResponse
}

type Chunk struct {
	ChunkSize  uint64
	CheckPoint *pb.CheckpointState
}

type Peer struct {
	PeerID       string
	TimeoutCount uint64
}

type PrepareData struct {
	Data any
}

type WrapEpochChange struct {
	Epcs  []*consensus.EpochChange
	Error error
}

type SnapCommitData struct {
	Data       []CommitData
	EpochState *consensus.QuorumCheckpoint
}

type syncRequestMessage interface {
	MarshalVT() (dAtA []byte, err error)
	UnmarshalVT(dAtA []byte) error
	GetHeight() uint64
}
type CommitData interface {
	GetParentHash() string
	GetHash() string
	GetHeight() uint64

	GetEpoch() uint64
}

var CommitDataRequestConstructor = map[SyncMode]syncRequestMessage{
	SyncModeFull:     &pb.SyncBlockRequest{},
	SyncModeSnapshot: &pb.SyncChainDataRequest{},
}

type BlockData struct {
	Block *types.Block
}

func (b *BlockData) GetParentHash() string {
	return b.Block.BlockHeader.ParentHash.String()
}

func (b *BlockData) GetHash() string {
	return b.Block.BlockHash.String()
}

func (b *BlockData) GetHeight() uint64 {
	return b.Block.Height()
}

func (b *BlockData) GetEpoch() uint64 {
	return b.Block.BlockHeader.Epoch
}

type ChainData struct {
	Block    *types.Block
	Receipts []*types.Receipt
}

func (c *ChainData) GetParentHash() string {
	return c.Block.BlockHeader.ParentHash.String()
}

func (c *ChainData) GetHash() string {
	return c.Block.BlockHash.String()
}

func (c *ChainData) GetHeight() uint64 {
	return c.Block.Height()
}

func (c *ChainData) GetEpoch() uint64 {
	return c.Block.BlockHeader.Epoch
}

func (c *Chunk) FillCheckPoint(chunkMaxHeight uint64, checkpoint *pb.CheckpointState) {
	if chunkMaxHeight >= checkpoint.Height {
		c.CheckPoint = checkpoint
	}
}
