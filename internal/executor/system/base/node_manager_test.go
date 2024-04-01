package base

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newMockLedger(t *testing.T) ledger.StateLedger {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(2, types.NewAddressByStr(common.GovernanceContractAddr))
	account.SetBalance(big.NewInt(3000000000000000000))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(0).AnyTimes()
	stateLedger.EXPECT().Snapshot().AnyTimes()
	stateLedger.EXPECT().Commit().Return(types.NewHash([]byte("")), nil).AnyTimes()
	stateLedger.EXPECT().Clear().AnyTimes()
	stateLedger.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
	stateLedger.EXPECT().SetNonce(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Finalise().AnyTimes()
	stateLedger.EXPECT().Snapshot().Return(0).AnyTimes()
	stateLedger.EXPECT().RevertToSnapshot(0).AnyTimes()
	stateLedger.EXPECT().SetTxContext(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetLogs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	stateLedger.EXPECT().PrepareBlock(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetBalance(gomock.Any()).Return(big.NewInt(3000000000000000000)).AnyTimes()
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().SubBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetCodeHash(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().Exist(gomock.Any()).AnyTimes()
	stateLedger.EXPECT().GetRefund().AnyTimes()
	stateLedger.EXPECT().GetCode(gomock.Any()).AnyTimes()
	return stateLedger
}

func TestNodeManager_LifeCycleOfNode(t *testing.T) {
	mockLedger := newMockLedger(t)
	mockNodeManager := NewNodeManager(nil)
	mockAccount := types.NewAddressByStr("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013").ETHAddress()
	epochAccount := types.NewAddressByStr(common.EpochManagerContractAddr).ETHAddress()
	accountCtx := &common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
		CurrentUser:   &mockAccount,
	}
	epochCtx := &common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
		CurrentUser:   &epochAccount,
	}
	mockNodeManager.SetContext(accountCtx)

	nodeId, err := mockNodeManager.InternalRegisterNode(NodeInfo{
		NodePubKey: "123",
		MetaData: NodeMetaData{
			Name:       "mockName",
			Desc:       "mockDesc",
			ImageURL:   "https://example.com/image.png",
			WebsiteURL: "https://example.com/",
		},
		OperatorAddress: mockAccount.String(),
	})
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(epochCtx)
	nodeId, err = mockNodeManager.InternalRegisterNode(NodeInfo{
		NodePubKey: "123",
		MetaData: NodeMetaData{
			Name:       "mockName",
			Desc:       "mockDesc",
			ImageURL:   "https://example.com/image.png",
			WebsiteURL: "https://example.com/",
		},
		OperatorAddress: mockAccount.String(),
	})
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nodeId)

	info, err := mockNodeManager.GetNodeInfo(nodeId)
	assert.Nil(t, err)
	assert.Equal(t, "123", info.NodePubKey)
	assert.Equal(t, StatusSyncing, info.Status)
	assert.Equal(t, "mockName", info.MetaData.Name)
	assert.Equal(t, mockAccount.String(), info.OperatorAddress)
	dataSyncingSet, err := mockNodeManager.GetDataSyncerSet()
	assert.Nil(t, err)
	assert.Equal(t, []NodeInfo{info}, dataSyncingSet)

	err = mockNodeManager.JoinCandidateSet(nodeId)
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(accountCtx)
	err = mockNodeManager.JoinCandidateSet(nodeId)
	assert.Nil(t, err)
	dataSyncingSet, err = mockNodeManager.GetDataSyncerSet()
	assert.Nil(t, err)
	assert.Equal(t, []NodeInfo(nil), dataSyncingSet)
	candidateSet, err := mockNodeManager.GetCandidateSet()
	assert.Nil(t, err)
	info.Status = StatusCandidate
	assert.Equal(t, []NodeInfo{info}, candidateSet)
	candidates, err := mockNodeManager.InternalGetConsensusCandidateNodeIDs()
	assert.Nil(t, err)
	assert.Equal(t, []uint64{nodeId}, candidates)

	votingPowers := ConsensusVotingPower{
		NodeID:               nodeId,
		ConsensusVotingPower: 10,
	}
	err = mockNodeManager.InternalUpdateActiveValidatorSet([]ConsensusVotingPower{votingPowers})
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(epochCtx)
	err = mockNodeManager.InternalUpdateActiveValidatorSet([]ConsensusVotingPower{votingPowers})
	assert.Nil(t, err)
	candidateSet, err = mockNodeManager.GetCandidateSet()
	assert.Nil(t, err)
	assert.Equal(t, []NodeInfo(nil), candidateSet)
	activeSet, returnVotingPowers, err := mockNodeManager.GetActiveValidatorSet()
	assert.Nil(t, err)
	info.Status = StatusActive
	assert.Equal(t, []NodeInfo{info}, activeSet)
	assert.Equal(t, []ConsensusVotingPower{votingPowers}, returnVotingPowers)

	err = mockNodeManager.LeaveValidatorSet(nodeId)
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(accountCtx)
	err = mockNodeManager.LeaveValidatorSet(nodeId)
	assert.Nil(t, err)
	activeSet, _, err = mockNodeManager.GetActiveValidatorSet()
	assert.Nil(t, err)
	assert.Equal(t, []NodeInfo(nil), activeSet)
	pendingInactiveSet, err := mockNodeManager.GetPendingInactiveSet()
	assert.Nil(t, err)
	info.Status = StatusPendingInactive
	assert.Equal(t, []NodeInfo{info}, pendingInactiveSet)

	err = mockNodeManager.InternalProcessNodeLeave()
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(epochCtx)
	err = mockNodeManager.InternalProcessNodeLeave()
	assert.Nil(t, err)
	pendingInactiveSet, err = mockNodeManager.GetPendingInactiveSet()
	assert.Nil(t, err)
	assert.Equal(t, []NodeInfo(nil), pendingInactiveSet)
	exitedSet, err := mockNodeManager.GetExitedSet()
	assert.Nil(t, err)
	info.Status = StatusExited
	assert.Equal(t, []NodeInfo{info}, exitedSet)

	totalNodesNum := mockNodeManager.GetTotalNodeCount()
	assert.Equal(t, 1, totalNodesNum)
}

func TestNodeManager_UpdateInfo(t *testing.T) {
	mockLedger := newMockLedger(t)
	mockNodeManager := NewNodeManager(nil)
	mockAccount := types.NewAddressByStr("0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013").ETHAddress()
	epochAccount := types.NewAddressByStr(common.EpochManagerContractAddr).ETHAddress()
	newAccount := types.NewAddressByStr(common.ZeroAddress).ETHAddress()
	accountCtx := &common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
		CurrentUser:   &mockAccount,
	}
	newAccountCtx := &common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
		CurrentUser:   &newAccount,
	}
	epochCtx := &common.VMContext{
		StateLedger:   mockLedger,
		CurrentHeight: 100,
		CurrentUser:   &epochAccount,
	}
	mockNodeManager.SetContext(epochCtx)

	nodeId, err := mockNodeManager.InternalRegisterNode(NodeInfo{
		NodePubKey: "123",
		MetaData: NodeMetaData{
			Name:       "mockName",
			Desc:       "mockDesc",
			ImageURL:   "https://example.com/image.png",
			WebsiteURL: "https://example.com/",
		},
		OperatorAddress: mockAccount.String(),
	})
	assert.Nil(t, err)
	err = mockNodeManager.UpdateOperator(nodeId, common.ZeroAddress)
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(accountCtx)
	err = mockNodeManager.UpdateOperator(nodeId, common.ZeroAddress)
	assert.Nil(t, err)
	info, err := mockNodeManager.GetNodeInfo(nodeId)
	assert.Nil(t, err)
	assert.Equal(t, common.ZeroAddress, info.OperatorAddress)

	err = mockNodeManager.UpdateMetaData(nodeId, NodeMetaData{})
	assert.EqualError(t, ErrPermissionDenied, err.Error())
	mockNodeManager.SetContext(newAccountCtx)
	err = mockNodeManager.UpdateMetaData(nodeId, NodeMetaData{
		Name:       "newName",
		Desc:       "newDesc",
		ImageURL:   "https://example.com/image.png",
		WebsiteURL: "https://example.com/",
	})
	assert.Nil(t, err)
	info, err = mockNodeManager.GetNodeInfo(nodeId)
	assert.Nil(t, err)
	assert.Equal(t, "newName", info.MetaData.Name)
}
