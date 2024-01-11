package sync

import (
	"fmt"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types/pb"
	network "github.com/axiomesh/axiom-p2p"
)

type Response interface {
	MarshalVT() ([]byte, error)
}

func (sm *SyncManager) handleSyncState(s network.Stream, msg *pb.Message) {
	stateResp := sm.constructSyncStateResponse(msg)
	sm.sendResponse(s, pb.Message_SYNC_STATE_RESPONSE, stateResp)
}

func (sm *SyncManager) handleFetchEpochState(s network.Stream, msg *pb.Message) {
	stateResp := sm.constructEpochStateResponse(msg)
	sm.sendResponse(s, pb.Message_FETCH_EPOCH_STATE_RESPONSE, stateResp)
}

func (sm *SyncManager) sendResponse(s network.Stream, typ pb.Message_Type, stateResp Response) {
	data, err := stateResp.MarshalVT()
	if err != nil {
		err = fmt.Errorf("marshal fetch epoch state response failed: %v", err)
		sm.logger.Error(err)
		return
	}

	err = retry.Retry(func(attempt uint) error {
		err := sm.network.SendWithStream(s, &pb.Message{
			From: sm.network.PeerID(),
			Type: typ,
			Data: data,
		})
		if err != nil {
			sm.logger.WithFields(logrus.Fields{
				"err": err,
			}).Errorf("Send %v response failed", typ)
		}
		return err
	}, strategy.Limit(common.MaxRetryCount), strategy.Wait(500*time.Millisecond))

	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("Send %v response failed", typ)
	}
}

func (sm *SyncManager) constructSyncStateResponse(msg *pb.Message) Response {
	syncStateRequest := &pb.SyncStateRequest{}
	var err error

	if msg.Type != pb.Message_SYNC_STATE_REQUEST {
		err = fmt.Errorf("wrong message type: %v", msg.Type)
		sm.logger.WithFields(logrus.Fields{
			"type": msg.Type,
		}).Error(err)

		return wrapFailedStateResp(pb.Message_SYNC_STATE_RESPONSE, err)
	}

	if err := syncStateRequest.UnmarshalVT(msg.Data); err != nil {
		err = fmt.Errorf("unmarshal sync state request failed: %v", err)
		sm.logger.Error(err)
		return wrapFailedStateResp(pb.Message_SYNC_STATE_RESPONSE, err)
	}

	block, err := sm.getBlockFunc(syncStateRequest.Height)
	if err != nil {
		err = fmt.Errorf("get block failed: %v", err)
		sm.logger.Error(err)
		// if block not found, return latest block height
		return &pb.SyncStateResponse{
			Status: pb.Status_ERROR,
			Error:  err.Error(),
			CheckpointState: &pb.CheckpointState{
				LatestHeight: sm.getChainMetaFunc().Height,
			},
		}
	}

	// set checkpoint state in current commitDataCache
	checkpointState := &pb.CheckpointState{
		Height:       block.Height(),
		Digest:       block.BlockHash.String(),
		LatestHeight: sm.getChainMetaFunc().Height,
	}
	return wrapSuccessStateResp(pb.Message_SYNC_STATE_RESPONSE, checkpointState)
}

func (sm *SyncManager) constructEpochStateResponse(msg *pb.Message) Response {
	req := &pb.FetchEpochStateRequest{}
	if msg.Type != pb.Message_FETCH_EPOCH_STATE_REQUEST {
		err := fmt.Errorf("wrong message type: %v", msg.Type)
		sm.logger.Error(err)
		return wrapFailedStateResp(pb.Message_FETCH_EPOCH_STATE_RESPONSE, err)
	}

	if err := req.UnmarshalVT(msg.Data); err != nil {
		err = fmt.Errorf("unmarshal fetch epoch state request failed: %v", err)
		sm.logger.Error(err)
		return wrapFailedStateResp(pb.Message_FETCH_EPOCH_STATE_RESPONSE, err)
	}

	eps, err := rbft.GetEpochQuorumCheckpoint(sm.getEpochStateFunc, req.Epoch)
	if err != nil {
		err = fmt.Errorf("get epoch state failed: %v", err)
		sm.logger.Error(err.Error())
		return wrapFailedStateResp(pb.Message_FETCH_EPOCH_STATE_RESPONSE, err)
	}

	data, err := eps.MarshalVT()
	if err != nil {
		err = fmt.Errorf("marshal fetch epoch state response failed: %v", err)
		sm.logger.Error(err.Error())
		return wrapFailedStateResp(pb.Message_FETCH_EPOCH_STATE_RESPONSE, err)
	}

	return wrapSuccessStateResp(pb.Message_FETCH_EPOCH_STATE_RESPONSE, data)
}

func wrapFailedStateResp(typ pb.Message_Type, err error) Response {
	switch typ {
	case pb.Message_FETCH_EPOCH_STATE_RESPONSE:
		return &pb.FetchEpochStateResponse{
			Status: pb.Status_ERROR,
			Error:  err.Error(),
		}
	case pb.Message_SYNC_STATE_RESPONSE:
		return &pb.SyncStateResponse{
			Status: pb.Status_ERROR,
			Error:  err.Error(),
		}
	}
	return nil
}

func wrapSuccessStateResp(typ pb.Message_Type, data any) Response {
	switch typ {
	case pb.Message_FETCH_EPOCH_STATE_RESPONSE:
		return &pb.FetchEpochStateResponse{
			Status: pb.Status_SUCCESS,
			Data:   data.([]byte),
		}
	case pb.Message_SYNC_STATE_RESPONSE:
		return &pb.SyncStateResponse{
			Status:          pb.Status_SUCCESS,
			CheckpointState: data.(*pb.CheckpointState),
		}
	}
	return nil
}
