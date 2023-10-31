package block_sync

import (
	"context"
	"time"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	network "github.com/axiomesh/axiom-p2p"
)

type requester struct {
	peerID      string
	blockHeight uint64

	quitCh     chan struct{}
	retryCh    chan string      // retry peerID
	invalidCh  chan *invalidMsg // a channel which send invalid msg to syncer
	gotBlock   bool
	gotBlockCh chan struct{}

	syncBlockRequestPipe network.Pipe
	block                *types.Block

	ctx context.Context
}

func newRequester(ctx context.Context, peerID string, height uint64, invalidCh chan *invalidMsg, pipe network.Pipe) *requester {
	return &requester{
		peerID:               peerID,
		blockHeight:          height,
		invalidCh:            invalidCh,
		quitCh:               make(chan struct{}, 1),
		retryCh:              make(chan string, 1),
		gotBlockCh:           make(chan struct{}, 1),
		syncBlockRequestPipe: pipe,
		ctx:                  ctx,
	}
}

func (r *requester) start(requestRetryTimeout time.Duration) {
	go r.requestRoutine(requestRetryTimeout)
}

// Responsible for making more requests as necessary
// Returns only when a block is found (e.g. AddBlock() is called)
func (r *requester) requestRoutine(requestRetryTimeout time.Duration) {
OUTER_LOOP:
	for {
		r.gotBlock = false
		ticker := time.NewTicker(requestRetryTimeout)
		// Send request and wait
		req := &pb.SyncBlockRequest{Height: r.blockHeight}
		data, err := req.MarshalVT()
		if err != nil {
			r.invalidCh <- &invalidMsg{nodeID: r.peerID, height: r.blockHeight, errMsg: err, typ: syncMsgType_ErrorMsg}
		}
		if err = r.syncBlockRequestPipe.Send(r.ctx, r.peerID, data); err != nil {
			r.invalidCh <- &invalidMsg{nodeID: r.peerID, height: r.blockHeight, errMsg: err, typ: syncMsgType_ErrorMsg}
		}

	WAIT_LOOP:
		for {
			select {
			case <-r.quitCh:
				return
			case <-r.ctx.Done():
				return
			case newPeer := <-r.retryCh:
				// Retry the new peer
				r.peerID = newPeer
				continue OUTER_LOOP
			case <-ticker.C:
				if r.gotBlock {
					continue WAIT_LOOP
				}
				// Timeout
				r.invalidCh <- &invalidMsg{nodeID: r.peerID, height: r.blockHeight, typ: syncMsgType_TimeoutBlock}

			case <-r.gotBlockCh:
				// We got a block!
				// Continue the for-loop and wait til Quit.
				r.gotBlock = true
				continue WAIT_LOOP
			}
		}
	}
}

func (r *requester) setBlock(block *types.Block) {
	r.block = block

	select {
	case r.gotBlockCh <- struct{}{}:
	default:
	}
}

func (r *requester) clearBlock() {
	r.block = nil
}
