package archive

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-bft/common/consensus"
	rbfttypes "github.com/axiomesh/axiom-bft/types"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/txpool/mock_txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/rbft/adaptor"
	"github.com/axiomesh/axiom-ledger/internal/consensus/rbft/testutil"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	sync_comm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/internal/sync/common/mock_sync"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	p2p "github.com/axiomesh/axiom-p2p"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func mockConfig(t *testing.T, ctrl *gomock.Controller) (map[int32]p2p.Pipe, *common.Config, map[uint64]map[string]p2p.Pipe) {
	err := storagemgr.Initialize(repo.KVStorageTypeLeveldb, repo.KVStorageCacheSize, repo.KVStorageSync, false)
	assert.Nil(t, err)
	logger := log.NewWithModule("adaptor")
	s, err := types.GenerateSigner()
	assert.Nil(t, err)

	fn := func() *types.ChainMeta {
		return &types.ChainMeta{
			Height:    1,
			BlockHash: types.NewHashByStr(hex.EncodeToString([]byte("block1"))),
		}
	}
	genesisEpochInfo := repo.GenesisEpochInfo(false)
	rep := t.TempDir()

	p2pID, err := repo.KeyToNodeID(s.Sk)
	assert.Nil(t, err)
	selfId := uint64(len(genesisEpochInfo.DataSyncerSet)+len(genesisEpochInfo.ValidatorSet)+len(genesisEpochInfo.CandidateSet)) + 1
	genesisEpochInfo.DataSyncerSet = append(genesisEpochInfo.DataSyncerSet, rbft.NodeInfo{
		ID:                   selfId,
		P2PNodeID:            p2pID,
		AccountAddress:       s.Addr.String(),
		ConsensusVotingPower: 0,
	})

	epochStore, err := storagemgr.Open(repo.GetStoragePath(rep, storagemgr.Epoch))
	require.Nil(t, err)
	conf := &common.Config{
		RepoRoot:             rep,
		Config:               repo.DefaultConsensusConfig(),
		Logger:               logger,
		ConsensusType:        "",
		ConsensusStorageType: repo.ConsensusStorageTypeMinifile,
		PrivKey:              s.Sk,
		GenesisEpochInfo:     genesisEpochInfo,
		Applied:              0,
		Digest:               "",
		GenesisDigest:        "genesisDigest",
		GetEpochInfoFromEpochMgrContractFunc: func(epoch uint64) (*rbft.EpochInfo, error) {
			return genesisEpochInfo, nil
		},

		GetAccountNonce: func(address *types.Address) uint64 {
			return 0
		},
		GetCurrentEpochInfoFromEpochMgrContractFunc: func() (*rbft.EpochInfo, error) {
			return genesisEpochInfo, nil
		},

		GetChainMetaFunc: fn,

		EpochStore: epochStore,
	}

	peers := getPeers(genesisEpochInfo)
	manager := p2p.GenMockHostManager(peers)
	mockNetwork, err := mock_network.NewMiniNetwork(selfId, p2pID, manager)
	assert.Nil(t, err)

	conf.Network = mockNetwork

	mockBlockSync := testutil.MockMiniBlockSync(ctrl)
	conf.BlockSync = mockBlockSync

	mockTxpool := mock_txpool.NewMockMinimalTxPool[types.Transaction, *types.Transaction](500, ctrl)
	conf.TxPool = mockTxpool

	consensusMsgPipes := make(map[int32]p2p.Pipe)
	for id, name := range consensus.Type_name {
		msgPipe, err := mockNetwork.CreatePipe(context.Background(), "mock_pipe"+name)
		assert.Nil(t, err)
		consensusMsgPipes[id] = msgPipe
	}

	pipes := prepareNetworks(manager, genesisEpochInfo, t)
	archivePipeIds := lo.FlatMap(common.ArchivePipeName, func(name string, _ int) []int32 {
		return []int32{consensus.Type_value[name]}
	})
	pipes[selfId] = lo.MapKeys(lo.PickByKeys(consensusMsgPipes, archivePipeIds), func(_ p2p.Pipe, val int32) string {
		return consensus.Type_name[val]
	})

	err = mockNetwork.Start()
	assert.Nil(t, err)
	return consensusMsgPipes, conf, pipes
}

