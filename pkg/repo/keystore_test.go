package repo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsensusKeystore(t *testing.T) {
	tests := []struct {
		privateKeyStr string
		password      string
	}{
		{
			privateKeyStr: "",
			password:      "",
		},
		{
			privateKeyStr: "",
			password:      "123",
		},
		{
			privateKeyStr: "0x10bf2dacd178ae68b3cb72b5b85d0e28c7092e665fb05ca2611c351a0aae273f",
			password:      "",
		},
		{
			privateKeyStr: "0x10bf2dacd178ae68b3cb72b5b85d0e28c7092e665fb05ca2611c351a0aae273f",
			password:      "123",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			repoPath := t.TempDir()

			_, err := ReadConsensusKeystore(repoPath)
			assert.ErrorContains(t, err, "not exist")

			consensusKeystore, err := GenerateConsensusKeystore(repoPath, tt.privateKeyStr, tt.password)
			assert.Nil(t, err)

			fmt.Println("PrivateKey", consensusKeystore.PrivateKey.String())
			fmt.Println("PublicKey", consensusKeystore.PublicKey.String())

			err = consensusKeystore.Write()
			assert.Nil(t, err)

			readConsensusKeystore, err := ReadConsensusKeystore(repoPath)
			assert.Nil(t, err)

			err = readConsensusKeystore.DecryptPrivateKey(tt.password + "err")
			assert.ErrorContains(t, err, "failed to decrypt")

			err = readConsensusKeystore.DecryptPrivateKey(tt.password)
			assert.Nil(t, err)
			assert.EqualValues(t, consensusKeystore.PrivateKey.String(), readConsensusKeystore.PrivateKey.String())
			assert.EqualValues(t, consensusKeystore.PublicKey.String(), readConsensusKeystore.PublicKey.String())
		})
	}
}

func TestP2PKeystore(t *testing.T) {
	tests := []struct {
		privateKeyStr string
		password      string
	}{
		{
			privateKeyStr: "",
			password:      "",
		},
		{
			privateKeyStr: "",
			password:      "123",
		},
		{
			privateKeyStr: "0x052cbbe64f2d02e09c088befdcd0d2940715ccda7e259aa6b00e3e3df74d084e",
			password:      "",
		},
		{
			privateKeyStr: "0x052cbbe64f2d02e09c088befdcd0d2940715ccda7e259aa6b00e3e3df74d084e",
			password:      "123",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			repoPath := t.TempDir()

			_, err := ReadP2PKeystore(repoPath)
			assert.ErrorContains(t, err, "not exist")

			p2pKeystore, err := GenerateP2PKeystore(repoPath, tt.privateKeyStr, tt.password)
			assert.Nil(t, err)

			fmt.Println("PrivateKey", p2pKeystore.PrivateKey.String())
			fmt.Println("PublicKey", p2pKeystore.PublicKey.String())

			err = p2pKeystore.Write()
			assert.Nil(t, err)

			readP2PKeystore, err := ReadP2PKeystore(repoPath)
			assert.Nil(t, err)

			err = readP2PKeystore.DecryptPrivateKey(tt.password + "err")
			assert.ErrorContains(t, err, "failed to decrypt")

			err = readP2PKeystore.DecryptPrivateKey(tt.password)
			assert.Nil(t, err)
			assert.EqualValues(t, p2pKeystore.PrivateKey.String(), readP2PKeystore.PrivateKey.String())
			assert.EqualValues(t, p2pKeystore.PublicKey.String(), readP2PKeystore.PublicKey.String())
			assert.Equal(t, p2pKeystore.P2PID(), readP2PKeystore.P2PID())
		})
	}
}
