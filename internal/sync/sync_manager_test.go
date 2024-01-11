package sync

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	network2 "github.com/axiomesh/axiom-p2p"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
)

func TestNewSyncManager(t *testing.T) {
	logger := loggers.Logger(loggers.BlockSync)
	getChainMetaFn := func() *types.ChainMeta {
		return nil
	}
	getBlockFn := func(height uint64) (*types.Block, error) {
		return nil, nil
	}
	getReceiptsFn := func(height uint64) ([]*types.Receipt, error) {
		return nil, nil
	}
	getEpochStateFn := func(key []byte) []byte {
		return nil
	}

	ctrl := gomock.NewController(t)
	net := mock_network.NewMockNetwork(ctrl)

	cnf := repo.Sync{
		TimeoutCountLimit: 1,
	}

	var createPipeErrCount = map[string]int{
		common.SyncBlockRequestPipe:      1,
		common.SyncBlockResponsePipe:     1,
		common.SyncChainDataRequestPipe:  1,
		common.SyncChainDataResponsePipe: 1,
	}

	net.EXPECT().CreatePipe(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, name string) (network2.Pipe, error) {
		if createPipeErrCount[name] > 0 {
			createPipeErrCount[name]--
			return nil, errors.New("create pipe err")
		}
		return mock_network.NewMockPipe(ctrl), nil
	}).AnyTimes()

	for i := 0; i < len(createPipeErrCount); i++ {
		_, err := NewSyncManager(logger, getChainMetaFn, getBlockFn, getReceiptsFn, getEpochStateFn, net, cnf)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "create pipe err")
	}
}
func TestStartSync(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 10)

	for _, sync := range syncs {
		sync.conf.TimeoutCountLimit = 1
	}
	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()
	remoteBlockHash := ledgers[remoteId].GetChainMeta().BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 10,
				Digest: remoteBlockHash,
			},
		},
	}

	// test switch sync status err
	err := syncs[0].switchSyncStatus(true)
	require.Nil(t, err)
	syncTaskDoneCh := make(chan error, 1)
	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 10, quorumCkpt), syncTaskDoneCh)
	actualErr := <-syncTaskDoneCh
	require.NotNil(t, actualErr)
	require.Equal(t, actualErr, err)
	require.Contains(t, err.Error(), "status is already true")
	err = syncs[0].switchSyncStatus(false)
	require.Nil(t, err)

	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 10, quorumCkpt), syncTaskDoneCh)
	require.Nil(t, err)

	// wait for reset Peers
	time.Sleep(1 * time.Second)
	// start sync model
	for i := 0; i < n; i++ {
		_, err = syncs[i].Prepare()
		require.Nil(t, err)
	}

	data := <-syncs[0].Commit()
	blocks := data.([]common.CommitData)
	require.Equal(t, 9, len(blocks))
	require.Equal(t, uint64(10), blocks[len(blocks)-1].GetHeight())
}

func TestStartSyncWithRemoteSendBlockResponseError(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendSyncBlockResponse, 0, 2)
	defer stopSyncs(syncs)

	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 10)

	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}

	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()
	remoteBlockHash := ledgers[remoteId].GetChainMeta().BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 10,
				Digest: remoteBlockHash,
			},
		},
	}

	syncTaskDoneCh := make(chan error, 1)

	err := syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 10, quorumCkpt), syncTaskDoneCh)
	require.Nil(t, err)
	<-syncTaskDoneCh
	data := <-syncs[0].Commit()
	blocks := data.([]common.CommitData)
	require.Equal(t, 9, len(blocks))
	require.Equal(t, uint64(10), blocks[len(blocks)-1].GetHeight())
}

