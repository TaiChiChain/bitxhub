package data_syncer

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/components/status"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	dagbft_common "github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/common"
	"github.com/bcds/go-hpc-dagbft/common/errors"
	"github.com/bcds/go-hpc-dagbft/common/types/events"
	"github.com/bcds/go-hpc-dagbft/common/utils"
	"github.com/bcds/go-hpc-dagbft/common/utils/concurrency"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	eth_event "github.com/ethereum/go-ethereum/event"
	"github.com/samber/lo"

	common_metrics "github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	"github.com/axiomesh/axiom-ledger/internal/consensus/dagbft/adaptor"
	"github.com/axiomesh/axiom-ledger/internal/consensus/epochmgr"
	consensus_types "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/internal/network"
	synccomm "github.com/axiomesh/axiom-ledger/internal/sync/common"
	dagtypes "github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/sirupsen/logrus"
)

type Node struct {
	client               *client
	network              network.Network
	bindNode             bindNode
	logger               logrus.FieldLogger
	chainState           *chainstate.ChainState
	getBlockFn           func(height uint64) (*types.Block, error)
	executedEventCh      chan *events.ExecutedEvent
	stateUpdatedEventCh  chan *events.StateUpdatedEvent
	recvNewTxSetCh       chan []protocol.Transaction
	recvNewAttestationCh chan *consensus_types.Attestation
	currentHeight        uint64
	verifier             *verifier
	recvEventCh          chan *localEvent
	wg                   *sync.WaitGroup
	ap                   *concurrency.AsyncPool
	statusMgr            *status.StatusMgr
	sync                 synccomm.Sync
	epochService         *epochmgr.EpochManager
	sendToExecuteCh      chan<- *consensus_types.CommitEvent

	notifyStateUpdateCh chan<- containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]
	waitEpochChangeCh   chan struct{}
	sendReadyC          chan<- *adaptor.Ready

	closeCh chan struct{}

	metrics *metrics

	recvProofCache  *ProofCache
	AttestationFeed *eth_event.Feed
	useBls          bool
}

func NewNode(config *common.Config, verifier protocol.ValidatorVerifier, logger logrus.FieldLogger,
	sendReadyCh chan<- *adaptor.Ready, sendCommitEvent chan<- *consensus_types.CommitEvent,
	sendNotifyStateUpdateCh chan<- containers.Tuple[dagtypes.Height, *dagtypes.QuorumCheckpoint, chan<- *events.StateUpdatedEvent]) (*Node, error) {

	dataSyncerConf := config.Repo.ConsensusConfig.Dagbft.DataSyncerConfigs
	ap, err := newAsyncPool(dataSyncerConf.GoroutineLimit, logger)
	if err != nil {
		return nil, err
	}

	bindN := bindNode{
		wsAddr: fmt.Sprintf("%s:%d", dataSyncerConf.BindNode.Host, dataSyncerConf.BindNode.Port),
		p2pID:  dataSyncerConf.BindNode.Pid,
		nodeId: uint64(dataSyncerConf.BindNode.Id),
	}
	closeCh := make(chan struct{})

	return &Node{
		network:              config.Network,
		bindNode:             bindN,
		chainState:           config.ChainState,
		getBlockFn:           config.GetBlockFunc,
		verifier:             newVerifier(logger, verifier, closeCh),
		statusMgr:            status.NewStatusMgr(),
		sync:                 config.BlockSync,
		epochService:         config.EpochStore,
		recvProofCache:       newProofCache(dataSyncerConf.MaxCacheSize),
		sendToExecuteCh:      sendCommitEvent,
		notifyStateUpdateCh:  sendNotifyStateUpdateCh,
		sendReadyC:           sendReadyCh,
		closeCh:              closeCh,
		waitEpochChangeCh:    make(chan struct{}, 1),
		executedEventCh:      make(chan *events.ExecutedEvent, common.MaxChainSize),
		stateUpdatedEventCh:  make(chan *events.StateUpdatedEvent, 1),
		recvNewTxSetCh:       make(chan []protocol.Transaction, common.MaxChainSize),
		recvNewAttestationCh: make(chan *consensus_types.Attestation, common.MaxChainSize),
		recvEventCh:          make(chan *localEvent, common.MaxChainSize),
		ap:                   ap,
		wg:                   &sync.WaitGroup{},
		logger:               logger,
		metrics:              newMetrics(),
		AttestationFeed:      new(eth_event.Feed),
		useBls:               config.Repo.Config.Consensus.UseBlsKey,
	}, nil
}

