package network

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	rbft "github.com/axiomesh/axiom-bft"
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/base"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	p2p "github.com/axiomesh/axiom-p2p"
)

func TestSwarm_OtherPeers(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}
}

func TestVersionCheck(t *testing.T) {
	peerCnt := 4
	// pass true to change the last Node's version
	swarms := newMockSwarms(t, peerCnt, true)
	defer stopSwarms(t, swarms)
	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}

	msg := &pb.Message{Type: pb.Message_SYNC_STATE_REQUEST, Data: []byte(strconv.Itoa(1))}
	_, err := swarms[0].Send(swarms[1].PeerID(), msg)
	assert.NotNil(t, err, "swarms[1] must register msg handler")

	err = swarms[1].RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, func(stream p2p.Stream, msg *pb.Message) {
		resp := &pb.Message{
			Type: pb.Message_SYNC_STATE_RESPONSE,
			Data: []byte("response aaa"),
		}
		data, err := resp.MarshalVT()
		require.Nil(t, err)
		err = stream.AsyncSend(data)
		require.Nil(t, err)
	})
	_, err = swarms[0].Send(swarms[1].PeerID(), msg)
	assert.Nil(t, err)

	_, err = swarms[0].Send(swarms[peerCnt-1].PeerID(), msg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "protocols not supported",
		"err should be protocols not supported")
}

func TestSwarm_OnConnected(t *testing.T) {
	config := generateMockConfig(t)
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	// mock data for ledger
	chainMeta := &types.ChainMeta{
		Height:    1,
		BlockHash: types.NewHashByStr("0x3f9d18f7c3a6e5e4c0b877fe3e688ab08840b997"),
	}
	chainLedger.EXPECT().GetChainMeta().Return(chainMeta).AnyTimes()

	jsonBytes, err := json.Marshal(config.Genesis.EpochInfo)
	assert.Nil(t, err)

	stateLedger.EXPECT().GetState(gomock.Any(), gomock.Any()).DoAndReturn(func(addr *types.Address, key []byte) (bool, []byte) {
		return true, jsonBytes
	}).AnyTimes()

	stateLedger.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(addr *types.Address, key []byte, value []byte) {},
	).AnyTimes()

	mockLedger.StateLedger.SetState(types.NewAddressByStr(common.NodeManagerContractAddr), []byte(common.NodeManagerContractAddr), jsonBytes)

	var peerID = "16Uiu2HAmRypzJbdbUNYsCV2VVgv9UryYS5d7wejTJXT73mNLJ8AK"

	success, data := mockLedger.StateLedger.GetState(types.NewAddressByStr(common.NodeManagerContractAddr), []byte(common.NodeManagerContractAddr))
	if success {
		stringData := strings.Split(string(data), ",")
		for _, nodeID := range stringData {
			if peerID == nodeID {
				t.Log("exist nodeMembernodeID: " + nodeID)
				break
			}
		}
	} else {
		t.Log("get nodeMember err")
	}
}

func generateMockConfig(t *testing.T) *repo.Config {
	r, err := repo.Default(t.TempDir())
	assert.Nil(t, err)
	config := r.Config

	for i := 0; i < 4; i++ {
		config.Genesis.Admins = append(config.Genesis.Admins, &repo.Admin{
			Address: types.NewAddress([]byte{byte(1)}).String(),
		})
	}

	return config
}

