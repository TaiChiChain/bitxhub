package repo

import (
	"path"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	tempPath := t.TempDir()

	keyPath := path.Join(tempPath, "node.key")
	k, err := GenerateKey()
	assert.Nil(t, err)

	err = WriteKey(keyPath, k)
	assert.Nil(t, err)

	readKey, err := ReadKey(keyPath)
	assert.Nil(t, err)
	assert.True(t, k.Equal(readKey))
}

func TestDefaultKeys(t *testing.T) {
	for i, key := range DefaultNodeKeys {
		k, err := ParseKey([]byte(key))
		assert.Nil(t, err)
		addr := ethcrypto.PubkeyToAddress(k.PublicKey)
		assert.Equal(t, DefaultNodeAddrs[i], addr.String())
		nodeID, err := KeyToNodeID(k)
		assert.Nil(t, err)
		assert.Equal(t, defaultNodeIDs[i], nodeID)
	}
}

func TestP2PID2PubKey(t *testing.T) {
	k, err := GenerateKey()
	assert.Nil(t, err)

	priKey, err := Libp2pKeyFromECDSAKey(k)
	assert.Nil(t, err)

	id, err := KeyToNodeID(k)
	assert.Nil(t, err)

	pubKey, err := Libp2pIDToPubKey(id)
	assert.Nil(t, err)
	assert.True(t, pubKey.Equals(priKey.GetPublic()))
}