func TestMultiEpochSync(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 300)

	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}

	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()
	remoteBlockHash := ledgers[remoteId].GetChainMeta().BlockHash.String()
	quorumCkpt300 := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 300,
				Digest: remoteBlockHash,
			},
		},
	}

	block100, err := ledgers[remoteId].GetBlock(100)
	require.Nil(t, err)
	block200, err := ledgers[remoteId].GetBlock(200)
	require.Nil(t, err)
	block300, err := ledgers[remoteId].GetBlock(300)
	require.Nil(t, err)
	epc1 := &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block100.Height(),
					Digest: block100.BlockHash.String(),
				},
			},
		},
	}
	epc2 := &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block200.Height(),
					Digest: block200.BlockHash.String(),
				},
			},
		},
	}

	epc3 := &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block300.Height(),
					Digest: block300.BlockHash.String(),
				},
			},
		},
	}

	syncTaskDoneCh := make(chan error, 1)
	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 300, quorumCkpt300, epc1, epc2, epc3), syncTaskDoneCh)
	require.Nil(t, err)
	err = <-syncTaskDoneCh
	require.Nil(t, err)

	data := <-syncs[0].Commit()
	blocks1 := data.([]common.CommitData)
	require.Equal(t, 99, len(blocks1))
	require.Equal(t, uint64(100), blocks1[len(blocks1)-1].GetHeight())
	data = <-syncs[0].Commit()
	blocks2 := data.([]common.CommitData)
	require.Equal(t, 100, len(blocks2))
	require.Equal(t, uint64(200), blocks2[len(blocks2)-1].GetHeight())
	data = <-syncs[0].Commit()
	blocks3 := data.([]common.CommitData)
	require.Equal(t, 100, len(blocks3))
	require.Equal(t, uint64(300), blocks3[len(blocks3)-1].GetHeight())

	require.False(t, syncs[0].syncStatus.Load())
}

