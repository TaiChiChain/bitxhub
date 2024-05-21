package governance

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/framework"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func initGovernance(t *testing.T, repoSetters ...func(rep *repo.Repo)) (*common.TestNVM, *Governance) {
	testNVM := common.NewTestNVM(t)
	testNVM.Rep.GenesisConfig.EpochInfo.FinanceParams.MinGasPrice = types.CoinNumberByAxc(0)
	testNVM.Rep.GenesisConfig.CouncilMembers = []*repo.CouncilMember{
		{
			Address: admin1.String(),
			Weight:  1,
			Name:    "111",
		},
		{
			Address: admin2.String(),
			Weight:  1,
			Name:    "222",
		},
		{
			Address: admin3.String(),
			Weight:  1,
			Name:    "333",
		},
		{
			Address: admin4.String(),
			Weight:  1,
			Name:    "444",
		},
	}
	for _, setter := range repoSetters {
		setter(testNVM.Rep)
	}
	gov := BuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	epochManager := framework.EpochManagerBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	testNVM.GenesisInit(epochManager, gov)
	return testNVM, gov
}

func TestGovernance_Propose(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testcases := []struct {
		input struct {
			user         string
			proposalType ProposalType
			title        string
			desc         string
			BlockNumber  uint64
		}
		err    error
		hasErr bool
	}{
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         admin1.String(),
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			hasErr: false,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: ProposalType(250),
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			err: ErrProposalType,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "",
				desc:         "test desc",
				BlockNumber:  10000,
			},
			err: ErrTitle,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "",
				BlockNumber:  10000,
			},
			err: ErrDesc,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  0,
			},
			err: ErrBlockNumber,
		},
		{
			input: struct {
				user         string
				proposalType ProposalType
				title        string
				desc         string
				BlockNumber  uint64
			}{
				user:         "0x1000000000000000000000000000000000000000",
				proposalType: NodeUpgrade,
				title:        "test title",
				desc:         "test desc",
				BlockNumber:  2,
			},
			err: ErrNotFoundCouncilMember,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, ethcommon.HexToAddress(test.input.user), func() error {
				arg := NodeUpgradeExtraArgs{
					DownloadUrls: []string{"http://127.0.0.1:10000"},
					CheckHash:    "",
				}
				data, err := json.Marshal(arg)
				assert.Nil(t, err)

				err = gov.Propose(uint8(test.input.proposalType), test.input.title, test.input.desc, test.input.BlockNumber, data)
				if test.err != nil {
					assert.Contains(t, err.Error(), test.err.Error())
				} else {
					if test.hasErr {
						assert.NotNil(t, err)
					} else {
						assert.Nil(t, err)
					}
				}
				return err
			})
		})
	}
}

func TestGovernance_Vote(t *testing.T) {
	testNVM, gov := initGovernance(t)

	addr := types.NewAddressByStr("0x10010000000").ETHAddress()

	testNVM.RunSingleTX(gov, admin1, func() error {
		arg := NodeUpgradeExtraArgs{
			DownloadUrls: []string{"http://127.0.0.1:10000"},
			CheckHash:    "",
		}
		data, err := json.Marshal(arg)
		assert.Nil(t, err)
		err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
		assert.Nil(t, err)
		return err
	})

	testcases := []struct {
		from   ethcommon.Address
		id     uint64
		result uint8
		err    error
		hasErr bool
	}{
		{
			from:   admin2,
			id:     100,
			result: uint8(Reject),
			err:    ErrNotFoundProposal,
		},
		{
			from:   admin2,
			id:     1,
			result: uint8(100),
			err:    ErrVoteResult,
		},
		{
			from:   admin1,
			id:     1,
			result: uint8(Pass),
			err:    ErrUseHasVoted,
		},
		{
			from:   addr,
			id:     1,
			result: uint8(Pass),
			err:    ErrNotFoundCouncilMember,
		},
		{
			// vote success
			from:   admin2,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			from:   admin3,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			// has finish the proposal
			from:   admin4,
			id:     1,
			result: uint8(Pass),
			err:    ErrProposalFinished,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.from, func() error {
				err := gov.Vote(test.id, test.result)
				if test.err != nil {
					assert.Contains(t, err.Error(), test.err.Error())
				} else {
					if test.hasErr {
						assert.NotNil(t, err)
					} else {
						assert.Nil(t, err)
					}
				}
				return err
			})
		})
	}
}

func TestGovernance_GetLatestProposalID(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.Call(gov, admin1, func() {
		proposalID, err := gov.GetLatestProposalID()
		assert.Nil(t, err)
		assert.EqualValues(t, 0, proposalID)
	})

	testNVM.RunSingleTX(gov, admin1, func() error {
		arg := NodeUpgradeExtraArgs{
			DownloadUrls: []string{"http://127.0.0.1:10000"},
			CheckHash:    "",
		}
		data, err := json.Marshal(arg)
		assert.Nil(t, err)
		err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(gov, admin1, func() {
		proposalID, err := gov.GetLatestProposalID()
		assert.Nil(t, err)
		assert.EqualValues(t, 1, proposalID)
	})
}

func TestGovernance_Proposal(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.RunSingleTX(gov, admin1, func() error {
		arg := NodeUpgradeExtraArgs{
			DownloadUrls: []string{"http://127.0.0.1:10000"},
			CheckHash:    "",
		}
		data, err := json.Marshal(arg)
		assert.Nil(t, err)

		err = gov.Propose(uint8(NodeUpgrade), "test title", "test desc", uint64(10000), data)
		assert.Nil(t, err)
		return err
	})

	testNVM.Call(gov, admin1, func() {
		proposal, err := gov.Proposal(1)
		assert.Nil(t, err)
		assert.EqualValues(t, Voting, proposal.Status)
		assert.EqualValues(t, 1, proposal.CreatedBlockNumber)
	})

	testcases := []struct {
		from   ethcommon.Address
		id     uint64
		result uint8
		err    error
		hasErr bool
	}{
		{
			// vote success
			from:   admin2,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
		{
			from:   admin3,
			id:     1,
			result: uint8(Pass),
			hasErr: false,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.from, func() error {
				gov.Ctx.BlockNumber = 2
				err := gov.Vote(test.id, test.result)
				if test.err != nil {
					assert.Contains(t, err.Error(), test.err.Error())
				} else {
					if test.hasErr {
						assert.NotNil(t, err)
					} else {
						assert.Nil(t, err)
					}
				}
				return err
			})
		})
	}

	testNVM.Call(gov, admin1, func() {
		proposal, err := gov.Proposal(1)
		assert.Nil(t, err)
		assert.EqualValues(t, Approved, proposal.Status)
		assert.EqualValues(t, 1, proposal.CreatedBlockNumber)
		assert.EqualValues(t, 2, proposal.EffectiveBlockNumber)
	})
}
