package solo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-bft/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func init() {
	repo.Register(repo.ConsensusTypeSolo, false)
}

type Node struct {
	config           *common.Config
	isTimed          bool
	commitC          chan *common.CommitEvent                                             // block channel
	logger           logrus.FieldLogger                                                   // logger
	txpool           txpool.TxPool[types.Transaction, *types.Transaction]                 // transaction pool
	batchDigestM     map[uint64]string                                                    // mapping blockHeight to batch digest
	poolFull         int32                                                                // pool full symbol
	recvCh           chan consensusEvent                                                  // receive message from consensus engine
	blockCh          chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction] // receive batch from txpool
	batchMgr         *timerManager
	noTxBatchTimeout time.Duration   // generate no-tx block period
	batchTimeout     time.Duration   // generate block period
	lastExec         uint64          // the index of the last-applied block
	network          network.Network // network manager
	checkpoint       uint64
	txPreCheck       precheck.PreCheck

	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
	txFeed event.Feed
}

func NewNode(config *common.Config) (*Node, error) {
	fn := func(addr string) uint64 {
		account := types.NewAddressByStr(addr)
		return config.GetAccountNonce(account)
	}

	txpoolConf := txpool.Config{
		Logger:              &common.Logger{FieldLogger: loggers.Logger(loggers.TxPool)},
		BatchSize:           config.GenesisEpochInfo.ConsensusParams.BlockMaxTxNum,
		PoolSize:            config.Config.TxPool.PoolSize,
		GetAccountNonce:     fn,
		IsTimed:             config.GenesisEpochInfo.ConsensusParams.EnableTimedGenEmptyBlock,
		ToleranceRemoveTime: config.Config.TxPool.ToleranceRemoveTime.ToDuration(),
	}
	txpoolInst := txpool.NewTxPool[types.Transaction, *types.Transaction](txpoolConf)
	// init batch timer manager
	recvCh := make(chan consensusEvent)
	batchTimerMgr := NewTimerManager(recvCh, config.Logger)
	batchTimerMgr.newTimer(Batch, config.Config.TxPool.BatchTimeout.ToDuration())
	batchTimerMgr.newTimer(NoTxBatch, config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration())
	batchTimerMgr.newTimer(RemoveTx, config.Config.TxPool.ToleranceRemoveTime.ToDuration())

	ctx, cancel := context.WithCancel(context.Background())
	soloNode := &Node{
		config:           config,
		isTimed:          txpoolConf.IsTimed,
		noTxBatchTimeout: config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration(),
		batchTimeout:     config.Config.TxPool.BatchTimeout.ToDuration(),
		blockCh:          make(chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction], maxChanSize),
		commitC:          make(chan *common.CommitEvent, maxChanSize),
		batchDigestM:     make(map[uint64]string),
		checkpoint:       config.Config.Solo.CheckpointPeriod,
		poolFull:         0,
		recvCh:           recvCh,
		lastExec:         config.Applied,
		txpool:           txpoolInst,
		batchMgr:         batchTimerMgr,
		network:          config.Network,
		ctx:              ctx,
		cancel:           cancel,
		txPreCheck:       precheck.NewTxPreCheckMgr(ctx, config.EVMConfig, config.Logger, config.GetAccountBalance),
		logger:           config.Logger,
	}
	soloNode.logger.Infof("SOLO lastExec = %d", soloNode.lastExec)
	soloNode.logger.Infof("SOLO epoch period = %d", config.GenesisEpochInfo.EpochPeriod)
	soloNode.logger.Infof("SOLO checkpoint period = %d", config.GenesisEpochInfo.ConsensusParams.CheckpointPeriod)
	soloNode.logger.Infof("SOLO enable timed gen empty block = %v", config.GenesisEpochInfo.ConsensusParams.EnableTimedGenEmptyBlock)
	soloNode.logger.Infof("SOLO no-tx batch timeout = %v", config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration())
	soloNode.logger.Infof("SOLO batch timeout = %v", config.Config.TxPool.BatchTimeout.ToDuration())
	soloNode.logger.Infof("SOLO batch size = %d", config.GenesisEpochInfo.ConsensusParams.BlockMaxTxNum)
	soloNode.logger.Infof("SOLO pool size = %d", config.Config.TxPool.PoolSize)
	soloNode.logger.Infof("SOLO tolerance time = %v", config.Config.TxPool.ToleranceTime.ToDuration())
	soloNode.logger.Infof("SOLO tolerance remove time = %v", config.Config.TxPool.ToleranceRemoveTime.ToDuration())
	soloNode.logger.Infof("SOLO tolerance nonce gap = %d", config.Config.TxPool.ToleranceNonceGap)
	return soloNode, nil
}