func mockArchiveNode(consensusMsgPipes map[int32]p2p.Pipe, conf *common.Config, t *testing.T) *Node[types.Transaction, *types.Transaction] {
	rbftAdaptor, err := adaptor.NewRBFTAdaptor(conf)
	assert.Nil(t, err)

	rbftAdaptor.SetMsgPipes(consensusMsgPipes)

	err = rbftAdaptor.UpdateEpoch()
	assert.Nil(t, err)

	logger := log.NewWithModule("archive")
	logger.Logger.SetLevel(logrus.DebugLevel)
	rbftConfig := rbft.Config{
		SelfP2PNodeID:           conf.GenesisEpochInfo.DataSyncerSet[0].P2PNodeID,
		SyncStateTimeout:        1 * time.Minute,
		SyncStateRestartTimeout: 1 * time.Second,
	}
	node, err := NewArchiveNode[types.Transaction, *types.Transaction](rbftConfig, rbftAdaptor, conf.GetChainMetaFunc, conf.TxPool, conf.GenesisDigest, logger)
	assert.Nil(t, err)

	err = node.Init()
	assert.Nil(t, err)
	return node
}

func prepareNetworks(manager *p2p.MockHostManager, epochInfo *rbft.EpochInfo, t *testing.T) map[uint64]map[string]p2p.Pipe {
	nets := make([]*mock_network.MiniNetwork, 0)
	lo.ForEach(epochInfo.ValidatorSet, func(info rbft.NodeInfo, _ int) {
		net, err := mock_network.NewMiniNetwork(info.ID, info.P2PNodeID, manager)
		assert.Nil(t, err)
		err = net.Start()
		assert.Nil(t, err)
		nets = append(nets, net)
	})

	netCache := make(map[uint64]map[string]p2p.Pipe)
	lo.ForEach(nets, func(net *mock_network.MiniNetwork, _ int) {
		consensusMsgPipes := make(map[string]p2p.Pipe)
		for _, name := range consensus.Type_name {
			msgPipe, err := net.CreatePipe(context.Background(), "mock_pipe"+name)
			assert.Nil(t, err)
			consensusMsgPipes[name] = msgPipe
		}
		netCache[net.GetSelfId()] = consensusMsgPipes
	})

	return netCache
}

func getPeers(epochInfo *rbft.EpochInfo) []string {
	allNodes := append(epochInfo.ValidatorSet, epochInfo.CandidateSet...)
	allNodes = append(allNodes, epochInfo.DataSyncerSet...)
	return lo.FlatMap(allNodes, func(info rbft.NodeInfo, _ int) []string {
		return []string{info.P2PNodeID}
	})
}

func receiveMsg(t *testing.T, pipesM map[string]p2p.Pipe, selfId uint64, handler func(msg *consensus.ConsensusMessage, pipe p2p.Pipe, to string, id uint64) error) {
	for _, pipe := range pipesM {
		go func(pipe p2p.Pipe) {
			for {
				msg := pipe.Receive(context.Background())
				if msg == nil {
					return
				}
				consensusMsg := &consensus.ConsensusMessage{}
				err := consensusMsg.UnmarshalVT(msg.Data)
				assert.Nil(t, err)
				err = handler(consensusMsg, pipesM[getRespPipe(consensusMsg.Type)], msg.From, selfId)
			}
		}(pipe)
	}

}

func getRespPipe(typ consensus.Type) string {
	switch typ {
	case consensus.Type_SYNC_STATE:
		return consensus.Type_name[int32(consensus.Type_SYNC_STATE_RESPONSE)]
	case consensus.Type_FETCH_MISSING_REQUEST:
		return consensus.Type_name[int32(consensus.Type_FETCH_MISSING_RESPONSE)]
	case consensus.Type_EPOCH_CHANGE_REQUEST:
		return consensus.Type_name[int32(consensus.Type_EPOCH_CHANGE_PROOF)]
	}
	return ""
}

