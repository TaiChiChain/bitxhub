package full_sync

import (
	"context"

	"github.com/axiomesh/axiom-ledger/internal/sync/common"
)

var _ common.ISyncConstructor = (*FullSync)(nil)

type FullSync struct {
	recvCommitDataCh chan []common.CommitData
	commitCh         chan any
	ctx              context.Context
}

func (s *FullSync) Mode() common.SyncMode {
	return common.SyncModeFull
}

func (s *FullSync) PostCommitData(data []common.CommitData) {
	s.recvCommitDataCh <- data
}

func (s *FullSync) Commit() chan any {
	return s.commitCh
}

func NewFullSync(ctx context.Context) common.ISyncConstructor {
	return &FullSync{
		recvCommitDataCh: make(chan []common.CommitData, 100),
		commitCh:         make(chan any, 100),
		ctx:              ctx,
	}
}

func (s *FullSync) Prepare(_ *common.Config) (*common.PrepareData, error) {
	go s.listenCommitData()
	return nil, nil
}

func (s *FullSync) listenCommitData() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case data := <-s.recvCommitDataCh:
			s.commitCh <- data
		}
	}
}