func (n *Node) GetPendingTxByHash(hash *types.Hash) *types.Transaction {
	request := &getTxReq{
		Hash: hash.String(),
		Resp: make(chan *types.Transaction),
	}
	n.recvCh <- request
	return <-request.Resp
}

func (n *Node) GetTotalPendingTxCount() uint64 {
	req := &getTotalPendingTxCountReq{
		Resp: make(chan uint64),
	}
	n.recvCh <- req
	return <-req.Resp
}

func (n *Node) GetLowWatermark() uint64 {
	req := &getLowWatermarkReq{
		Resp: make(chan uint64),
	}
	n.recvCh <- req
	return <-req.Resp
}

func (n *Node) Start() error {
	n.logger.Info("Consensus started")
	if n.isTimed {
		n.batchMgr.startTimer(NoTxBatch)
	}
	n.batchMgr.startTimer(RemoveTx)
	n.txPreCheck.Start()
	go n.listenValidTxs()
	go n.listenEvent()
	go n.listenReadyBlock()
	return nil
}

func (n *Node) Stop() {
	n.cancel()
	n.logger.Info("Consensus stopped")
}

func (n *Node) GetPendingTxCountByAccount(account string) uint64 {
	request := &getNonceReq{
		account: account,
		Resp:    make(chan uint64),
	}
	n.recvCh <- request
	return <-request.Resp
}

func (n *Node) Prepare(tx *types.Transaction) error {
	if err := n.Ready(); err != nil {
		return fmt.Errorf("node get ready failed: %w", err)
	}
	txWithResp := &common.TxWithResp{
		Tx:     tx,
		RespCh: make(chan *common.TxResp),
	}
	n.recvCh <- txWithResp
	resp := <-txWithResp.RespCh
	if !resp.Status {
		return fmt.Errorf(resp.ErrorMsg)
	}
	return nil
}

func (n *Node) Commit() chan *common.CommitEvent {
	return n.commitC
}

func (n *Node) Step([]byte) error {
	return nil
}

func (n *Node) Ready() error {
	if n.isPoolFull() {
		return fmt.Errorf(ErrPoolFull)
	}
	return nil
}

func (n *Node) ReportState(height uint64, blockHash *types.Hash, txHashList []*types.Hash, _ *consensus.Checkpoint) {
	state := &chainState{
		Height:     height,
		BlockHash:  blockHash,
		TxHashList: txHashList,
	}
	n.recvCh <- state
}

func (n *Node) Quorum() uint64 {
	return 1
}

func (n *Node) SubscribeTxEvent(events chan<- []*types.Transaction) event.Subscription {
	return n.txFeed.Subscribe(events)
}

func (n *Node) SubmitTxsFromRemote(_ [][]byte) error {
	return nil
}