func TestMultiEpochSyncWithWrongBlock(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 200)

	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}

	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()

	block100, err := ledgers[remoteId].GetBlock(100)
	require.Nil(t, err)

	epc1 := &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block100.Height(),
					Digest: block100.BlockHash.String(),
				},
			},
		},
	}

	// wrong commitDataCache is not epoch commitDataCache
	// mock wrong commitDataCache
	wrongHeight := uint64(7)
	oldRightBlock, err := ledgers[remoteId].GetBlock(wrongHeight)
	require.Nil(t, err)
	parentBlock, err := ledgers[remoteId].GetBlock(wrongHeight - 1)
	require.Nil(t, err)
	wrongBlock := &types.Block{
		BlockHeader: &types.BlockHeader{
			Number:     wrongHeight,
			ParentHash: parentBlock.BlockHash,
		},
		BlockHash: types.NewHash([]byte("wrong_block")),
	}

	idx := wrongHeight % uint64(len(peers))
	wrongRemoteId := peers[idx]

	err = ledgers[wrongRemoteId].PersistExecutionResult(wrongBlock, genReceipts(wrongBlock))
	require.Nil(t, err)

	block10, err := ledgers[remoteId].GetBlock(10)
	require.Nil(t, err)

	quorumCkpt10 := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: block10.Height(),
				Digest: block10.BlockHash.String(),
			},
		},
	}
	// start sync
	syncTaskDoneCh := make(chan error, 1)
	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 10, quorumCkpt10), syncTaskDoneCh)
	require.Nil(t, err)
	err = <-syncTaskDoneCh
	require.Nil(t, err)

	data := <-syncs[0].Commit()
	blocks1 := data.([]common.CommitData)
	require.Equal(t, 9, len(blocks1))
	require.Equal(t, uint64(10), blocks1[len(blocks1)-1].GetHeight())
	require.Equal(t, block10.BlockHash.String(), blocks1[len(blocks1)-1].GetHash())
	require.False(t, syncs[0].syncStatus.Load())

	// reset right block
	err = ledgers[wrongRemoteId].PersistExecutionResult(oldRightBlock, genReceipts(oldRightBlock))
	require.Nil(t, err)

	require.False(t, syncs[0].syncStatus.Load())

	t.Run("wrong block is epoch block", func(t *testing.T) {
		// mock wrong commitDataCache
		wrongHeight = uint64(100)
		oldRightBlock, err = ledgers[remoteId].GetBlock(wrongHeight)
		require.Nil(t, err)
		parentBlock, err = ledgers[remoteId].GetBlock(wrongHeight - 1)
		require.Nil(t, err)
		wrongBlock = &types.Block{
			BlockHeader: &types.BlockHeader{
				Number:     wrongHeight,
				ParentHash: parentBlock.BlockHash,
			},
			BlockHash: types.NewHash([]byte("wrong_block")),
		}
		wrongRemoteId = syncs[0].peers[int(wrongHeight)%len(syncs[0].peers)].PeerID
		err = ledgers[wrongRemoteId].PersistExecutionResult(wrongBlock, genReceipts(wrongBlock))
		require.Nil(t, err)

		block101, err := ledgers[remoteId].GetBlock(101)
		require.Nil(t, err)

		quorumCkpt101 := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block101.Height(),
					Digest: block101.BlockHash.String(),
				},
			},
		}

		// start sync
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 101, quorumCkpt101, epc1), syncTaskDoneCh)
		require.Nil(t, err)
		err = <-syncTaskDoneCh
		require.Nil(t, err)

		data = <-syncs[0].Commit()
		blocks1 = data.([]common.CommitData)
		require.Equal(t, 99, len(blocks1))
		require.Equal(t, uint64(100), blocks1[len(blocks1)-1].GetHeight())
		data = <-syncs[0].Commit()
		blocks2 := data.([]common.CommitData)
		require.Equal(t, 1, len(blocks2))
		require.Equal(t, uint64(101), blocks2[len(blocks2)-1].GetHeight())

		// reset right block
		err = ledgers[wrongRemoteId].PersistExecutionResult(oldRightBlock, genReceipts(oldRightBlock))
		require.Nil(t, err)
		require.False(t, syncs[0].syncStatus.Load())
	})
}
func TestMultiEpochSyncWithWrongCheckpoint(t *testing.T) {
	n := 4
	t.Parallel()
	t.Run("wrong checkpoint in first epoch", func(t *testing.T) {
		syncs, ledgers := newMockBlockSyncs(t, n)
		defer stopSyncs(syncs)

		localId := "0"
		// store blocks expect node 0
		prepareLedger(t, ledgers, localId, 100)

		// start sync model
		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}

		// node0 start sync commitData
		peers := []string{"1", "2", "3"}
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()

		startHeight := ledgers[localId].GetChainMeta().Height + 1
		targetHeight := startHeight + 10
		wrongQuorumCkpt := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: targetHeight,
					Digest: "wrong digest",
				},
			},
		}
		syncTaskDoneCh := make(chan error, 1)
		err := syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, startHeight, targetHeight, wrongQuorumCkpt), syncTaskDoneCh)
		require.Nil(t, err)
		err = <-syncTaskDoneCh
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "quorum checkpoint is not equal to current hash")
	})

	t.Run("wrong checkpoint in second epoch", func(t *testing.T) {
		syncs, ledgers := newMockBlockSyncs(t, n)
		defer stopSyncs(syncs)

		localId := "0"
		// store blocks expect node 0
		prepareLedger(t, ledgers, localId, 200)

		// start sync model
		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}

		// node0 start sync commitData
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()

		startHeight := ledgers[localId].GetChainMeta().Height + 1
		targetHeight := ledgers[remoteId].GetChainMeta().Height
		wrongQuorumCkpt := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: targetHeight,
					Digest: "wrong digest",
				},
			},
		}
		block100, err := ledgers[remoteId].GetBlock(100)
		require.Nil(t, err)
		epc1 := &consensus.EpochChange{
			Checkpoint: &consensus.QuorumCheckpoint{
				Checkpoint: &consensus.Checkpoint{
					ExecuteState: &consensus.Checkpoint_ExecuteState{
						Height: block100.Height(),
						Digest: block100.BlockHash.String(),
					},
				},
			},
		}
		syncTaskDoneCh := make(chan error, 1)
		param := genSyncParams(peers, latestBlockHash, 2, startHeight, targetHeight, wrongQuorumCkpt, epc1)
		err = syncs[0].StartSync(param, syncTaskDoneCh)
		require.Nil(t, err)
		blocks := <-syncs[0].Commit()
		require.NotNil(t, blocks)

		err = <-syncTaskDoneCh
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "quorum checkpoint is not equal to current hash")
	})
}

