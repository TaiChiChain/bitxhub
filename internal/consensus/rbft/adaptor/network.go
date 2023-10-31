package adaptor

import (
	"context"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-bft/common/consensus"
	network "github.com/axiomesh/axiom-p2p"
)

func (a *RBFTAdaptor) Broadcast(ctx context.Context, msg *consensus.ConsensusMessage) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "broadcast type[%s] failed", msg.Type.String())
		}
	}()

	data, err := msg.MarshalVTStrict()
	if err != nil {
		return err
	}

	if a.config.Config.Rbft.EnableMultiPipes {
		pipe, ok := a.msgPipes[int32(msg.Type)]
		if !ok {
			return errors.New("unsupported broadcast msg type")
		}

		return pipe.Broadcast(ctx, a.broadcastNodes, data)
	}

	return a.globalMsgPipe.Broadcast(ctx, a.broadcastNodes, data)
}

func (a *RBFTAdaptor) Unicast(ctx context.Context, msg *consensus.ConsensusMessage, to string) error {
	doUnicast, err := a.unicastCheck(ctx, msg, to)
	if err != nil {
		return err
	}
	go func() {
		err := doUnicast()
		if err != nil {
			a.logger.Error(err)
		}
	}()
	return nil
}

func (a *RBFTAdaptor) unicastCheck(ctx context.Context, msg *consensus.ConsensusMessage, to string) (doUnicast func() error, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "unicast msg[%s] to %s failed", msg.Type.String(), to)
		}
	}()

	data, err := msg.MarshalVTStrict()
	if err != nil {
		return nil, err
	}

	var pipe network.Pipe
	if a.config.Config.Rbft.EnableMultiPipes {
		var ok bool
		pipe, ok = a.msgPipes[int32(msg.Type)]
		if !ok {
			return nil, errors.New("unsupported unicast msg type")
		}
	} else {
		pipe = a.globalMsgPipe
	}

	return func() error {
		err := pipe.Send(ctx, to, data)
		if err != nil {
			return errors.Wrapf(err, "unicast msg[%s] to %s failed", msg.Type.String(), to)
		}
		return nil
	}, nil
}