func (n *Node) Start() error {
	if err := n.network.RegisterMultiMsgHandler([]pb.Message_Type{
		pb.Message_EPOCH_REQUEST, pb.Message_EPOCH_REQUEST_RESPONSE}, n.handleEpochMsg); err != nil {
		return err
	}

	if err := n.network.RegisterMultiMsgHandler([]pb.Message_Type{
		pb.Message_TRANSACTION_PROPAGATION}, n.handleTxsMsg); err != nil {
		return err
	}

	cli, err := newClient(n.bindNode.wsAddr, n.closeCh, n.recvNewAttestationCh)
	if err != nil {
		return err
	}
	n.client = cli

	if err = n.client.start(n.wg, n.ap); err != nil {
		return err
	}

	if err = n.initState(); err != nil {
		return err
	}

	if err = utils.AssertNoError(
		n.ap.AsyncDo(func() { n.listenExecutedEvent() }, n.wg),     // listen executed event
		n.ap.AsyncDo(func() { n.listenStateUpdatedEvent() }, n.wg), // listen state updated event
		n.ap.AsyncDo(func() { n.listenNewTxSet() }, n.wg),          // listen new txs event
		n.ap.AsyncDo(func() { n.listenNewAttestation() }, n.wg),    // listen new attestation from remote bind node
		n.ap.AsyncDo(func() { n.listenEvent() }, n.wg),             // listen local event
	); err != nil {
		n.ap.Release()
		return fmt.Errorf("%w: failed to create AsyncTask for Data syncer, %s", errors.ErrGoroutine, err)
	}

	return nil
}

func (n *Node) IsNormal() bool {
	return n.statusMgr.In(Normal)
}

func (n *Node) initState() error {
	request := &pb.Message{
		From: n.chainState.SelfNodeInfo.P2PID,
		Type: pb.Message_FETCH_STATE,
	}

	if err := retry.Retry(func(attempt uint) error {
		resp, err := n.network.Send(n.bindNode.p2pID, request)
		if err != nil {
			return err
		}
		if resp.Type != pb.Message_FETCH_STATE_RESPONSE {
			return fmt.Errorf("invalid response type: %s", resp.Type)
		}
		if resp.From != n.bindNode.p2pID {
			return fmt.Errorf("invalid response from: %s", resp.From)
		}

		return n.recvFetchStateResponse(resp.Data)
	}, strategy.Limit(5), strategy.Wait(1*time.Second)); err != nil {
		panic(fmt.Errorf("retry fetch state failed: %v", err))
	}

	return nil
}

func (n *Node) listenEvent() {
	defer n.wg.Done()
	for {
		select {
		case <-n.closeCh:
			n.logger.Info("data syncer stopped")
			return
		case next := <-n.recvEventCh:
			nexts := make([]*localEvent, 0)
			for {
				select {
				case <-n.closeCh:
					n.logger.Info("data syncer stopped")
					return
				default:
				}
				evs := n.processEvent(next)
				nexts = append(nexts, evs...)
				if len(nexts) == 0 {
					break
				}
				next = nexts[0]
				nexts = nexts[1:]
			}
		}
	}
}

