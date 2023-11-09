package base

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	vm "github.com/axiomesh/eth-kit/evm"
)

func prepareLedger(t *testing.T) ledger.StateLedger {
	mockCtl := gomock.NewController(t)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	account := ledger.NewMockAccount(1, types.NewAddressByStr(common.EpochManagerContractAddr))
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()
	return stateLedger
}

func TestEpochManager(t *testing.T) {
	stateLedger := prepareLedger(t)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	err := InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	currentEpoch, err := GetCurrentEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, currentEpoch.Epoch)
	assert.EqualValues(t, 1, currentEpoch.StartBlock)

	nextEpoch, err := GetNextEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, nextEpoch.Epoch)
	assert.EqualValues(t, 101, nextEpoch.StartBlock)

	epoch1, err := GetEpochInfo(stateLedger, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, epoch1.Epoch)

	_, err = GetEpochInfo(stateLedger, 2)
	assert.Error(t, err)

	newCurrentEpoch, err := TurnIntoNewEpoch([]byte{}, stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, newCurrentEpoch.Epoch)
	assert.EqualValues(t, 101, newCurrentEpoch.StartBlock)

	currentEpoch, err = GetCurrentEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, currentEpoch.Epoch)
	assert.EqualValues(t, 101, currentEpoch.StartBlock)

	nextEpoch, err = GetNextEpochInfo(stateLedger)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, nextEpoch.Epoch)
	assert.EqualValues(t, 201, nextEpoch.StartBlock)

	epoch2, err := GetEpochInfo(stateLedger, 2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, epoch2.Epoch)

	_, err = GetEpochInfo(stateLedger, 3)
	assert.Error(t, err)

	epochMgr := NewEpochManager(&common.SystemContractConfig{
		Logger: logrus.New(),
	})
	epochMgr.Reset(1, stateLedger)
	_, err = epochMgr.EstimateGas(&types.CallArgs{
		Data: &hexutil.Bytes{},
	})
	assert.Error(t, err)

	t.Run("query currentEpoch", func(t *testing.T) {
		method := epochManagerABI.Methods["currentEpoch"]

		inputs, err := method.Inputs.Pack()
		assert.Nil(t, err)
		inputs = append(method.ID, inputs...)

		_, err = epochMgr.EstimateGas(&types.CallArgs{
			Data: (*hexutil.Bytes)(&inputs),
		})
		assert.Nil(t, err)
		res, err := epochMgr.Run(&vm.Message{
			Data: inputs,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Err, string(res.ReturnData))

		unpacked, err := method.Outputs.Unpack(res.ReturnData)
		assert.Nil(t, err)
		decodedEpoch := &rbft.EpochInfo{}
		err = method.Outputs.Copy(&decodedEpoch, unpacked)
		assert.Nil(t, err)

		assert.EqualValues(t, currentEpoch, decodedEpoch)
	})

	t.Run("query nextEpoch", func(t *testing.T) {
		method := epochManagerABI.Methods["nextEpoch"]

		inputs, err := method.Inputs.Pack()
		assert.Nil(t, err)
		inputs = append(method.ID, inputs...)

		_, err = epochMgr.EstimateGas(&types.CallArgs{
			Data: (*hexutil.Bytes)(&inputs),
		})
		assert.Nil(t, err)
		res, err := epochMgr.Run(&vm.Message{
			Data: inputs,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Err, string(res.ReturnData))

		unpacked, err := method.Outputs.Unpack(res.ReturnData)
		assert.Nil(t, err)
		decodedEpoch := &rbft.EpochInfo{}
		err = method.Outputs.Copy(&decodedEpoch, unpacked)
		assert.Nil(t, err)

		assert.EqualValues(t, nextEpoch, decodedEpoch)
	})

	t.Run("query historyEpoch", func(t *testing.T) {
		method := epochManagerABI.Methods["historyEpoch"]

		inputs, err := method.Inputs.Pack(uint64(2))
		assert.Nil(t, err)
		inputs = append(method.ID, inputs...)

		_, err = epochMgr.EstimateGas(&types.CallArgs{
			Data: (*hexutil.Bytes)(&inputs),
		})
		assert.Nil(t, err)
		res, err := epochMgr.Run(&vm.Message{
			Data: inputs,
		})
		assert.Nil(t, err)
		assert.Nil(t, res.Err, string(res.ReturnData))

		unpacked, err := method.Outputs.Unpack(res.ReturnData)
		assert.Nil(t, err)
		decodedEpoch := &rbft.EpochInfo{}
		err = method.Outputs.Copy(&decodedEpoch, unpacked)
		assert.Nil(t, err)

		assert.EqualValues(t, epoch2, decodedEpoch)
	})
}

func TestInitEpochInfo_InvalidNodeInfo(t *testing.T) {
	stateLedger := prepareLedger(t)

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1

	tests := []struct {
		name      string
		epochInfo *rbft.EpochInfo
	}{
		{
			name: "invalid node account address",
			epochInfo: func() *rbft.EpochInfo {
				e := g.Clone()
				e.ValidatorSet[0].AccountAddress = "invalid"
				return e
			}(),
		},
		{
			name: "invalid node p2p id",
			epochInfo: func() *rbft.EpochInfo {
				e := g.Clone()
				e.ValidatorSet[0].P2PNodeID = "invalid"
				return e
			}(),
		},
		{
			name: "duplicate node id",
			epochInfo: func() *rbft.EpochInfo {
				e := g.Clone()
				e.ValidatorSet[0].ID = e.ValidatorSet[1].ID
				return e
			}(),
		},
		{
			name: "duplicate node account addr",
			epochInfo: func() *rbft.EpochInfo {
				e := g.Clone()
				e.CandidateSet[0].AccountAddress = e.ValidatorSet[1].AccountAddress
				return e
			}(),
		},
		{
			name: "duplicate p2p node id",
			epochInfo: func() *rbft.EpochInfo {
				e := g.Clone()
				e.DataSyncerSet = append(e.DataSyncerSet, rbft.NodeInfo{
					ID:             100,
					AccountAddress: "0xD1AEFdf2195f2457A6a675068Cad98B67Eb54e68",
					P2PNodeID:      e.ValidatorSet[0].P2PNodeID,
				})
				return e
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, InitEpochInfo(stateLedger, tt.epochInfo))
		})
	}
}