func getAddr(p2p p2p.Network) (peer.AddrInfo, error) {
	realAddr := fmt.Sprintf("%s/p2p/%s", p2p.LocalAddr(), p2p.PeerID())
	multiaddr, err := ma.NewMultiaddr(realAddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	addrInfo, err := peer.AddrInfoFromP2pAddr(multiaddr)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return *addrInfo, nil
}

func newMockSwarms(t *testing.T, peerCnt int, versionChange bool) []*networkImpl {
	var swarms []*networkImpl
	mockCtl := gomock.NewController(t)
	chainLedger := mock_ledger.NewMockChainLedger(mockCtl)
	stateLedger := mock_ledger.NewMockStateLedger(mockCtl)
	mockLedger := &ledger.Ledger{
		ChainLedger: chainLedger,
		StateLedger: stateLedger,
	}

	chainLedger.EXPECT().GetBlock(gomock.Any()).Return(&types.Block{
		BlockHeader: &types.BlockHeader{
			Number: 1,
		},
	}, nil).AnyTimes()

	chainLedger.EXPECT().GetBlockSign(gomock.Any()).Return([]byte("sign"), nil).AnyTimes()
	chainLedger.EXPECT().GetTransaction(gomock.Any()).Return(&types.Transaction{}, nil).AnyTimes()
	stateLedger.EXPECT().NewView().Return(stateLedger).AnyTimes()
	stateLedger.EXPECT().GetState(gomock.Any(), gomock.Any()).DoAndReturn(func(addr *types.Address, key []byte) (bool, []byte) {
		return false, nil
	}).AnyTimes()

	accountCache, err := ledger.NewAccountCache()
	assert.Nil(t, err)
	repoRoot := t.TempDir()
	ld, err := leveldb.New(filepath.Join(repoRoot, "network"), nil)
	assert.Nil(t, err)
	account := ledger.NewAccount(ld, accountCache, types.NewAddressByStr(common.EpochManagerContractAddr), ledger.NewChanger())
	stateLedger.EXPECT().GetOrCreateAccount(gomock.Any()).Return(account).AnyTimes()

	epochInfo := repo.GenesisEpochInfo(false)
	epochInfo.CandidateSet = append(epochInfo.CandidateSet, &rbft.NodeInfo{
		ID:             9,
		P2PNodeID:      "16Uiu2HAmSBJ7tARZkRT3KS41KPuEbGYZvDXdSzTj8b31gQYYGs9a",
		AccountAddress: "0xD1AEFdf2195f2457A6a675068Cad98B67Eb54e68",
	})
	err = base.InitEpochInfo(mockLedger.StateLedger, epochInfo)
	assert.Nil(t, err)

	var addrs []peer.AddrInfo
	for i := 0; i < peerCnt; i++ {
		rep, err := repo.DefaultWithNodeIndex(t.TempDir(), i, true)
		require.Nil(t, err)
		if versionChange && i == peerCnt-1 {
			repo.BuildVersionSecret = "Shanghai"
		}

		rep.Config.Port.P2P = 0
		rep.Config.Genesis.EpochInfo.P2PBootstrapNodeAddresses = []string{}
		rep.EpochInfo.P2PBootstrapNodeAddresses = []string{}
		swarm, err := newNetworkImpl(rep, log.NewWithModule(fmt.Sprintf("swarm%d", i)), mockLedger)
		require.Nil(t, err)
		err = swarm.Start()
		require.Nil(t, err)

		swarms = append(swarms, swarm)
		addr, err := getAddr(swarm.p2p)
		require.Nil(t, err)
		addrs = append(addrs, addr)
	}

	for i := 0; i < peerCnt; i++ {
		for j := 0; j < peerCnt; j++ {
			if i != j {
				err = swarms[i].p2p.Connect(addrs[j])
				require.Nil(t, err)
			}
		}
	}

	return swarms
}

func stopSwarms(t *testing.T, swarms []*networkImpl) error {
	for _, swarm := range swarms {
		err := swarm.Stop()
		assert.Nil(t, err)
	}
	return nil
}

func TestSwarm_Gater(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}
	gater := newConnectionGater(swarms[0].logger, swarms[0].ledger)
	require.False(t, gater.InterceptPeerDial("1"))
	for _, validator := range swarms[0].repo.EpochInfo.ValidatorSet {
		peerID, err := peer.Decode(validator.P2PNodeID)
		require.Nil(t, err)
		require.True(t, gater.InterceptPeerDial(peerID))
	}
	for _, candidate := range swarms[0].repo.EpochInfo.CandidateSet {
		peerID, err := peer.Decode(candidate.P2PNodeID)
		require.Nil(t, err)
		require.True(t, gater.InterceptPeerDial(peerID))
	}
	require.True(t, gater.InterceptAccept(nil))
}

func TestSwarm_Send(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}

	err := swarms[1].RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, func(stream p2p.Stream, msg *pb.Message) {
		resp := &pb.Message{
			Type: pb.Message_SYNC_STATE_RESPONSE,
			Data: []byte("response aaa"),
		}
		err := swarms[1].SendWithStream(stream, resp)
		require.Nil(t, err)
	})

	require.Nil(t, err)

	msg := &pb.Message{
		Type: pb.Message_SYNC_STATE_REQUEST,
		Data: []byte("aaa"),
	}

	var res *pb.Message
	err = retry.Retry(func(attempt uint) error {
		res, err = swarms[0].Send(swarms[1].PeerID(), msg)
		if err != nil {
			swarms[0].logger.Errorf(err.Error())
			return err
		}
		return nil
	}, strategy.Wait(50*time.Millisecond))
	require.Nil(t, err)
	require.Equal(t, pb.Message_SYNC_STATE_RESPONSE, res.Type)
	require.Equal(t, "response aaa", string(res.Data))
}

