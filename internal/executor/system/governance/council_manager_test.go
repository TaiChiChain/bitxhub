package governance

import (
	"encoding/json"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	ErrNotFoundCouncilMember = errors.New("no permission")

	admin1 = ethcommon.HexToAddress("0x1210000000000000000000000000000000000000")
	admin2 = ethcommon.HexToAddress("0x1220000000000000000000000000000000000000")
	admin3 = ethcommon.HexToAddress("0x1230000000000000000000000000000000000000")
	admin4 = ethcommon.HexToAddress("0x1240000000000000000000000000000000000000")
)

func TestRunForPropose(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testcases := []struct {
		Caller   ethcommon.Address
		ExtraArg *Candidates
		Err      error
		HasErr   bool
	}{
		{
			Caller:   admin1,
			ExtraArg: nil,
			HasErr:   true,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrMinCouncilMembersCount,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
					{
						Address: admin1.String(),
						Weight:  1,
						Name:    "111",
					},
					{
						Address: admin1.String(),
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
				},
			},
			Err: ErrRepeatedAddress,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
						Name:    "333",
					},
				},
			},
			Err: ErrRepeatedName,
		},
		{
			Caller: ethcommon.Address{},
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrNotFoundCouncilMember,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			HasErr: false,
		},
		{
			Caller: admin1,
			ExtraArg: &Candidates{
				Candidates: []CouncilMember{
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
				},
			},
			Err: ErrExistNotFinishedProposal,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				data, err := json.Marshal(test.ExtraArg)
				assert.Nil(t, err)

				err = gov.Propose(uint8(CouncilElect), "test", "test desc", 100, data)
				if test.Err != nil {
					assert.Contains(t, err.Error(), test.Err.Error())
				} else {
					if test.HasErr {
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

func TestVoteExecute(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(Candidates{
			Candidates: []CouncilMember{
				{
					Address: admin1.String(),
					Weight:  2,
					Name:    "111",
				},
				{
					Address: admin2.String(),
					Weight:  2,
					Name:    "222",
				},
				{
					Address: admin3.String(),
					Weight:  2,
					Name:    "333",
				},
				{
					Address: admin4.String(),
					Weight:  2,
					Name:    "444",
				},
			},
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(CouncilElect), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	var latestProposalID uint64
	testNVM.Call(gov, admin1, func() {
		var err error
		latestProposalID, err = gov.GetLatestProposalID()
		assert.Nil(t, err)
	})

	testcases := []struct {
		Caller     ethcommon.Address
		ProposalID uint64
		Res        VoteResult
		Err        error
	}{
		{
			Caller:     admin2,
			ProposalID: latestProposalID,
			Res:        Pass,
			Err:        nil,
		},
		{
			Caller:     admin3,
			ProposalID: latestProposalID,
			Res:        Pass,
			Err:        nil,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(test.ProposalID, uint8(test.Res))
				assert.Equal(t, test.Err, err)
				return err
			})
		})
	}
}