func TestAddNode(t *testing.T) {
	stateLedger := prepareLedger(t)
	var newNodeID uint64
	t.Run("next epoch info is not initialized", func(t *testing.T) {
		newNodeID, err := AddNode(stateLedger, rbft.NodeInfo{
			AccountAddress:       "0xD1AEFdf2195f2457A6a675068Cad98B67Eb54e68",
			P2PNodeID:            "16Uiu2HAmSBJ7tARZkRT3KS41KPuEbGYZvDXdSzTj8b31gQYYGs9a",
			ConsensusVotingPower: 100,
		})
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, newNodeID)
		}
	})

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	g.DataSyncerSet = append(g.DataSyncerSet, rbft.NodeInfo{
		ID:                   9,
		AccountAddress:       "0x88E9A1cE92b4D6e4d860CFBB5bB7aC44d9b548f8",
		P2PNodeID:            "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
		ConsensusVotingPower: 100,
	})
	err := InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	t.Run("add correct node info", func(t *testing.T) {
		newNodeID, err = AddNode(stateLedger, rbft.NodeInfo{
			AccountAddress:       "0xD1AEFdf2195f2457A6a675068Cad98B67Eb54e68",
			P2PNodeID:            "16Uiu2HAmSBJ7tARZkRT3KS41KPuEbGYZvDXdSzTj8b31gQYYGs9a",
			ConsensusVotingPower: 100,
		})
		assert.Nil(t, err)
		assert.NotNil(t, newNodeID)
		ne, err := GetNextEpochInfo(stateLedger)
		assert.Nil(t, err)
		assert.EqualValues(t, len(g.CandidateSet)+1, len(ne.CandidateSet))
		assert.EqualValues(t, 10, ne.CandidateSet[len(ne.CandidateSet)-1].ID)
	})

	t.Run("add next correct node info", func(t *testing.T) {
		newNodeID, err = AddNode(stateLedger, rbft.NodeInfo{
			AccountAddress:       "0x7D9428f0cE5c89dA907Ae6860F93861BD99Fbf0d",
			P2PNodeID:            "16Uiu2HAmTYQW5Tp2cXxyENCAy8cTNRyVrmxshUvS8fXWGakbUJep",
			ConsensusVotingPower: 100,
		})
		assert.Nil(t, err)
		assert.NotNil(t, newNodeID)
		ne, err := GetNextEpochInfo(stateLedger)
		assert.Nil(t, err)
		assert.EqualValues(t, len(g.CandidateSet)+2, len(ne.CandidateSet))
		assert.EqualValues(t, 11, ne.CandidateSet[len(ne.CandidateSet)-1].ID)
	})

	exceptionTests := []struct {
		name    string
		newNode rbft.NodeInfo
	}{
		{
			name: "invalid node account address",
			newNode: rbft.NodeInfo{
				AccountAddress:       "invalid",
				P2PNodeID:            "16Uiu2HAmLDLMYKSAP67UazgNfxg2neKg3crbihuS4TEZ87F5ePGg",
				ConsensusVotingPower: 100,
			},
		},
		{
			name: "invalid node p2p id",
			newNode: rbft.NodeInfo{
				AccountAddress:       "0xA681B4E0CFA5bf0a068d5512b5E130bff2Ce6593",
				P2PNodeID:            "invalid",
				ConsensusVotingPower: 100,
			},
		},
		{
			name: "duplicate node account addr",
			newNode: rbft.NodeInfo{
				AccountAddress:       g.ValidatorSet[0].AccountAddress,
				P2PNodeID:            "16Uiu2HAmLDLMYKSAP67UazgNfxg2neKg3crbihuS4TEZ87F5ePGg",
				ConsensusVotingPower: 100,
			},
		},
		{
			name: "duplicate p2p node id",
			newNode: rbft.NodeInfo{
				AccountAddress:       "0xA681B4E0CFA5bf0a068d5512b5E130bff2Ce6593",
				P2PNodeID:            g.CandidateSet[0].P2PNodeID,
				ConsensusVotingPower: 100,
			},
		},
		{
			name: "duplicate p2p node id with DataSyncer node",
			newNode: rbft.NodeInfo{
				AccountAddress:       "0xA681B4E0CFA5bf0a068d5512b5E130bff2Ce6593",
				P2PNodeID:            g.DataSyncerSet[0].P2PNodeID,
				ConsensusVotingPower: 100,
			},
		},
	}
	for _, tt := range exceptionTests {
		t.Run(tt.name, func(t *testing.T) {
			newNodeID, err = AddNode(stateLedger, tt.newNode)
			assert.Error(t, err)
			assert.NotNil(t, newNodeID)
		})
	}
}

