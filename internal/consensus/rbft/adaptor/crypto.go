package adaptor

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// TODO: use edd25519 improve performance
func (a *RBFTAdaptor) Sign(msg []byte) ([]byte, error) {
	h := crypto.Keccak256Hash(msg)
	return crypto.Sign(h[:], a.priv)
}

func (a *RBFTAdaptor) Verify(address string, signature []byte, msg []byte) error {
	h := crypto.Keccak256Hash(msg)
	pubKey, err := crypto.SigToPub(h[:], signature)
	if err != nil {
		return err
	}

	if address != crypto.PubkeyToAddress(*pubKey).String() {
		return errors.New("invalid signature")
	}
	return nil
}
