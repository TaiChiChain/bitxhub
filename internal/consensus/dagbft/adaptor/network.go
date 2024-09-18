package adaptor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/chainstate"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common/metrics"
	consensustypes "github.com/axiomesh/axiom-ledger/internal/consensus/types"
	"github.com/axiomesh/axiom-ledger/internal/network"
	p2p "github.com/axiomesh/axiom-p2p"
	dagbft "github.com/bcds/go-hpc-dagbft"
	"github.com/bcds/go-hpc-dagbft/common/config"
	"github.com/bcds/go-hpc-dagbft/common/types"
	"github.com/bcds/go-hpc-dagbft/common/types/messages"
	"github.com/bcds/go-hpc-dagbft/common/types/protos"
	"github.com/bcds/go-hpc-dagbft/common/utils/channel"
	"github.com/bcds/go-hpc-dagbft/common/utils/containers"
	"github.com/bcds/go-hpc-dagbft/common/utils/results"
	"github.com/bcds/go-hpc-dagbft/protocol"
	"github.com/bcds/go-hpc-dagbft/protocol/layer"
	"github.com/gammazero/workerpool"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type (
	Pid  = string
	ID   = uint64
	Port = int
)

type NetworkProtocol interface {
	layer.NetworkFactory
}

type NetworkFactory struct {
	networks map[types.Host]*Network
}

func (n *NetworkFactory) Start(engine dagbft.DagBFT) {
	for _, net := range n.networks {
		net.Start(engine)
	}
}

func NewNetworkFactory(cnf *common.Config, ctx context.Context) *NetworkFactory {
	pid := cnf.ChainState.SelfNodeInfo.P2PID
	workers := lo.SliceToMap(cnf.ChainState.SelfNodeInfo.Workers, func(n types.Host) (types.Host, Pid) {
		return n, pid
	})
	remoteNodeIDs := lo.Filter(cnf.ChainState.ValidatorSet, func(n chainstate.ValidatorInfo, _ int) bool {
		return n.ID != cnf.ChainState.SelfNodeInfo.ID
	})
	remotePrimarys := lo.SliceToMap(remoteNodeIDs, func(n chainstate.ValidatorInfo) (types.Host, Pid) {
		info, err := cnf.ChainState.GetNodeInfo(n.ID)
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("%s%d", chainstate.PrimaryPrefix, n.ID), info.P2PID
	})
	remoteWorkers := lo.SliceToMap(remoteNodeIDs, func(n chainstate.ValidatorInfo) (types.Host, Pid) {
		info, err := cnf.ChainState.GetNodeInfo(n.ID)
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("%s%d", chainstate.WorkerPrefix, n.ID), info.P2PID
	})

	localPrimaryRequestCh := make(chan *wrapRequest, common.MaxChainSize)
	localWorkersRequestCh := lo.SliceToMap(cnf.ChainState.SelfNodeInfo.Workers, func(n types.Host) (types.Host, chan *wrapRequest) {
		return n, make(chan *wrapRequest, common.MaxChainSize)
	})

	networkConfig := *DefaultNetworkConfig()
	networkConfig.LocalWorkers = workers
	localPrimary := containers.Pack2(cnf.ChainState.SelfNodeInfo.Primary, cnf.ChainState.SelfNodeInfo.P2PID)
	networkConfig.LocalPrimary = localPrimary

	primaryNet := &Network{
		isPrimary:          true,
		Network:            cnf.Network,
		localHost:          localPrimary,
		remotes:            lo.Assign(remoteWorkers, remotePrimarys),
		networkConfig:      networkConfig,
		sendRequestCh:      make(chan *wrapRequest, common.MaxChainSize),
		recvLocalRequestCh: localPrimaryRequestCh,
		sendLocalWorkerCh:  lo.MapValues(localWorkersRequestCh, func(ch chan *wrapRequest, _ types.Host) chan<- *wrapRequest { return ch }),
		ctx:                ctx,
		logger:             cnf.Logger,
	}

	err := primaryNet.Network.RegisterMultiMsgHandler(PrimaryMessageTypes, primaryNet.handlePrimaryMsg)
	if err != nil {
		panic(err)
	}

	err = primaryNet.Network.RegisterMultiMsgHandler(EpochMessageTypes, primaryNet.handleEpochMsg)
	if err != nil {
		panic(err)
	}
	networks := map[types.Host]*Network{
		cnf.ChainState.SelfNodeInfo.Primary: primaryNet,
	}

	primaryNet.metrics = newNetworkMessageMetrics()
	lo.ForEach(cnf.ChainState.SelfNodeInfo.Workers, func(n types.Host, _ int) {
		workerNet := &Network{
			isPrimary:          false,
			Network:            cnf.Network,
			localHost:          containers.Pack2(n, pid),
			remotes:            lo.Assign(remoteWorkers, remotePrimarys),
			networkConfig:      networkConfig,
			sendRequestCh:      make(chan *wrapRequest, common.MaxChainSize),
			sendLocalPrimaryCh: localPrimaryRequestCh,
			recvLocalRequestCh: localWorkersRequestCh[n], // todo: forbid multi local workers communication?
			ctx:                ctx,
			logger:             cnf.Logger,
		}
		err = workerNet.Network.RegisterMultiMsgHandler(WorkerMessageTypes, workerNet.handleWorkerMsg)
		if err != nil {
			panic(err)
		}
		err = workerNet.Network.RegisterMultiMsgHandler(EpochMessageTypes, workerNet.handleEpochMsg)
		if err != nil {
			panic(err)
		}
		networks[n] = workerNet
	})

	return &NetworkFactory{
		networks: networks,
	}
}

