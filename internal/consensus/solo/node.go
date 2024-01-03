package solo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/consensus/precheck"
	"github.com/axiomesh/axiom-ledger/internal/network"
	"github.com/axiomesh/axiom-ledger/pkg/events"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func init() {
	repo.Register(repo.ConsensusTypeSolo, false)
}

type Node struct {
	config           *common.Config
	proposerAccount  string
	isTimed          bool
	commitC          chan *common.CommitEvent                                             // block channel
	logger           logrus.FieldLogger                                                   // logger
	txpool           txpool.TxPool[types.Transaction, *types.Transaction]                 // transaction pool
	batchDigestM     map[uint64]string                                                    // mapping blockHeight to batch digest
	poolFull         int32                                                                // pool full symbol
	recvCh           chan consensusEvent                                                  // receive message from consensus engine
	blockCh          chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction] // receive batch from txpool
	batchMgr         *batchTimerManager
	noTxBatchTimeout time.Duration   // generate no-tx block period
	batchTimeout     time.Duration   // generate block period
	lastExec         uint64          // the index of the last-applied block
	network          network.Network // network manager
	checkpoint       uint64
	txPreCheck       precheck.PreCheck

	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
	txFeed        event.Feed
	mockBlockFeed event.Feed
}

func NewNode(config *common.Config) (*Node, error) {
	currentEpoch, err := config.GetCurrentEpochInfoFromEpochMgrContractFunc()
	if err != nil {
		return nil, err
	}

	proposerAccount, err := repo.KeyToNodeID(config.PrivKey)
	if err != nil {
		return nil, err
	}

	// init batch timer manager
	recvCh := make(chan consensusEvent, maxChanSize)

	ctx, cancel := context.WithCancel(context.Background())
	soloNode := &Node{
		config:           config,
		proposerAccount:  proposerAccount,
		isTimed:          currentEpoch.ConsensusParams.EnableTimedGenEmptyBlock,
		noTxBatchTimeout: config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration(),
		batchTimeout:     config.Config.Solo.BatchTimeout.ToDuration(),
		blockCh:          make(chan *txpool.RequestHashBatch[types.Transaction, *types.Transaction], maxChanSize),
		commitC:          make(chan *common.CommitEvent, maxChanSize),
		batchDigestM:     make(map[uint64]string),
		checkpoint:       config.Config.Solo.CheckpointPeriod,
		poolFull:         0,
		recvCh:           recvCh,
		lastExec:         config.Applied,
		txpool:           config.TxPool,
		network:          config.Network,
		ctx:              ctx,
		cancel:           cancel,
		txPreCheck:       precheck.NewTxPreCheckMgr(ctx, config),
		logger:           config.Logger,
	}
	batchTimerMgr := &batchTimerManager{Timer: timer.NewTimerManager(config.Logger)}

	err = batchTimerMgr.CreateTimer(timer.Batch, config.Config.Solo.BatchTimeout.ToDuration(), soloNode.handleTimeoutEvent)
	if err != nil {
		return nil, err
	}
	err = batchTimerMgr.CreateTimer(timer.NoTxBatch, config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration(), soloNode.handleTimeoutEvent)
	if err != nil {
		return nil, err
	}
	soloNode.batchMgr = batchTimerMgr
	soloNode.logger.Infof("SOLO lastExec = %d", soloNode.lastExec)
	soloNode.logger.Infof("SOLO epoch period = %d", config.GenesisEpochInfo.EpochPeriod)
	soloNode.logger.Infof("SOLO checkpoint period = %d", config.GenesisEpochInfo.ConsensusParams.CheckpointPeriod)
	soloNode.logger.Infof("SOLO enable timed gen empty block = %v", config.GenesisEpochInfo.ConsensusParams.EnableTimedGenEmptyBlock)
	soloNode.logger.Infof("SOLO no-tx batch timeout = %v", config.Config.TimedGenBlock.NoTxBatchTimeout.ToDuration())
	soloNode.logger.Infof("SOLO batch timeout = %v", config.Config.Solo.BatchTimeout.ToDuration())
	soloNode.logger.Infof("SOLO batch size = %d", config.GenesisEpochInfo.ConsensusParams.BlockMaxTxNum)
	soloNode.logger.Infof("SOLO pool size = %d", config.Config.TxPool.PoolSize)
	soloNode.logger.Infof("SOLO tolerance time = %v", config.Config.TxPool.ToleranceTime.ToDuration())
	soloNode.logger.Infof("SOLO tolerance remove time = %v", config.Config.TxPool.ToleranceRemoveTime.ToDuration())
	soloNode.logger.Infof("SOLO tolerance nonce gap = %d", config.Config.TxPool.ToleranceNonceGap)
	return soloNode, nil
}

func (n *Node) GetLowWatermark() uint64 {
	req := &getLowWatermarkReq{
		Resp: make(chan uint64),
	}
	n.postMsg(req)
	return <-req.Resp
}

func (n *Node) Start() error {
	n.txpool.Init(txpool.ConsensusConfig{
		NotifyGenerateBatchFn: n.notifyGenerateBatch,
	})
	err := n.txpool.Start()
	if err != nil {
		return err
	}
	err = n.batchMgr.StartTimer(timer.Batch)
	if err != nil {
		return err
	}

	if n.isTimed {
		err = n.batchMgr.StartTimer(timer.NoTxBatch)
		if err != nil {
			return err
		}
	}
	n.txPreCheck.Start()
	go n.listenEvent()
	go n.listenReadyBlock()
	n.logger.Info("Consensus started")
	return nil
}

func (n *Node) Stop() {
	n.cancel()
	n.logger.Info("Consensus stopped")
}

