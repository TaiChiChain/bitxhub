package crypto

import (
	"crypto/rand"

	"github.com/pkg/errors"
	blst "github.com/supranational/blst/bindings/go"

	"github.com/axiomesh/axiom-kit/hexutil"
)

var bls12381Dist = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

type Bls12381Signature = blst.P1Affine

type Bls12381AggregateSignature = blst.P1Aggregate

type Bls12381AggregatePublicKey = blst.P2Aggregate

var _ KeystoreKey = (*Bls12381PrivateKey)(nil)

var _ KeystoreKey = (*Bls12381PublicKey)(nil)

type Bls12381PrivateKey struct {
	PrivateKey *blst.SecretKey
}

func GenerateBls12381PrivateKey() (*Bls12381PrivateKey, error) {
	var ikm [32]byte
	_, err := rand.Reader.Read(ikm[:])
	if err != nil {
		return nil, errors.Wrap(err, "bls12381: failed to generate random seed")
	}
	secretKey := blst.KeyGen(ikm[:])
	if IsZeroBytes(secretKey.Serialize()) {
		return nil, errors.Errorf("bls12381: private key is zero")
	}
	return &Bls12381PrivateKey{
		PrivateKey: secretKey,
	}, nil
}

func (e *Bls12381PrivateKey) Sign(msg []byte) ([]byte, error) {
	sig := new(Bls12381Signature).Sign(e.PrivateKey, msg, bls12381Dist)
	return sig.Compress(), nil
}

func (e *Bls12381PrivateKey) Type() string {
	return KeyTypeBls12381
}

func (e *Bls12381PrivateKey) String() string {
	return hexutil.Encode(e.PrivateKey.Serialize())
}

func (e *Bls12381PrivateKey) PublicKey() *Bls12381PublicKey {
	return &Bls12381PublicKey{PublicKey: new(blst.P2Affine).From(e.PrivateKey)}
}

func (e *Bls12381PrivateKey) Marshal() ([]byte, error) {
	return e.PrivateKey.Serialize(), nil
}

func (e *Bls12381PrivateKey) Unmarshal(bytes []byte) error {
	if len(bytes) != blst.BLST_SCALAR_BYTES {
		return errors.Errorf("bls12381: bad private key length %d, want %d", len(bytes), blst.BLST_SCALAR_BYTES)
	}
	if IsZeroBytes(bytes) {
		return errors.New("bls12381: private key is zero")
	}
	secKey := new(blst.SecretKey).Deserialize(bytes)
	if secKey == nil {
		return errors.New("bls12381: bad private key")
	}
	e.PrivateKey = secKey
	return nil
}

type Bls12381PublicKey struct {
	PublicKey *blst.P2Affine
}

func (e *Bls12381PublicKey) Verify(msg []byte, sig []byte) bool {
	signature, err := signatureFromBytes(sig)
	if err != nil {
		return false
	}
	return signature.Verify(false, e.PublicKey, false, msg, bls12381Dist)
}

func (e *Bls12381PublicKey) Type() string {
	return KeyTypeBls12381
}

func (e *Bls12381PublicKey) String() string {
	return hexutil.Encode(e.PublicKey.Compress())
}

func (e *Bls12381PublicKey) Marshal() ([]byte, error) {
	return e.PublicKey.Compress(), nil
}

func (e *Bls12381PublicKey) Unmarshal(bytes []byte) error {
	if len(bytes) != blst.BLST_P2_COMPRESS_BYTES {
		return errors.Errorf("bls12381: bad public key length %d, want %d", len(bytes), blst.BLST_P2_COMPRESS_BYTES)
	}
	if IsZeroBytes(bytes) {
		return errors.New("bls12381: public key is zero")
	}
	pubKey := new(blst.P2Affine).Uncompress(bytes)
	if pubKey == nil {
		return errors.New("bls12381: bad public key")
	}
	e.PublicKey = pubKey
	return nil
}

func signatureFromBytes(sig []byte) (*Bls12381Signature, error) {
	if len(sig) != blst.BLST_P1_COMPRESS_BYTES {
		return nil, errors.Errorf("bls12381: bad signature length %d, want %d", len(sig), blst.BLST_P1_COMPRESS_BYTES)
	}
	signature := new(Bls12381Signature).Uncompress(sig)
	if signature == nil {
		return nil, errors.New("bls12381: invalid signature")
	}
	return signature, nil
}