func (n *NetworkFactory) GetNetwork(peer types.Peer) layer.Network {
	return n.networks[peer.Host]
}

type NetworkConfig struct {
	LocalPrimary    containers.Pair[types.Host, Pid]
	LocalWorkers    map[types.Host]Pid
	ConcurrentLimit int
	RetryConfig
}

type RetryConfig struct {
	RetryAttempts int
	RetryTimeout  time.Duration
}

func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		ConcurrentLimit: 10,
		RetryConfig: RetryConfig{
			RetryAttempts: 5,
			RetryTimeout:  500 * time.Millisecond,
		},
	}
}

type wrapRequest struct {
	local  bool
	to     Pid
	index  int
	msg    *pb.Message
	respCh chan<- protocol.MessageResult
}

type Network struct {
	network.Network
	engine dagbft.DagBFT

	isPrimary          bool
	localHost          containers.Pair[types.Host, Pid] // dagbft name -> pid
	recvLocalRequestCh <-chan *wrapRequest
	sendLocalPrimaryCh chan<- *wrapRequest
	sendLocalWorkerCh  map[types.Host]chan<- *wrapRequest

	networkConfig NetworkConfig

	remotes map[types.Host]Pid

	sendRequestCh chan *wrapRequest

	ctx context.Context

	logger logrus.FieldLogger

	metrics *networkMessageMetrics
}

func (n *Network) asyncSendResponseMsg(stream p2p.Stream, typ pb.Message_Type, data []byte) error {
	msg := &pb.Message{
		From: n.localHost.First,
		Type: typ,
		Data: data,
	}
	marshalMsg, err := msg.MarshalVT()
	if err != nil {
		return err
	}
	return stream.AsyncSend(marshalMsg)
}

func (n *Network) handleEpochMsg(stream p2p.Stream, msg *pb.Message) {
	data, responseType, err := n.processEpochMsg(msg)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return
	}
	if err = n.asyncSendResponseMsg(stream, responseType, data); err != nil {
		n.logger.Errorf("send response err: %v", err)
	}
}
func (n *Network) processEpochMsg(msg *pb.Message) ([]byte, pb.Message_Type, error) {
	requestMsg, err := decodeMessageFrom(msg, n.localHost.First)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return nil, pb.Message_Consensus_ERROR, err
	}
	resultCh := n.engine.ReceiveMessage(context.Background(), requestMsg)
	select {
	case <-n.ctx.Done():
		return nil, pb.Message_Consensus_ERROR, fmt.Errorf("context done")

	case res := <-resultCh:
		var (
			data         []byte
			responseType pb.Message_Type
			nullMsg      = &messages.NullMessage{}
		)
		if res.Err() != nil {
			n.logger.Errorf("receive response err: %v", res.Err())
			responseType = pb.Message_Consensus_ERROR
			errMsg := &pb.ConsensusError{Error: res.Error.Error(), RequestType: msg.Type}
			data, err = errMsg.MarshalVT()
		} else {
			switch msg.GetType() {
			case pb.Message_EPOCH_REQUEST:
				responseType = pb.Message_EPOCH_REQUEST_RESPONSE
				data, err = res.Val().(*messages.EpochChangeResponse).Marshal()

			case pb.Message_EPOCH_PROOF:
				responseType = pb.Message_EPOCH_PROOF_RESPONSE
				data, err = nullMsg.Marshal()
			}
		}

		if err != nil {
			n.logger.Errorf("send response err: %v", err)
			return nil, pb.Message_Consensus_ERROR, err
		}
		return data, responseType, nil
	}
}

