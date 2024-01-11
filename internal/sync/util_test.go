package sync

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/internal/sync/common"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	network "github.com/axiomesh/axiom-p2p"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	wrongTypeSendSyncState = iota
	wrongTypeSendSyncBlockRequest
	wrongTypeSendSyncBlockResponse
	wrongTypeSendStream
	latencyTypeSendState
)

const (
	epochPrefix = "epoch." + rbft.EpochStatePrefix
	epochPeriod = 100
)

var (
	blockCache  = make([]*types.Block, 0)
	gensisBlock = genGenesisBlock()
)

type mockLedger struct {
	*mock_ledger.MockChainLedger
	blockDb      map[uint64]*types.Block
	chainMeta    *types.ChainMeta
	receiptsDb   map[uint64][]*types.Receipt
	epochStateDb map[string]*consensus.QuorumCheckpoint

	stateResponse      chan *pb.Message
	epochStateResponse chan *pb.Message
}

func genReceipts(block *types.Block) []*types.Receipt {
	receipts := make([]*types.Receipt, 0)
	for _, tx := range block.Transactions {
		receipts = append(receipts, &types.Receipt{
			Status: types.ReceiptSUCCESS,
			TxHash: tx.GetHash(),
		})
	}
	return receipts
}

func newMockMinLedger(t *testing.T, genesisBlock *types.Block) *mockLedger {
	genesis := genesisBlock.Clone()
	mockLg := &mockLedger{
		blockDb: map[uint64]*types.Block{
			genesis.Height(): genesis,
		},
		receiptsDb:         make(map[uint64][]*types.Receipt),
		epochStateDb:       make(map[string]*consensus.QuorumCheckpoint),
		stateResponse:      make(chan *pb.Message, 1),
		epochStateResponse: make(chan *pb.Message, 1),
		chainMeta: &types.ChainMeta{
			Height:    genesis.Height(),
			BlockHash: genesis.BlockHash,
		},
	}
	ctrl := gomock.NewController(t)
	mockLg.MockChainLedger = mock_ledger.NewMockChainLedger(ctrl)

	mockLg.EXPECT().GetBlock(gomock.Any()).DoAndReturn(func(height uint64) (*types.Block, error) {
		if mockLg.blockDb[height] == nil {
			return nil, errors.New("block not found")
		}
		return mockLg.blockDb[height], nil
	}).AnyTimes()

	mockLg.EXPECT().GetChainMeta().DoAndReturn(func() *types.ChainMeta {
		return mockLg.chainMeta
	}).AnyTimes()

	mockLg.EXPECT().GetReceiptsByHeight(gomock.Any()).DoAndReturn(func(height uint64) ([]*types.Receipt, error) {
		if mockLg.receiptsDb[height] == nil {
			return nil, fmt.Errorf("receipts not found:[height:%d]", height)
		}
		return mockLg.receiptsDb[height], nil
	}).AnyTimes()

	mockLg.EXPECT().PersistExecutionResult(gomock.Any(), gomock.Any()).DoAndReturn(func(block *types.Block, receipts []*types.Receipt) error {
		h := block.Height()
		mockLg.blockDb[h] = block
		if mockLg.chainMeta.Height <= h {
			mockLg.chainMeta.Height = h
			mockLg.chainMeta.BlockHash = block.BlockHash
		}
		mockLg.receiptsDb[h] = receipts

		if h%epochPeriod == 0 {
			epoch := h / epochPeriod
			key := fmt.Sprintf(epochPrefix+"%d", epoch)
			mockLg.epochStateDb[key] = &consensus.QuorumCheckpoint{
				Checkpoint: &consensus.Checkpoint{
					Epoch: h / epochPeriod,
					ExecuteState: &consensus.Checkpoint_ExecuteState{
						Height: h,
						Digest: block.BlockHash.String(),
					},
				},
			}
		}
		return nil
	}).AnyTimes()
	return mockLg
}

