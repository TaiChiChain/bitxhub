package block_sync

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
)

func TestStartSync(t *testing.T) {
	n := 4
	syncs := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 10)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}

	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()
	remoteBlockHash := getMockChainMeta(1).BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 10,
				Digest: remoteBlockHash,
			},
		},
	}

	t.Run("test switch sync status err", func(t *testing.T) {
		err := syncs[0].switchSyncStatus(true)
		require.Nil(t, err)
		err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 10, quorumCkpt)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "status is already true")
		err = syncs[0].switchSyncStatus(false)
		require.Nil(t, err)
	})

	err := syncs[0].StartSync(peers, latestBlockHash, 2, 2, 10, quorumCkpt)
	require.Nil(t, err)
	defer func() {
		err = syncs[0].StopSync()
		require.Nil(t, err)
	}()

	blocks := <-syncs[0].Commit()
	require.Equal(t, 9, len(blocks))
	require.Equal(t, uint64(10), blocks[len(blocks)-1].Height())
}

func TestStartSyncWithRemoteSendBlockResponseError(t *testing.T) {
	n := 4
	syncs := newMockBlockSyncs(t, n, wrongTypeSendSyncBlockResponse, 0, 2)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 10)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}

	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()
	remoteBlockHash := getMockChainMeta(1).BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 10,
				Digest: remoteBlockHash,
			},
		},
	}

	err := syncs[0].StartSync(peers, latestBlockHash, 2, 2, 10, quorumCkpt)
	require.Nil(t, err)
	defer func() {
		err = syncs[0].StopSync()
		require.Nil(t, err)
	}()

	blocks := <-syncs[0].Commit()
	require.Equal(t, 9, len(blocks))
	require.Equal(t, uint64(10), blocks[len(blocks)-1].Height())
}

func TestMultiEpochSync(t *testing.T) {
	n := 4
	syncs := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 300)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}

	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()
	remoteBlockHash := getMockChainMeta(1).BlockHash.String()
	quorumCkpt300 := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 300,
				Digest: remoteBlockHash,
			},
		},
	}

	block100, err := getMockBlockLedger(100, 1)
	require.Nil(t, err)
	block200, err := getMockBlockLedger(200, 1)
	require.Nil(t, err)
	block300, err := getMockBlockLedger(300, 1)
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

	err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 300, quorumCkpt300, epc1, epc2, epc3)
	require.Nil(t, err)

	blocks1 := <-syncs[0].Commit()
	require.Equal(t, 99, len(blocks1))
	require.Equal(t, uint64(100), blocks1[len(blocks1)-1].Height())
	require.True(t, syncs[0].syncStatus.Load())
	blocks2 := <-syncs[0].Commit()
	require.Equal(t, 100, len(blocks2))
	require.Equal(t, uint64(200), blocks2[len(blocks2)-1].Height())
	blocks3 := <-syncs[0].Commit()
	require.Equal(t, 100, len(blocks3))
	require.Equal(t, uint64(300), blocks3[len(blocks3)-1].Height())

	err = syncs[0].StopSync()
	require.Nil(t, err)
	require.False(t, syncs[0].syncStatus.Load())

	t.Run("TestMultiEpochSyncWithWrongBlock, wrong block is not epoch block", func(t *testing.T) {
		// mock wrong block
		wrongHeight := uint64(7)
		oldRightBlock, err := getMockBlockLedger(wrongHeight, 1)
		require.Nil(t, err)
		parentBlock, err := getMockBlockLedger(wrongHeight-1, 1)
		require.Nil(t, err)
		wrongBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number:     wrongHeight,
				ParentHash: parentBlock.BlockHash,
			},
			BlockHash: types.NewHash([]byte("wrong_block")),
		}
		id, err := strconv.Atoi(syncs[0].pickPeer(wrongHeight))
		require.Nil(t, err)
		setMockBlockLedger(wrongBlock, id)

		block10, err := getMockBlockLedger(10, 1)
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
		err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 10, quorumCkpt10)
		require.Nil(t, err)

		blocks1 = <-syncs[0].Commit()
		require.Equal(t, 9, len(blocks1))
		require.Equal(t, uint64(10), blocks1[len(blocks1)-1].Height())
		require.Equal(t, block10.BlockHash, blocks1[len(blocks1)-1].BlockHash)
		require.True(t, syncs[0].syncStatus.Load())

		// reset right block
		setMockBlockLedger(oldRightBlock, id)
		err = syncs[0].StopSync()
		require.Nil(t, err)
		require.False(t, syncs[0].syncStatus.Load())
	})

	t.Run("TestMultiEpochSyncWithWrongBlock, wrong block is epoch block", func(t *testing.T) {
		// mock wrong block
		wrongHeight := uint64(100)
		oldRightBlock, err := getMockBlockLedger(wrongHeight, 1)
		require.Nil(t, err)
		parentBlock, err := getMockBlockLedger(wrongHeight-1, 1)
		require.Nil(t, err)
		wrongBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number:     wrongHeight,
				ParentHash: parentBlock.BlockHash,
			},
			BlockHash: types.NewHash([]byte("wrong_block")),
		}
		id, err := strconv.Atoi(syncs[0].pickPeer(wrongHeight))
		require.Nil(t, err)
		setMockBlockLedger(wrongBlock, id)

		block101, err := getMockBlockLedger(101, 1)
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
		err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 101, quorumCkpt101, epc1)
		require.Nil(t, err)

		blocks1 = <-syncs[0].Commit()
		require.Equal(t, 99, len(blocks1))
		require.Equal(t, uint64(100), blocks1[len(blocks1)-1].Height())
		require.True(t, syncs[0].syncStatus.Load())
		blocks2 = <-syncs[0].Commit()
		require.Equal(t, 1, len(blocks2))
		require.Equal(t, uint64(101), blocks2[len(blocks2)-1].Height())

		// reset right block
		setMockBlockLedger(oldRightBlock, id)
		err = syncs[0].StopSync()
		require.Nil(t, err)
		require.False(t, syncs[0].syncStatus.Load())
	})
}