func generateSignedCheckpoint(t *testing.T, selfId, height uint64, blockHash, batchDigest string) *consensus.SignedCheckpoint {
	signedCheckpoint := &consensus.SignedCheckpoint{
		Author: selfId,
	}

	checkpoint := &consensus.Checkpoint{
		NeedUpdateEpoch: height%100 == 0,
		Epoch:           ((height - 1) / 100) + 1,
		ExecuteState: &consensus.Checkpoint_ExecuteState{
			Height:      height,
			Digest:      blockHash,
			BatchDigest: batchDigest,
		},
		ViewChange: &consensus.ViewChange{},
	}

	vcBasis := &consensus.VcBasis{
		ReplicaId: selfId,
		View:      1,
		H:         10,
		Pset:      []*consensus.VcPq{},
		Qset:      []*consensus.VcPq{},
		Cset:      []*consensus.SignedCheckpoint{},
	}
	checkpoint.ViewChange.Basis = vcBasis
	signedCheckpoint.Checkpoint = checkpoint
	key := repo.DefaultNodeKeys[int(selfId)-1]
	sk, err := repo.ParseKey([]byte(key))
	assert.Nil(t, err)
	privKey, err := repo.Libp2pKeyFromECDSAKey(sk)
	assert.Nil(t, err)

	msg := checkpoint.Hash()
	signature, err := privKey.Sign(msg)
	assert.Nil(t, err)
	signedCheckpoint.Signature = signature
	return signedCheckpoint
}

func genQuorumSignCheckpoint(t *testing.T, height uint64, blockHash, batchDigest string) *consensus.QuorumCheckpoint {
	signM := make(map[uint64][]byte)
	list := make([]*consensus.SignedCheckpoint, 0)
	for i := 1; i <= 4; i++ {
		sckpt := generateSignedCheckpoint(t, uint64(i), height, blockHash, batchDigest)
		signM[uint64(i)] = sckpt.Signature
		list = append(list, sckpt)
	}
	return &consensus.QuorumCheckpoint{
		Checkpoint: list[0].GetCheckpoint(),
		Signatures: signM,
	}
}

func catchExpectLogOut(t *testing.T, log logrus.FieldLogger, expectResult string, taskDoneCh chan bool) {
	ticker := time.NewTicker(100 * time.Second)
	// Setup log output capturing
	lg := log.(*logrus.Entry).Logger
	originalOutput := log.(*logrus.Entry).Logger.Out
	var logOutput bytes.Buffer
	lg.SetOutput(&logOutput)
	go func(ch chan bool) {
		for {
			select {
			case <-ticker.C:
				t.Errorf("Expected log message not found in output: \n%s", logOutput.String())
				ch <- false
				return
			default:
				if strings.Contains(logOutput.String(), expectResult) {
					// Restore log output
					lg.SetOutput(originalOutput)
					ch <- true
					return
				}
			}
		}
	}(taskDoneCh)
}

func remoteProposal(t *testing.T, id uint64, broadcastNodes []string, pipes map[uint64]map[string]p2p.Pipe, reqBatch *txpool.RequestHashBatch[types.Transaction, *types.Transaction], view, seqNo uint64) {
	var executePipe p2p.Pipe
	for i, pipeM := range pipes {
		if i == id {
			pipe, ok := pipeM["PRE_PREPARE"]
			assert.True(t, ok)
			executePipe = pipe
		}
	}

	hashBatch := &consensus.HashBatch{
		RequestHashList: reqBatch.TxHashList,
		Timestamp:       reqBatch.Timestamp,
		Proposer:        id,
	}

	preprepare := &consensus.PrePrepare{
		View:           view,
		SequenceNumber: seqNo,
		BatchDigest:    reqBatch.BatchHash,
		HashBatch:      hashBatch,
		ReplicaId:      id,
	}

	payload, err := preprepare.MarshalVTStrict()
	assert.Nil(t, err)
	consensusMsg := &consensus.ConsensusMessage{
		Type:    consensus.Type_PRE_PREPARE,
		Payload: payload,
		Epoch:   ((seqNo - 1) / 100) + 1,
		From:    id,
	}
	data, err := consensusMsg.MarshalVT()
	assert.Nil(t, err)
	err = executePipe.Broadcast(context.Background(), broadcastNodes, data)
	assert.Nil(t, err)
}

