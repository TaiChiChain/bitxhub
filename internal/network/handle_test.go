package network

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/types/pb"
	p2p "github.com/axiomesh/axiom-p2p"
)

func TestHandleMessage(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	stateResponse := make([]*pb.Message, 0)
	doneCh := make(chan struct{})
	err := swarms[1].RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, func(stream p2p.Stream, msg *pb.Message) {
		resp := &pb.Message{
			Type: pb.Message_SYNC_STATE_RESPONSE,
			Data: []byte("response aaa"),
		}
		stateResponse = append(stateResponse, resp)
		doneCh <- struct{}{}
	})

	require.Nil(t, err)

	t.Run("handle wrong message data", func(t *testing.T) {
		swarms[0].handleMessage(nil, nil)
		require.Equal(t, 0, len(stateResponse), "should not return response with wrong message data")
	})

	t.Run("handle wrong message type", func(t *testing.T) {
		msg := &pb.Message{
			Type: pb.Message_SYNC_BLOCK_REQUEST,
			Data: []byte("request aaa"),
		}
		req, err := msg.MarshalVT()
		require.Nil(t, err)
		swarms[0].handleMessage(nil, req)
		require.Equal(t, 0, len(stateResponse), "should not return response with wrong message type")
	})

	t.Run("handle wrong message handler", func(t *testing.T) {
		wrongHandler := func() {}
		swarms[0].msgHandlers.Store(pb.Message_SYNC_BLOCK_REQUEST, wrongHandler) // wrong handler type
		msg := &pb.Message{
			Type: pb.Message_SYNC_BLOCK_REQUEST,
			Data: []byte("request aaa"),
		}
		req, err := msg.MarshalVT()
		require.Nil(t, err)
		swarms[0].handleMessage(nil, req)
		require.Equal(t, 0, len(stateResponse), "should not return response with wrong message type")
	})
}
