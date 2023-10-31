package network

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/types/pb"
	network "github.com/axiomesh/axiom-p2p"
)

func (swarm *networkImpl) handleMessage(s network.Stream, data []byte) {
	m := &pb.Message{}
	if err := m.UnmarshalVT(data); err != nil {
		swarm.logger.Error(err)
		return
	}

	handler, ok := swarm.msgHandlers.Load(m.Type)
	if !ok {
		swarm.logger.WithFields(logrus.Fields{
			"error": fmt.Errorf("can't handle msg[type: %v]", m.Type),
			"type":  m.Type.String(),
		}).Error("Handle message")
		return
	}

	msgHandler, ok := handler.(func(network.Stream, *pb.Message))
	if !ok {
		swarm.logger.WithFields(logrus.Fields{
			"error": fmt.Errorf("invalid handler for msg [type: %v]", m.Type),
			"type":  m.Type.String(),
		}).Error("Handle message")
		return
	}

	msgHandler(s, m)
}