func TestHandleTimeoutBlockMsg(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync request error
	syncs, ledgers := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 200)

	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}
	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()

	remoteId := "1"
	block100, err := ledgers[remoteId].GetBlock(100)
	require.Nil(t, err)

	epc1 := &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block100.Height(),
					Digest: block100.BlockHash.String(),
				},
			},
		},
	}

	// timeout with one time
	timeoutBlockHeight := uint64(7)
	idx := int(timeoutBlockHeight % uint64(len(peers)))
	wrongId := fmt.Sprintf("%d", idx+1)

	oldRightBlock, err := ledgers[wrongId].GetBlock(timeoutBlockHeight)
	require.Nil(t, err)
	delete(ledgers[wrongId].blockDb, timeoutBlockHeight)

	block10, err := ledgers[wrongId].GetBlock(10)
	require.Nil(t, err)

	quorumCkpt10 := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: block10.Height(),
				Digest: block10.BlockHash.String(),
			},
		},
	}

	// start sync
	syncTaskDoneCh := make(chan error, 1)
	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 10, quorumCkpt10), syncTaskDoneCh)
	require.Nil(t, err)
	err = <-syncTaskDoneCh
	require.Nil(t, err)

	data := <-syncs[0].Commit()
	blocks1 := data.([]common.CommitData)
	require.Equal(t, wrongId, syncs[0].peers[idx].PeerID)
	require.Equal(t, uint64(1), syncs[0].peers[idx].TimeoutCount, "record timeout count")
	require.Equal(t, 3, len(syncs[0].peers), "not remove timeout peer because timeoutCount < timeoutCountLimit")
	require.Equal(t, 9, len(blocks1))
	require.Equal(t, uint64(10), blocks1[len(blocks1)-1].GetHeight())
	require.Equal(t, oldRightBlock.BlockHash.String(), blocks1[timeoutBlockHeight-2].GetHash())

	// reset right commitDataCache
	err = ledgers[wrongId].PersistExecutionResult(oldRightBlock, genReceipts(oldRightBlock))
	require.Nil(t, err)
	require.False(t, syncs[0].syncStatus.Load())

	t.Run("TestSyncTimeoutBlock with many times, bigger than timeoutCount", func(t *testing.T) {
		syncs[1].Stop()
		quorumCkpt100 := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block100.Height(),
					Digest: block100.BlockHash.String(),
				},
			},
		}
		// start sync
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 100, quorumCkpt100, epc1), syncTaskDoneCh)
		require.Nil(t, err)
		data = <-syncs[0].Commit()
		blocks1 = data.([]common.CommitData)
		require.Equal(t, 99, len(blocks1))
		require.Equal(t, uint64(100), blocks1[len(blocks1)-1].GetHeight())
		require.Equal(t, 2, len(syncs[0].peers), "remove timeout peer")

		require.False(t, syncs[0].syncStatus.Load())
	})
}

func TestHandleSyncErrMsg(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync request error
	syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendSyncBlockRequest, 0, 1)
	defer stopSyncs(syncs)
	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 100)

	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}
	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()
	remoteBlockHash := ledgers[remoteId].GetChainMeta().BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 100,
				Digest: remoteBlockHash,
			},
		},
	}
	syncTaskDoneCh := make(chan error, 1)
	err := syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 100, quorumCkpt), syncTaskDoneCh)
	require.Nil(t, err)

	data := <-syncs[0].Commit()
	blocks := data.([]common.CommitData)
	require.Equal(t, 99, len(blocks))
	require.Equal(t, uint64(100), blocks[len(blocks)-1].GetHeight())
}

func TestValidateChunk(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n)

	localId := "0"
	syncs[0].latestCheckedState = &pb.CheckpointState{
		Height: ledgers[localId].GetChainMeta().Height,
		Digest: ledgers[localId].GetChainMeta().BlockHash.String(),
	}
	syncs[0].curHeight = 1
	syncs[0].chunk = &common.Chunk{
		ChunkSize: 1,
	}

	// get requester err: nil
	_, err := syncs[0].validateChunk()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "requester[height:1] is nil")

	oldRequest := &requester{
		peerID:      "1",
		blockHeight: 1,
		quitCh:      make(chan struct{}, 1),
	}

	syncs[0].increaseRequester(oldRequest, 1)

	newRequest := &requester{
		peerID:      "2",
		blockHeight: 1,
		quitCh:      make(chan struct{}, 1),
	}

	syncs[0].increaseRequester(newRequest, 1)
	<-oldRequest.quitCh

	t.Run("get requester commitDataCache err: commitDataCache is nil", func(t *testing.T) {
		invalidMsgs, err := syncs[0].validateChunk()
		require.Nil(t, err)
		require.Equal(t, 1, len(invalidMsgs))
	})
}