func TestSwarm_RegisterMsgHandler(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}

	t.Run("register msg handler with nil handler", func(t *testing.T) {
		err := swarms[1].RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, nil)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "empty handler")
	})

	t.Run("register msg handler with invalid message type", func(t *testing.T) {
		err := swarms[1].RegisterMsgHandler(pb.Message_Type(100), func(stream p2p.Stream, msg *pb.Message) {})
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "invalid message type")
	})

	t.Run("register msg handler success", func(t *testing.T) {
		err := swarms[1].RegisterMsgHandler(pb.Message_SYNC_STATE_REQUEST, func(stream p2p.Stream, msg *pb.Message) {
			resp := &pb.Message{
				Type: pb.Message_SYNC_STATE_RESPONSE,
				Data: []byte("response aaa"),
			}
			err := swarms[0].SendWithStream(stream, resp)
			require.Nil(t, err)
		})
		require.Nil(t, err)

		msg := &pb.Message{
			Type: pb.Message_SYNC_STATE_REQUEST,
			Data: []byte("aaa"),
		}

		var res *pb.Message
		err = retry.Retry(func(attempt uint) error {
			res, err = swarms[0].Send(swarms[1].PeerID(), msg)
			if err != nil {
				swarms[0].logger.Errorf(err.Error())
				return err
			}
			return nil
		}, strategy.Wait(50*time.Millisecond))
		require.Nil(t, err)
		require.Equal(t, pb.Message_SYNC_STATE_RESPONSE, res.Type)
		require.Equal(t, "response aaa", string(res.Data))
	})
}

func TestSwarm_RegisterMultiMsgHandler(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer stopSwarms(t, swarms)

	for swarms[0].CountConnectedPeers() != 3 {
		time.Sleep(100 * time.Millisecond)
	}

	err := swarms[1].RegisterMultiMsgHandler([]pb.Message_Type{pb.Message_SYNC_BLOCK_REQUEST, pb.Message_SYNC_STATE_REQUEST}, func(stream p2p.Stream, msg *pb.Message) {
		var resp *pb.Message
		switch msg.Type {
		case pb.Message_SYNC_BLOCK_REQUEST:
			resp = &pb.Message{
				Type: pb.Message_SYNC_BLOCK_RESPONSE,
				Data: []byte("response block aaa"),
			}
		case pb.Message_SYNC_STATE_REQUEST:
			resp = &pb.Message{
				Type: pb.Message_SYNC_STATE_RESPONSE,
				Data: []byte("response state bbb"),
			}
		}
		err := swarms[1].SendWithStream(stream, resp)
		require.Nil(t, err)
	})

	require.Nil(t, err)

	msg1 := &pb.Message{
		Type: pb.Message_SYNC_BLOCK_REQUEST,
		Data: []byte("aaa"),
	}

	var res *pb.Message
	err = retry.Retry(func(attempt uint) error {
		res, err = swarms[0].Send(swarms[1].PeerID(), msg1)
		if err != nil {
			swarms[0].logger.Errorf(err.Error())
			return err
		}
		return nil
	}, strategy.Wait(50*time.Millisecond))
	require.Nil(t, err)
	require.Equal(t, pb.Message_SYNC_BLOCK_RESPONSE, res.Type)

	msg2 := &pb.Message{
		Type: pb.Message_SYNC_STATE_REQUEST,
		Data: []byte("aaa"),
	}

	err = retry.Retry(func(attempt uint) error {
		res, err = swarms[0].Send(swarms[1].PeerID(), msg2)
		if err != nil {
			swarms[0].logger.Errorf(err.Error())
			return err
		}
		return nil
	}, strategy.Wait(50*time.Millisecond))
	require.Nil(t, err)
	require.Equal(t, pb.Message_SYNC_STATE_RESPONSE, res.Type)

	t.Run("register msg handler with nil handler", func(t *testing.T) {
		err := swarms[1].RegisterMultiMsgHandler([]pb.Message_Type{pb.Message_SYNC_STATE_REQUEST}, nil)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "empty handler")
	})
}
