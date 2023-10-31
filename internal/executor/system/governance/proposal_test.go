package governance

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

func TestProposal_CheckAndUpdateState(t *testing.T) {
	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	assert.Nil(t, err)
	ld, err := leveldb.New(filepath.Join(repoRoot, "proposal"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.NodeManagerContractAddr), ledger.NewChanger())

	updateErr := errors.New("update error")
	testcases := []struct {
		BlockNumber         uint64
		Status              ProposalStatus
		DeadlineBlockNumber uint64
		SaveProposal        func(proposal *BaseProposal) ([]byte, error)
		ExpectedErr         error
	}{
		{
			BlockNumber:         0,
			Status:              Voting,
			DeadlineBlockNumber: 1,
			SaveProposal:        func(proposal *BaseProposal) ([]byte, error) { return nil, nil },
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         1,
			Status:              Voting,
			DeadlineBlockNumber: 1,
			SaveProposal:        func(proposal *BaseProposal) ([]byte, error) { return nil, nil },
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         2,
			Status:              Voting,
			DeadlineBlockNumber: 2,
			SaveProposal:        func(proposal *BaseProposal) ([]byte, error) { return nil, updateErr },
			ExpectedErr:         updateErr,
		},
		{
			BlockNumber:         1,
			Status:              Approved,
			DeadlineBlockNumber: 1,
			SaveProposal:        func(proposal *BaseProposal) ([]byte, error) { return nil, updateErr },
			ExpectedErr:         nil,
		},
		{
			BlockNumber:         1,
			Status:              Rejected,
			DeadlineBlockNumber: 1,
			SaveProposal:        func(proposal *BaseProposal) ([]byte, error) { return nil, updateErr },
			ExpectedErr:         nil,
		},
	}

	for _, testcase := range testcases {
		baseProposal := &BaseProposal{BlockNumber: testcase.BlockNumber, Status: testcase.Status}
		b, err := json.Marshal(baseProposal)
		assert.Nil(t, err)
		account.SetState([]byte("/proposalkey"), b)

		err = CheckAndUpdateState[BaseProposal, *BaseProposal](testcase.DeadlineBlockNumber, account, "/proposalkey", testcase.SaveProposal)
		assert.Equal(t, testcase.ExpectedErr, err)

		account.SetState([]byte("/proposalkey"), nil)
	}
}