func TestRequestState(t *testing.T) {
	t.Parallel()
	t.Run("test request state with wrong localState", func(t *testing.T) {
		n := 4
		// mock syncs[0] which send sync request error
		syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendSyncBlockRequest, 0, 1)
		defer stopSyncs(syncs)
		localId := "0"
		// store blocks expect node 0
		prepareLedger(t, ledgers, localId, 10)

		// start sync model
		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}
		// node0 start sync commitDataCache
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		wrongLatestBlockHash := "wrong hash"
		block10, err := ledgers[remoteId].GetBlock(10)
		require.Nil(t, err)
		quorumCkpt10 := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block10.Height(),
					Digest: block10.BlockHash.String(),
				},
			},
		}
		// start sync
		syncTaskDoneCh := make(chan error, 1)
		err = syncs[0].StartSync(genSyncParams(peers, wrongLatestBlockHash, 2, 2, 10, quorumCkpt10), syncTaskDoneCh)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "quorum state is not equal to current state")
	})

	t.Run("request state with different state", func(t *testing.T) {
		n := 4
		// peer1 will latency send state request
		syncs, ledgers := newMockBlockSyncs(t, n, latencyTypeSendState, 0, 1)
		defer stopSyncs(syncs)

		// peer2 will send wrong state request
		wrongRemoteId := "2"
		localId := "0"
		// store blocks expect node 0
		prepareLedger(t, ledgers, localId, 10)
		wrongGensisBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 1,
			},
			BlockHash: types.NewHashByStr("wrong hash"),
		}
		err := ledgers[wrongRemoteId].PersistExecutionResult(wrongGensisBlock, genReceipts(wrongGensisBlock))
		require.Nil(t, err)

		// start sync model
		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}
		// node0 start sync commitDataCache
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash
		block10, err := ledgers[remoteId].GetBlock(10)
		require.Nil(t, err)
		quorumCkpt10 := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block10.Height(),
					Digest: block10.BlockHash.String(),
				},
			},
		}
		// start sync
		syncTaskDoneCh := make(chan error, 1)
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash.String(), 2, 2, 10, quorumCkpt10), syncTaskDoneCh)
		require.Nil(t, err)

		err = <-syncTaskDoneCh
		require.Nil(t, err)
	})
}

func TestSwitchMode(t *testing.T) {
	n := 1
	syncs, _ := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	originMode := syncs[0].mode
	require.Equal(t, common.SyncModeFull, originMode)

	err := syncs[0].SwitchMode(originMode)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "current mode is same as switch mode")

	syncs[0].syncStatus.Store(true)
	err = syncs[0].SwitchMode(common.SyncModeSnapshot)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "sync status is true")

	syncs[0].syncStatus.Store(false)
	err = syncs[0].SwitchMode(common.SyncModeSnapshot)
	require.Nil(t, err)
	require.Equal(t, syncs[0].mode, common.SyncModeSnapshot)

	err = syncs[0].SwitchMode(common.SyncModeFull)
	require.Nil(t, err)
	require.Equal(t, syncs[0].mode, common.SyncModeFull)
}

