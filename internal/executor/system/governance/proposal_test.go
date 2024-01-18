package governance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
)

func TestProposal_ProposalID(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.ProposalIDContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	proposalID := NewProposalID(stateLedger)
	assert.EqualValues(t, 1, proposalID.GetID())
	id, err := proposalID.GetAndAddID()
	assert.Nil(t, err)
	assert.EqualValues(t, 1, id)
	assert.EqualValues(t, 2, proposalID.GetID())
}

func TestProposal_Addr2NameSystem(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.Addr2NameContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	addr2NameSystem := NewAddr2NameSystem(stateLedger)
	isExist, _ := addr2NameSystem.GetName("test")
	assert.False(t, isExist)
	addr2NameSystem.SetName("test", "name")
	isExist, name := addr2NameSystem.GetName("test")
	assert.True(t, isExist)
	assert.Equal(t, "name", name)

	isExist, addr := addr2NameSystem.GetAddr("name")
	assert.True(t, isExist)
	assert.Equal(t, "test", addr)
}

func TestProposal_NotFinishedProposalMgr(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NotFinishedProposalContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	notFinishedProposalMgr := NewNotFinishedProposalMgr(stateLedger)
	err := notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  10,
		DeadlineBlockNumber: 100,
	})
	assert.Nil(t, err)

	proposals, err := notFinishedProposalMgr.GetProposals()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(proposals))

	err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  30,
		DeadlineBlockNumber: 100,
	})
	assert.Nil(t, err)
	proposals, err = notFinishedProposalMgr.GetProposals()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(proposals))

	err = notFinishedProposalMgr.RemoveProposal(10)
	assert.Nil(t, err)
	proposals, err = notFinishedProposalMgr.GetProposals()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(proposals))

	err = notFinishedProposalMgr.RemoveProposal(50)
	assert.Equal(t, ErrNotFoundProposal, err)
}
