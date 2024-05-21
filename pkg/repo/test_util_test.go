package repo

import (
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-ledger/pkg/crypto"
)

func TestMockRepo(t *testing.T) {
	for i, skStr := range MockOperatorKeys {
		sk, err := ethcrypto.HexToECDSA(skStr)
		assert.Nil(t, err)
		assert.EqualValues(t, MockOperatorAddrs[i], ethcrypto.PubkeyToAddress(sk.PublicKey).String())
	}

	for i, skStr := range MockConsensusKeys {
		sk := &crypto.Bls12381PrivateKey{}
		err := sk.Unmarshal(hexutil.Decode(skStr))
		assert.Nil(t, err)
		assert.EqualValues(t, MockConsensusPubKeys[i], sk.PublicKey().String())
	}

	for i, skStr := range MockP2PKeys {
		sk := &crypto.Ed25519PrivateKey{}
		err := sk.Unmarshal(hexutil.Decode(skStr))
		assert.Nil(t, err)
		assert.EqualValues(t, MockP2PPubKeys[i], sk.PublicKey().String())
		p2pID, err := P2PPubKeyToID(sk.PublicKey())
		assert.Nil(t, err)
		assert.EqualValues(t, MockP2PIDs[i], p2pID)
	}
}
