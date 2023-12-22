package snap_sync

import (
	"context"
	"fmt"
	"testing"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSnapSync_Mode(t *testing.T) {
	sync, cancel := mockSnapSync(nil)
	defer stopSnapSync(t, cancel)
	require.Equal(t, common.SyncModeSnapshot, sync.Mode())
}

func TestSnapSync_Commit(t *testing.T) {
	commitData := []common.CommitData{
		&common.ChainData{Block: &types.Block{BlockHeader: &types.BlockHeader{Number: 1, Epoch: 1}}},
		&common.ChainData{Block: &types.Block{BlockHeader: &types.BlockHeader{Number: 2, Epoch: 1}}},
	}

	sync, cancel := mockSnapSync(nil)
	defer stopSnapSync(t, cancel)

	epochStates := make(map[uint64]*consensus.QuorumCheckpoint)
	epochStates[1] = &consensus.QuorumCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 2,
				Digest: "test2",
			},
		},
	}
	sync.(*SnapSync).epochStateCache = epochStates

	go func() {
		sync.(*SnapSync).listenCommitData()
	}()
	sync.PostCommitData(commitData)

	res := <-sync.Commit()
	require.Equal(t, "test2", res.(*common.SnapCommitData).EpochState.Checkpoint.Digest())
}

func TestSnapSync_Prepare(t *testing.T) {
	peersSet := []string{"peer1", "peer2", "peer3"}

	epcStates := make(map[string]map[uint64]*consensus.QuorumCheckpoint)
	epcStates[peersSet[0]] = make(map[uint64]*consensus.QuorumCheckpoint)
	epcStates[peersSet[0]][1] = &consensus.QuorumCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			Epoch: 1,
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 100,
				Digest: "test100",
			},
		},
	}
	epcStates[peersSet[0]][2] = &consensus.QuorumCheckpoint{
		Checkpoint: &consensus.Checkpoint{
			Epoch: 2,
			ExecuteState: &consensus.Checkpoint_ExecuteState{
				Height: 200,
				Digest: "test200",
			},
		},
	}

	t.Run("test fetch epoch state failed", func(t *testing.T) {
		latestBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 133,
				Epoch:  2,
			},
		}
		conf := &common.Config{
			Peers:               peersSet,
			StartEpochChangeNum: latestBlock.BlockHeader.Epoch,
			SnapPersistedEpoch:  1,
			LatestPersistEpoch:  0,
		}
		sync, cancel := mockSnapSync(mockNetwork(t, nil))
		defer stopSnapSync(t, cancel)
		data, err := sync.Prepare(conf)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "all peers invalid")
		require.Nil(t, data)
	})
	t.Run("test fill epoch change failed", func(t *testing.T) {
		latestBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 133,
				Epoch:  1,
			},
		}
		conf := &common.Config{
			Peers:               peersSet,
			StartEpochChangeNum: latestBlock.BlockHeader.Epoch,
			SnapPersistedEpoch:  2,
			LatestPersistEpoch:  2,
		}
		sync, cancel := mockSnapSync(mockNetwork(t, epcStates))
		defer stopSnapSync(t, cancel)
		data, err := sync.Prepare(conf)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "epoch 1 not found")
		require.Nil(t, data)
	})

	t.Run("test snap sync need not prepare epoch change", func(t *testing.T) {
		latestBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 133,
				Epoch:  2,
			},
		}
		conf := &common.Config{
			Peers:               peersSet,
			StartEpochChangeNum: latestBlock.BlockHeader.Epoch,
			SnapPersistedEpoch:  1,
			LatestPersistEpoch:  1,
		}
		sync, cancel := mockSnapSync(mockNetwork(t, epcStates))
		defer stopSnapSync(t, cancel)
		data, err := sync.Prepare(conf)
		require.Nil(t, err)
		require.Equal(t, 0, len(data.Data.([]*consensus.EpochChange)))
	})

	t.Run("test snap sync need prepare epoch change", func(t *testing.T) {
		// test snap sync need prepare epoch change
		latestBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 1,
				Epoch:  1,
			},
		}
		conf := &common.Config{
			Peers:               peersSet,
			StartEpochChangeNum: latestBlock.BlockHeader.Epoch,
			SnapPersistedEpoch:  2,
			LatestPersistEpoch:  0,
		}
		sync, cancel := mockSnapSync(mockNetwork(t, epcStates))
		defer stopSnapSync(t, cancel)

		data, err := sync.Prepare(conf)
		require.Nil(t, err)
		require.Equal(t, data.Data.([]*consensus.EpochChange)[0].Checkpoint.Digest(), "test100")
		require.Equal(t, data.Data.([]*consensus.EpochChange)[1].Checkpoint.Digest(), "test200")
	})

	t.Run("test latestBlock EpochHeight is bigger than snapPersistedEpoch", func(t *testing.T) {
		sync, cancel := mockSnapSync(mockNetwork(t, epcStates))
		defer stopSnapSync(t, cancel)

		latestBlock := &types.Block{
			BlockHeader: &types.BlockHeader{
				Number: 1,
				Epoch:  1,
			},
		}
		conf := &common.Config{
			Peers:               peersSet,
			StartEpochChangeNum: latestBlock.BlockHeader.Epoch,
			SnapPersistedEpoch:  2,
			LatestPersistEpoch:  0,
		}
		data, err := sync.Prepare(conf)
		require.Nil(t, err)
		require.Equal(t, data.Data.([]*consensus.EpochChange)[0].Checkpoint.Digest(), "test100")
		require.Equal(t, data.Data.([]*consensus.EpochChange)[1].Checkpoint.Digest(), "test200")
	})
}

func mockNetwork(t *testing.T, mockEpcStates map[string]map[uint64]*consensus.QuorumCheckpoint) network.Network {
	ctrl := gomock.NewController(t)
	net := mock_network.NewMockNetwork(ctrl)
	net.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(func(peerId string, msg *pb.Message) (*pb.Message, error) {
		req := &pb.FetchEpochStateRequest{}
		err := req.UnmarshalVT(msg.GetData())
		require.Nil(t, err)

		if states, ok := mockEpcStates[peerId]; ok {
			epcState, ok := states[req.Epoch]
			if !ok {
				return nil, fmt.Errorf("no state for epoch %d", req.Epoch)
			}
			epcStateBytes, err := epcState.MarshalVT()
			require.Nil(t, err)
			resp := &pb.FetchEpochStateResponse{
				Data:   epcStateBytes,
				Status: pb.Status_SUCCESS,
			}
			respBytes, err := resp.MarshalVT()
			require.Nil(t, err)
			return &pb.Message{
				Type: pb.Message_FETCH_EPOCH_STATE_RESPONSE,
				Data: respBytes,
			}, nil
		}
		return nil, fmt.Errorf("no state for peer %s", peerId)
	}).AnyTimes()
	return net
}

func mockSnapSync(network network.Network) (common.ISyncConstructor, context.CancelFunc) {
	logger := log.NewWithModule("snap_sync_test")
	logger.Logger.SetLevel(logrus.DebugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	return NewSnapSync(logger, ctx, network), cancel
}

func stopSnapSync(t *testing.T, cancel context.CancelFunc) {
	cancel()
}