func (n *Node) listenEvent() {
	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("----- Exit listen event -----")
			return

		case ev := <-n.recvCh:
			switch e := ev.(type) {
			// handle report state
			case *chainState:
				if e.Height%n.checkpoint == 0 {
					n.logger.WithFields(logrus.Fields{
						"height": e.Height,
						"hash":   e.BlockHash.String(),
					}).Info("Report checkpoint")
					digestList := make([]string, 0)
					for i := e.Height; i > e.Height-n.checkpoint; i-- {
						for h, d := range n.batchDigestM {
							if i == h {
								digestList = append(digestList, d)
								delete(n.batchDigestM, i)
							}
						}
					}

					n.txpool.RemoveBatches(digestList)
					if !n.txpool.IsPoolFull() {
						n.setPoolNotFull()
					}
				}

			// receive tx from api
			case *common.TxWithResp:
				unCheckedEv := &common.UncheckedTxEvent{
					EventType: common.LocalTxEvent,
					Event:     e,
				}
				n.txPreCheck.PostUncheckedTxEvent(unCheckedEv)

			case *precheck.ValidTxs:
				if !e.Local {
					n.logger.Errorf("Receive remote type tx")
					continue
				}
				if n.txpool.IsPoolFull() {
					n.logger.Warn("TxPool is full")
					n.setPoolFull()
					e.LocalRespCh <- &common.TxResp{
						Status:   false,
						ErrorMsg: ErrPoolFull,
					}
					continue
				}
				// stop no-tx batch timer when this node receives the first transaction
				n.batchMgr.stopTimer(NoTxBatch)
				// start batch timer when this node receives the first transaction
				if !n.batchMgr.isTimerActive(Batch) {
					n.batchMgr.startTimer(Batch)
				}

				if len(e.Txs) != singleTx {
					n.logger.Warningf("Receive wrong txs length from local, expect:%d, actual:%d", singleTx, len(e.Txs))
				}

				if batches, _ := n.txpool.AddNewRequests(e.Txs, true, true, false, true); batches != nil {
					n.batchMgr.stopTimer(Batch)
					if len(batches) != 1 {
						n.logger.Errorf("Batch size is not 1, actual: %d", len(batches))
						continue
					}
					n.postProposal(batches[0])
					// start no-tx batch timer when this node handle the last transaction
					if n.isTimed && !n.txpool.HasPendingRequestInPool() {
						n.batchMgr.startTimer(NoTxBatch)
					}
				}

				// post tx event to websocket
				go n.txFeed.Send(e.Txs)
				e.LocalRespCh <- &common.TxResp{Status: true}

			// handle timeout event
			case batchTimeoutEvent:
				if err := n.processBatchTimeout(e); err != nil {
					n.logger.Errorf("Process batch timeout failed: %v", err)
					continue
				}
			case *getTxReq:
				e.Resp <- n.txpool.GetPendingTxByHash(e.Hash)
			case *getNonceReq:
				e.Resp <- n.txpool.GetPendingTxCountByAccount(e.account)
			case *getTotalPendingTxCountReq:
				e.Resp <- n.txpool.GetTotalPendingTxCount()
			case *getLowWatermarkReq:
				e.Resp <- n.lastExec
			}
		}
	}
}

func (n *Node) processBatchTimeout(e batchTimeoutEvent) error {
	switch e {
	case Batch:
		n.batchMgr.stopTimer(Batch)
		n.logger.Debug("Batch timer expired, try to create a batch")
		if n.txpool.HasPendingRequestInPool() {
			if batches := n.txpool.GenerateRequestBatch(); batches != nil {
				now := time.Now().UnixNano()
				if n.batchMgr.lastBatchTime != 0 {
					interval := time.Duration(now - n.batchMgr.lastBatchTime).Seconds()
					batchInterval.WithLabelValues("timeout").Observe(interval)
					if n.batchMgr.minTimeoutBatchTime == 0 || interval < n.batchMgr.minTimeoutBatchTime {
						n.logger.Debugf("update min timeoutBatch Time[height:%d, interval:%f, lastBatchTime:%v]",
							n.lastExec+1, interval, time.Unix(0, n.batchMgr.lastBatchTime))
						minBatchIntervalDuration.WithLabelValues("timeout").Set(interval)
						n.batchMgr.minTimeoutBatchTime = interval
					}
				}
				n.batchMgr.lastBatchTime = now
				for _, batch := range batches {
					n.postProposal(batch)
				}
				n.batchMgr.startTimer(Batch)

				// check if there is no tx in the txpool, start the no tx batch timer
				if n.isTimed && !n.txpool.HasPendingRequestInPool() {
					if !n.batchMgr.isTimerActive(NoTxBatch) {
						n.batchMgr.startTimer(NoTxBatch)
					}
				}
			}
		} else {
			n.logger.Debug("The length of priorityIndex is 0, skip the batch timer")
		}
	case NoTxBatch:
		if !n.isTimed {
			return errors.New("the node is not support the no-tx batch, skip the timer")
		}
		if !n.txpool.HasPendingRequestInPool() {
			n.batchMgr.stopTimer(NoTxBatch)
			n.logger.Debug("Start create empty block")
			batches := n.txpool.GenerateRequestBatch()
			if batches == nil {
				return errors.New("create empty block failed, the length of batches is 0")
			}
			if len(batches) != 1 {
				return fmt.Errorf("create empty block failed, the expect length of batches is 1, but actual is %d", len(batches))
			}

			now := time.Now().UnixNano()
			if n.batchMgr.lastBatchTime != 0 {
				interval := time.Duration(now - n.batchMgr.lastBatchTime).Seconds()
				batchInterval.WithLabelValues("timeout_no_tx").Observe(interval)
				if n.batchMgr.minNoTxTimeoutBatchTime == 0 || interval < n.batchMgr.minNoTxTimeoutBatchTime {
					n.logger.Debugf("update min noTxTimeoutBatch Time[height:%d, interval:%f, lastBatchTime:%v]",
						n.lastExec+1, interval, time.Unix(0, n.batchMgr.lastBatchTime))
					minBatchIntervalDuration.WithLabelValues("timeout_no_tx").Set(interval)
					n.batchMgr.minNoTxTimeoutBatchTime = interval
				}
			}
			n.batchMgr.lastBatchTime = now

			n.postProposal(batches[0])
			if !n.batchMgr.isTimerActive(NoTxBatch) {
				n.batchMgr.startTimer(NoTxBatch)
			}
		}
	case RemoveTx:
		n.batchMgr.stopTimer(RemoveTx)
		n.processNeedRemoveReqs()
		n.batchMgr.startTimer(RemoveTx)
	}
	return nil
}

