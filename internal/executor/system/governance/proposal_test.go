package governance

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
)

func TestProposal_CheckAndUpdateState(t *testing.T) {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "proposal"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

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

		notFinishedProposalMgr.SetProposal(&NotFinishedProposal{
			ID:                  baseProposal.ID,
			DeadlineBlockNumber: baseProposal.BlockNumber,
			ContractAddr:        common.CouncilManagerContractAddr,
		})

		err = CheckAndUpdateState(testcase.DeadlineBlockNumber, stateLedger)
		assert.Equal(t, testcase.ExpectedErr, err)

		proposal, err := loadCouncilProposal(stateLedger, 1)
		assert.Nil(t, err)
		assert.Equal(t, testcase.ExpectedResult, proposal.Status)

		// clear
		account.SetState(notFinishedProposalsKey(), nil)
	}
}
