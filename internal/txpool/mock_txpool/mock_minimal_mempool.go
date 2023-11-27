package mock_txpool

import (
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
)

// NewMockMinimalTxPool returns a minimal implement of MockTxPool which accepts
// any kinds of input and returns 'zero value' as all outputs.
// Users can defines custom MockTxPool like this:
// func NewMockCustomTxPool(ctrl *gomock.Controller) *MockTxPool {...}
// in which users must specify output for all functions.
func NewMockMinimalTxPool[T any, Constraint types.TXConstraint[T]](ctrl *gomock.Controller) *MockTxPool[T, Constraint] {
	mock := NewMockTxPool[T, Constraint](ctrl)
	mock.EXPECT().GenerateRequestBatch(gomock.Any()).Return(nil, nil).AnyTimes()
	mock.EXPECT().AddLocalTx(gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().AddRemoteTxs(gomock.Any()).AnyTimes()
	mock.EXPECT().RemoveBatches(gomock.Any()).Return().AnyTimes()
	mock.EXPECT().IsPoolFull().Return(false).AnyTimes()
	mock.EXPECT().HasPendingRequestInPool().Return(false).AnyTimes()
	mock.EXPECT().RestoreOneBatch(gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().GetRequestsByHashList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, nil, nil).AnyTimes()
	mock.EXPECT().SendMissingRequests(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mock.EXPECT().ReceiveMissingRequests(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().FilterOutOfDateRequests(gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().RestorePool().Return().AnyTimes()
	mock.EXPECT().ReConstructBatchByOrder(gomock.Any()).Return(nil, nil).AnyTimes()
	mock.EXPECT().Stop().AnyTimes()
	mock.EXPECT().Start().Return(nil).AnyTimes()
	mock.EXPECT().Init(gomock.Any()).AnyTimes()
	return mock
}