// processNeedRemoveReqs process the checkPoolRemove timeout requests in requestPool, get the remained reqs from pool,
// then remove these txs in local pool
func (n *Node) processNeedRemoveReqs() {
	n.logger.Info("RemoveTx timer expired, start remove tx in local txpool")
	reqLen, err := n.txpool.RemoveTimeoutRequests()
	if err != nil {
		n.logger.Warningf("Node get the remained reqs failed, error: %v", err)
	}

	if reqLen == 0 {
		n.logger.Infof("Node in normal finds 0 tx to remove")
		return
	}

	if !n.txpool.IsPoolFull() {
		n.setPoolNotFull()
	}
	n.logger.Warningf("Node successful remove %d tx in local txpool ", reqLen)
}

// Schedule to collect txs to the listenReadyBlock channel
func (n *Node) listenReadyBlock() {
	for {
		select {
		case <-n.ctx.Done():
			n.logger.Info("----- Exit listen ready block loop -----")
			return
		case e := <-n.blockCh:
			n.logger.WithFields(logrus.Fields{
				"batch_hash": e.BatchHash,
				"tx_count":   len(e.TxList),
			}).Debugf("Receive proposal from txcache")
			n.logger.Infof("======== Call execute, height=%d", n.lastExec+1)

			block := &types.Block{
				BlockHeader: &types.BlockHeader{
					Epoch:           1,
					Number:          n.lastExec + 1,
					Timestamp:       e.Timestamp / int64(time.Second),
					ProposerAccount: n.config.SelfAccountAddress,
				},
				Transactions: e.TxList,
			}
			localList := make([]bool, len(e.TxList))
			for i := 0; i < len(e.TxList); i++ {
				localList[i] = true
			}
			executeEvent := &common.CommitEvent{
				Block: block,
			}
			n.commitC <- executeEvent
			n.batchDigestM[block.Height()] = e.BatchHash
			n.lastExec++
		}
	}
}

func (n *Node) postProposal(batch *txpool.RequestHashBatch[types.Transaction, *types.Transaction]) {
	n.blockCh <- batch
}

func (n *Node) listenValidTxs() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case requests := <-n.txPreCheck.CommitValidTxs():
			n.postValidTx(requests)
		}
	}
}

func (n *Node) postValidTx(txs *precheck.ValidTxs) {
	n.recvCh <- txs
}

func (n *Node) isPoolFull() bool {
	return atomic.LoadInt32(&n.poolFull) == 1
}

func (n *Node) setPoolNotFull() {
	atomic.StoreInt32(&n.poolFull, 0)
}

func (n *Node) setPoolFull() {
	atomic.StoreInt32(&n.poolFull, 1)
}