func (n *Node) processEvent(ev *localEvent) []*localEvent {
	n.logger.WithField("event type", EventTypeM[ev.EventType]).Infof("process event")
	start := time.Now()
	defer n.logger.WithField("event type", EventTypeM[ev.EventType]).Infof("process event cost %s", time.Since(start))
	var (
		nextEvent []*localEvent
		err       error
	)
	switch ev.EventType {
	case newAttestation:
		at, ok := ev.Event.(*consensus_types.Attestation)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}
		nextEvent, err = n.handleNewAttestation(at)
		if err != nil {
			n.logger.WithField("event", ev).Errorf("handleNewAttestation err: %v", err)
			return nil
		}
	case newTxSet:
		txSet, ok := ev.Event.([]protocol.Transaction)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}
		if err = n.handleAsyncSendNewTxSet(txSet); err != nil {
			n.logger.WithField("event", ev).Errorf("handleAsyncSendNewTxSet err: %v", err)
			return nil
		}

	case executed:
		ex, ok := ev.Event.(*events.ExecutedEvent)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}
		if err = n.handleExecutedEvent(ex); err != nil {
			n.logger.WithField("event", ev).Errorf("handleExecutedEvent err: %v", err)
			return nil
		}

	case syncEpoch:
		startEpoch := n.chainState.EpochInfo.Epoch
		endEpoch, ok := ev.Event.(uint64)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}

		if err = n.handleEpochRequest(startEpoch, endEpoch); err != nil {
			n.logger.WithField("event", ev).Errorf("handleEpochRequest err: %v", err)
			return nil
		}
		n.statusMgr.On(InEpochSync)

	case finishSyncEpoch:
		epochChanges, ok := ev.Event.(protos.EpochChangeResponse)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}

		target := n.peakProof()

		var targetCheckpoint *dagtypes.QuorumCheckpoint

		if epochChanges.Proof.More != 0 {
			lastEpochProof := epochChanges.Proof.Checkpoints[len(epochChanges.Proof.Checkpoints)-1]
			targetCheckpoint = &dagtypes.QuorumCheckpoint{QuorumCheckpoint: *lastEpochProof}
		} else {
			// finish sync epoch, turn off InEpochSync status
			targetCheckpoint = target.QuorumCheckpoint
			n.statusMgr.Off(InEpochSync)
		}

		stateEvent := &stateUpdateEvent{
			targetCheckpoint: targetCheckpoint,
			EpochChanges: lo.Map(epochChanges.Proof.Checkpoints, func(cp *protos.QuorumCheckpoint, _ int) *dagtypes.QuorumCheckpoint {
				return &dagtypes.QuorumCheckpoint{QuorumCheckpoint: *cp}
			}),
		}

		nextEvent = append(nextEvent, &localEvent{
			EventType: stateUpdate,
			Event:     stateEvent,
		})

	case stateUpdate:
		st, ok := ev.Event.(*stateUpdateEvent)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}

		if n.statusMgr.In(InSync) {
			n.logger.Warningf("node is already in sync, just ignore it...")
			return nil
		}

		if len(st.EpochChanges) > 0 {
			n.logger.Infof("store epoch state at epoch %d", st.targetCheckpoint.Epoch())
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(waitG *sync.WaitGroup) {
				defer waitG.Done()
				defer n.wg.Done()
				lo.ForEach(st.EpochChanges, func(ckpt *dagtypes.QuorumCheckpoint, index int) {
					if ckpt != nil {
						epochChange := &consensus_types.DagbftQuorumCheckpoint{QuorumCheckpoint: ckpt}
						if err = n.epochService.StoreEpochState(epochChange); err != nil {
							n.logger.Errorf("failed to store epoch state at epoch %d: %s", ckpt.Epoch(), err)
						}
					}
				})
			}(wg)
		}

		targetHeight := st.targetCheckpoint.Height()
		targetCheckpoint := st.targetCheckpoint
		n.logger.Infof("node start state updating to target height: %d", targetHeight)
		n.notifyConsensusStateUpdate(targetHeight, targetCheckpoint, n.stateUpdatedEventCh)

		if err = n.handleStateUpdate(st.targetCheckpoint, st.EpochChanges...); err != nil {
			n.logger.WithField("event", ev).Errorf("handleStateUpdate err: %v", err)
			return nil
		}
		n.statusMgr.On(InSync)

	case finishStateUpdate:
		se, ok := ev.Event.(*events.StateUpdatedEvent)
		if !ok {
			n.logger.WithFields(logrus.Fields{
				"event": ev,
				"type":  EventTypeM[ev.EventType],
			}).Errorf("invalid event type")
			n.metrics.errCounter.WithLabelValues(ErrorType).Inc()
			return nil
		}
		n.logger.Infof("node finish state updating to target height: %d", se.Checkpoint.Height())
		p := n.popProof(se.Checkpoint.Height())
		if p == nil {
			panic(fmt.Errorf("proof not found at height: %d", se.Checkpoint.Height()))
		}
		if p.StateRoot().String() != se.Checkpoint.StateRoot().String() {
			panic(fmt.Sprintf("state root not match, expect: %s, got: %s", se.Checkpoint.StateRoot().String(), p.StateRoot().String()))
		}

		// update latest proof in chainstate and db
		if err = n.updateLatestProof(se.Checkpoint); err != nil {
			return nil
		}
		// update attestation
		if err = dagbft_common.PostAttestationEvent(se.Checkpoint, n.AttestationFeed, n.getBlockFn, n.epochService); err != nil {
			n.logger.WithField("event", ev).Errorf("PostAttestationEvent err: %v", err)
			return nil
		}

		for _, at := range n.popProofList() {
			nextEvent = append(nextEvent, &localEvent{
				EventType: newAttestation,
				Event:     at,
			})
		}
		if len(nextEvent) > 0 {
			n.logger.Infof("node continue handle new blocks from height: %d, newAttestation event count: %d",
				se.Checkpoint.Height(), len(nextEvent))
		}
		n.updateCurrentHeight(se.Checkpoint.Height())
		n.statusMgr.Off(InSync)

		if !n.statusMgr.In(Normal) {
			n.statusMgr.On(Normal)
		}
	}

	return nextEvent
}