func TestStartSyncWithSnapshotMode(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync request error
	syncs, ledgers := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)
	localId := "0"
	// store blocks expect node 0
	prepareLedger(t, ledgers, localId, 300)

	// node0 start sync commitDataCache
	peers := []string{"1", "2", "3"}
	remoteId := "1"
	latestBlockHash := ledgers[localId].GetChainMeta().BlockHash.String()
	remoteBlockHash := ledgers[remoteId].GetChainMeta().BlockHash.String()
	quorumCkpt300 := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 300,
				Digest: remoteBlockHash,
			},
		},
	}

	err := syncs[0].SwitchMode(common.SyncModeSnapshot)
	require.Nil(t, err)

	//ctrl := gomock.NewController(t)
	//mockSnapConstructor := mock_sync.NewMockISyncConstructor(ctrl)
	//syncs[0] =
	// todo: mock snap sync
	startEpcNum := uint64(1)
	data, err := syncs[0].Prepare(common.WithPeers(peers),
		common.WithStartEpochChangeNum(startEpcNum),
		common.WithLatestPersistEpoch(0),
		common.WithSnapCurrentEpoch(3),
	)
	require.Nil(t, err)
	require.NotNil(t, data)
	epcs := data.Data.([]*consensus.EpochChange)
	require.Equal(t, 3, len(epcs))

	// start sync model except node0 (other nodes should not start sync, but need prepare be listener)
	for i := 1; i < n; i++ {
		_, err = syncs[i].Prepare()
		require.Nil(t, err)
	}

	// start sync
	syncTaskDoneCh := make(chan error, 1)
	err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash, 2, 2, 300, quorumCkpt300, epcs...), syncTaskDoneCh)
	require.Nil(t, err)
	err = <-syncTaskDoneCh
	require.Nil(t, err)

	chainData := <-syncs[0].Commit()
	require.NotNil(t, chainData)
	require.Equal(t, uint64(100), chainData.(*common.SnapCommitData).EpochState.Checkpoint.Height())

	chainData = <-syncs[0].Commit()
	require.NotNil(t, chainData)
	require.Equal(t, uint64(200), chainData.(*common.SnapCommitData).EpochState.Checkpoint.Height())

	chainData = <-syncs[0].Commit()
	require.NotNil(t, chainData)
	require.Equal(t, uint64(300), chainData.(*common.SnapCommitData).EpochState.Checkpoint.Height())

}

// todo: refactor it later
func TestTps(t *testing.T) {
	t.Skip()
	round := 100
	localId := 0
	begin := uint64(1)
	syncCount := 10000
	end := begin + uint64(syncCount) - 1
	epochInternal := 100
	n := 4
	roundDuration := make([]time.Duration, round)
	syncs, epochChanges := prepareBlockSyncs(t, epochInternal, localId, n, begin, end)
	// start sync model
	for i := 0; i < n; i++ {
		_, err := syncs[i].Prepare()
		require.Nil(t, err)
	}

	for i := 0; i < round; i++ {
		// start sync
		peers := []string{"1", "2", "3"}
		latestBlock, err := syncs[localId].getBlockFunc(begin - 1)
		require.Nil(t, err)
		latestBlockHash := latestBlock.BlockHash.String()
		remoteId := (localId + 1) % n
		endBlock, err := syncs[remoteId].getBlockFunc(end)
		require.Nil(t, err)
		quorumCkpt := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: end,
					Digest: endBlock.BlockHash.String(),
				},
			},
		}
		now := time.Now()
		// start sync commitDataCache
		syncTaskDone := make(chan error, 1)
		err = syncs[localId].StartSync(genSyncParams(peers, latestBlockHash, 2, begin, end, quorumCkpt, epochChanges...), syncTaskDone)
		require.Nil(t, err)
		err = <-syncTaskDone
		require.Nil(t, err)

		var taskDone bool
		for {
			select {
			case data := <-syncs[localId].Commit():
				blocks := data.([]common.CommitData)
				if blocks[len(blocks)-1].GetHeight() == end {
					taskDone = true
				}
			}
			if taskDone {
				break
			}
		}
		roundDuration[i] = time.Since(now)
		fmt.Printf("round%d cost time: %v\n", i, roundDuration[i])
		time.Sleep(1 * time.Millisecond)
	}
	var sum time.Duration
	lo.ForEach(roundDuration, func(duration time.Duration, _ int) {
		sum += duration
	})
	fmt.Printf("tps: %f\n", float64(syncCount*round)/sum.Seconds())
}

