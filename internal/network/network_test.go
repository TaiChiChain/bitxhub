package network

import (
	"encoding/json"
	"fmt"
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

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/executor/system/common"
	"github.com/axiomesh/axiom-ledger/internal/ledger"
	"github.com/axiomesh/axiom-ledger/internal/ledger/mock_ledger"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
	p2p "github.com/axiomesh/axiom-p2p"
)

func TestVersionCheck(t *testing.T) {
	peerCnt := 4
	// pass true to change the last Node's version
	swarms := newMockSwarms(t, peerCnt, true)
	defer func() {
		_ = stopSwarms(t, swarms)
	}()

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
	genesisConfig := generateMockConfig(t)
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

	jsonBytes, err := json.Marshal(genesisConfig.EpochInfo)
	assert.Nil(t, err)

	stateLedger.EXPECT().GetState(gomock.Any(), gomock.Any()).DoAndReturn(func(addr *types.Address, key []byte) (bool, []byte) {
		return true, jsonBytes
	}).AnyTimes()

	stateLedger.EXPECT().SetState(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(addr *types.Address, key []byte, value []byte) {},
	).AnyTimes()

	mockLedger.StateLedger.SetState(types.NewAddressByStr(common.GovernanceContractAddr), []byte(common.GovernanceContractAddr), jsonBytes)

	var peerID = "16Uiu2HAmRypzJbdbUNYsCV2VVgv9UryYS5d7wejTJXT73mNLJ8AK"

	success, data := mockLedger.StateLedger.GetState(types.NewAddressByStr(common.GovernanceContractAddr), []byte(common.GovernanceContractAddr))
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

func generateMockConfig(t *testing.T) *repo.GenesisConfig {
	r := repo.MockRepo(t)
	config := r.GenesisConfig

	for i := 0; i < 4; i++ {
		config.CouncilMembers = append(config.CouncilMembers, &repo.CouncilMember{
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
	var addrs []peer.AddrInfo
	for i := 0; i < peerCnt; i++ {
		rep := repo.MockRepoWithNodeID(t, peerCnt, uint64(i+1), repo.MockDefaultIsDataSyncers, repo.MockDefaultStakeNumbers)
		if versionChange && i == peerCnt-1 {
			repo.BuildVersionSecret = "Shanghai"
		}

		rep.Config.Port.P2P = 0
		swarm, err := newNetworkImpl(rep, log.NewWithModule(fmt.Sprintf("swarm%d", i)))
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
				err := swarms[i].p2p.Connect(addrs[j])
				require.Nil(t, err)
			}
		}
	}

	for i := 0; i < peerCnt; i++ {
		for len(swarms[i].p2p.GetPeers()) != 4 {
			time.Sleep(50 * time.Millisecond)
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

func TestSwarm_Send(t *testing.T) {
	peerCnt := 4
	swarms := newMockSwarms(t, peerCnt, false)
	defer func() {
		_ = stopSwarms(t, swarms)
	}()

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
	defer func() {
		_ = stopSwarms(t, swarms)
	}()

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
	defer func() {
		_ = stopSwarms(t, swarms)
	}()

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
