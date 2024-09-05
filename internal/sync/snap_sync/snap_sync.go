package snap_sync

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-kit/types"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/gammazero/workerpool"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
)

type SnapSync struct {
	lock             sync.Mutex
	epochStateCache  *epochStateCache
	epochStateRecvCh chan containers.Pair[uint64, types.QuorumCheckpoint]
	currentEpoch     uint64
	recvCommitDataCh chan []common.CommitData
	commitCh         chan any
	network          network.Network
	consensusType    string
	logger           logrus.FieldLogger

	ctx    context.Context
	cancel context.CancelFunc

	cnf *common.Config
}

func (s *SnapSync) Stop() {
	s.cancel()
}

func (s *SnapSync) Start() {
	go s.listenCommitData()
}

func (s *SnapSync) Mode() common.SyncMode {
	return common.SyncModeSnapshot
}

func (s *SnapSync) listenCommitData() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case data := <-s.recvCommitDataCh:
			last, err := lo.Last(data)
			if err != nil {
				s.logger.Errorf("failed to get last commit data, err: %v", err)
				continue
			}

			snapData := &common.SnapCommitData{
				Data:         data,
				TargetHeight: last.GetHeight(),
				TargetHash:   last.GetHash(),
			}
			s.commitCh <- snapData
		}
	}
}

func (s *SnapSync) PostCommitData(data []common.CommitData) {
	s.recvCommitDataCh <- data
}

func (s *SnapSync) Commit() chan any {
	return s.commitCh
}