func TestPickPeer(t *testing.T) {
	t.Parallel()
	t.Run("test update peers, latestHeight equal target height", func(t *testing.T) {
		n := 4
		localId := "0"
		latestHeight := 10
		syncs, ledgers := newMockBlockSyncs(t, n)
		defer stopSyncs(syncs)
		prepareLedger(t, ledgers, localId, latestHeight)

		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}
		// node0 start sync commitDataCache
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash
		block10, err := ledgers[remoteId].GetBlock(10)
		require.Nil(t, err)
		quorumCkpt10 := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block10.Height(),
					Digest: block10.BlockHash.String(),
				},
			},
		}
		// start sync
		syncTaskDoneCh := make(chan error, 1)
		targetHeight := uint64(latestHeight)
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash.String(), 2, 2, targetHeight, quorumCkpt10), syncTaskDoneCh)
		require.Nil(t, err)
		err = <-syncTaskDoneCh
		require.Nil(t, err)

		// init peers' latest height in targetHeight
		lo.ForEach(syncs[0].peers, func(peer *common.Peer, _ int) {
			require.Equal(t, targetHeight, peer.LatestHeight)
		})
	})

	t.Run("test update peers, latestHeight is bigger than target height", func(t *testing.T) {
		n := 4
		localId := "0"
		latestHeight := uint64(10)
		syncs, ledgers := newMockBlockSyncs(t, n)
		defer stopSyncs(syncs)
		prepareLedger(t, ledgers, localId, int(latestHeight))

		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}
		// node0 start sync commitDataCache
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash
		targetHeight := latestHeight - 1
		block, err := ledgers[remoteId].GetBlock(targetHeight)
		require.Nil(t, err)
		quorumCkpt := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block.Height(),
					Digest: block.BlockHash.String(),
				},
			},
		}
		// start sync
		syncTaskDoneCh := make(chan error, 1)
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash.String(), 2, 2, targetHeight, quorumCkpt), syncTaskDoneCh)
		require.Nil(t, err)
		err = <-syncTaskDoneCh
		require.Nil(t, err)

		// peers' latest height had been updated(count is bigger than quorum)
		updatedCount := uint64(0)
		lo.ForEach(syncs[0].peers, func(peer *common.Peer, _ int) {
			if peer.LatestHeight == latestHeight {
				updatedCount++
			}
		})
		require.True(t, updatedCount >= syncs[0].quorum)
	})

	t.Run("test update peers, latestHeight is smaller than target height", func(t *testing.T) {
		n := 4
		localId := "0"
		latestHeight := uint64(10)
		syncs, ledgers := newMockBlockSyncs(t, n)
		defer stopSyncs(syncs)
		prepareLedger(t, ledgers, localId, int(latestHeight))

		for i := 0; i < n; i++ {
			_, err := syncs[i].Prepare()
			require.Nil(t, err)
		}
		// node0 start sync commitDataCache
		peers := []string{"1", "2", "3"}
		remoteId := "1"
		latestBlockHash := ledgers[localId].GetChainMeta().BlockHash
		targetHeight := latestHeight - 1
		block, err := ledgers[remoteId].GetBlock(targetHeight)
		require.Nil(t, err)

		// remove targetHeight block in remote peers(N - removeCount < quorum)
		N := len(peers)
		quorum := 2
		removeCount := N - quorum + 1

		// decrease latestHeight(smaller than targetHeight)
		for id, ledger := range ledgers {
			if id == localId {
				continue
			}
			ledger.chainMeta.Height = targetHeight - 1
			removeCount--
			if removeCount == 0 {
				break
			}
		}

		quorumCkpt := &consensus.SignedCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: block.Height(),
					Digest: block.BlockHash.String(),
				},
			},
		}
		// start sync
		syncTaskDoneCh := make(chan error, 1)
		err = syncs[0].StartSync(genSyncParams(peers, latestBlockHash.String(), uint64(quorum), 2, targetHeight, quorumCkpt), syncTaskDoneCh)
		require.Nil(t, err)
		err = <-syncTaskDoneCh
		require.Nil(t, err)

		// peers' latest height had been updated(count is bigger than quorum)
		updatedCount := uint64(0)
		lo.ForEach(syncs[0].peers, func(peer *common.Peer, _ int) {
			if peer.LatestHeight == latestHeight {
				updatedCount++
			}
		})
		require.Equal(t, 1, int(updatedCount))
	})
}