func (n *Network) handlePrimaryMsg(stream p2p.Stream, msg *pb.Message) {
	data, responseType, err := n.processPrimaryMsg(msg)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return
	}
	if err = n.asyncSendResponseMsg(stream, responseType, data); err != nil {
		n.logger.Errorf("send response err: %v", err)
	}
}

func (n *Network) processPrimaryMsg(msg *pb.Message) ([]byte, pb.Message_Type, error) {
	requestMsg, err := decodeMessageFrom(msg, n.localHost.First)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return nil, pb.Message_Consensus_ERROR, err
	}
	resultCh := n.engine.ReceiveMessage(context.Background(), requestMsg)
	select {
	case <-n.ctx.Done():
		return nil, pb.Message_Consensus_ERROR, fmt.Errorf("context done")

	case res := <-resultCh:
		var (
			data         []byte
			responseType pb.Message_Type
			nullMsg      = &messages.NullMessage{}
		)
		if res.Err() != nil {
			n.logger.Errorf("receive response err: %v", res.Err())
			responseType = pb.Message_Consensus_ERROR
			errMsg := &pb.ConsensusError{Error: res.Error.Error(), RequestType: msg.Type}
			data, err = errMsg.MarshalVT()
		} else {
			switch msg.GetType() {
			case pb.Message_REQUEST_VOTE:
				responseType = pb.Message_REQUEST_VOTE_RESPONSE
				data, err = res.Val().(*messages.RequestVoteResponse).Marshal()

			case pb.Message_SEND_CERTIFICATE:
				responseType = pb.Message_SEND_CERTIFICATE_RESPONSE
				data, err = res.Val().(*messages.SendCertificateResponse).Marshal()
			case pb.Message_FETCH_CERTIFICATES:
				responseType = pb.Message_FETCH_CERTIFICATES_RESPONSE
				data, err = res.Val().(*messages.FetchCertificatesResponse).Marshal()
			case pb.Message_FETCH_CHECKPOINT:
				responseType = pb.Message_FETCH_CHECKPOINT_RESPONSE
				data, err = res.Val().(*messages.FetchCheckpointResponse).Marshal()
			case pb.Message_FETCH_QUORUM_CHECKPOINT:
				responseType = pb.Message_FETCH_QUORUM_CHECKPOINT_RESPONSE
				data, err = res.Val().(*messages.FetchQuorumCheckpointResponse).Marshal()
			case pb.Message_FETCH_COMMIT_SUB_DAG:
				responseType = pb.Message_FETCH_COMMIT_SUB_DAG_RESPONSE
				data, err = res.Val().(*messages.FetchCommitSummaryResponse).Marshal()
			case pb.Message_SEND_CHECKPOINT:
				responseType = pb.Message_SEND_CHECKPOINT_RESPONSE
				data, err = nullMsg.Marshal()
			case pb.Message_WORKER_OUR_BATCH:
				responseType = pb.Message_WORKER_OUR_BATCH_RESPONSE
				data, err = nullMsg.Marshal()
			case pb.Message_WORKER_OTHERS_BATCH:
				responseType = pb.Message_WORKER_OTHERS_BATCH_RESPONSE
				data, err = nullMsg.Marshal()
			}

		}

		if err != nil {
			n.logger.Errorf("receive message from %s err: %s", msg.From, err)
			return nil, pb.Message_Consensus_ERROR, err
		}

		return data, responseType, nil
	}
}

func (n *Network) getRemoteHost(pid Pid) (types.Host, bool) {
	return lo.FindKey(n.remotes, pid)
}

