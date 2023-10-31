package block_sync

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	network "github.com/axiomesh/axiom-p2p"
)

const (
	wrongTypeSendSyncState = iota
	wrongTypeSendSyncBlockRequest
	wrongTypeSendSyncBlockResponse
	wrongTypeSendStream
)

var (
	lock                  = &sync.RWMutex{}
	mockBlockLedger       = make(map[int]map[uint64]*types.Block)
	mockBlockResponsePipe = &sync.Map{}
	mockBlockRequestPipe  = &sync.Map{}
	mockStateResponseM    = make(map[string]map[uint64]*pb.Message)
)

func clean() {
	lock.Lock()
	defer lock.Unlock()
	mockBlockLedger = make(map[int]map[uint64]*types.Block)
	mockBlockResponsePipe = &sync.Map{}
	mockBlockRequestPipe = &sync.Map{}
	mockStateResponseM = make(map[string]map[uint64]*pb.Message)
}

func getMockStateResponse(id string, height uint64) *pb.Message {
	lock.RLock()
	defer lock.RUnlock()
	if m, ok := mockStateResponseM[id]; ok {
		if msg, ok := m[height]; ok {
			return msg
		}
	}
	return nil
}

func cleanMockStateResponse() {
	lock.Lock()
	defer lock.Unlock()
	mockStateResponseM = make(map[string]map[uint64]*pb.Message)
}

func getMockChainMeta(id int) *types.ChainMeta {
	lock.RLock()
	defer lock.RUnlock()
	if cl, ok := mockBlockLedger[id]; ok {
		// find key:max height,get value
		var max uint64
		for k := range cl {
			if k > max {
				max = k
			}
		}
		if block, ok := cl[max]; ok {
			return &types.ChainMeta{Height: max, BlockHash: block.BlockHash}
		}
	}
	return nil
}

func setMockBlockLedger(block *types.Block, id int) {
	lock.Lock()
	defer lock.Unlock()
	if cl, ok := mockBlockLedger[id]; ok {
		cl[block.Height()] = block
	} else {
		mockBlockLedger[id] = map[uint64]*types.Block{block.Height(): block}
	}
}

func deleteMockBlockLedger(height uint64, id int) {
	lock.Lock()
	defer lock.Unlock()
	if cl, ok := mockBlockLedger[id]; ok {
		delete(cl, height)
	}
}

func getMockBlockLedger(height uint64, id int) (*types.Block, error) {
	lock.RLock()
	defer lock.RUnlock()
	if cl, ok := mockBlockLedger[id]; ok {
		if block, ok := cl[height]; ok {
			return block, nil
		}
	}
	return nil, errors.New("block not found")
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
	}
	return &types.Block{
		BlockHash:    blockHash,
		BlockHeader:  header,
		Transactions: []*types.Transaction{},
	}
}

func ConstructBlocks(start, end uint64) []*types.Block {
	blockList := make([]*types.Block, 0)
	parentHash := getMockChainMeta(1).BlockHash
	for i := start; i <= end; i++ {
		block := ConstructBlock(i, parentHash)
		blockList = append(blockList, block)
		parentHash = block.BlockHash
	}
	return blockList
}

func MockMiniLedger(ctrl *gomock.Controller, id int) *ledger.Ledger {
	chainLedger := mock_ledger.NewMockChainLedger(ctrl)
	stateLedger := mock_ledger.NewMockStateLedger(ctrl)
	chainLedger.EXPECT().GetChainMeta().DoAndReturn(func() *types.ChainMeta {
		return getMockChainMeta(id)
	}).AnyTimes()

	chainLedger.EXPECT().PersistExecutionResult(gomock.Any(), gomock.Any()).DoAndReturn(
		func(block *types.Block, receipts []*types.Receipt) error {
			setMockBlockLedger(block, id)
			return nil
		}).AnyTimes()

	chainLedger.EXPECT().GetBlock(gomock.Any()).DoAndReturn(
		func(height uint64) (*types.Block, error) {
			return getMockBlockLedger(height, id)
		}).AnyTimes()

	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}
	return mockLedger
}

func newMockBlockRequestPipe(ctrl *gomock.Controller, localId string, wrongPipeId ...int) network.Pipe {
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
			msg := &pb.SyncBlockRequest{}
			if err := msg.UnmarshalVT(data); err != nil {
				return fmt.Errorf("unmarshal message failed: %w", err)
			}

			if wrongRemoteId != "" && wrongRemoteId == to {
				return fmt.Errorf("send remote peer err: %s", to)
			}

			ch, loaded := mockBlockRequestPipe.Load(to)
			if !loaded {
				ch = make(chan *network.PipeMsg, 1024)
				mockBlockRequestPipe.Store(to, ch)
			}
			ch.(chan *network.PipeMsg) <- &network.PipeMsg{
				From: localId,
				Data: data,
			}
			return nil
		}).AnyTimes()

	mockPipe.EXPECT().Receive(gomock.Any()).DoAndReturn(
		func(ctx context.Context) *network.PipeMsg {
			ch, _ := mockBlockRequestPipe.Load(localId)
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-ch.(chan *network.PipeMsg):
					return msg
				}
			}
		}).AnyTimes()

	return mockPipe
}

