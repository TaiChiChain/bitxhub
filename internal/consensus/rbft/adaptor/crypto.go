package adaptor

import (
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

// TODO: use edd25519 improve performance
func (a *RBFTAdaptor) Sign(msg []byte) ([]byte, error) {
	return a.libp2pKey.Sign(msg)
}

func (a *RBFTAdaptor) Verify(p2pID string, signature []byte, msg []byte) error {
	pubKey, err := a.getP2PPubKey(p2pID)
	if err != nil {
		return err
	}

	ok, err := pubKey.Verify(msg, signature)
	if err != nil {
		return errors.Wrap(err, "invalid signature")
	}
	if !ok {
		return errors.New("invalid signature")
	}
	return nil
}

func (a *RBFTAdaptor) getP2PPubKey(p2pID string) (libp2pcrypto.PubKey, error) {
	a.p2pID2PubKeyCacheLock.RLock()
	pk, ok := a.p2pID2PubKeyCache[p2pID]
	a.p2pID2PubKeyCacheLock.RUnlock()
	if ok {
		return pk, nil
	}

	pk, err := repo.Libp2pIDToPubKey(p2pID)
	if err != nil {
		return nil, err
	}
	a.p2pID2PubKeyCacheLock.Lock()
	a.p2pID2PubKeyCache[p2pID] = pk
	a.p2pID2PubKeyCacheLock.Unlock()

	return pk, nil
}