func (n *Network) handleWorkerMsg(stream p2p.Stream, msg *pb.Message) {
	data, responseType, err := n.processWorkerMsg(msg)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return
	}
	if err = n.asyncSendResponseMsg(stream, responseType, data); err != nil {
		n.logger.Errorf("send response err: %v", err)
	}
}
func (n *Network) processWorkerMsg(msg *pb.Message) ([]byte, pb.Message_Type, error) {
	requestMsg, err := decodeMessageFrom(msg, n.localHost.First)
	if err != nil {
		host, exist := n.getRemoteHost(msg.From)
		if exist {
			n.logger.Errorf("receive message from %s err: decode message error: %v", host, err)
		} else {
			n.logger.Errorf("receive message from %s err: decode message error: %v", msg.From, err)
		}
		return nil, pb.Message_Consensus_ERROR, err
	}
	resultCh := n.engine.ReceiveMessage(context.Background(), requestMsg)
	select {
	case <-n.ctx.Done():
		return nil, pb.Message_Consensus_ERROR, fmt.Errorf("ctx done")

	case res := <-resultCh:
		var (
			data         []byte
			responseType pb.Message_Type
			nullMsg      = &messages.NullMessage{}
		)
		if res.Err() != nil {
			n.logger.Errorf("receive response err: %v", res.Err())
			responseType = pb.Message_Consensus_ERROR
			errMsg := &pb.ConsensusError{Error: res.Error.Error(), RequestType: msg.Type}
			data, err = errMsg.MarshalVT()
		} else {
			switch msg.GetType() {
			case pb.Message_NEW_BATCH:
				responseType = pb.Message_NEW_BATCH_RESPONSE
				data, err = nullMsg.Marshal()
			case pb.Message_WORKER_SYNCHRONIZE:
				responseType = pb.Message_WORKER_SYNCHRONIZE_RESPONSE
				data, err = nullMsg.Marshal()
			case pb.Message_PRIMARY_STATE:
				responseType = pb.Message_PRIMARY_STATE_RESPONSE
				data, err = nullMsg.Marshal()
			case pb.Message_REQUEST_BATCH:
				responseType = pb.Message_REQUEST_BATCH_RESPONSE
				data, err = res.Val().(*messages.RequestBatchResponse).Marshal()
			case pb.Message_REQUEST_BATCHES:
				responseType = pb.Message_REQUEST_BATCHES_RESPONSE
				data, err = res.Val().(*messages.RequestBatchesResponse).Marshal()
			case pb.Message_FETCH_BATCHES:
				responseType = pb.Message_FETCH_BATCHES_RESPONSE
				data, err = res.Val().(*messages.FetchBatchesResponse).Marshal()
			}
		}

		if err != nil {
			host, exist := n.getRemoteHost(msg.From)
			if exist {
				n.logger.Errorf("receive message from %s err: decode message error: %s", host, err)
			} else {
				n.logger.Errorf("receive message from %s err: decode message error: %s", msg.From, err)
			}
			return nil, pb.Message_Consensus_ERROR, err
		}

		return data, responseType, nil
	}
}

func (n *Network) Start(engine dagbft.DagBFT) {
	n.engine = engine
	go n.listenRequestToSubmit()
	go n.listenLocalRequest()
}

func (n *Network) listenLocalRequest() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case req := <-n.recvLocalRequestCh:
			var (
				responseNetMsg = &pb.Message{}
				result         protocol.MessageResult
				respData       []byte
				respType       pb.Message_Type
				err            error
			)

			if !lo.Contains(lo.Union(PrimaryMessageTypes, WorkerMessageTypes, EpochMessageTypes), req.msg.Type) {
				result = genWrongResponse(fmt.Errorf("invalid message type: %d", req.msg.Type), req.index)
				channel.TrySend(req.respCh, result)
				continue
			}

			n.logger.Info("receive message from local", "type", req.msg.Type, "network", n.isPrimary)
			// send request to local consensus engine, wait for response
			switch {
			case lo.Contains(PrimaryMessageTypes, req.msg.Type):
				respData, respType, err = n.processPrimaryMsg(req.msg)
			case lo.Contains(WorkerMessageTypes, req.msg.Type):
				respData, respType, err = n.processWorkerMsg(req.msg)
			case lo.Contains(EpochMessageTypes, req.msg.Type):
				respData, respType, err = n.processEpochMsg(req.msg)
			}
			responseNetMsg.Data = respData
			responseNetMsg.Type = respType
			if err != nil {
				result = genWrongResponse(err, req.index)
				channel.TrySend(req.respCh, result)
				continue
			}
			resp, err := decodeMessageResult(responseNetMsg, n.localHost.First)
			if err != nil {
				n.logger.Errorf("receive message from %s err: decode message error: %v", req.to, err)
				result = genWrongResponse(err, req.index)
				channel.TrySend(req.respCh, result)
				continue
			}

			result = protocol.MessageResult{
				Index:  req.index,
				Result: results.OK(resp),
			}
			channel.TrySend(req.respCh, result)
		}
	}
}