func TestHandleTimeoutBlockMsg(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync request error
	syncs := newMockBlockSyncs(t, n)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 100)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}
	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()

	block100, err := getMockBlockLedger(100, 1)
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

	t.Run("TestSyncTimeoutBlock with one time", func(t *testing.T) {
		timeoutBlockHeight := uint64(7)
		idx := int(timeoutBlockHeight % uint64(len(peers)))
		wrongId := idx + 1

		oldRightBlock, err := getMockBlockLedger(timeoutBlockHeight, wrongId)
		require.Nil(t, err)
		deleteMockBlockLedger(timeoutBlockHeight, wrongId)

		block10, err := getMockBlockLedger(10, 1)
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
		err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 10, quorumCkpt10)
		require.Nil(t, err)
		blocks1 := <-syncs[0].Commit()
		require.Equal(t, strconv.Itoa(wrongId), syncs[0].peers[idx].peerID)
		require.Equal(t, uint64(1), syncs[0].peers[idx].timeoutCount, "record timeout count")
		require.Equal(t, 3, len(syncs[0].peers), "not remove timeout peer because timeoutCount < timeoutCountLimit")
		require.Equal(t, 9, len(blocks1))
		require.Equal(t, uint64(10), blocks1[len(blocks1)-1].Height())
		require.Equal(t, oldRightBlock.BlockHash.String(), blocks1[timeoutBlockHeight-2].BlockHash.String())

		// reset right block
		setMockBlockLedger(oldRightBlock, wrongId)
		err = syncs[0].StopSync()
		require.Nil(t, err)
		require.False(t, syncs[0].syncStatus.Load())
	})

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
		err = syncs[0].StartSync(peers, latestBlockHash, 2, 2, 100, quorumCkpt100, epc1)
		require.Nil(t, err)
		blocks1 := <-syncs[0].Commit()
		require.Equal(t, 99, len(blocks1))
		require.Equal(t, uint64(100), blocks1[len(blocks1)-1].Height())
		require.Equal(t, 2, len(syncs[0].peers), "remove timeout peer")

		err = syncs[0].StopSync()
		require.Nil(t, err)
		require.False(t, syncs[0].syncStatus.Load())
	})
}

func TestHandleSyncErrMsg(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync request error
	syncs := newMockBlockSyncs(t, n, wrongTypeSendSyncBlockRequest, 0, 1)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 100)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}
	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()
	remoteBlockHash := getMockChainMeta(1).BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 100,
				Digest: remoteBlockHash,
			},
		},
	}
	err := syncs[0].StartSync(peers, latestBlockHash, 2, 2, 100, quorumCkpt)
	require.Nil(t, err)
	defer func() {
		err = syncs[0].StopSync()
		require.Nil(t, err)
	}()

	blocks := <-syncs[0].Commit()
	require.Equal(t, 99, len(blocks))
	require.Equal(t, uint64(100), blocks[len(blocks)-1].Height())
}

func TestSendSyncStateError(t *testing.T) {
	n := 4
	// mock syncs[0] which send sync state request error
	syncs := newMockBlockSyncs(t, n, wrongTypeSendSyncState, 0, 1)
	defer stopSyncs(syncs)

	// store blocks expect node 0
	prepareLedger(n, 100)

	// start sync model
	for i := 0; i < n; i++ {
		err := syncs[i].Start()
		require.Nil(t, err)
	}
	// node0 start sync block
	peers := []string{"1", "2", "3"}
	latestBlockHash := getMockChainMeta(0).BlockHash.String()
	remoteBlockHash := getMockChainMeta(1).BlockHash.String()
	quorumCkpt := &consensus.SignedCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 100,
				Digest: remoteBlockHash,
			},
		},
	}
	err := syncs[0].StartSync(peers, latestBlockHash, 2, 2, 100, quorumCkpt)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "timeout send sync request")

	err = syncs[0].StopSync()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "status is already false")
}

func TestValidateChunk(t *testing.T) {
	n := 4
	syncs := newMockBlockSyncs(t, n)

	syncs[0].latestCheckedState = &pb.CheckpointState{
		Height: getMockChainMeta(0).Height,
		Digest: getMockChainMeta(0).BlockHash.String(),
	}
	syncs[0].curHeight = 1
	syncs[0].chunk = &chunk{
		chunkSize: 1,
	}

	t.Run("get requester err: nil", func(t *testing.T) {
		_, err := syncs[0].validateChunk()
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "requester[height:1] is nil")
	})

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

	t.Run("get requester block err: block is nil", func(t *testing.T) {
		invalidMsgs, err := syncs[0].validateChunk()
		require.Nil(t, err)
		require.Equal(t, 1, len(invalidMsgs))
	})
}
