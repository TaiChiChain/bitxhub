package data_syncer

import (
	"fmt"
	"reflect"

	"github.com/axiomesh/axiom-kit/types/pb"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	network "github.com/axiomesh/axiom-p2p"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/messages"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/consensus/epocher"
)

// handleTxsMsg for data syncer, if receiving txs msg, just forward to bind node(until bind node is validator)
func (n *Node) handleTxsMsg(_ network.Stream, msg *pb.Message) {
	selfMsg := &pb.Message{
		From: n.chainState.SelfNodeInfo.P2PID,
		Type: pb.Message_TRANSACTION_PROPAGATION,
		Data: msg.Data,
	}
	if err := n.network.AsyncSend(n.bindNode.p2pID, selfMsg); err != nil {
		n.logger.Errorf("failed to send tx propagation: %v", err)
	}
}

func (n *Node) handleEpochMsg(stream network.Stream, msg *pb.Message) {
	switch msg.Type {
	case pb.Message_EPOCH_REQUEST:
		var req protos.EpochChangeRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			n.logger.Errorf("failed to unmarshal epoch request: %v", err)
			return
		}
		if err = checkEpochRequest(req); err != nil {
			n.logger.Errorf("invalid epoch request: %v", err)
			return
		}
		proof, err := n.pagingGetEpochChangeProof(req.GetStartEpoch(), req.GetTargetEpoch(), epocher.MaxNumEpochEndingCheckpoint)
		if err != nil {
			n.logger.Errorf("failed to get epoch change proof: %v", err)
			return
		}
		resp := &messages.EpochChangeResponse{EpochChangeResponse: protos.EpochChangeResponse{Proof: proof.Proto()}}
		respBytes, err := resp.Marshal()
		if err != nil {
			n.logger.Errorf("failed to marshal epoch response: %v", err)
			return
		}

		respMsg := &pb.Message{
			Type: pb.Message_EPOCH_REQUEST_RESPONSE,
			Data: respBytes,
		}
		msgBytes, err := respMsg.MarshalVT()
		if err != nil {
			n.logger.Errorf("failed to marshal epoch response: %v", err)
			return
		}

		if err = stream.AsyncSend(msgBytes); err != nil {
			n.logger.Errorf("failed to send epoch response: %v", err)
			return
		}

	case pb.Message_EPOCH_REQUEST_RESPONSE:
		var resp protos.EpochChangeResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			n.logger.Errorf("failed to unmarshal epoch response: %v", err)
			return
		}
		n.postLocalEvent(&localEvent{finishSyncEpoch, resp})
	}
}

func checkEpochRequest(request protos.EpochChangeRequest) error {
	if request.GetStartEpoch() >= request.GetTargetEpoch() {
		return fmt.Errorf("reject epoch change request for illegal change from %d to %d", request.GetStartEpoch(), request.GetTargetEpoch())
	}
	return nil
}

func (n *Node) pagingGetEpochChangeProof(startEpoch, endEpoch types.Epoch, pageLimit uint64) (*types.EpochChangeProof, error) {
	pagingEpoch := endEpoch
	more := types.Epoch(0)

	if pagingEpoch-startEpoch > pageLimit {
		more = pagingEpoch
		pagingEpoch = startEpoch + pageLimit
	}

	checkpoints := make([]*types.QuorumCheckpoint, 0)
	for epoch := startEpoch; epoch < pagingEpoch; epoch++ {
		raw, err := n.epochService.ReadEpochState(epoch)
		if err != nil {
			n.logger.Errorf("failed to read epoch %d quorum chkpt: %v", epoch, err)
			return nil, err
		}
		cp, ok := raw.(*consensustypes.DagbftQuorumCheckpoint)
		if !ok {
			n.logger.Errorf("failed to read epoch %d quorum chkpt: type assertion failed: %v", epoch, reflect.TypeOf(raw))
			return nil, err
		}
		if cp == nil {
			return nil, fmt.Errorf("cannot find epoch change for epoch %d", epoch)
		}
		checkpoints = append(checkpoints, cp.QuorumCheckpoint)
	}

	return types.NewEpochChangeProof(more, checkpoints...), nil
}

func (n *Node) recvFetchStateResponse(data []byte) error {
	proof := &consensustypes.DagbftQuorumCheckpoint{}
	if err := proof.Unmarshal(data); err != nil {
		return err
	}
	if proof.Height() == repo.GenesisBlockNumber {
		n.logger.Infof("start at genesis block, need not sync")
		n.statusMgr.On(Normal)
		return nil
	}
	if err := n.verifier.verifyProof(proof); err != nil {
		return err
	}

	if err := n.pushProof(data); err != nil {
		return err
	}

	if proof.Epoch() > n.chainState.EpochInfo.Epoch {
		channel.SafeSend(n.recvEventCh, &localEvent{syncEpoch, proof.Epoch()}, n.closeCh)
		return nil
	} else {
		channel.SafeSend(n.recvEventCh, &localEvent{stateUpdate, &stateUpdateEvent{targetCheckpoint: proof.QuorumCheckpoint}}, n.closeCh)
	}

	return nil
}
