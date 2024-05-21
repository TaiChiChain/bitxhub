package network

import (
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

// ConnectionGater
type connectionGater struct {
	logger logrus.FieldLogger
	ledger *ledger.Ledger
}

func newConnectionGater(logger logrus.FieldLogger) *connectionGater {
	return &connectionGater{
		logger: logger,
	}
}

func (g *connectionGater) VerifyRemotePeer(peerID string) (allow bool) {
	return true
}

func (g *connectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return g.VerifyRemotePeer(p.String())
}

func (g *connectionGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) (allow bool) {
	return g.VerifyRemotePeer(p.String())
}

func (g *connectionGater) InterceptAccept(addr network.ConnMultiaddrs) (allow bool) {
	return true
}

func (g *connectionGater) InterceptSecured(d network.Direction, p peer.ID, addr network.ConnMultiaddrs) (allow bool) {
	return g.VerifyRemotePeer(p.String())
}

func (g *connectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}
