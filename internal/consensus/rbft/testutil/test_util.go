package testutil

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-bft/common/consensus"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/block_sync/mock_block_sync"
	"github.com/axiomesh/axiom-ledger/internal/consensus/common"
	"github.com/axiomesh/axiom-ledger/internal/network/mock_network"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

var (
	mockBlockLedger      = make(map[uint64]*types.Block)
	mockLocalBlockLedger = make(map[uint64]*types.Block)
	mockChainMeta        *types.ChainMeta
	blockCacheChan       = make(chan []*types.Block, 1024)
)

func SetMockChainMeta(chainMeta *types.ChainMeta) {
	mockChainMeta = chainMeta
}

func ResetMockChainMeta() {
	block := ConstructBlock("block1", uint64(1))
	mockChainMeta = &types.ChainMeta{Height: uint64(1), BlockHash: block.BlockHash}
}

func SetMockBlockLedger(block *types.Block, local bool) {
	if local {
		mockLocalBlockLedger[block.Height()] = block
	} else {
		mockBlockLedger[block.Height()] = block
	}
}

func getRemoteMockBlockLedger(height uint64) (*types.Block, error) {
	if block, ok := mockBlockLedger[height]; ok {
		return block, nil
	}
	return nil, errors.New("block not found")
}

func ResetMockBlockLedger() {
	mockBlockLedger = make(map[uint64]*types.Block)
	mockLocalBlockLedger = make(map[uint64]*types.Block)
}

func ConstructBlock(blockHashStr string, height uint64) *types.Block {
	from := make([]byte, 0)
	strLen := len(blockHashStr)
	for i := 0; i < 32; i++ {
		from = append(from, blockHashStr[i%strLen])
	}
	fromStr := hex.EncodeToString(from)
	blockHash := types.NewHashByStr(fromStr)
	header := &types.BlockHeader{
		Number:     height,
		ParentHash: blockHash,
		Timestamp:  time.Now().Unix(),
	}
	return &types.Block{
		BlockHash:    blockHash,
		BlockHeader:  header,
		Transactions: []*types.Transaction{},
	}
}

func MockMiniNetwork(ctrl *gomock.Controller, selfAddr string) *mock_network.MockNetwork {
	mock := mock_network.NewMockNetwork(ctrl)
	mockPipe := mock_network.NewMockPipe(ctrl)
	mockPipe.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockPipe.EXPECT().Broadcast(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockPipe.EXPECT().Receive(gomock.Any()).Return(nil).AnyTimes()

	mock.EXPECT().CreatePipe(gomock.Any(), gomock.Any()).Return(mockPipe, nil).AnyTimes()

	N := 3
	f := (N - 1) / 3
	mock.EXPECT().CountConnectedPeers().Return(uint64((N + f + 2) / 2)).AnyTimes()
	mock.EXPECT().PeerID().Return(selfAddr).AnyTimes()
	return mock
}

func MockMiniBlockSync(ctrl *gomock.Controller) *mock_block_sync.MockSync {
	blockCacheChan = make(chan []*types.Block, 1024)
	mock := mock_block_sync.NewMockSync(ctrl)
	mock.EXPECT().StartSync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(peers []string, curBlockHash string, quorum, startHeight, targetHeight uint64, quorumCheckpoint *consensus.SignedCheckpoint, epc ...*consensus.EpochChange) error {
			blockCache := make([]*types.Block, 0)
			for i := startHeight; i <= targetHeight; i++ {
				block, err := getRemoteMockBlockLedger(i)
				if err != nil {
					return err
				}
				blockCache = append(blockCache, block)
			}
			blockCacheChan <- blockCache
			return nil
		}).AnyTimes()

	mock.EXPECT().Commit().Return(blockCacheChan).AnyTimes()
	mock.EXPECT().StopSync().Return(nil).AnyTimes()
	return mock
}

func MockConsensusConfig(logger logrus.FieldLogger, ctrl *gomock.Controller, t *testing.T) *common.Config {
	s, err := types.GenerateSigner()
	assert.Nil(t, err)

	genesisEpochInfo := repo.GenesisEpochInfo(false)
	conf := &common.Config{
		RepoRoot:           t.TempDir(),
		Config:             repo.DefaultConsensusConfig(),
		Logger:             logger,
		ConsensusType:      "",
		PrivKey:            s.Sk,
		SelfAccountAddress: genesisEpochInfo.ValidatorSet[0].AccountAddress,
		GenesisEpochInfo:   genesisEpochInfo,
		Applied:            0,
		Digest:             "",
		GetEpochInfoFromEpochMgrContractFunc: func(epoch uint64) (*rbft.EpochInfo, error) {
			return genesisEpochInfo, nil
		},
		GetChainMetaFunc: GetChainMetaFunc,
		GetBlockFunc: func(height uint64) (*types.Block, error) {
			if block, ok := mockLocalBlockLedger[height]; ok {
				return block, nil
			} else {
				return nil, errors.New("block not found")
			}
		},
		GetAccountNonce: func(address *types.Address) uint64 {
			return 0
		},
		GetCurrentEpochInfoFromEpochMgrContractFunc: func() (*rbft.EpochInfo, error) {
			return genesisEpochInfo, nil
		},
	}

	mockNetwork := MockMiniNetwork(ctrl, conf.SelfAccountAddress)
	conf.Network = mockNetwork

	mockBlockSync := MockMiniBlockSync(ctrl)
	conf.BlockSync = mockBlockSync

	return conf
}

func GetChainMetaFunc() *types.ChainMeta {
	if mockChainMeta == nil {
		ResetMockChainMeta()
	}
	return mockChainMeta
}

func ConstructBlocks(height uint64, num int) []*types.Block {
	blockHashStr := fmt.Sprintf("block%d", height)
	blocks := make([]*types.Block, 0)
	for i := 0; i < num; i++ {
		blocks = append(blocks, ConstructBlock(blockHashStr, height+uint64(i)))
	}
	return blocks
}