func (n *Node) notifyConsensusStateUpdate(targetHeight dagtypes.Height, targetCheckpoint *dagtypes.QuorumCheckpoint, evCh chan<- *events.StateUpdatedEvent) {
	stateResult := containers.Pack3(targetHeight, targetCheckpoint, evCh)
	channel.SafeSend(n.notifyStateUpdateCh, stateResult, n.closeCh)
}

func (n *Node) postLocalEvent(ev *localEvent) {
	channel.SafeSend(n.recvEventCh, ev, n.closeCh)
}

func (n *Node) listenStateUpdatedEvent() {
	for {
		select {
		case <-n.closeCh:
			return
		case ev := <-n.stateUpdatedEventCh:
			n.postLocalEvent(&localEvent{EventType: finishStateUpdate, Event: ev})
		}
	}
}

func (n *Node) listenNewAttestation() {
	for {
		select {
		case <-n.closeCh:
			return
		case ev := <-n.recvNewAttestationCh:
			n.postLocalEvent(&localEvent{EventType: newAttestation, Event: ev})
		}
	}
}

func (n *Node) listenNewTxSet() {
	for {
		select {
		case <-n.closeCh:
			return
		case ev := <-n.recvNewTxSetCh:
			n.postLocalEvent(&localEvent{EventType: newTxSet, Event: ev})
		}
	}
}

func (n *Node) listenExecutedEvent() {
	for {
		select {
		case <-n.closeCh:
			return
		case ev := <-n.executedEventCh:
			n.postLocalEvent(&localEvent{EventType: executed, Event: ev})
		}
	}
}

func (n *Node) SendNewTxSet(txSet []protocol.Transaction) {
	channel.SafeSend(n.recvNewTxSetCh, txSet, n.closeCh)
}