func (n *Network) Author() types.Host {
	return n.localHost.First
}

func genWrongResponse(err error, index int) protocol.MessageResult {
	return protocol.MessageResult{
		Index:  index,
		Result: results.Error[protocol.Message](err),
	}
}

func (n *Network) listenRequestToSubmit() {
	wp := workerpool.New(n.networkConfig.ConcurrentLimit)
	for {
		select {
		case <-n.ctx.Done():
			wp.StopWait()
			return
		case req := <-n.sendRequestCh:
			wp.Submit(func() {
				var result protocol.MessageResult
				if err := retry.Retry(func(attempt uint) error {
					now := time.Now()
					responseMsg, err := n.Send(req.to, req.msg)
					if err != nil {
						if strings.Contains(err.Error(), p2p.WaitMsgTimeout.Error()) {
							return err
						} else {
							result = genWrongResponse(err, req.index)
							return nil
						}
					}

					resp, err := decodeMessageResult(responseMsg, n.localHost.First)
					if err != nil {
						host, exist := n.getRemoteHost(req.to)
						if exist {
							n.logger.Errorf("receive message from %s err: decode message error: %s", host, err)
						} else {
							n.logger.Errorf("receive message from %s err: decode message error: %s", req.to, err)
						}
						result = genWrongResponse(err, req.index)
						return nil
					}

					result = protocol.MessageResult{
						Index:  req.index,
						Result: results.OK(resp),
					}
					metrics.ProcessConsensusMessageDuration.With(prometheus.Labels{"consensus": consensustypes.Dagbft, "event": req.msg.Type.String()}).Observe(time.Since(now).Seconds())
					return nil
				}, strategy.Limit(uint(n.networkConfig.RetryTimeout)), strategy.Wait(200*time.Millisecond)); err != nil {
					result = genWrongResponse(err, req.index)
				}

				req.respCh <- result
			})
		}
	}
}

func (n *Network) sendLocal(req *wrapRequest, host types.Host) bool {
	var find bool
	if ch, ok := n.sendLocalWorkerCh[host]; ok {
		find = true
		channel.TrySend(ch, req)
	}
	if !n.isPrimary && n.networkConfig.LocalPrimary.First == host {
		find = true
		channel.TrySend(n.sendLocalPrimaryCh, req)
	}
	return find
}

func (n *Network) Unicast(req protocol.Message, peer types.Peer, retry config.RetryConfig) (<-chan protocol.MessageResult, chan bool) {
	respCh := make(chan protocol.MessageResult, 1)
	cancel := make(chan bool)
	n.unicast(req, peer, retry, respCh, cancel)
	return respCh, cancel
}