func ConstructBlock(height uint64, parentHash *types.Hash) *types.Block {
	blockHashStr := "block" + strconv.FormatUint(height, 10)
	from := make([]byte, 0)
	strLen := len(blockHashStr)
	for i := 0; i < 32; i++ {
		from = append(from, blockHashStr[i%strLen])
	}
	fromStr := hex.EncodeToString(from)
	blockHash := types.NewHashByStr(fromStr)
	header := &types.BlockHeader{
		Number:     height,
		ParentHash: parentHash,
		Timestamp:  time.Now().Unix(),
		Epoch:      ((height - 1) / epochPeriod) + 1,
	}
	return &types.Block{
		BlockHash:    blockHash,
		BlockHeader:  header,
		Transactions: []*types.Transaction{},
	}
}

func ConstructBlocks(count int, parentHash *types.Hash) []*types.Block {
	blocks := make([]*types.Block, 0)
	parent := parentHash
	for i := 2; i <= count; i++ {
		blocks = append(blocks, ConstructBlock(uint64(i), parent))
		parent = blocks[len(blocks)-1].BlockHash
	}
	return blocks
}

func ConstructBlockWithTxs(t *testing.T, height uint64, parentHash *types.Hash, txCount uint64) *types.Block {
	blockHashStr := "commitData" + strconv.FormatUint(height, 10)
	from := make([]byte, 0)
	strLen := len(blockHashStr)
	for i := 0; i < 32; i++ {
		from = append(from, blockHashStr[i%strLen])
	}
	fromStr := hex.EncodeToString(from)
	blockHash := types.NewHashByStr(fromStr)
	header := &types.BlockHeader{
		Number:     height,
		ParentHash: parentHash,
		Timestamp:  time.Now().Unix(),
	}
	txs := make([]*types.Transaction, 0)

	for i := uint64(0); i < txCount; i++ {
		tx, err := types.GenerateEmptyTransactionAndSigner()
		require.Nil(t, err)
		txs = append(txs, tx)
	}
	return &types.Block{
		BlockHash:    blockHash,
		BlockHeader:  header,
		Transactions: txs,
	}
}
func (net *mockMiniNetwork) newMockBlockRequestPipe(nets map[string]*mockMiniNetwork, mode common.SyncMode, ctrl *gomock.Controller, localId string, wrongPipeId ...int) network.Pipe {
	var wrongRemoteId string
	if len(wrongPipeId) > 0 {
		if len(wrongPipeId) != 2 {
			panic("wrong pipe id must be 2, local id + remote id")
		}
		if strconv.Itoa(wrongPipeId[0]) == localId {
			wrongRemoteId = strconv.Itoa(wrongPipeId[1])
		}
	}
	mockPipe := mock_network.NewMockPipe(ctrl)
	mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, to string, data []byte) error {
			switch mode {
			case common.SyncModeFull:
				msg := &pb.SyncBlockRequest{}
				if err := msg.UnmarshalVT(data); err != nil {
					return fmt.Errorf("unmarshal message failed: %w", err)
				}

				if wrongRemoteId != "" && wrongRemoteId == to {
					return fmt.Errorf("send remote peer err: %s", to)
				}

				nets[to].blockReqPipeDb <- &network.PipeMsg{
					From: localId,
					Data: data,
				}
			case common.SyncModeSnapshot:
				msg := &pb.SyncChainDataRequest{}
				if err := msg.UnmarshalVT(data); err != nil {
					return fmt.Errorf("unmarshal message failed: %w", err)
				}

				if wrongRemoteId != "" && wrongRemoteId == to {
					return fmt.Errorf("send remote peer err: %s", to)
				}

				nets[to].chainDataReqPipeDb <- &network.PipeMsg{
					From: localId,
					Data: data,
				}
			}
			return nil
		}).AnyTimes()

	var ch chan *network.PipeMsg
	switch mode {
	case common.SyncModeFull:
		ch = net.blockReqPipeDb
	case common.SyncModeSnapshot:
		ch = net.chainDataReqPipeDb
	}
	mockPipe.EXPECT().Receive(gomock.Any()).DoAndReturn(
		func(ctx context.Context) *network.PipeMsg {
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-ch:
					return msg
				}
			}
		}).AnyTimes()

	return mockPipe
}