func (n *Node) handleExecutedEvent(ev *events.ExecutedEvent) error {
	height := ev.ExecutedState.ExecuteState.Height
	ct := n.popProof(height)
	if ct == nil {
		return fmt.Errorf("failed to pop attestation at height: %d", height)
	}

	if err := n.updateLatestProof(ct.QuorumCheckpoint); err != nil {
		return err
	}
	return dagbft_common.PostAttestationEvent(ct.QuorumCheckpoint, n.AttestationFeed, n.getBlockFn, n.epochService)
}

func (n *Node) updateLatestProof(checkpoint *dagtypes.QuorumCheckpoint) error {
	proof := &consensus_types.DagbftQuorumCheckpoint{QuorumCheckpoint: checkpoint}
	proofData, err := proof.Marshal()
	if err != nil {
		n.logger.Errorf("failed to marshal proof data: %v", err)
		return err
	}

	validators, err := dagbft_common.GetValidators(n.chainState, n.useBls)
	if err != nil {
		n.logger.Errorf("failed to get validators: %v", err)
		return err
	}

	n.chainState.UpdateCheckpoint(&pb.QuorumCheckpoint{
		Epoch: checkpoint.Epoch(),
		State: &pb.ExecuteState{
			Height: checkpoint.Checkpoint().ExecuteState().StateHeight(),
			Digest: checkpoint.Checkpoint().ExecuteState().StateRoot().String(),
		},
		Proof: proofData,
		ValidatorSet: lo.Map(validators, func(v *protos.Validator, _ int) *pb.QuorumCheckpoint_Validator {
			p2pId, err := n.chainState.GetP2PIDByNodeID(uint64(v.ValidatorId))
			if err != nil {
				n.logger.Errorf("failed to get p2p id: %v", err)
				return nil
			}
			return &pb.QuorumCheckpoint_Validator{
				Id:      uint64(v.ValidatorId),
				P2PId:   p2pId,
				Primary: v.Hostname,
				Workers: v.Workers,
			}
		}),
	})

	return n.epochService.StoreLatestProof(proofData)
}

func (n *Node) handleEpochRequest(start, end uint64) error {
	epochRequest := protos.EpochChangeRequest{
		StartEpoch:  start,
		TargetEpoch: end,
	}

	data, err := epochRequest.Marshal()
	if err != nil {
		return err
	}
	msg := &pb.Message{
		Type: pb.Message_EPOCH_REQUEST,
		Data: data,
	}

	if err = n.network.AsyncSend(n.bindNode.p2pID, msg); err != nil {
		return err
	}
	return nil

}