func remoteReportCheckpoint(t *testing.T, id uint64, broadcastNodes []string, pipes map[uint64]map[string]p2p.Pipe, batchDigest string, seqNo uint64) {
	var executePipe p2p.Pipe
	for i, pipeM := range pipes {
		if i == id {
			pipe, ok := pipeM["SIGNED_CHECKPOINT"]
			assert.True(t, ok)
			executePipe = pipe
		}
	}

	block := testutil.ConstructBlock(fmt.Sprintf("block%d", seqNo), seqNo)
	sckpt := generateSignedCheckpoint(t, id, seqNo, block.Hash().String(), batchDigest)
	payload, err := sckpt.MarshalVT()
	assert.Nil(t, err)

	consensusMsg := &consensus.ConsensusMessage{
		Type:    consensus.Type_SIGNED_CHECKPOINT,
		Payload: payload,
		Epoch:   ((seqNo - 1) / 100) + 1,
	}

	data, err := consensusMsg.MarshalVT()
	assert.Nil(t, err)
	err = executePipe.Broadcast(context.Background(), broadcastNodes, data)
	assert.Nil(t, err)
}

func generateBatch(t *testing.T) (*txpool.RequestHashBatch[types.Transaction, *types.Transaction], []*types.Transaction) {
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	txs := testutil.ConstructTxs(s, 10)
	txHashList := lo.Map(txs, func(tx *types.Transaction, _ int) string {
		return tx.RbftGetTxHash()
	})
	localList := lo.RepeatBy(len(txs), func(index int) bool {
		return true
	})
	reqBatch := &txpool.RequestHashBatch[types.Transaction, *types.Transaction]{
		TxHashList: txHashList,
		TxList:     txs,
		Timestamp:  time.Now().UnixNano(),
		LocalList:  localList,
	}
	reqBatch.BatchHash = reqBatch.GenerateBatchHash()
	return reqBatch, txs
}

func mockSyncInStack(t *testing.T, node *Node[types.Transaction, *types.Transaction], consensusMsgPipes map[int32]p2p.Pipe, cnf *common.Config, ctrl *gomock.Controller) {
	mockSync := mock_sync.NewMockSync(ctrl)
	mockSync.EXPECT().StartSync(gomock.Any(), gomock.Any()).DoAndReturn(func(params *sync_comm.SyncParams, syncTaskDoneCh chan error) error {
		syncTaskDoneCh <- nil
		return nil
	}).AnyTimes()

	block2 := testutil.ConstructBlock("block2", uint64(2))
	blockChan := make(chan any, 1)
	d := &sync_comm.BlockData{
		Block: block2,
	}
	blocks := make([]sync_comm.CommitData, 0)
	blocks = append(blocks, d)
	blockChan <- blocks

	mockSync.EXPECT().Commit().Return(blockChan).AnyTimes()
	cnf.BlockSync = mockSync

	cnf.RepoRoot = filepath.Join(t.TempDir(), "repo")
	newStack, err := adaptor.NewRBFTAdaptor(cnf)
	assert.Nil(t, err)

	newStack.SetMsgPipes(consensusMsgPipes)

	err = newStack.UpdateEpoch()
	assert.Nil(t, err)
	node.stack = newStack

	go func() {
		for {
			select {
			case block := <-newStack.BlockC:
				assert.Equal(t, block2.Height(), block.Block.Height())
				assert.Equal(t, block2.Hash().String(), block.Block.Hash().String())
				state := &rbfttypes.ServiceSyncState{}
				state.ServiceState.MetaState = &rbfttypes.MetaState{
					Height: block.Block.Height(),
					Digest: block.Block.Hash().String(),
				}
				state.EpochChanged = false
				state.Epoch = 1

				node.ReportStateUpdated(state)
			}
		}
	}()
}