func (net *mockMiniNetwork) newMockBlockResponsePipe(nets map[string]*mockMiniNetwork, mode common.SyncMode, ctrl *gomock.Controller, localId string, wrongPid ...int) network.Pipe {
	var wrongSendBlockResponse bool
	if len(wrongPid) > 0 {
		if len(wrongPid) != 2 {
			panic("wrong pipe id must be 2, local id + remote id")
		}
		if strconv.Itoa(wrongPid[1]) == localId {
			wrongSendBlockResponse = true
		}
	}
	mockPipe := mock_network.NewMockPipe(ctrl)
	if wrongSendBlockResponse {
		mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("send commitDataCache response error")).AnyTimes()
	} else {
		mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, to string, data []byte) error {
				msg := &pb.Message{}
				if err := msg.UnmarshalVT(data); err != nil {
					return fmt.Errorf("unmarshal message failed: %w", err)
				}
				resp := &network.PipeMsg{
					From: localId,
					Data: data,
				}
				switch msg.Type {
				case pb.Message_SYNC_BLOCK_RESPONSE:
					nets[to].blockRespPipeDb <- resp
				case pb.Message_SYNC_CHAIN_DATA_RESPONSE:
					nets[to].chainDataRespPipeDb <- resp
				default:
					return fmt.Errorf("invalid message type: %v", msg.Type)
				}
				return nil
			}).AnyTimes()
	}

	var ch chan *network.PipeMsg
	switch mode {
	case common.SyncModeFull:
		ch = net.blockRespPipeDb
	case common.SyncModeSnapshot:
		ch = net.chainDataRespPipeDb
	}
	mockPipe.EXPECT().Receive(gomock.Any()).DoAndReturn(
		func(ctx context.Context) *network.PipeMsg {
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-ch:
					return msg
				}
			}
		}).AnyTimes()

	return mockPipe
}

type pipeMsgIndexM struct {
	lock           sync.RWMutex
	blockReqCache  map[string]chan *network.PipeMsg
	blockRespCache map[string]chan *network.PipeMsg

	chainDataReqCache  map[string]chan *network.PipeMsg
	chainDataRespCache map[string]chan *network.PipeMsg
}

type mockMiniNetwork struct {
	*mock_network.MockNetwork
	blockReqPipeDb      chan *network.PipeMsg
	blockRespPipeDb     chan *network.PipeMsg
	chainDataReqPipeDb  chan *network.PipeMsg
	chainDataRespPipeDb chan *network.PipeMsg
}

func newMockMiniNetworks(n int) map[string]*mockMiniNetwork {
	nets := make(map[string]*mockMiniNetwork)
	for i := 0; i < n; i++ {
		mock := &mockMiniNetwork{
			blockReqPipeDb:      make(chan *network.PipeMsg, 1000),
			blockRespPipeDb:     make(chan *network.PipeMsg, 1000),
			chainDataReqPipeDb:  make(chan *network.PipeMsg, 1000),
			chainDataRespPipeDb: make(chan *network.PipeMsg, 1000),
		}
		nets[strconv.Itoa(i)] = mock
	}
	return nets
}