func (n *Node) handleStateUpdate(target *dagtypes.QuorumCheckpoint, epochChanges ...*dagtypes.QuorumCheckpoint) error {
	targetHeight := target.Height()
	localHeight := n.chainState.GetCurrentCheckpointState().Height
	latestBlockHash := n.chainState.GetCurrentCheckpointState().Digest
	peerM := dagbft_common.GetRemotePeers(epochChanges, n.chainState)

	syncTaskDoneCh := make(chan error, 1)
	if err := retry.Retry(func(attempt uint) error {
		params := &synccomm.SyncParams{
			Peers:           peerM,
			LatestBlockHash: latestBlockHash,
			// ensure sync remote count including at least one correct node
			Quorum:       common.CalFaulty(uint64(len(peerM) + 1)),
			CurHeight:    localHeight + 1,
			TargetHeight: targetHeight,
			QuorumCheckpoint: &consensus_types.DagbftQuorumCheckpoint{
				QuorumCheckpoint: target,
			},
			EpochChanges: lo.Map(epochChanges, func(qckpt *dagtypes.QuorumCheckpoint, index int) *consensus_types.EpochChange {
				return &consensus_types.EpochChange{
					QuorumCheckpoint: &consensus_types.DagbftQuorumCheckpoint{QuorumCheckpoint: qckpt},
				}
			}),
		}
		n.logger.WithFields(logrus.Fields{
			"target":       params.TargetHeight,
			"target_hash":  params.QuorumCheckpoint.GetStateDigest(),
			"start":        params.CurHeight,
			"epochChanges": epochChanges,
		}).Info("State update start")
		err := n.sync.StartSync(params, syncTaskDoneCh)
		if err != nil {
			n.logger.Errorf("start sync failed[local:%d, target:%d]: %s", localHeight, targetHeight, err)
			return err
		}
		return nil
	}, strategy.Limit(5), strategy.Wait(500*time.Microsecond)); err != nil {
		panic(fmt.Errorf("retry start sync failed: %v", err))
	}

	var stateUpdatedCheckpoint *consensus_types.Checkpoint
	// wait for the sync to finish
	for {
		select {
		case <-n.closeCh:
			n.logger.Info("state update is canceled!!!!!!")
			return nil
		case syncErr := <-syncTaskDoneCh:
			if syncErr != nil {
				return syncErr
			}
		case data := <-n.sync.Commit():
			endSync := false
			blockCache, ok := data.([]synccomm.CommitData)
			if !ok {
				panic("state update failed: invalid commit data")
			}

			n.logger.Infof("fetch chunk: start: %d, end: %d", blockCache[0].GetHeight(), blockCache[len(blockCache)-1].GetHeight())

			for _, commitData := range blockCache {
				// if the block is the target block, we should resign the stateUpdatedCheckpoint in CommitEvent
				// and send the quitSync signal to sync module
				if commitData.GetHeight() == targetHeight {
					stateUpdatedCheckpoint = &consensus_types.Checkpoint{
						Epoch:  target.Epoch(),
						Height: target.Height(),
						Digest: target.Checkpoint().ExecuteState().StateRoot().String(),
					}
					endSync = true
				}
				block, ok := commitData.(*synccomm.BlockData)
				if !ok {
					panic("state update failed: invalid commit data")
				}
				commitEvent := &consensus_types.CommitEvent{
					Block:                  block.Block,
					StateUpdatedCheckpoint: stateUpdatedCheckpoint,
				}
				channel.SafeSend(n.sendToExecuteCh, commitEvent, n.closeCh)

				if endSync {
					n.logger.Infof("State update finished, target height: %d", targetHeight)
					return nil
				}
			}
		}
	}

}

func (n *Node) handleAsyncSendNewTxSet(txSet []protocol.Transaction) error {
	raw := &pb.BytesSlice{
		Slice: convertTransactions(txSet),
	}
	data, err := raw.MarshalVT()
	if err != nil {
		return err
	}

	p2pMsg := &pb.Message{
		From: n.chainState.SelfNodeInfo.P2PID,
		Type: pb.Message_TRANSACTION_PROPAGATION,
		Data: data,
	}

	if err = n.network.AsyncSend(n.bindNode.p2pID, p2pMsg); err != nil {
		return err
	}
	n.metrics.txCounter.Add(float64(len(txSet)))
	return nil
}

// convertTransactions converts a slice of protocol.Transaction to a slice of byte slices.
func convertTransactions(transactions []protocol.Transaction) [][]byte {
	var result [][]byte
	for _, tx := range transactions {
		result = append(result, tx)
	}
	return result
}