func (n *Network) unicast(request protocol.Message, peer types.Peer, retryConf config.RetryConfig, respCh chan protocol.MessageResult, cancelCh chan bool) {
	isLocal := func(peerHost types.Host) bool {
		if n.isPrimary {
			if _, find := n.networkConfig.LocalWorkers[peerHost]; find {
				return true
			}
		} else {
			return n.networkConfig.LocalPrimary.First == peerHost
		}
		return false
	}
	defer func() {
		n.logger.WithFields(logrus.Fields{
			"to":        peer.Host,
			"type":      reflect.TypeOf(request),
			"isLocal":   isLocal(peer.Host),
			"isPrimary": n.isPrimary,
		}).Info("start unicast")
	}()
	var (
		enc   []byte
		err   error
		index = 0
	)

	if m, ok := request.(proto.Marshaler); ok {
		enc, err = m.Marshal()
	} else {
		enc, err = json.Marshal(m)
	}

	if err != nil {
		failedResp := protocol.MessageResult{
			Index:  0,
			Result: results.Error[protocol.Message](err),
		}

		channel.SafeSend(respCh, failedResp, cancelCh)
		return
	}

	var (
		pid   Pid
		ok    bool
		local bool
	)
	if isLocal(peer.Host) {
		local = true
		pid = n.localHost.First
	} else {
		pid, ok = n.remotes[peer.Host]
		if !ok {
			failedResp := protocol.MessageResult{
				Index:  0,
				Result: results.Error[protocol.Message](fmt.Errorf("peer not found: %s", peer.Host)),
			}

			channel.SafeSend(respCh, failedResp, cancelCh)
			return
		}
	}

	netMsg := &pb.Message{
		From: n.Author(),
		Data: enc,
		Type: dispatchMsgToType(request),
	}

	wr := &wrapRequest{local: local, to: pid, index: index, msg: netMsg, respCh: respCh}
	if local {
		if success := n.sendLocal(wr, peer.Host); !success {
			n.logger.Warnf("send local request failed: %s", peer.Host)
		}
	} else {
		channel.SafeSend(n.sendRequestCh, wr, cancelCh)
	}
	return
}

func (n *Network) Broadcast(m protocol.Message, peers []types.Peer, retry config.RetryConfig) (<-chan protocol.MessageResult, []chan bool) {
	cancels := channel.MakeChannels[bool](len(peers))
	responses := make(chan protocol.MessageResult, len(peers))
	lo.ForEach(peers, func(peer types.Peer, index int) {
		n.unicast(m, peer, retry, responses, cancels[index])
	})
	return responses, cancels
}

var WorkerMessageTypes = []pb.Message_Type{
	pb.Message_NEW_BATCH,
	pb.Message_NEW_BATCH_RESPONSE,
	pb.Message_WORKER_SYNCHRONIZE,
	pb.Message_WORKER_SYNCHRONIZE_RESPONSE,
	pb.Message_REQUEST_BATCH,
	pb.Message_REQUEST_BATCH_RESPONSE,
	pb.Message_REQUEST_BATCHES,
	pb.Message_REQUEST_BATCHES_RESPONSE,
	pb.Message_FETCH_BATCHES,
	pb.Message_FETCH_BATCHES_RESPONSE,
	pb.Message_PRIMARY_STATE,
	pb.Message_PRIMARY_STATE_RESPONSE,
}

var PrimaryMessageTypes = []pb.Message_Type{
	pb.Message_REQUEST_VOTE,
	pb.Message_REQUEST_VOTE_RESPONSE,
	pb.Message_SEND_CERTIFICATE,
	pb.Message_SEND_CERTIFICATE_RESPONSE,
	pb.Message_FETCH_CERTIFICATES,
	pb.Message_FETCH_CERTIFICATES_RESPONSE,
	pb.Message_FETCH_COMMIT_SUB_DAG,
	pb.Message_FETCH_COMMIT_SUB_DAG_RESPONSE,
	pb.Message_WORKER_OUR_BATCH,
	pb.Message_WORKER_OUR_BATCH_RESPONSE,
	pb.Message_WORKER_OTHERS_BATCH,
	pb.Message_WORKER_OTHERS_BATCH_RESPONSE,
	pb.Message_SEND_CHECKPOINT,
	pb.Message_SEND_CHECKPOINT_RESPONSE,
	pb.Message_FETCH_CHECKPOINT,
	pb.Message_FETCH_CHECKPOINT_RESPONSE,
	pb.Message_FETCH_QUORUM_CHECKPOINT,
	pb.Message_FETCH_QUORUM_CHECKPOINT_RESPONSE,
}

var EpochMessageTypes = []pb.Message_Type{
	pb.Message_EPOCH_REQUEST,
	pb.Message_EPOCH_PROOF,
	pb.Message_EPOCH_REQUEST_RESPONSE,
	pb.Message_EPOCH_PROOF_RESPONSE,
}

