package diff_sync

import (
	"context"

	"github.com/axiomesh/axiom-ledger/internal/sync/common"
)

var _ common.ISyncConstructor = (*DiffSync)(nil)

type DiffSync struct {
	recvCommitDataCh chan []common.CommitData
	commitCh         chan any
	ctx              context.Context
	cancel           context.CancelFunc
}

func (s *DiffSync) Start() {
	go s.listenCommitData()
}

func (s *DiffSync) Stop() {
	s.cancel()
}

func (s *DiffSync) Mode() common.SyncMode {
	return common.SyncModeDiff
}

func (s *DiffSync) PostCommitData(data []common.CommitData) {
	s.recvCommitDataCh <- data
}

func (s *DiffSync) Commit() chan any {
	return s.commitCh
}

func NewDiffSync() common.ISyncConstructor {
	ctx, cancel := context.WithCancel(context.Background())
	return &DiffSync{
		recvCommitDataCh: make(chan []common.CommitData, 1),
		commitCh:         make(chan any, 1),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (s *DiffSync) Prepare(_ *common.Config) (*common.PrepareData, error) {
	return nil, nil
}

func (s *DiffSync) listenCommitData() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case data := <-s.recvCommitDataCh:
			s.commitCh <- data
		}
	}
}