func (n *Node) handleNewAttestation(a *consensus_types.Attestation) ([]*localEvent, error) {
	proof, err := a.GetProof()
	if err != nil {
		n.logger.Errorf("failed to decode quorum checkpoint: %v", err)
		return nil, err
	}
	quorumCheckpoint, ok := proof.(*consensus_types.DagbftQuorumCheckpoint)
	if !ok {
		return nil, fmt.Errorf("failed to cast quorum checkpoint")
	}
	newBlock, err := a.GetBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %v", err)
	}
	if newBlock.Height() < n.currentHeight {
		n.logger.Warningf("receive new block height %d is less than current height %d, just ignore it...",
			quorumCheckpoint.GetHeight(), n.currentHeight)
		return nil, nil
	}

	if a.GetEpoch() > n.chainState.EpochInfo.Epoch {
		n.logger.Infof("receive new block epoch %d is greater than current epoch %d, "+
			"should wait for finishing epoch change", a.GetEpoch(), n.chainState.EpochInfo.Epoch)

		// if cache is full, should dismiss the new attestation until cache is empty
		if err = n.pushProof(a.Proof); err != nil {
			n.logger.WithFields(logrus.Fields{"height": quorumCheckpoint.GetHeight()}).Warningf("push attestation failed: %v", err)
			return nil, nil
		}

		if newBlock.Height() > n.currentHeight && !n.statusMgr.In(InEpochSync) {
			nextEvent := &localEvent{
				EventType: syncEpoch,
				Event:     a.GetEpoch(),
			}
			return []*localEvent{nextEvent}, nil
		}

		return nil, nil
	}

	// start verify the attestation
	err = n.verifier.verifyAttestation(a)
	if err != nil {
		n.logger.WithFields(logrus.Fields{"height": quorumCheckpoint.GetHeight()}).Errorf("verify attestation failed: %v", err)
		return nil, err
	}

	if err = n.pushProof(a.Proof); err != nil {
		return nil, err
	}

	// This situation often occurs when the bound validator had performed synchronization,
	// validator would lose some Attestation data during the synchronization,
	// therefore, the data syncer node should also follow a synchronization process.
	demandHeight := n.currentHeight + 1
	if newBlock.Height() > demandHeight && !n.statusMgr.In(InSync) {
		nextEvent := &localEvent{
			EventType: stateUpdate,
			Event:     &stateUpdateEvent{targetCheckpoint: quorumCheckpoint.QuorumCheckpoint},
		}
		return []*localEvent{nextEvent}, nil
	}

	if err = n.execute(newBlock, quorumCheckpoint, n.executedEventCh); err != nil {
		return nil, err
	}

	if quorumCheckpoint.EndEpoch() {
		if err = n.epochService.StoreEpochState(quorumCheckpoint); err != nil {
			return nil, err
		}
		n.metrics.epochChangeNum.Inc()
	}

	n.updateCurrentHeight(newBlock.Height())
	return nil, nil
}

func (n *Node) updateCurrentHeight(height uint64) {
	n.currentHeight = height
	n.logger.Debugf("update current height to %d", height)
}

func (n *Node) execute(block *types.Block, checkpoint types.QuorumCheckpoint, eventCh chan<- *events.ExecutedEvent) error {
	ckp, ok := checkpoint.(*consensus_types.DagbftQuorumCheckpoint)
	if !ok {
		return fmt.Errorf("failed to convert checkpoint to *consensus_types.DagbftQuorumCheckpoint: %v", checkpoint)
	}
	n.logger.Infof("Execute Consensus Output %d at height %d, txs: %d",
		ckp.CommitSequence(), block.Height(), len(block.Transactions))

	isConfigured := common.NeedChangeEpoch(block.Height(), n.chainState.GetCurrentEpochInfo())

	n.asyncSendNewBlock(block, ckp, eventCh, isConfigured)

	if isConfigured {
		start := time.Now()
		select {
		case <-n.waitEpochChangeCh:
			common_metrics.WaitEpochTime.WithLabelValues(consensus_types.Dagbft).Observe(time.Since(start).Seconds())
			n.logger.WithFields(logrus.Fields{
				"duration": time.Since(start),
			}).Info("wait epoch change done")
			return nil
		case <-n.closeCh:
			return nil
		//todo: configurable timeout
		case <-time.After(60 * time.Second):
			return fmt.Errorf("wait epoch change timeout")
		}
	}

	return nil
}

