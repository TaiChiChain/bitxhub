package sync

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/types/pb"
)

func TestHandleSyncState(t *testing.T) {
	n := 4
	syncs, ledgers := newMockBlockSyncs(t, n, wrongTypeSendStream, 0, 2)
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
}