func NewSnapSync(logger logrus.FieldLogger, net network.Network, consensusTyp string) common.ISyncConstructor {
	ctx, cancel := context.WithCancel(context.Background())
	return &SnapSync{
		commitCh:         make(chan any, 1),
		recvCommitDataCh: make(chan []common.CommitData, repo.ChannelSize),
		network:          net,
		logger:           logger,
		consensusType:    consensusTyp,
		epochStateRecvCh: make(chan containers.Pair[uint64, types.QuorumCheckpoint], repo.ChannelSize),
		epochStateCache:  newEpochStateCache(),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (s *SnapSync) Prepare(config *common.Config) error {
	s.cnf = config
	s.fetchEpochState()
	return nil
}

func (s *SnapSync) fetchEpochState() {
	latestPersistEpoch := s.cnf.LatestPersistEpoch
	snapPersistedEpoch := s.cnf.SnapPersistedEpoch

	// fetch epoch state
	// latestPersistEpoch means the local's latest persisted epoch,
	// snapCurrentEpoch means the current epoch in snapshot,
	// the latestPersistEpoch is typically smaller than snapCurrentEpoch, because our ledger behind the height of the snap state,
	// and we need to synchronize to the block height of the snap.
	start := latestPersistEpoch + 1
	end := snapPersistedEpoch
	if start > end {
		s.logger.WithFields(logrus.Fields{
			"latest": start - 1,
			"end":    end,
		}).Infof("local latest persist epoch %d is same as end epoch %d", start-1, end)
		channel.TrySend(s.cnf.EpochChangeSendCh, containers.Pack3[types.QuorumCheckpoint, error, bool](nil, nil, true))
		return
	}

	errCh := s.fetchEpochQuorumCheckpoint(start, end)
	go func(startTime time.Time) {
		for {
			select {
			case <-s.ctx.Done():
			case err := <-errCh:
				if err != nil {
					s.logger.WithError(err).Errorf("fetch epoch state failed")
					channel.TrySend(s.cnf.EpochChangeSendCh, containers.Pack3[types.QuorumCheckpoint, error, bool](nil, err, true))
					return
				}
			case state := <-s.epochStateRecvCh:
				epoch, epochState := state.Unpack()
				taskDone := func() bool {
					return s.currentEpoch >= snapPersistedEpoch
				}
				if s.currentEpoch == epoch {
					channel.TrySend(s.cnf.EpochChangeSendCh, containers.Pack3[types.QuorumCheckpoint, error, bool](epochState, nil, taskDone()))
					s.currentEpoch++
					for s.epochStateCache.getTopEpoch() == s.currentEpoch {
						epoch, epochState = s.epochStateCache.popQuorumCheckpoint().Unpack()
						channel.TrySend(s.cnf.EpochChangeSendCh, containers.Pack3[types.QuorumCheckpoint, error, bool](epochState, nil, taskDone()))
						s.currentEpoch++
					}

					if taskDone() {
						s.logger.WithFields(logrus.Fields{
							"cost": time.Since(startTime),
							"end":  snapPersistedEpoch,
						}).Infof("fetch all epoch state success")
						return
					}
				} else {
					s.epochStateCache.pushQuorumCheckpoint(epoch, epochState)
				}
				return
			}
		}
	}(time.Now())
}

func (s *SnapSync) fetchEpochQuorumCheckpoint(start, end uint64) <-chan error {
	taskNum := int(end - start + 1)

	s.currentEpoch = start
	wp := workerpool.New(concurrencyLimit)

	errCh := make(chan error, taskNum)
	for i := start; i <= end; i++ {
		wp.Submit(func() {
			epoch := i
			var (
				err  error
				resp *pb.Message
			)
			defer func() {
				errCh <- err
			}()
			peer := s.pickPeer(int(epoch))
			req := &pb.FetchEpochStateRequest{
				Epoch: epoch,
			}
			data, err := req.MarshalVT()
			if err != nil {
				s.logger.WithFields(logrus.Fields{
					"peer": peer,
					"err":  err,
				}).Error("Marshal fetch epoch state request failed")
				return
			}

			invalidPeers := make([]*common.Node, 0)
			if err = retry.Retry(func(attempt uint) error {
				s.logger.WithFields(logrus.Fields{
					"peer":  peer.Id,
					"epoch": epoch,
				}).Info("Send fetch epoch state request")
				resp, err = s.network.Send(peer.PeerID, &pb.Message{
					Type: pb.Message_FETCH_EPOCH_STATE_REQUEST,
					Data: data,
				})
				if err != nil {
					s.logger.WithFields(logrus.Fields{
						"peer": peer.Id,
						"err":  err,
					}).Error("Send fetch epoch state request failed")

					invalidPeers = append(invalidPeers, peer)
					var pickErr error
					peer, pickErr = s.pickRandomPeer(invalidPeers...)
					// reset invalidPeers
					if pickErr != nil {
						invalidPeers = make([]*common.Node, 0)
						peer, _ = s.pickRandomPeer()
					}
					return err
				}

				s.logger.WithFields(logrus.Fields{
					"peer":  peer.Id,
					"epoch": epoch,
				}).Info("Receive fetch epoch state response")
				return nil
			}, strategy.Limit(uint(len(s.cnf.Peers))), strategy.Wait(200*time.Millisecond)); err != nil {
				err = fmt.Errorf("retry send fetch epoch state request failed, all peers invalid: %v", s.cnf.Peers)
				s.logger.WithFields(logrus.Fields{
					"peer":  peer.Id,
					"err":   err,
					"epoch": epoch,
				}).Error("Send fetch epoch state request failed")
				return
			}

			epcStateResp := &pb.FetchEpochStateResponse{}
			if err = epcStateResp.UnmarshalVT(resp.Data); err != nil {
				s.logger.WithFields(logrus.Fields{
					"peer": peer.Id,
					"err":  err,
				}).Error("Unmarshal fetch epoch state response failed")
				return
			}

			epochState := consensustypes.QuorumCheckpointConstructor[s.consensusType]()

			if err = epochState.Unmarshal(epcStateResp.Data); err != nil {
				s.logger.WithFields(logrus.Fields{
					"peer": peer.Id,
					"err":  err,
				}).Error("Unmarshal epoch state failed")
				return
			}

			channel.TrySend(s.epochStateRecvCh, containers.Pack2(epoch, epochState))

		})
	}
	return errCh
}

func (s *SnapSync) pickPeer(taskNum int) *common.Node {
	index := taskNum % len(s.cnf.Peers)
	return s.cnf.Peers[index]
}

func (s *SnapSync) pickRandomPeer(exceptPeerIds ...*common.Node) (*common.Node, error) {
	if len(exceptPeerIds) > 0 {
		newPeers := lo.Filter(s.cnf.Peers, func(p *common.Node, _ int) bool {
			return !lo.Contains(exceptPeerIds, p)
		})
		if len(newPeers) == 0 {
			return nil, errors.New("all peers are excluded")
		}
		return newPeers[rand.Intn(len(newPeers))], nil
	}
	return s.cnf.Peers[rand.Intn(len(s.cnf.Peers))], nil
}
