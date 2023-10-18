// Code generated by MockGen. DO NOT EDIT.
// Source: precheck.go
//
// Generated by this command:
//
//	mockgen -destination mock_precheck/mock_precheck.go -package mock_precheck -source precheck.go -typed
//

// Package mock_precheck is a generated GoMock package.
package mock_precheck

import (
	reflect "reflect"

	common "github.com/axiomesh/axiom-ledger/internal/consensus/common"
	precheck "github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	gomock "go.uber.org/mock/gomock"
)

// MockPreCheck is a mock of PreCheck interface.
type MockPreCheck struct {
	ctrl     *gomock.Controller
	recorder *MockPreCheckMockRecorder
}

// MockPreCheckMockRecorder is the mock recorder for MockPreCheck.
type MockPreCheckMockRecorder struct {
	mock *MockPreCheck
}

// NewMockPreCheck creates a new mock instance.
func NewMockPreCheck(ctrl *gomock.Controller) *MockPreCheck {
	mock := &MockPreCheck{ctrl: ctrl}
	mock.recorder = &MockPreCheckMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPreCheck) EXPECT() *MockPreCheckMockRecorder {
	return m.recorder
}

// CommitValidTxs mocks base method.
func (m *MockPreCheck) CommitValidTxs() chan *precheck.ValidTxs {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitValidTxs")
	ret0, _ := ret[0].(chan *precheck.ValidTxs)
	return ret0
}

// CommitValidTxs indicates an expected call of CommitValidTxs.
func (mr *MockPreCheckMockRecorder) CommitValidTxs() *PreCheckCommitValidTxsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitValidTxs", reflect.TypeOf((*MockPreCheck)(nil).CommitValidTxs))
	return &PreCheckCommitValidTxsCall{Call: call}
}

// PreCheckCommitValidTxsCall wrap *gomock.Call
type PreCheckCommitValidTxsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *PreCheckCommitValidTxsCall) Return(arg0 chan *precheck.ValidTxs) *PreCheckCommitValidTxsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *PreCheckCommitValidTxsCall) Do(f func() chan *precheck.ValidTxs) *PreCheckCommitValidTxsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *PreCheckCommitValidTxsCall) DoAndReturn(f func() chan *precheck.ValidTxs) *PreCheckCommitValidTxsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// PostUncheckedTxEvent mocks base method.
func (m *MockPreCheck) PostUncheckedTxEvent(ev *common.UncheckedTxEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PostUncheckedTxEvent", ev)
}

// PostUncheckedTxEvent indicates an expected call of PostUncheckedTxEvent.
func (mr *MockPreCheckMockRecorder) PostUncheckedTxEvent(ev any) *PreCheckPostUncheckedTxEventCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostUncheckedTxEvent", reflect.TypeOf((*MockPreCheck)(nil).PostUncheckedTxEvent), ev)
	return &PreCheckPostUncheckedTxEventCall{Call: call}
}

// PreCheckPostUncheckedTxEventCall wrap *gomock.Call
type PreCheckPostUncheckedTxEventCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *PreCheckPostUncheckedTxEventCall) Return() *PreCheckPostUncheckedTxEventCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *PreCheckPostUncheckedTxEventCall) Do(f func(*common.UncheckedTxEvent)) *PreCheckPostUncheckedTxEventCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *PreCheckPostUncheckedTxEventCall) DoAndReturn(f func(*common.UncheckedTxEvent)) *PreCheckPostUncheckedTxEventCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Start mocks base method.
func (m *MockPreCheck) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockPreCheckMockRecorder) Start() *PreCheckStartCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockPreCheck)(nil).Start))
	return &PreCheckStartCall{Call: call}
}

// PreCheckStartCall wrap *gomock.Call
type PreCheckStartCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *PreCheckStartCall) Return() *PreCheckStartCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *PreCheckStartCall) Do(f func()) *PreCheckStartCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *PreCheckStartCall) DoAndReturn(f func()) *PreCheckStartCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
