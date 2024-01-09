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

func (sm *SyncManager) handleSyncState(s network.Stream, msg *pb.Message) {
	syncStateRequest := &pb.SyncStateRequest{}

	if msg.Type != pb.Message_SYNC_STATE_REQUEST {
		sm.logger.WithFields(logrus.Fields{
			"type": msg.Type,
		}).Error("Handle sync state request failed, wrong message type")
		return
	}

	if err := syncStateRequest.UnmarshalVT(msg.Data); err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unmarshal sync state request failed")
		return
	}

	block, err := sm.getBlockFunc(syncStateRequest.Height)
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Get block failed")

		return
	}

	// set checkpoint state in current commitDataCache
	stateResp := &pb.SyncStateResponse{
		CheckpointState: &pb.CheckpointState{
			Height: block.Height(),
			Digest: block.BlockHash.String(),
		},
	}

	data, err := stateResp.MarshalVT()
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Marshal sync state response failed")

		return
	}

	// send sync state response
	if err = retry.Retry(func(attempt uint) error {
		err = sm.network.SendWithStream(s, &pb.Message{
			From: sm.network.PeerID(),
			Type: pb.Message_SYNC_STATE_RESPONSE,
			Data: data,
		})
		if err != nil {
			sm.logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Send sync state response failed")
			return err
		}
		return nil
	}, strategy.Limit(common.MaxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Send sync state response failed")
		return
	}
}

func (sm *SyncManager) handleFetchEpochState(s network.Stream, msg *pb.Message) {
	stateResp := sm.constructEpochStateResponse(msg)

	data, err := stateResp.MarshalVT()
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Marshal epoch state response failed")
		return
	}

	// send fetch epoch state response
	if err = retry.Retry(func(attempt uint) error {
		err = sm.network.SendWithStream(s, &pb.Message{
			From: sm.network.PeerID(),
			Type: pb.Message_FETCH_EPOCH_STATE_RESPONSE,
			Data: data,
		})
		if err != nil {
			sm.logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Send epoch state response failed")
			return err
		}
		return nil
	}, strategy.Limit(common.MaxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Send epoch state response failed")
		return
	}
}

func (sm *SyncManager) constructEpochStateResponse(msg *pb.Message) *pb.FetchEpochStateResponse {
	req := &pb.FetchEpochStateRequest{}
	if msg.Type != pb.Message_FETCH_EPOCH_STATE_REQUEST {
		err := fmt.Errorf("handle fetch epoch state request failed, wrong message type:%v", msg.Type)
		sm.logger.Error(err)
		return wrapFailedSyncStateResp(err.Error())
	}

	if err := req.UnmarshalVT(msg.Data); err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unmarshal fetch epoch state request failed")
		return wrapFailedSyncStateResp(fmt.Errorf("unmarshal fetch epoch state request failed: %v", err).Error())
	}

	eps, err := rbft.GetEpochQuorumCheckpoint(sm.getEpochStateFunc, req.Epoch)
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("get epochState failed")
		return wrapFailedSyncStateResp(fmt.Errorf("get epochState failed: %v", err).Error())
	}

	data, err := eps.MarshalVT()
	if err != nil {
		sm.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Marshal fetch epoch state response failed")
		return wrapFailedSyncStateResp(err.Error())
	}

	return wrapSuccessSyncStateResp(data)
}

func wrapFailedSyncStateResp(err string) *pb.FetchEpochStateResponse {
	return &pb.FetchEpochStateResponse{
		Status: pb.Status_ERROR,
		Error:  err,
	}
}

func wrapSuccessSyncStateResp(data []byte) *pb.FetchEpochStateResponse {
	return &pb.FetchEpochStateResponse{
		Data:   data,
		Status: pb.Status_SUCCESS,
	}
}