func (n *Node) Prepare(tx *types.Transaction) error {
	defer n.txFeed.Send([]*types.Transaction{tx})
	if err := n.Ready(); err != nil {
		return fmt.Errorf("node get ready failed: %w", err)
	}
	txWithResp := &common.TxWithResp{
		Tx:      tx,
		CheckCh: make(chan *common.TxResp),
		PoolCh:  make(chan *common.TxResp),
	}
	n.postMsg(txWithResp)
	resp := <-txWithResp.CheckCh
	if !resp.Status {
		return errors.Wrap(common.ErrorPreCheck, resp.ErrorMsg)
	}

	resp = <-txWithResp.PoolCh
	if !resp.Status {
		return errors.Wrap(common.ErrorAddTxPool, resp.ErrorMsg)
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
	if n.txpool.IsPoolFull() {
		return fmt.Errorf(ErrPoolFull)
	}
	return nil
}

func (n *Node) ReportState(height uint64, blockHash *types.Hash, txPointerList []*events.TxPointer, _ *consensus.Checkpoint, _ bool) {
	txHashList := make([]*types.Hash, len(txPointerList))
	lo.ForEach(txPointerList, func(item *events.TxPointer, i int) {
		txHashList[i] = item.Hash
	})
	state := &chainState{
		Height:     height,
		BlockHash:  blockHash,
		TxHashList: txHashList,
	}
	n.postMsg(state)
}

func (n *Node) Quorum() uint64 {
	return 1
}

func (n *Node) SubscribeTxEvent(events chan<- []*types.Transaction) event.Subscription {
	return n.txFeed.Subscribe(events)
}

func (n *Node) SubscribeMockBlockEvent(ch chan<- events.ExecutedEvent) event.Subscription {
	return n.mockBlockFeed.Subscribe(ch)
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
				}

			// receive tx from api
			case *common.TxWithResp:
				unCheckedEv := &common.UncheckedTxEvent{
					EventType: common.LocalTxEvent,
					Event:     e,
				}
				n.txPreCheck.PostUncheckedTxEvent(unCheckedEv)

			// handle timeout event
			case timer.TimeoutEvent:
				if err := n.processBatchTimeout(e); err != nil {
					n.logger.Errorf("Process batch timeout failed: %v", err)
				}
				if err := n.batchMgr.RestartTimer(timer.Batch); err != nil {
					n.logger.Errorf("restart batch timeout failed: %v", err)
				}

				// check if there is no tx in the txpool, start the no tx batch timer
				if n.isTimed && !n.txpool.HasPendingRequestInPool() {
					if !n.batchMgr.IsTimerActive(timer.NoTxBatch) {
						if err := n.batchMgr.StartTimer(timer.NoTxBatch); err != nil {
							n.logger.Errorf("start no-tx batch timeout failed: %v", err)
						}
					}
				}
			case *getLowWatermarkReq:
				e.Resp <- n.lastExec
			case *genBatchReq:
				n.batchMgr.StopTimer(timer.Batch)
				batch, err := n.txpool.GenerateRequestBatch(e.typ)
				if err != nil {
					n.logger.Errorf("Generate batch failed: %v", err)
				} else if batch != nil {
					n.postProposal(batch)
				}
				// start no-tx batch timer when this node handle the last transaction
				if n.isTimed && !n.txpool.HasPendingRequestInPool() {
					if err = n.batchMgr.RestartTimer(timer.NoTxBatch); err != nil {
						n.logger.Errorf("restart no-tx batch timeout failed: %v", err)
					}
				}
				if err = n.batchMgr.RestartTimer(timer.Batch); err != nil {
					n.logger.Errorf("restart batch timeout failed: %v", err)
				}
			}
		}
	}
}

func (n *Node) processBatchTimeout(e timer.TimeoutEvent) error {
	switch e {
	case timer.Batch:
		n.batchMgr.StopTimer(timer.Batch)
		if n.txpool.HasPendingRequestInPool() {
			batch, err := n.txpool.GenerateRequestBatch(txpool.GenBatchTimeoutEvent)
			if err != nil {
				return err
			}
			if batch != nil {
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
				n.postProposal(batch)
				n.logger.Debugf("batch timeout, post proposal: %v", batch)
			}
		}
	case timer.NoTxBatch:
		if n.txpool.HasPendingRequestInPool() {
			n.logger.Debugf("TxPool is not empty, skip handle the no-tx batch timer event")
			return nil
		}
		if !n.isTimed {
			n.batchMgr.StopTimer(timer.NoTxBatch)
			return errors.New("the node is not support the no-tx batch, skip the timer")
		}
		n.batchMgr.StopTimer(timer.NoTxBatch)
		batch, err := n.txpool.GenerateRequestBatch(txpool.GenBatchNoTxTimeoutEvent)
		if err != nil {
			return err
		}
		if batch != nil {
			n.logger.Debug("Start create empty block")
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

			n.postProposal(batch)
			if err = n.batchMgr.RestartTimer(timer.NoTxBatch); err != nil {
				n.logger.Errorf("restart no-tx batch timeout failed: %v", err)
			}
			n.logger.Debugf("batch no-tx timeout, post proposal: %v", batch)
		}
	}
	return nil
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
					ProposerAccount: n.proposerAccount,
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

func (n *Node) notifyGenerateBatch(typ int) {
	req := &genBatchReq{typ: typ}
	n.postMsg(req)
}

func (n *Node) postMsg(ev consensusEvent) {
	n.recvCh <- ev
}

func (n *Node) handleTimeoutEvent(typ timer.TimeoutEvent) {
	switch typ {
	case timer.Batch:
		n.postMsg(timer.Batch)
	case timer.NoTxBatch:
		n.postMsg(timer.NoTxBatch)
	default:
		n.logger.Errorf("receive wrong timeout event type: %s", typ)
	}
}