func initMockMiniNetwork(t *testing.T, ledgers map[string]*mockLedger, nets map[string]*mockMiniNetwork, ctrl *gomock.Controller, localId string, wrong ...int) {
	mock := nets[localId]
	mock.MockNetwork = mock_network.NewMockNetwork(ctrl)
	var (
		wrongSendStream        bool
		latencySendState       bool
		latencyRemoteId        int
		wrongSendStateRequest  bool
		wrongSendBlockRequest  bool
		wrongSendBlockResponse bool
	)

	if len(wrong) > 0 {
		if len(wrong) != 3 {
			panic("wrong pipe id must be 3, wrong type + local id + remote id")
		}
		switch wrong[0] {
		case wrongTypeSendSyncState:
			if strconv.Itoa(wrong[1]) == localId {
				wrongSendStateRequest = true
			}
		case wrongTypeSendStream:
			if strconv.Itoa(wrong[2]) == localId {
				wrongSendStream = true
			}
		case latencyTypeSendState:
			if strconv.Itoa(wrong[1]) == localId {
				latencySendState = true
				latencyRemoteId = wrong[2]
			}
		case wrongTypeSendSyncBlockRequest:
			wrong = wrong[1:]
			wrongSendBlockRequest = true
		case wrongTypeSendSyncBlockResponse:
			wrong = wrong[1:]
			wrongSendBlockResponse = true
		}
	}

	mock.EXPECT().RegisterMsgHandler(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mock.EXPECT().CreatePipe(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, pipeID string) (network.Pipe, error) {
			switch pipeID {
			case common.SyncBlockRequestPipe, common.SyncChainDataRequestPipe:
				mode := common.SyncModeFull
				if pipeID == common.SyncChainDataRequestPipe {
					mode = common.SyncModeSnapshot
				}
				if !wrongSendBlockRequest {
					return mock.newMockBlockRequestPipe(nets, mode, ctrl, localId), nil
				}
				return mock.newMockBlockRequestPipe(nets, mode, ctrl, localId, wrong...), nil
			case common.SyncBlockResponsePipe, common.SyncChainDataResponsePipe:
				mode := common.SyncModeFull
				if pipeID == common.SyncChainDataResponsePipe {
					mode = common.SyncModeSnapshot
				}
				if !wrongSendBlockResponse {
					return mock.newMockBlockResponsePipe(nets, mode, ctrl, localId), nil
				}
				return mock.newMockBlockResponsePipe(nets, mode, ctrl, localId, wrong...), nil
			default:
				return nil, fmt.Errorf("invalid pipe id: %s", pipeID)
			}
		}).AnyTimes()

	mock.EXPECT().PeerID().Return(localId).AnyTimes()

	if wrongSendStateRequest {
		mock.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil, errors.New("send error")).AnyTimes()
	} else {
		mock.EXPECT().Send(gomock.Any(), gomock.Any()).DoAndReturn(
			func(to string, msg *pb.Message) (*pb.Message, error) {
				if msg.Type == pb.Message_FETCH_EPOCH_STATE_REQUEST {
					req := &pb.FetchEpochStateRequest{}
					if err := req.UnmarshalVT(msg.Data); err != nil {
						return nil, fmt.Errorf("unmarshal fetch epoch state request failed: %w", err)
					}

					key := fmt.Sprintf(epochPrefix+"%d", req.Epoch)
					eps, ok := ledgers[to].epochStateDb[key]
					if !ok {
						return nil, fmt.Errorf("epoch %d not found", req.Epoch)
					}

					v, err := eps.MarshalVT()
					require.Nil(t, err)
					epsResp := &pb.FetchEpochStateResponse{
						Data: v,
					}
					data, err := epsResp.MarshalVT()
					require.Nil(t, err)
					resp := &pb.Message{From: to, Type: pb.Message_FETCH_EPOCH_STATE_RESPONSE, Data: data}
					return resp, nil
				}

				req := &pb.SyncStateRequest{}
				if err := req.UnmarshalVT(msg.Data); err != nil {
					return nil, fmt.Errorf("unmarshal sync state request failed: %w", err)
				}

				block, err := ledgers[to].GetBlock(req.Height)
				if err != nil {
					return nil, fmt.Errorf("get block with height %d failed: %w", req.Height, err)
				}

				stateResp := &pb.SyncStateResponse{
					Status: pb.Status_SUCCESS,
					CheckpointState: &pb.CheckpointState{
						Height:       block.Height(),
						Digest:       block.BlockHash.String(),
						LatestHeight: ledgers[to].GetChainMeta().Height,
					},
				}

				data, err := stateResp.MarshalVT()
				if err != nil {
					return nil, fmt.Errorf("marshal sync state response failed: %w", err)
				}
				resp := &pb.Message{From: to, Type: pb.Message_SYNC_STATE_RESPONSE, Data: data}

				if latencySendState && to == strconv.Itoa(latencyRemoteId) {
					time.Sleep(200 * time.Millisecond)
				}
				return resp, nil
			}).AnyTimes()
	}

	mock.EXPECT().SendWithStream(gomock.Any(), gomock.Any()).DoAndReturn(func(stream network.Stream, msg *pb.Message) error {
		if wrongSendStream {
			return errors.New("send stream error")
		}
		if msg.Type == pb.Message_FETCH_EPOCH_STATE_RESPONSE {
			resp := &pb.FetchEpochStateResponse{}
			if err := resp.UnmarshalVT(msg.Data); err != nil {
				return fmt.Errorf("unmarshal fetch epoch state response failed: %w", err)
			}
			ledgers[msg.From].epochStateResponse <- msg
			return nil
		}

		resp := &pb.SyncStateResponse{}
		if err := resp.UnmarshalVT(msg.Data); err != nil {
			return fmt.Errorf("unmarshal sync state response failed: %w", err)
		}
		ledgers[msg.From].stateResponse <- msg
		return nil
	}).AnyTimes()
}