func (n *Node) asyncSendNewBlock(block *types.Block, ckp *consensus_types.DagbftQuorumCheckpoint, executedCh chan<- *events.ExecutedEvent, isConfigured bool) {

	ready := &adaptor.Ready{
		Txs:                    block.Transactions,
		Height:                 block.Height(),
		RecvConsensusTimestamp: time.Now().UnixNano(),
		Timestamp:              block.Header.Timestamp,
		ProposerNodeID:         block.Header.ProposerNodeID,
		ExecutedCh:             executedCh,
		CommitState:            &dagtypes.CommitState{CommitState: *ckp.QuorumCheckpoint.GetCheckpoint().GetCommitState()},
	}
	if isConfigured {
		ready.EpochChangedCh = n.waitEpochChangeCh
	}
	channel.SafeSend(n.sendReadyC, ready, n.closeCh)
}

func (n *Node) Stop() {
	close(n.closeCh)
	n.wg.Wait()
	n.logger.Infof("data syncer stopped")
}

func (ac *ProofCache) checkStatus() {
	if len(ac.quorumCheckpointM) >= ac.maxSize && !ac.status.In(cacheFull) {
		ac.status.On(cacheFull)
		ac.status.On(cacheAbnormal)
	}

	if len(ac.quorumCheckpointM) < ac.maxSize && ac.status.In(cacheFull) {
		ac.status.Off(cacheFull)
	}

	if len(ac.quorumCheckpointM) == 0 && ac.status.In(cacheAbnormal) {
		ac.status.Off(cacheAbnormal)
	}
}

func (ac *ProofCache) isNormal() bool {
	return !ac.status.InOne(cacheFull, cacheAbnormal)
}

func (n *Node) pushProof(p []byte) error {
	quorumCkpt, err := dagbft_common.DecodeProof(p)
	if err != nil {
		return err
	}
	n.logger.Debugf("push proof %d", quorumCkpt.Height())
	return n.recvProofCache.pushProof(quorumCkpt, n.metrics)
}

func (n *Node) peakProof() *consensus_types.DagbftQuorumCheckpoint {
	h := n.recvProofCache.heightIndex.PeekItem()
	if h == math.MaxUint64 {
		return nil
	}
	return n.recvProofCache.quorumCheckpointM[h]
}

func (n *Node) popProofList() []*consensus_types.DagbftQuorumCheckpoint {
	defer n.recvProofCache.checkStatus()
	matched := make([]*consensus_types.DagbftQuorumCheckpoint, 0)
	for h := n.recvProofCache.heightIndex.PeekItem(); h != math.MaxUint64 && h == n.currentHeight; n.recvProofCache.heightIndex.PopItem() {
		qt, ok := n.recvProofCache.quorumCheckpointM[h]
		if !ok {
			n.logger.Errorf("block %d not found in cache, but exists in heightIndex", h)
			return nil
		}
		if qt.Epoch() != n.chainState.EpochInfo.Epoch {
			n.logger.Infof("block %d epoch %d not match current epoch %d, need wait for next epoch",
				h, qt.Epoch(), n.chainState.EpochInfo.Epoch)
			return matched
		}
		matched = append(matched, n.recvProofCache.quorumCheckpointM[h])
		n.logger.Debugf("pop proof %d", h)
		delete(n.recvProofCache.quorumCheckpointM, h)
		n.metrics.attestationCacheNum.Desc()
	}
	return matched
}

func (n *Node) popProof(height uint64) *consensus_types.DagbftQuorumCheckpoint {
	defer n.recvProofCache.checkStatus()
	if n.peakProof() != nil && n.peakProof().Height() == height {
		matched := n.recvProofCache.quorumCheckpointM[height]
		delete(n.recvProofCache.quorumCheckpointM, height)
		n.metrics.attestationCacheNum.Desc()
		n.logger.Debugf("pop proof %d", height)
		return matched
	}
	return nil
}
