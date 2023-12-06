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

func TestProposal_CheckAndUpdateState(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	testcases := []struct {
		BlockNumber         uint64
		Status              ProposalStatus
		DeadlineBlockNumber uint64
		ExpectedResult      ProposalStatus
		ExpectedErr         error
	}{
		{
			BlockNumber:         0,
			Status:              Voting,
			DeadlineBlockNumber: 1,
			ExpectedResult:      Rejected,
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         2,
			Status:              Voting,
			DeadlineBlockNumber: 1,
			ExpectedResult:      Voting,
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         1,
			Status:              Voting,
			DeadlineBlockNumber: 1,
			ExpectedResult:      Rejected,
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         1,
			Status:              Approved,
			DeadlineBlockNumber: 1,
			ExpectedResult:      Approved,
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         1,
			Status:              Rejected,
			DeadlineBlockNumber: 1,
			ExpectedResult:      Rejected,
			ExpectedErr:         nil,
		},
	}

	notFinishedProposalMgr := NewNotFinishedProposalMgr(stateLedger)

	for _, testcase := range testcases {
		baseProposal := &BaseProposal{ID: 1, BlockNumber: testcase.BlockNumber, Status: testcase.Status}

		_, err := saveCouncilProposal(stateLedger, baseProposal)
		assert.Nil(t, err)

		err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
			ID:                  baseProposal.ID,
			DeadlineBlockNumber: baseProposal.BlockNumber,
			ContractAddr:        common.CouncilManagerContractAddr,
		})
		assert.Nil(t, err)

		err = CheckAndUpdateState(testcase.DeadlineBlockNumber, stateLedger)
		assert.Equal(t, testcase.ExpectedErr, err)

		proposal, err := loadCouncilProposal(stateLedger, 1)
		assert.Nil(t, err)
		assert.Equal(t, testcase.ExpectedResult, proposal.Status)

		// clear
		account.SetState(notFinishedProposalsKey(), nil)
	}
}

func TestProposal_CheckAndUpdateState_AllProposals(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.NodeManagerContractAddr))

	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	stateLedger.EXPECT().SetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	stateLedger.EXPECT().AddLog(gomock.Any()).AnyTimes()

	notFinishedProposalMgr := NewNotFinishedProposalMgr(stateLedger)

	// test council
	baseProposal := &BaseProposal{ID: 1, BlockNumber: 10, Status: Voting}
	_, err := saveCouncilProposal(stateLedger, baseProposal)
	assert.Nil(t, err)

	err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  baseProposal.ID,
		DeadlineBlockNumber: baseProposal.BlockNumber,
		ContractAddr:        common.CouncilManagerContractAddr,
	})
	assert.Nil(t, err)

	err = CheckAndUpdateState(20, stateLedger)
	assert.Nil(t, err)

	proposal, err := loadCouncilProposal(stateLedger, baseProposal.ID)
	assert.Nil(t, err)
	assert.Equal(t, Rejected, proposal.Status)

	// test node
	baseProposal = &BaseProposal{ID: 2, BlockNumber: 10, Status: Voting}
	_, err = saveNodeProposal(stateLedger, baseProposal)
	assert.Nil(t, err)

	err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  baseProposal.ID,
		DeadlineBlockNumber: baseProposal.BlockNumber,
		ContractAddr:        common.NodeManagerContractAddr,
	})
	assert.Nil(t, err)

	err = CheckAndUpdateState(20, stateLedger)
	assert.Nil(t, err)

	nodeProposal, err := loadNodeProposal(stateLedger, baseProposal.ID)
	assert.Nil(t, err)
	assert.Equal(t, Rejected, nodeProposal.Status)

	// test white list provider
	baseProposal = &BaseProposal{ID: 3, BlockNumber: 10, Status: Voting}
	_, err = saveProviderProposal(stateLedger, baseProposal)
	assert.Nil(t, err)

	err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  baseProposal.ID,
		DeadlineBlockNumber: baseProposal.BlockNumber,
		ContractAddr:        common.WhiteListProviderManagerContractAddr,
	})
	assert.Nil(t, err)

	err = CheckAndUpdateState(20, stateLedger)
	assert.Nil(t, err)

	providerProposal, err := loadProviderProposal(stateLedger, baseProposal.ID)
	assert.Nil(t, err)
	assert.Equal(t, Rejected, providerProposal.Status)

	// test gas
	baseProposal = &BaseProposal{ID: 4, BlockNumber: 10, Status: Voting}
	_, err = saveProviderProposal(stateLedger, baseProposal)
	assert.Nil(t, err)

	err = notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
		ID:                  baseProposal.ID,
		DeadlineBlockNumber: baseProposal.BlockNumber,
		ContractAddr:        common.GasManagerContractAddr,
	})
	assert.Nil(t, err)

	err = CheckAndUpdateState(20, stateLedger)
	assert.Nil(t, err)

	gasProposal, err := loadProviderProposal(stateLedger, baseProposal.ID)
	assert.Nil(t, err)
	assert.Equal(t, Rejected, gasProposal.Status)
}
