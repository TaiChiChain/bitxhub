package sync

import (
	"strconv"
	"testing"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/stretchr/testify/require"
)

func TestHandleSyncState(t *testing.T) {
	t.Parallel()
	t.Run("handle sync state success", func(t *testing.T) {
		n := 4
		syncs, ledgers := newMockBlockSyncs(t, n)
		_, err := syncs[0].Prepare()
		require.Nil(t, err)

		localId := 0
		remoteId := 1
		requestHeight := uint64(1)

		req := &pb.SyncStateRequest{
			Height: requestHeight,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: reqData,
		}

		prepareLedger(t, ledgers, strconv.Itoa(localId), 100)
		syncs[remoteId].handleSyncState(nil, msg)
		resp := <-ledgers[strconv.Itoa(remoteId)].stateResponse
		require.NotNil(t, resp, "after request, the state response should not be nil")
		require.Equal(t, pb.Message_SYNC_STATE_RESPONSE, resp.Type)

		stateResp := &pb.SyncStateResponse{}
		err = stateResp.UnmarshalVT(resp.Data)
		require.Nil(t, err)
		require.Equal(t, requestHeight, stateResp.CheckpointState.Height)
		expectBlock, err := ledgers[strconv.Itoa(remoteId)].GetBlock(stateResp.CheckpointState.Height)
		require.Nil(t, err)
		require.Equal(t, expectBlock.BlockHash.String(), stateResp.CheckpointState.Digest)
	})
	t.Run("wrong message type", func(t *testing.T) {
		n := 4
		syncs, _ := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, 2)
		wrongMsg := &pb.Message{
			From: strconv.Itoa(0),
			Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
		}
		syncs[0].handleSyncState(nil, wrongMsg)
	})
	t.Run("wrong message data", func(t *testing.T) {
		n := 4
		syncs, _ := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, 2)
		wrongMsg := &pb.Message{
			From: strconv.Itoa(0),
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: []byte("wrong data"),
		}
		syncs[0].handleSyncState(nil, wrongMsg)
	})

	t.Run("get state failed", func(t *testing.T) {
		n := 4
		syncs, _ := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, 2)
		requestHeight := uint64(100)

		req := &pb.SyncStateRequest{
			Height: requestHeight,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(1),
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: reqData,
		}
		syncs[0].handleSyncState(nil, msg)
	})

	t.Run("send response failed", func(t *testing.T) {
		n := 4
		wrongId := 2
		syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, wrongId)
		_, err := syncs[0].Prepare()
		require.Nil(t, err)

		localId := 0
		remoteId := wrongId
		requestHeight := uint64(1)

		req := &pb.SyncStateRequest{
			Height: requestHeight,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: reqData,
		}

		prepareLedger(t, ledgers, strconv.Itoa(localId), 100)
		syncs[remoteId].handleSyncState(nil, msg)
	})
}

func TestHandleFetchEpochState(t *testing.T) {
	t.Parallel()
	t.Run("handle fetch epoch state success", func(t *testing.T) {
		n := 4
		syncs, ledgers := newMockBlockSyncs(t, n)
		_, err := syncs[0].Prepare()
		require.Nil(t, err)

		localId := 0
		remoteId := 1
		epoch := uint64(1)

		req := &pb.FetchEpochStateRequest{
			Epoch: epoch,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
			Data: reqData,
		}

		prepareLedger(t, ledgers, strconv.Itoa(localId), 100)
		syncs[remoteId].handleFetchEpochState(nil, msg)
		resp := <-ledgers[strconv.Itoa(remoteId)].epochStateResponse
		require.NotNil(t, resp, "after request, the epoch state response should not be nil")
		require.Equal(t, pb.Message_FETCH_EPOCH_STATE_RESPONSE, resp.Type)

		epsResp := &pb.FetchEpochStateResponse{}
		err = epsResp.UnmarshalVT(resp.Data)
		require.Nil(t, err)
		require.Equal(t, pb.Status_SUCCESS, epsResp.Status)
		epochState := &consensus.QuorumCheckpoint{}
		err = epochState.UnmarshalVT(epsResp.Data)
		require.Nil(t, err)
		require.Equal(t, epoch, epochState.Epoch())
	})
	t.Run("wrong message type", func(t *testing.T) {
		n := 4
		localId := 0
		remoteId := 1
		syncs, ledgers := newMockBlockSyncs(t, n)
		wrongMsg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_SYNC_STATE_REQUEST,
		}
		syncs[remoteId].handleFetchEpochState(nil, wrongMsg)
		resp := <-ledgers[strconv.Itoa(remoteId)].epochStateResponse
		require.NotNil(t, resp)
		require.Equal(t, pb.Message_FETCH_EPOCH_STATE_RESPONSE, resp.Type)

		epsResp := &pb.FetchEpochStateResponse{}
		err := epsResp.UnmarshalVT(resp.Data)
		require.Nil(t, err)
		require.Equal(t, pb.Status_ERROR, epsResp.Status)
		require.Contains(t, epsResp.Error, "wrong message type")
	})
	t.Run("wrong message data", func(t *testing.T) {
		n := 4
		localId := 0
		remoteId := 1
		syncs, ledgers := newMockBlockSyncs(t, n)
		wrongMsg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
			Data: []byte("wrong data"),
		}
		syncs[remoteId].handleFetchEpochState(nil, wrongMsg)

		resp := <-ledgers[strconv.Itoa(remoteId)].epochStateResponse
		require.NotNil(t, resp)
		require.Equal(t, pb.Message_FETCH_EPOCH_STATE_RESPONSE, resp.Type)

		epsResp := &pb.FetchEpochStateResponse{}
		err := epsResp.UnmarshalVT(resp.Data)
		require.Nil(t, err)
		require.Equal(t, pb.Status_ERROR, epsResp.Status)
		require.Contains(t, epsResp.Error, "unmarshal fetch epoch state request failed")
	})

	t.Run("get epoch state failed", func(t *testing.T) {
		n := 4
		localId := 0
		remoteId := 1
		syncs, ledgers := newMockBlockSyncs(t, n)
		epoch := uint64(1)

		req := &pb.FetchEpochStateRequest{
			Epoch: epoch,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
			Data: reqData,
		}
		syncs[remoteId].handleFetchEpochState(nil, msg)

		resp := <-ledgers[strconv.Itoa(remoteId)].epochStateResponse
		require.NotNil(t, resp)
		require.Equal(t, pb.Message_FETCH_EPOCH_STATE_RESPONSE, resp.Type)

		epsResp := &pb.FetchEpochStateResponse{}
		err = epsResp.UnmarshalVT(resp.Data)
		require.Nil(t, err)
		require.Equal(t, pb.Status_ERROR, epsResp.Status)
		require.Contains(t, epsResp.Error, "get epochState failed")
	})

	t.Run("send response failed", func(t *testing.T) {
		n := 4
		wrongId := 2
		syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, wrongId)
		_, err := syncs[0].Prepare()
		require.Nil(t, err)

		epoch := uint64(1)
		localId := 0
		remoteId := wrongId
		req := &pb.FetchEpochStateRequest{
			Epoch: epoch,
		}
		reqData, err := req.MarshalVT()
		require.Nil(t, err)

		msg := &pb.Message{
			From: strconv.Itoa(localId),
			Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
			Data: reqData,
		}

		prepareLedger(t, ledgers, strconv.Itoa(localId), 100)
		syncs[remoteId].handleFetchEpochState(nil, msg)
	})
}
