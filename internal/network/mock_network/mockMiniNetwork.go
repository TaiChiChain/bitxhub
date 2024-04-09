package mock_network

import (
	"context"
	"fmt"
	"sync"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/axiomesh/axiom-ledger/internal/network"
	p2p "github.com/axiomesh/axiom-p2p"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ network.Network = (*MiniNetwork)(nil)

type MiniNetwork struct {
	p2p         *p2p.MockP2P
	id          uint64
	msgHandlers sync.Map
	logger      logrus.FieldLogger
}

func NewMiniNetwork(id uint64, p2pNodeId string, manager *p2p.MockHostManager) (*MiniNetwork, error) {
	p2pLog := log.NewWithModule(fmt.Sprintf("p2p%d", id))
	p2p, err := p2p.NewMockP2P(p2pNodeId, manager, p2pLog)
	if err != nil {
		return nil, err
	}
	return &MiniNetwork{p2p: p2p, logger: p2pLog, id: id}, nil
}

func (m *MiniNetwork) CreatePipe(ctx context.Context, pipeID string) (p2p.Pipe, error) {
	return m.p2p.CreatePipe(ctx, pipeID)
}

func (m *MiniNetwork) Start() error {
	m.p2p.SetMessageHandler(m.handleMessage)
	return m.p2p.Start()
}

func (m *MiniNetwork) Stop() error {
	return m.p2p.Stop()
}

func (m *MiniNetwork) PeerID() string {
	return m.p2p.PeerID()
}

func (m *MiniNetwork) Send(to string, message *pb.Message) (*pb.Message, error) {
	msg, err := message.MarshalVT()
	if err != nil {
		return nil, err
	}
	resp, err := m.p2p.Send(to, msg)
	if err != nil {
		return nil, err
	}
	respMsg := &pb.Message{}
	if err = respMsg.UnmarshalVT(resp); err != nil {
		return nil, err
	}
	return respMsg, nil
}

func (m *MiniNetwork) SendWithStream(stream p2p.Stream, message *pb.Message) error {
	msg, err := message.MarshalVT()
	if err != nil {
		return err
	}
	return stream.AsyncSend(msg)
}

func (m *MiniNetwork) CountConnectedValidators() uint64 {
	// TODO implement me
	panic("implement me")
}

func (m *MiniNetwork) handleMessage(s p2p.Stream, data []byte) {
	msg := &pb.Message{}
	if err := msg.UnmarshalVT(data); err != nil {
		m.logger.Errorf("unmarshal message error: %w", err)
		return
	}

	handler, ok := m.msgHandlers.Load(msg.Type)
	if !ok {
		m.logger.WithFields(logrus.Fields{
			"error": fmt.Errorf("can't handle msg[type: %v]", msg.Type),
			"type":  msg.Type.String(),
		}).Error("Handle message")
		return
	}

	msgHandler, ok := handler.(func(p2p.Stream, *pb.Message))
	if !ok {
		m.logger.WithFields(logrus.Fields{
			"error": fmt.Errorf("invalid handler for msg [type: %v]", msg.Type),
			"type":  msg.Type.String(),
		}).Error("Handle message")
		return
	}

	msgHandler(s, msg)
}

func (m *MiniNetwork) RegisterMsgHandler(messageType pb.Message_Type, handler func(p2p.Stream, *pb.Message)) error {
	if handler == nil {
		return errors.New("register msg handler: empty handler")
	}

	for msgType := range pb.Message_Type_name {
		if msgType == int32(messageType) {
			m.msgHandlers.Store(messageType, handler)
			return nil
		}
	}

	return errors.New("register msg handler: invalid message type")
}

func (m *MiniNetwork) RegisterMultiMsgHandler(messageTypes []pb.Message_Type, handler func(p2p.Stream, *pb.Message)) error {
	for _, typ := range messageTypes {
		if err := m.RegisterMsgHandler(typ, handler); err != nil {
			return err
		}
	}

	return nil
}

func (m *MiniNetwork) GetSelfId() uint64 {
	return m.id
}