func decodeMessageFrom(msg *pb.Message, author types.Host) (protocol.Message, error) {
	switch msg.Type {
	// epoch message
	case pb.Message_EPOCH_REQUEST:
		var req protos.EpochChangeRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.EpochChangeRequest{
			EpochChangeRequest: req,
			IsMessageType:      protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_EPOCH_PROOF:
		var req protos.EpochChangeProof
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.SendEpochChangeProof{
			EpochChangeProof: req,
			IsMessageType:    protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil

	case pb.Message_REQUEST_VOTE:
		var req protos.RequestVoteRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestVoteRequest{
			RequestVoteRequest: req,
			IsMessageType:      protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_SEND_CERTIFICATE:
		var req protos.SendCertificateRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.SendCertificateRequest{
			SendCertificateRequest: req,
			IsMessageType:          protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_CERTIFICATES:
		var req protos.FetchCertificatesRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCertificatesRequest{
			FetchCertificatesRequest: req,
			IsMessageType:            protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_WORKER_OUR_BATCH:
		var req protos.WorkerOurBatchMsg
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.WorkerOurBatchMsg{
			WorkerOurBatchMsg: req,
			IsMessageType:     protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_WORKER_OTHERS_BATCH:
		var req protos.WorkerOthersBatchMsg
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.WorkerOthersBatchMsg{
			WorkerOthersBatchMsg: req,
			IsMessageType:        protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_SEND_CHECKPOINT:
		var req protos.SignedCheckpoint
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.SendCheckpointRequest{
			SignedCheckpoint: req,
			IsMessageType:    protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_CHECKPOINT:
		var req protos.FetchCheckpointRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCheckpointRequest{
			FetchCheckpointRequest: req,
			IsMessageType:          protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_QUORUM_CHECKPOINT:
		var req protos.FetchQuorumCheckpointRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchQuorumCheckpointRequest{
			FetchQuorumCheckpointRequest: req,
			IsMessageType:                protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_COMMIT_SUB_DAG:
		var req protos.FetchCommitSummaryRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCommitSummaryRequest{
			FetchCommitSummaryRequest: req,
			IsMessageType:             protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil

	// handle worker msg
	case pb.Message_NEW_BATCH:
		var req protos.WorkerBatchMsg
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.WorkerBatchMsg{
			WorkerBatchMsg: req,
			IsMessageType:  protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_WORKER_SYNCHRONIZE:
		var req protos.WorkerSynchronizeMsg
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.WorkerSynchronizeMsg{
			WorkerSynchronizeMsg: req,
			IsMessageType:        protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_REQUEST_BATCH:
		var req protos.RequestBatchRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestBatchRequest{
			RequestBatchRequest: req,
			IsMessageType:       protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_REQUEST_BATCHES:
		var req protos.RequestBatchesRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestBatchesRequest{
			RequestBatchesRequest: req,
			IsMessageType:         protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_BATCHES:
		var req protos.FetchBatchesRequest
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchBatchesRequest{
			FetchBatchesRequest: req,
			IsMessageType:       protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_PRIMARY_STATE:
		var req protos.PrimaryStateMsg
		err := req.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.PrimaryStateMsg{
			PrimaryStateMsg: req,
			IsMessageType:   protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	}
	return nil, fmt.Errorf("unknown message type: %d", msg.Type)
}

func decodeMessageResult(msg *pb.Message, author types.Host) (protocol.Message, error) {
	switch msg.Type {
	case pb.Message_Consensus_ERROR:
		errMsg := &pb.ConsensusError{}
		err := errMsg.UnmarshalVT(msg.Data)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("receive message error: %s, type: %s", errMsg.GetError(), errMsg.GetRequestType().String())
	// epoch message
	case pb.Message_EPOCH_REQUEST_RESPONSE:
		var resp protos.EpochChangeResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.EpochChangeResponse{
			EpochChangeResponse: resp,
			IsMessageType:       protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_EPOCH_PROOF_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil

	// Primary messages
	case pb.Message_REQUEST_VOTE_RESPONSE:
		var resp protos.RequestVoteResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestVoteResponse{
			RequestVoteResponse: resp,
			IsMessageType:       protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_SEND_CERTIFICATE_RESPONSE:
		var resp protos.SendCertificateResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.SendCertificateResponse{
			SendCertificateResponse: resp,
			IsMessageType:           protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_CERTIFICATES_RESPONSE:
		var resp protos.FetchCertificatesResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCertificatesResponse{
			FetchCertificatesResponse: resp,
			IsMessageType:             protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_COMMIT_SUB_DAG_RESPONSE:
		var resp protos.FetchCommitSummaryResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCommitSummaryResponse{
			FetchCommitSummaryResponse: resp,
			IsMessageType:              protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_WORKER_OUR_BATCH_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	case pb.Message_WORKER_OTHERS_BATCH_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	case pb.Message_SEND_CHECKPOINT_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	case pb.Message_FETCH_CHECKPOINT_RESPONSE:
		var resp protos.SignedCheckpoint
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchCheckpointResponse{
			SignedCheckpoint: resp,
			IsMessageType:    protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_FETCH_QUORUM_CHECKPOINT_RESPONSE:
		var resp protos.QuorumCheckpoint
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchQuorumCheckpointResponse{
			QuorumCheckpoint: resp,
			IsMessageType:    protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil

	// workers messages
	case pb.Message_NEW_BATCH_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	case pb.Message_WORKER_SYNCHRONIZE_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	case pb.Message_FETCH_BATCHES_RESPONSE:
		var resp protos.FetchBatchesResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.FetchBatchesResponse{
			FetchBatchesResponse: resp,
			IsMessageType:        protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_REQUEST_BATCH_RESPONSE:
		var resp protos.RequestBatchResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestBatchResponse{
			RequestBatchResponse: resp,
			IsMessageType:        protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_REQUEST_BATCHES_RESPONSE:
		var resp protos.RequestBatchesResponse
		err := resp.Unmarshal(msg.Data)
		if err != nil {
			return nil, err
		}
		return &messages.RequestBatchesResponse{
			RequestBatchesResponse: resp,
			IsMessageType:          protocol.IsMessageType{Sender: msg.From, Receiver: author},
		}, nil
	case pb.Message_PRIMARY_STATE_RESPONSE:
		return &messages.NullMessage{IsMessageType: protocol.IsMessageType{Sender: msg.From, Receiver: author}}, nil
	}

	return nil, fmt.Errorf("wrong message type: %v", msg.Type)
}

func dispatchMsgToType(msg protocol.Message) pb.Message_Type {
	switch msg.(type) {
	// primary
	case *messages.RequestVoteRequest:
		return pb.Message_REQUEST_VOTE
	case *messages.SendCertificateRequest:
		return pb.Message_SEND_CERTIFICATE
	case *messages.FetchCertificatesRequest:
		return pb.Message_FETCH_CERTIFICATES
	case *messages.FetchCommitSummaryRequest:
		return pb.Message_FETCH_COMMIT_SUB_DAG
	case *messages.WorkerOurBatchMsg:
		return pb.Message_WORKER_OUR_BATCH
	case *messages.WorkerOthersBatchMsg:
		return pb.Message_WORKER_OTHERS_BATCH
	case *messages.SendCheckpointRequest:
		return pb.Message_SEND_CHECKPOINT
	case *messages.FetchCheckpointRequest:
		return pb.Message_FETCH_CHECKPOINT
	case *messages.FetchQuorumCheckpointRequest:
		return pb.Message_FETCH_QUORUM_CHECKPOINT
	// worker
	case *messages.WorkerBatchMsg:
		return pb.Message_NEW_BATCH
	case *messages.WorkerSynchronizeMsg:
		return pb.Message_WORKER_SYNCHRONIZE
	case *messages.FetchBatchesRequest:
		return pb.Message_FETCH_BATCHES
	case *messages.RequestBatchRequest:
		return pb.Message_REQUEST_BATCH
	case *messages.RequestBatchesRequest:
		return pb.Message_REQUEST_BATCHES
	case *messages.PrimaryStateMsg:
		return pb.Message_PRIMARY_STATE
	// epoch
	case *messages.EpochChangeRequest:
		return pb.Message_EPOCH_REQUEST
	case *messages.SendEpochChangeProof:
		return pb.Message_EPOCH_PROOF
	}

	return pb.Message_UNKNOWN
}