func genGenesisBlock() *types.Block {
	header := &types.BlockHeader{
		Number:      1,
		ReceiptRoot: &types.Hash{},
		StateRoot:   &types.Hash{},
		TxRoot:      &types.Hash{},
		Bloom:       new(types.Bloom),
		ParentHash:  types.NewHashByStr("0x00"),
		Timestamp:   time.Now().Unix(),
	}

	hash := make([]byte, 0)
	blockHashStr := "genesis_block"
	strLen := len(blockHashStr)
	for i := 0; i < 32; i++ {
		hash = append(hash, blockHashStr[i%strLen])
	}
	genesisBlock := &types.Block{
		BlockHash:    types.NewHashByStr(hex.EncodeToString(hash)),
		BlockHeader:  header,
		Transactions: []*types.Transaction{},
	}
	return genesisBlock
}

func newMockBlockSyncs(t *testing.T, n int, wrongPipeId ...int) ([]*SyncManager, map[string]*mockLedger) {
	ctrl := gomock.NewController(t)
	syncs := make([]*SyncManager, 0)
	genesis := gensisBlock.Clone()

	ledgers := make(map[string]*mockLedger)
	// 1. prepare all ledgers
	for i := 0; i < n; i++ {
		lg := newMockMinLedger(t, genesis)
		localId := strconv.Itoa(i)
		ledgers[localId] = lg
	}
	nets := newMockMiniNetworks(n)

	// 2. prepare all syncs
	for i := 0; i < n; i++ {
		localId := strconv.Itoa(i)
		logger := log.NewWithModule("sync" + strconv.Itoa(i))
		logger.Logger.SetLevel(logrus.DebugLevel)
		initMockMiniNetwork(t, ledgers, nets, ctrl, localId, wrongPipeId...)

		getBlockFn := func(height uint64) (*types.Block, error) {
			return ledgers[localId].GetBlock(height)
		}

		getChainMetaFn := func() *types.ChainMeta {
			return ledgers[localId].GetChainMeta()
		}

		conf := repo.Sync{
			RequesterRetryTimeout: repo.Duration(1 * time.Second),
			TimeoutCountLimit:     5,
			ConcurrencyLimit:      100,
			WaitStatesTimeout:     repo.Duration(100 * time.Second),
		}

		getReceiptsFn := func(height uint64) ([]*types.Receipt, error) {
			return ledgers[localId].GetReceiptsByHeight(height)
		}

		getEpochStateFn := func(key []byte) []byte {
			epochState := ledgers[localId].epochStateDb[string(key)]
			val, err := epochState.MarshalVT()
			require.Nil(t, err)
			return val
		}

		blockSync, err := NewSyncManager(logger, getChainMetaFn, getBlockFn, getReceiptsFn, getEpochStateFn, nets[strconv.Itoa(i)], conf)
		require.Nil(t, err)
		syncs = append(syncs, blockSync)
	}

	return syncs, ledgers
}

func stopSyncs(syncs []*SyncManager) {
	for _, s := range syncs {
		s.Stop()
	}
}

