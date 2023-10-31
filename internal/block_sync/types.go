package block_sync

import (
	"time"

	"github.com/axiomesh/axiom-kit/types/pb"
)

const (
	syncBlockRequestPipe  = "sync_block_pipe_v1_request"
	syncBlockResponsePipe = "sync_block_pipe_v1_response"

	maxRetryCount   = 5
	requestInterval = 500 * time.Millisecond
)

type syncMsgType int

const (
	syncMsgType_InvalidBlock syncMsgType = iota
	syncMsgType_TimeoutBlock
	syncMsgType_ErrorMsg
)

type invalidMsg struct {
	nodeID string
	height uint64
	errMsg error
	typ    syncMsgType
}

type wrapperStateResp struct {
	peerID string
	hash   string
	resp   *pb.SyncStateResponse
}

type chunk struct {
	chunkSize  uint64
	checkPoint *pb.CheckpointState
}

type peer struct {
	peerID       string
	timeoutCount uint64
}

func (c *chunk) fillCheckPoint(chunkMaxHeight uint64, checkpoint *pb.CheckpointState) {
	if chunkMaxHeight >= checkpoint.Height {
		c.checkPoint = checkpoint
	}
}