func newMockBlockResponsePipe(ctrl *gomock.Controller, localId string, wrongPid ...int) network.Pipe {
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
		mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("send block response error")).AnyTimes()
	} else {
		mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, to string, data []byte) error {
				msg := &pb.Message{}
				if err := msg.UnmarshalVT(data); err != nil {
					return fmt.Errorf("unmarshal message failed: %w", err)
				}
				if msg.Type != pb.Message_SYNC_BLOCK_RESPONSE {
					return fmt.Errorf("invalid message type: %v", msg.Type)
				}
				ch, _ := mockBlockResponsePipe.Load(to)
				ch.(chan *network.PipeMsg) <- &network.PipeMsg{
					From: localId,
					Data: data,
				}
				return nil
			}).AnyTimes()
	}

	mockPipe.EXPECT().Receive(gomock.Any()).DoAndReturn(
		func(ctx context.Context) *network.PipeMsg {
			ch, _ := mockBlockResponsePipe.Load(localId)
			for {
				select {
				case <-ctx.Done():
					return nil
				case msg := <-ch.(chan *network.PipeMsg):
					return msg
				}
			}
		}).AnyTimes()

	return mockPipe
}

func newMockMiniNetwork(ctrl *gomock.Controller, localId string, wrong ...int) *mock_network.MockNetwork {
	mock := mock_network.NewMockNetwork(ctrl)
	var (
		wrongSendStream        bool
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
			case syncBlockRequestPipe:
				if !wrongSendBlockRequest {
					return newMockBlockRequestPipe(ctrl, localId), nil
				}
				return newMockBlockRequestPipe(ctrl, localId, wrong...), nil
			case syncBlockResponsePipe:
				if !wrongSendBlockResponse {
					return newMockBlockResponsePipe(ctrl, localId), nil
				}
				return newMockBlockResponsePipe(ctrl, localId, wrong...), nil
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
				req := &pb.SyncStateRequest{}
				if err := req.UnmarshalVT(msg.Data); err != nil {
					return nil, fmt.Errorf("unmarshal sync state request failed: %w", err)
				}
				remoteID, err := strconv.Atoi(to)
				if err != nil {
					return nil, fmt.Errorf("invalid remote id: %w", err)
				}
				block, err := getMockBlockLedger(req.Height, remoteID)
				if err != nil {
					return nil, fmt.Errorf("get block with height %d failed: %w", req.Height, err)
				}

				stateResp := &pb.SyncStateResponse{
					CheckpointState: &pb.CheckpointState{
						Height: block.Height(),
						Digest: block.BlockHash.String(),
					},
				}

				data, err := stateResp.MarshalVT()
				if err != nil {
					return nil, fmt.Errorf("marshal sync state response failed: %w", err)
				}
				resp := &pb.Message{From: to, Type: pb.Message_SYNC_STATE_RESPONSE, Data: data}

				return resp, nil
			}).AnyTimes()
	}

	if wrongSendStream {
		mock.EXPECT().SendWithStream(gomock.Any(), gomock.Any()).Return(errors.New("send stream error")).AnyTimes()
	} else {
		mock.EXPECT().SendWithStream(gomock.Any(), gomock.Any()).DoAndReturn(
			func(s network.Stream, msg *pb.Message) error {
				resp := &pb.SyncStateResponse{}
				if err := resp.UnmarshalVT(msg.Data); err != nil {
					return fmt.Errorf("unmarshal sync state response failed: %w", err)
				}
				if _, ok := mockStateResponseM[localId]; !ok {
					mockStateResponseM[localId] = make(map[uint64]*pb.Message)
				}
				mockStateResponseM[localId][resp.CheckpointState.Height] = msg
				return nil
			}).AnyTimes()
	}

	return mock
}

func initLedger() {
	for i := 0; i < 4; i++ {
		mockBlockRequestPipe.Store(strconv.Itoa(i), make(chan *network.PipeMsg, 1024))
		mockBlockResponsePipe.Store(strconv.Itoa(i), make(chan *network.PipeMsg, 1024))
	}

	header := &types.BlockHeader{
		Number:     1,
		ParentHash: types.NewHashByStr("0x00"),
		Timestamp:  time.Now().Unix(),
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

	for i := 0; i < 4; i++ {
		setMockBlockLedger(genesisBlock, i)
	}
}

func newMockBlockSyncs(t *testing.T, n int, wrongPipeId ...int) []*BlockSync {
	initLedger()
	ctrl := gomock.NewController(t)
	syncs := make([]*BlockSync, 0)
	for i := 0; i < n; i++ {
		lg := MockMiniLedger(ctrl, i)
		localId := strconv.Itoa(i)
		mockNetwork := newMockMiniNetwork(ctrl, localId, wrongPipeId...)
		logger := log.NewWithModule("block_sync" + strconv.Itoa(i))

		getBlockFn := func(height uint64) (*types.Block, error) {
			return lg.ChainLedger.GetBlock(height)
		}

		conf := repo.Sync{
			WaitStateTimeout:      repo.Duration(2 * time.Second),
			RequesterRetryTimeout: repo.Duration(1 * time.Second),
			TimeoutCountLimit:     5,
			ConcurrencyLimit:      100,
		}

		blockSync, err := NewBlockSync(logger, getBlockFn, mockNetwork, conf)
		require.Nil(t, err)
		syncs = append(syncs, blockSync)
	}
	return syncs
}

func stopSyncs(syncs []*BlockSync) {
	for _, s := range syncs {
		s.Stop()
	}
	clean()
}

func prepareLedger(n int, endBlockHeight uint64) {
	blocks := ConstructBlocks(2, endBlockHeight)
	for i := 0; i < n; i++ {
		if i == 0 {
			continue
		}
		storeBlocks(blocks, i)
	}
}

func storeBlocks(blocks []*types.Block, id int) {
	for _, block := range blocks {
		setMockBlockLedger(block, id)
	}
}
