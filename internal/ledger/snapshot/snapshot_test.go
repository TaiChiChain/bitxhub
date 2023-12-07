package snapshot

import (
	"math/big"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
)

func TestNormalCase(t *testing.T) {
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	snapshot := NewSnapshot(pStateStorage)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
	addr3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))
	addr4 := types.NewAddress(LeftPadBytes([]byte{104}, 20))
	addr5 := types.NewAddress(LeftPadBytes([]byte{105}, 20))

	stateRoot := common.Hash{}
	storageRoot1 := common.Hash{1}
	storageRoot2 := common.Hash{2}
	storageRoot3 := common.Hash{3}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string][]byte)
	storageSet := make(map[string]map[string][]byte)

	destructSet[addr1.String()] = struct{}{}
	destructSet[addr2.String()] = struct{}{}

	account1 := &types.InnerAccount{
		Balance:     big.NewInt(1),
		Nonce:       1,
		StorageRoot: storageRoot1,
	}

	account2 := &types.InnerAccount{
		Balance:     big.NewInt(2),
		Nonce:       2,
		StorageRoot: storageRoot2,
	}

	account3 := &types.InnerAccount{
		Balance:     big.NewInt(3),
		Nonce:       3,
		StorageRoot: storageRoot3,
	}

	blob1, err := account1.Marshal()
	require.Nil(t, err)
	blob2, err := account2.Marshal()
	require.Nil(t, err)
	blob3, err := account3.Marshal()
	require.Nil(t, err)

	accountSet[addr1.String()] = blob1
	accountSet[addr2.String()] = blob2
	accountSet[addr3.String()] = blob3

	storageSet[addr4.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}
	storageSet[addr5.String()] = map[string][]byte{
		"key2": []byte("val22"),
		"key3": []byte("val33"),
	}

	err = snapshot.Update(stateRoot, destructSet, accountSet, storageSet)
	require.Nil(t, err)

	a1, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a1, account1))

	a2, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a2, account2))

	a3, err := snapshot.Account(addr3)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a3, account3))

	a4k1, err := snapshot.Storage(addr4, []byte("key1"))
	require.Nil(t, err)
	require.Equal(t, a4k1, []byte("val1"))

	a4k2, err := snapshot.Storage(addr4, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a4k2, []byte("val2"))

	a5k2, err := snapshot.Storage(addr5, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a5k2, []byte("val22"))

	a5k3, err := snapshot.Storage(addr5, []byte("key3"))
	require.Nil(t, err)
	require.Equal(t, a5k3, []byte("val33"))
}

func TestStateTransit(t *testing.T) {
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	snapshot := NewSnapshot(pStateStorage)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))

	stateRoot := common.Hash{}
	storageRoot1 := common.Hash{1}
	storageRoot2 := common.Hash{2}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string][]byte)
	storageSet := make(map[string]map[string][]byte)

	account1 := &types.InnerAccount{
		Balance:     big.NewInt(1),
		Nonce:       1,
		StorageRoot: storageRoot1,
	}

	account2 := &types.InnerAccount{
		Balance:     big.NewInt(2),
		Nonce:       2,
		StorageRoot: storageRoot2,
	}

	blob1, err := account1.Marshal()
	require.Nil(t, err)
	blob2, err := account2.Marshal()
	require.Nil(t, err)

	accountSet[addr1.String()] = blob1
	accountSet[addr2.String()] = blob2

	storageSet[addr2.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}

	err = snapshot.Update(stateRoot, destructSet, accountSet, storageSet)
	require.Nil(t, err)

	a1, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a1, account1))

	a2, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a2, account2))

	a2k1, err := snapshot.Storage(addr2, []byte("key1"))
	require.Nil(t, err)
	require.Equal(t, a2k1, []byte("val1"))

	a2k2, err := snapshot.Storage(addr2, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a2k2, []byte("val2"))

	// state transit

	accountSet2 := make(map[string][]byte)

	account11 := &types.InnerAccount{
		Balance:     big.NewInt(11),
		Nonce:       11,
		StorageRoot: storageRoot1,
	}

	account22 := &types.InnerAccount{
		Balance:     big.NewInt(22),
		Nonce:       22,
		StorageRoot: storageRoot2,
	}

	blob11, err := account11.Marshal()
	require.Nil(t, err)
	blob22, err := account22.Marshal()
	require.Nil(t, err)

	accountSet2[addr1.String()] = blob11
	accountSet2[addr2.String()] = blob22

	err = snapshot.Update(stateRoot, destructSet, accountSet2, storageSet)
	require.Nil(t, err)

	a11, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a11, account11))

	a22, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a22, account22))

}

func TestDestructAccount(t *testing.T) {
	repoRoot := t.TempDir()
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	snapshot := NewSnapshot(pStateStorage)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
	addr3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))
	addr4 := types.NewAddress(LeftPadBytes([]byte{104}, 20))
	addr5 := types.NewAddress(LeftPadBytes([]byte{105}, 20))

	stateRoot := common.Hash{}
	storageRoot1 := common.Hash{1}
	storageRoot2 := common.Hash{2}
	storageRoot3 := common.Hash{3}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string][]byte)
	storageSet := make(map[string]map[string][]byte)

	destructSet[addr1.String()] = struct{}{}
	destructSet[addr2.String()] = struct{}{}

	account1 := &types.InnerAccount{
		Balance:     big.NewInt(1),
		Nonce:       1,
		StorageRoot: storageRoot1,
	}

	account2 := &types.InnerAccount{
		Balance:     big.NewInt(2),
		Nonce:       2,
		StorageRoot: storageRoot2,
	}

	account3 := &types.InnerAccount{
		Balance:     big.NewInt(3),
		Nonce:       3,
		StorageRoot: storageRoot3,
	}

	blob1, err := account1.Marshal()
	require.Nil(t, err)
	blob2, err := account2.Marshal()
	require.Nil(t, err)
	blob3, err := account3.Marshal()
	require.Nil(t, err)

	accountSet[addr1.String()] = blob1
	accountSet[addr2.String()] = blob2
	accountSet[addr3.String()] = blob3

	storageSet[addr4.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}
	storageSet[addr5.String()] = map[string][]byte{
		"key2": []byte("val22"),
		"key3": []byte("val33"),
	}

	err = snapshot.Update(stateRoot, destructSet, accountSet, storageSet)
	require.Nil(t, err)

	a1, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a1, account1))

	a2, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a2, account2))

	a3, err := snapshot.Account(addr3)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a3, account3))

	a4k1, err := snapshot.Storage(addr4, []byte("key1"))
	require.Nil(t, err)
	require.Equal(t, a4k1, []byte("val1"))

	a4k2, err := snapshot.Storage(addr4, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a4k2, []byte("val2"))

	a5k2, err := snapshot.Storage(addr5, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a5k2, []byte("val22"))

	a5k3, err := snapshot.Storage(addr5, []byte("key3"))
	require.Nil(t, err)
	require.Equal(t, a5k3, []byte("val33"))
}

// LeftPadBytes zero-pads slice to the left up to length l.
func LeftPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)

	return padded
}

func isEqualAccount(a1 *types.InnerAccount, a2 *types.InnerAccount) bool {
	if a1 == nil && a2 == nil {
		return true
	}
	if a1 == nil && a2 != nil || a1 != nil && a2 == nil {
		return false
	}

	return a1.Balance.Int64() == a2.Balance.Int64() && a1.Nonce == a2.Nonce && a1.StorageRoot == a2.StorageRoot
}