func TestRemoveNode(t *testing.T) {
	stateLedger := prepareLedger(t)

	t.Run("next epoch info is not initialized", func(t *testing.T) {
		err := RemoveNode(stateLedger, 0)
		assert.Error(t, err)
	})

	g := repo.GenesisEpochInfo(true)
	g.EpochPeriod = 100
	g.StartBlock = 1
	g.DataSyncerSet = append(g.DataSyncerSet, rbft.NodeInfo{
		ID:                   9,
		AccountAddress:       "0x88E9A1cE92b4D6e4d860CFBB5bB7aC44d9b548f8",
		P2PNodeID:            "16Uiu2HAkwmNbfH8ZBdnYhygUHyG5mSWrWTEra3gwHWt9dGTUSRVV",
		ConsensusVotingPower: 100,
	})
	err := InitEpochInfo(stateLedger, g)
	assert.Nil(t, err)

	t.Run("remove not existing node", func(t *testing.T) {
		err = RemoveNode(stateLedger, 1000)
		assert.Error(t, err)
	})

	t.Run("remove validator", func(t *testing.T) {
		err = RemoveNode(stateLedger, 1)
		assert.Nil(t, err)
		ne, err := GetNextEpochInfo(stateLedger)
		assert.Nil(t, err)
		assert.EqualValues(t, len(g.ValidatorSet)-1, len(ne.ValidatorSet))
		assert.False(t, lo.ContainsBy(ne.ValidatorSet, func(v rbft.NodeInfo) bool {
			return v.ID == 1
		}))
	})

	t.Run("remove candidate", func(t *testing.T) {
		err = RemoveNode(stateLedger, 5)
		assert.Nil(t, err)
		ne, err := GetNextEpochInfo(stateLedger)
		assert.Nil(t, err)
		assert.EqualValues(t, len(g.CandidateSet)-1, len(ne.CandidateSet))
		assert.False(t, lo.ContainsBy(ne.CandidateSet, func(v rbft.NodeInfo) bool {
			return v.ID == 5
		}))
	})

	t.Run("remove dataSyncer", func(t *testing.T) {
		err = RemoveNode(stateLedger, 9)
		assert.Nil(t, err)
		ne, err := GetNextEpochInfo(stateLedger)
		assert.Nil(t, err)
		assert.EqualValues(t, len(g.DataSyncerSet)-1, len(ne.DataSyncerSet))
		assert.False(t, lo.ContainsBy(ne.DataSyncerSet, func(v rbft.NodeInfo) bool {
			return v.ID == 9
		}))
	})
}
