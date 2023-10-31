package block_sync

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/types/pb"
)

func TestHandleSyncState(t *testing.T) {
	n := 4
	syncs := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, 2)
	err := syncs[0].Start()
	require.Nil(t, err)

	localId := 0
	remoteId := 1
	requestHeight := uint64(1)
	resp := getMockStateResponse(strconv.Itoa(remoteId), requestHeight)
	require.Nil(t, resp, "before request, the state response should be nil")
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
	syncs[remoteId].handleSyncState(nil, msg)
	resp = getMockStateResponse(strconv.Itoa(remoteId), requestHeight)
	require.NotNil(t, resp, "after request, the state response should not be nil")
	require.Equal(t, pb.Message_SYNC_STATE_RESPONSE, resp.Type)

	stateResp := &pb.SyncStateResponse{}
	err = stateResp.UnmarshalVT(resp.Data)
	require.Nil(t, err)
	require.Equal(t, requestHeight, stateResp.CheckpointState.Height)
	require.Equal(t, getMockChainMeta(remoteId).BlockHash.String(), stateResp.CheckpointState.Digest)

	t.Run("test send wrong state request msg type", func(t *testing.T) {
		cleanMockStateResponse()
		wrongMsg := &pb.Message{
			Type: pb.Message_SYNC_BLOCK_REQUEST,
			Data: reqData,
		}
		syncs[remoteId].handleSyncState(nil, wrongMsg)
		resp = getMockStateResponse(strconv.Itoa(remoteId), requestHeight)
		require.Nil(t, resp, "after wrong request, the state response should be nil")
	})

	t.Run("test send wrong state request msg data", func(t *testing.T) {
		cleanMockStateResponse()
		wrongMsg := &pb.Message{
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: []byte("wrong data"),
		}
		syncs[remoteId].handleSyncState(nil, wrongMsg)
		resp = getMockStateResponse(strconv.Itoa(remoteId), requestHeight)
		require.Nil(t, resp, "after wrong request, the state response should be nil")
	})

	t.Run("test get local block failed", func(t *testing.T) {
		meta := getMockChainMeta(remoteId)
		wrongBlockHeight := meta.Height + 1

		req = &pb.SyncStateRequest{
			Height: wrongBlockHeight,
		}

		data, err := req.MarshalVT()
		require.Nil(t, err)
		wrongMsg := &pb.Message{
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: data,
		}
		syncs[remoteId].handleSyncState(nil, wrongMsg)
		resp = getMockStateResponse(strconv.Itoa(remoteId), wrongBlockHeight)
		require.Nil(t, resp, "after wrong request, the state response should be nil")
	})

	t.Run("test send state response failed", func(t *testing.T) {
		cleanMockStateResponse()
		syncs[2].handleSyncState(nil, msg)
		resp = getMockStateResponse(strconv.Itoa(2), requestHeight)
		require.Nil(t, resp, "after send request failed, the state response should be nil")
	})
}
