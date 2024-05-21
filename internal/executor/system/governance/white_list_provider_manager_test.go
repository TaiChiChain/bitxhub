package governance

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-ledger/internal/executor/system/access"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
)

var (
	WhiteListProvider1 = "0x1000000000000000000000000000000000000001"
	WhiteListProvider2 = "0x1000000000000000000000000000000000000002"
)

func generateProviderProposeData(t *testing.T, extraArgs WhitelistProviderArgs) []byte {
	data, err := json.Marshal(extraArgs)
	assert.Nil(t, err)

	return data
}

func TestWhiteListProviderManager_RunForPropose(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testcases := []struct {
		Caller ethcommon.Address
		Data   []byte
		Err    error
		HasErr bool
	}{
		{
			Caller: ethcommon.HexToAddress(WhiteListProvider1),
			Data: generateProviderProposeData(t, WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{
					{
						Addr: WhiteListProvider1,
					},
					{
						Addr: WhiteListProvider2,
					},
				},
			}),
			Err: ErrNotFoundCouncilMember,
		},
		{
			Caller: admin1,
			Data:   []byte{0, 1, 2, 3},
			Err:    errors.New("unmarshal provider extra arguments error"),
		},
		{
			Caller: admin1,
			Data: generateProviderProposeData(t, WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{
					{
						Addr: WhiteListProvider1,
					},
					{
						Addr: WhiteListProvider2,
					},
				},
			}),
			HasErr: false,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Propose(uint8(WhitelistProviderAdd), "test", "test desc", 100, test.Data)
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

func TestWhiteListProviderManager_RunForVoteAdd(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.RunSingleTX(gov, admin1, func() error {
		args, err := json.Marshal(WhitelistProviderArgs{
			Providers: []WhitelistProviderInfo{
				{
					Addr: WhiteListProvider1,
				},
				{
					Addr: WhiteListProvider2,
				},
			},
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(WhitelistProviderAdd), "test title", "test desc", 100, args)
		assert.Nil(t, err)
		return err
	})

	var proposalID uint64
	testNVM.Call(gov, admin1, func() {
		var err error
		proposalID, err = gov.GetLatestProposalID()
		assert.Nil(t, err)
	})

	testcases := []struct {
		Caller     ethcommon.Address
		ProposalID uint64
		Res        uint8
		Err        error
		HasErr     bool
	}{
		{
			Caller:     ethcommon.HexToAddress("0xfff0000000000000000000000000000000000000"),
			ProposalID: proposalID,
			Res:        uint8(Pass),
			Err:        ErrNotFoundCouncilMember,
		},
		{
			Caller:     admin2,
			ProposalID: proposalID,
			Res:        uint8(Pass),
			HasErr:     false,
		},
		{
			Caller:     admin3,
			ProposalID: proposalID,
			Res:        uint8(Pass),
			HasErr:     false,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(test.ProposalID, test.Res)
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

func TestWhiteListProviderManager_RunForVoteRemove(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.Rep.GenesisConfig.WhitelistProviders = []string{WhiteListProvider1, WhiteListProvider2, admin1.String(), admin2.String(), admin3.String()}
	whitelistContract := access.WhitelistBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	err := whitelistContract.GenesisInit(testNVM.Rep.GenesisConfig)
	assert.Nil(t, err)

	testNVM.RunSingleTX(gov, admin1, func() error {
		args, err := json.Marshal(WhitelistProviderArgs{
			Providers: []WhitelistProviderInfo{
				{
					Addr: WhiteListProvider1,
				},
				{
					Addr: WhiteListProvider2,
				},
			},
		})
		assert.Nil(t, err)

		err = gov.Propose(uint8(WhitelistProviderRemove), "test title", "test desc", 100, args)
		assert.Nil(t, err)
		return err
	})

	var proposalID uint64
	testNVM.Call(gov, admin1, func() {
		var err error
		proposalID, err = gov.GetLatestProposalID()
		assert.Nil(t, err)
	})

	testcases := []struct {
		Caller     ethcommon.Address
		ProposalID uint64
		Res        uint8
		Err        error
		HasErr     bool
	}{
		{
			Caller:     admin2,
			ProposalID: proposalID,
			Res:        uint8(Pass),
			HasErr:     false,
		},
		{
			Caller:     admin3,
			ProposalID: proposalID,
			Res:        uint8(Pass),
			HasErr:     false,
		},
	}

	for i, test := range testcases {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			testNVM.RunSingleTX(gov, test.Caller, func() error {
				err := gov.Vote(test.ProposalID, test.Res)
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

func TestWhiteListProviderManager_propose(t *testing.T) {
	testNVM, gov := initGovernance(t)

	testNVM.Rep.GenesisConfig.WhitelistProviders = []string{admin1.String(), admin2.String(), admin3.String()}
	whitelistContract := access.WhitelistBuildConfig.Build(common.NewTestVMContext(testNVM.StateLedger, ethcommon.Address{}))
	err := whitelistContract.GenesisInit(testNVM.Rep.GenesisConfig)
	assert.Nil(t, err)

	testcases := []struct {
		from     ethcommon.Address
		ptype    ProposalType
		args     *WhitelistProviderArgs
		expected error
	}{
		{
			from:  admin1,
			ptype: WhitelistProviderAdd,
			args: &WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{
					{
						Addr: admin1.String(),
					},
				},
			},
			expected: fmt.Errorf("provider already exists, %s", admin1),
		},
		{
			from:  admin2,
			ptype: WhitelistProviderRemove,
			args: &WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{
					{
						Addr: "0x0000000000000000000000000000000000000000",
					},
				},
			},
			expected: fmt.Errorf("provider does not exist, %s", "0x0000000000000000000000000000000000000000"),
		},
		{
			from:  admin3,
			ptype: WhitelistProviderAdd,
			args: &WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{},
			},
			expected: errors.New("empty providers"),
		},
		{
			from:  admin3,
			ptype: WhitelistProviderAdd,
			args: &WhitelistProviderArgs{
				Providers: []WhitelistProviderInfo{
					{
						Addr: "0x0000000000000000000000000000000000000000",
					},
					{
						Addr: "0x0000000000000000000000000000000000000000",
					},
				},
			},
			expected: errors.New("provider address repeated"),
		},
	}

	for _, testcase := range testcases {
		testNVM.RunSingleTX(gov, testcase.from, func() error {
			data, err := json.Marshal(testcase.args)
			assert.Nil(t, err)

			err = gov.Propose(uint8(testcase.ptype), "test", "test desc", 100, data)
			if testcase.expected == nil {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), testcase.expected.Error())
			}
			return err
		})
	}

	// test unfinished proposal
	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(WhitelistProviderArgs{
			Providers: []WhitelistProviderInfo{
				{
					Addr: WhiteListProvider1,
				},
			},
		})
		assert.Nil(t, err)
		err = gov.Propose(uint8(WhitelistProviderAdd), "test", "test desc", 100, data)
		assert.Nil(t, err)
		return err
	})

	testNVM.RunSingleTX(gov, admin1, func() error {
		data, err := json.Marshal(WhitelistProviderArgs{
			Providers: []WhitelistProviderInfo{
				{
					Addr: WhiteListProvider1,
				},
			},
		})
		assert.Nil(t, err)
		err = gov.Propose(uint8(WhitelistProviderAdd), "test", "test desc", 100, data)
		assert.ErrorContains(t, err, ErrExistVotingProposal.Error())
		return err
	})
}
