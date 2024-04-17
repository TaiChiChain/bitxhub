package repo

import (
	"fmt"
	"path"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/axiomesh/axiom-kit/fileutil"
	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
)

type P2PKeystore struct {
	crypto.Ed25519Keystore
}

type ConsensusKeystore struct {
	crypto.Bls12381Keystore
}

func (p *P2PKeystore) P2PID() string {
	return p.Extra[p2pKeystoreIDKey]
}

func GenerateP2PKeystore(repoPath string, privateKeyStr string, password string) (*P2PKeystore, error) {
	var privateKey *crypto.Ed25519PrivateKey
	var err error
	if privateKeyStr != "" {
		privateKey = &crypto.Ed25519PrivateKey{}
		if err = privateKey.Unmarshal(hexutil.Decode(privateKeyStr)); err != nil {
			return nil, err
		}
	} else {
		privateKey, err = crypto.GenerateEd25519PrivateKey()
		if err != nil {
			return nil, err
		}
	}

	pubKey := privateKey.PublicKey()
	p2pID, err := P2PPubKeyToID(pubKey)
	if err != nil {
		return nil, err
	}

	return &P2PKeystore{
		Ed25519Keystore: crypto.Ed25519Keystore{
			Path:        path.Join(repoPath, P2PKeystoreFileName),
			KeyType:     privateKey.Type(),
			Description: "used for p2p",
			PrivateKey:  privateKey,
			PublicKey:   pubKey,
			Password:    password,
			Extra: map[string]string{
				p2pKeystoreIDKey: p2pID,
			},
		}}, nil
}

func ReadP2PKeystore(repoPath string) (*P2PKeystore, error) {
	keystorePath := path.Join(repoPath, P2PKeystoreFileName)
	if !fileutil.Exist(keystorePath) {
		return nil, fmt.Errorf("%s not exist", keystorePath)
	}
	ks, err := crypto.ReadKeystore[*crypto.Ed25519PrivateKey, *crypto.Ed25519PublicKey](keystorePath)
	if err != nil {
		return nil, err
	}
	return &P2PKeystore{Ed25519Keystore: *ks}, nil
}

func GenerateConsensusKeystore(repoPath string, privateKeyStr string, password string) (*ConsensusKeystore, error) {
	if password == "" {
		password = DefaultKeystorePassword
		fmt.Println("keystore password is empty, will use default:", password)
	}
	var privateKey *crypto.Bls12381PrivateKey
	var err error
	if privateKeyStr != "" {
		privateKey = &crypto.Bls12381PrivateKey{}
		if err = privateKey.Unmarshal(hexutil.Decode(privateKeyStr)); err != nil {
			return nil, err
		}
	} else {
		privateKey, err = crypto.GenerateBls12381PrivateKey()
		if err != nil {
			return nil, err
		}
	}

	return &ConsensusKeystore{
		Bls12381Keystore: crypto.Bls12381Keystore{
			Path:        path.Join(repoPath, ConsensusKeystoreFileName),
			KeyType:     privateKey.Type(),
			Description: "used for consensus",
			PrivateKey:  privateKey,
			PublicKey:   privateKey.PublicKey(),
			Password:    password,
			Extra:       map[string]string{},
		}}, nil
}

func ReadConsensusKeystore(repoPath string) (*ConsensusKeystore, error) {
	keystorePath := path.Join(repoPath, ConsensusKeystoreFileName)
	if !fileutil.Exist(keystorePath) {
		return nil, fmt.Errorf("%s not exist", keystorePath)
	}

	ks, err := crypto.ReadKeystore[*crypto.Bls12381PrivateKey, *crypto.Bls12381PublicKey](keystorePath)
	if err != nil {
		return nil, err
	}
	return &ConsensusKeystore{Bls12381Keystore: *ks}, nil
}

func P2PPubKeyToID(p2pPubKey *crypto.Ed25519PublicKey) (string, error) {
	libp2pPubKey, err := libp2pcrypto.UnmarshalEd25519PublicKey(p2pPubKey.PublicKey)
	if err != nil {
		return "", err
	}

	p2pID, err := peer.IDFromPublicKey(libp2pPubKey)
	if err != nil {
		return "", err
	}
	return p2pID.String(), nil
}
