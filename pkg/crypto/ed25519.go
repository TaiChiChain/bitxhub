package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-kit/hexutil"
)

var _ KeystoreKey = (*Ed25519PrivateKey)(nil)

var _ KeystoreKey = (*Ed25519PublicKey)(nil)

type Ed25519PrivateKey struct {
	PrivateKey ed25519.PrivateKey
}

func GenerateEd25519PrivateKey() (*Ed25519PrivateKey, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "bls12381: failed to generate private key")
	}
	return &Ed25519PrivateKey{PrivateKey: privateKey}, nil
}

func (e *Ed25519PrivateKey) Sign(msg []byte) ([]byte, error) {
	return e.PrivateKey.Sign(rand.Reader, msg, &ed25519.Options{
		Hash:    0,
		Context: "",
	})
}

func (e *Ed25519PrivateKey) PublicKey() *Ed25519PublicKey {
	return &Ed25519PublicKey{PublicKey: e.PrivateKey.Public().(ed25519.PublicKey)}
}

func (e *Ed25519PrivateKey) Type() string {
	return KeyTypeEd25519
}

func (e *Ed25519PrivateKey) String() string {
	return hexutil.Encode(e.PrivateKey.Seed())
}

func (e *Ed25519PrivateKey) Marshal() ([]byte, error) {
	return e.PrivateKey.Seed(), nil
}

func (e *Ed25519PrivateKey) Unmarshal(bytes []byte) error {
	if len(bytes) != ed25519.SeedSize {
		return errors.Errorf("ed25519: bad private key length %d, want %d", len(bytes), ed25519.SeedSize)
	}
	if IsZeroBytes(bytes) {
		return errors.New("ed25519: private key is zero")
	}
	e.PrivateKey = ed25519.NewKeyFromSeed(bytes)
	return nil
}

type Ed25519PublicKey struct {
	PublicKey ed25519.PublicKey
}

func (e *Ed25519PublicKey) Verify(msg []byte, sig []byte) bool {
	return ed25519.Verify(e.PublicKey, msg, sig)
}

func (e *Ed25519PublicKey) Type() string {
	return KeyTypeEd25519
}

func (e *Ed25519PublicKey) String() string {
	return hexutil.Encode(e.PublicKey)
}

func (e *Ed25519PublicKey) Marshal() ([]byte, error) {
	return bytes.Clone(e.PublicKey), nil
}

func (e *Ed25519PublicKey) Unmarshal(raw []byte) error {
	if len(raw) != ed25519.PublicKeySize {
		return errors.Errorf("ed25519: bad public key length %d, want %d", len(raw), ed25519.PublicKeySize)
	}
	if IsZeroBytes(raw) {
		return errors.New("ed25519: public key is zero")
	}
	e.PublicKey = bytes.Clone(raw)
	return nil
}
