package block_sync

import (
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types/pb"
	network "github.com/axiomesh/axiom-p2p"
)

func (bs *BlockSync) handleSyncState(s network.Stream, msg *pb.Message) {
	syncStateRequest := &pb.SyncStateRequest{}

	if msg.Type != pb.Message_SYNC_STATE_REQUEST {
		bs.logger.WithFields(logrus.Fields{
			"type": msg.Type,
		}).Error("Handle sync state request failed, wrong message type")
		return
	}

	if err := syncStateRequest.UnmarshalVT(msg.Data); err != nil {
		bs.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Unmarshal sync state request failed")
		return
	}

	block, err := bs.getBlockFunc(syncStateRequest.Height)
	if err != nil {
		bs.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Get block failed")

		return
	}

	// set checkpoint state in current block
	stateResp := &pb.SyncStateResponse{
		CheckpointState: &pb.CheckpointState{
			Height: block.Height(),
			Digest: block.BlockHash.String(),
		},
	}

	data, err := stateResp.MarshalVT()
	if err != nil {
		bs.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Marshal sync state response failed")

		return
	}

	// send sync state response
	if err = retry.Retry(func(attempt uint) error {
		err = bs.network.SendWithStream(s, &pb.Message{
			From: bs.network.PeerID(),
			Type: pb.Message_SYNC_STATE_RESPONSE,
			Data: data,
		})
		if err != nil {
			bs.logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Send sync state response failed")
			return err
		}
		return nil
	}, strategy.Limit(maxRetryCount), strategy.Wait(500*time.Millisecond)); err != nil {
		bs.logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Send sync state response failed")
		return
	}
}