func prepareBlockSyncs(t *testing.T, epochInterval int, local, count int, begin, end uint64) ([]*SyncManager, []*consensus.EpochChange) {
	ctrl := gomock.NewController(t)
	syncs := make([]*SyncManager, 0)
	var (
		epochChanges       []*consensus.EpochChange
		needGenEpochChange bool
		ledgers            = make(map[string]*mockLedger)
	)

	remainder := (int(begin) + epochInterval) % epochInterval
	// prepare epoch change
	if begin+uint64(epochInterval)-uint64(remainder) < end {
		needGenEpochChange = true
	}

	gensis := gensisBlock.Clone()
	for i := 0; i < count; i++ {
		lg := newMockMinLedger(t, gensis)
		localId := strconv.Itoa(i)

		latestBlock := ConstructBlock(begin-1, types.NewHashByStr(fmt.Sprintf("blcok%d", begin-2)))
		err := lg.PersistExecutionResult(latestBlock, []*types.Receipt{})
		require.Nil(t, err)

		if i != local {
			if len(blockCache) == 0 {
				parentHash := lg.chainMeta.BlockHash
				for j := begin; j <= end; j++ {
					block := ConstructBlock(j, parentHash)
					blockCache = append(blockCache, block)
					parentHash = block.BlockHash
				}
			}
			lo.ForEach(blockCache, func(block *types.Block, _ int) {
				err = lg.PersistExecutionResult(block, genReceipts(block))
				require.Nil(t, err)
			})

			if needGenEpochChange && len(epochChanges) == 0 {
				nextEpochHeight := begin + uint64(epochInterval) - uint64(remainder)
				nextEpochHash := lg.blockDb[nextEpochHeight].BlockHash.String()
				for nextEpochHeight <= end {
					nextEpochHash = lg.blockDb[nextEpochHeight].BlockHash.String()
					epochChanges = append(epochChanges, prepareEpochChange(nextEpochHeight, nextEpochHash))
					nextEpochHeight += uint64(epochInterval)
				}
			}
		}
		ledgers[localId] = lg
	}
	nets := newMockMiniNetworks(count)
	for localId := range ledgers {
		lg := ledgers[localId]
		fmt.Printf("%s: %T", localId, lg)
		logger := log.NewWithModule("sync" + localId)

		getChainMetaFn := func() *types.ChainMeta {
			return lg.GetChainMeta()
		}
		getBlockFn := func(height uint64) (*types.Block, error) {
			return lg.GetBlock(height)
		}

		getReceiptsFn := func(height uint64) ([]*types.Receipt, error) {
			return lg.GetReceiptsByHeight(height)
		}

		getEpochStateFn := func(key []byte) []byte {
			epochState := lg.epochStateDb[string(key)]
			val, err := epochState.MarshalVT()
			require.Nil(t, err)
			return val
		}
		conf := repo.Sync{
			RequesterRetryTimeout: repo.Duration(1 * time.Second),
			TimeoutCountLimit:     5,
			ConcurrencyLimit:      100,
		}

		initMockMiniNetwork(t, ledgers, nets, ctrl, localId)

		blockSync, err := NewSyncManager(logger, getChainMetaFn, getBlockFn, getReceiptsFn, getEpochStateFn, nets[localId], conf)
		require.Nil(t, err)
		syncs = append(syncs, blockSync)
	}

	return syncs, epochChanges
}

func prepareEpochChange(height uint64, hash string) *consensus.EpochChange {
	return &consensus.EpochChange{
		Checkpoint: &consensus.QuorumCheckpoint{
			Checkpoint: &consensus.Checkpoint{
				ExecuteState: &consensus.Checkpoint_ExecuteState{
					Height: height,
					Digest: hash,
				},
			},
		},
	}
}

func genSyncParams(peers []string, latestBlockHash string, quorum uint64, curHeight, targetHeight uint64,
	quorumCkpt *consensus.SignedCheckpoint, epochChanges ...*consensus.EpochChange) *common.SyncParams {
	return &common.SyncParams{
		Peers:            peers,
		LatestBlockHash:  latestBlockHash,
		Quorum:           quorum,
		CurHeight:        curHeight,
		TargetHeight:     targetHeight,
		QuorumCheckpoint: quorumCkpt,
		EpochChanges:     epochChanges,
	}
}

func prepareLedger(t *testing.T, ledgers map[string]*mockLedger, exceptId string, endHeight int) {
	blocks := ConstructBlocks(endHeight, gensisBlock.BlockHash)
	for id, lg := range ledgers {
		if id == exceptId {
			continue
		}
		lo.ForEach(blocks, func(block *types.Block, _ int) {
			err := lg.PersistExecutionResult(block, genReceipts(block))
			require.Nil(t, err)
		})
	}
}
