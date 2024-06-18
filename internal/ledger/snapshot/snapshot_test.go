package snapshot

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNormalCase(t *testing.T) {
	logger := log.NewWithModule("snapshot_test")
	pStateStorage := kv.NewMemory()

	snapshot := NewSnapshot(createMockRepo(t), pStateStorage, logger)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
	addr3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))
	addr4 := types.NewAddress(LeftPadBytes([]byte{104}, 20))
	addr5 := types.NewAddress(LeftPadBytes([]byte{105}, 20))

	storageRoot1 := common.Hash{1}
	storageRoot2 := common.Hash{2}
	storageRoot3 := common.Hash{3}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string]*types.InnerAccount)
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

	accountSet[addr1.String()] = account1
	accountSet[addr2.String()] = account2
	accountSet[addr3.String()] = account3

	storageSet[addr4.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}
	batch := snapshot.Batch()
	batch.Put(utils.CompositeStorageKey(addr4, []byte("key3")), []byte("val3"))
	batch.Commit()

	storageSet[addr5.String()] = map[string][]byte{
		"key2": []byte("val22"),
		"key3": []byte("val33"),
	}

	_, err := snapshot.Update(1, nil, destructSet, accountSet, storageSet)
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

	a4k3, err := snapshot.Storage(addr4, []byte("key3"))
	require.Nil(t, err)
	require.Equal(t, a4k3, []byte("val3"))

	a5k2, err := snapshot.Storage(addr5, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a5k2, []byte("val22"))

	a5k3, err := snapshot.Storage(addr5, []byte("key3"))
	require.Nil(t, err)
	require.Equal(t, a5k3, []byte("val33"))
}

func TestStateTransit(t *testing.T) {
	logger := log.NewWithModule("snapshot_test")
	pStateStorage := kv.NewMemory()

	snapshot := NewSnapshot(createMockRepo(t), pStateStorage, logger)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))

	storageRoot1 := common.Hash{1}
	storageRoot2 := common.Hash{2}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string]*types.InnerAccount)
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

	accountSet[addr1.String()] = account1
	accountSet[addr2.String()] = account2

	storageSet[addr2.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}

	_, err := snapshot.Update(1, nil, destructSet, accountSet, storageSet)
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

	accountSet2 := make(map[string]*types.InnerAccount)

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

	accountSet2[addr1.String()] = account11
	accountSet2[addr2.String()] = account22

	_, err = snapshot.Update(1, nil, destructSet, accountSet2, storageSet)
	require.Nil(t, err)

	a11, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a11, account11))

	a22, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a22, account22))
}

func TestRollback(t *testing.T) {
	logger := log.NewWithModule("snapshot_test")
	pStateStorage := kv.NewMemory()

	snapshot := NewSnapshot(createMockRepo(t), pStateStorage, logger)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))

	emptyStorageRoot := common.Hash{}
	storageRoot2 := common.Hash{2}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string]*types.InnerAccount)
	storageSet := make(map[string]map[string][]byte)

	account1 := &types.InnerAccount{
		Balance:     big.NewInt(1),
		Nonce:       1,
		StorageRoot: emptyStorageRoot,
	}

	account2 := &types.InnerAccount{
		StorageRoot: storageRoot2,
	}

	accountSet[addr1.String()] = account1
	accountSet[addr2.String()] = account2

	storageSet[addr2.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}

	journal1 := &types.SnapshotJournal{}

	journal1.Journals = append(journal1.Journals, &types.SnapshotJournalEntry{
		Address:        addr1,
		PrevAccount:    nil,
		AccountChanged: true,
		PrevStates:     nil,
	})

	journal1.Journals = append(journal1.Journals, &types.SnapshotJournalEntry{
		Address:        addr2,
		PrevAccount:    nil,
		AccountChanged: true,
		PrevStates:     nil,
	})

	_, err := snapshot.Update(1, journal1, destructSet, accountSet, storageSet)
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
	// block1 -> block2

	accountSet2 := make(map[string]*types.InnerAccount)
	storageSet2 := make(map[string]map[string][]byte)
	storageRoot3 := common.Hash{3}

	account11 := &types.InnerAccount{
		Balance:     big.NewInt(11),
		Nonce:       11,
		StorageRoot: emptyStorageRoot,
	}

	account22 := &types.InnerAccount{
		StorageRoot: storageRoot3,
	}

	storageSet2[addr2.String()] = map[string][]byte{
		"key2": []byte("val22"),
		"key3": []byte("val3"),
	}

	accountSet2[addr1.String()] = account11
	accountSet2[addr2.String()] = account22

	journal2 := &types.SnapshotJournal{}

	journal2.Journals = append(journal2.Journals, &types.SnapshotJournalEntry{
		Address:        addr1,
		PrevAccount:    account1,
		AccountChanged: true,
		PrevStates:     nil,
	})

	journal2.Journals = append(journal2.Journals, &types.SnapshotJournalEntry{
		Address:        addr2,
		PrevAccount:    account2,
		AccountChanged: true,
		PrevStates: map[string][]byte{
			"key1": []byte("val1"),
			"key2": []byte("val2"),
			"key3": nil,
			"key4": make([]byte, maxBatchSize+1),
		},
	})

	_, err = snapshot.Update(2, journal2, destructSet, accountSet2, storageSet2)
	require.Nil(t, err)

	a11, err := snapshot.Account(addr1)
	require.Nil(t, err)

	require.True(t, isEqualAccount(a11, account11))

	a22, err := snapshot.Account(addr2)
	require.Nil(t, err)
	require.True(t, isEqualAccount(a22, account22))

	a2k1, err = snapshot.Storage(addr2, []byte("key1"))
	require.Nil(t, err)
	require.Equal(t, a2k1, []byte("val1"))

	a2k2, err = snapshot.Storage(addr2, []byte("key2"))
	require.Nil(t, err)
	require.Equal(t, a2k2, []byte("val22"))

	a2k3, err := snapshot.Storage(addr2, []byte("key3"))
	require.Nil(t, err)
	require.Equal(t, a2k3, []byte("val3"))

	t.Run("rollback to state 1", func(t *testing.T) {
		err = snapshot.Rollback(1)
		require.Nil(t, err)

		a1, err = snapshot.Account(addr1)
		require.Nil(t, err)

		require.True(t, isEqualAccount(a1, account1))

		a2, err = snapshot.Account(addr2)
		require.Nil(t, err)
		require.True(t, isEqualAccount(a2, account2))

		a2k1, err = snapshot.Storage(addr2, []byte("key1"))
		require.Nil(t, err)
		require.Equal(t, a2k1, []byte("val1"))

		a2k2, err = snapshot.Storage(addr2, []byte("key2"))
		require.Nil(t, err)
		require.Equal(t, a2k2, []byte("val2"))

		a2k3, err = snapshot.Storage(addr2, []byte("key3"))
		require.Nil(t, err)
		require.Equal(t, a2k3, []byte(nil))
	})

	t.Run("rollback to state 0", func(t *testing.T) {
		err = snapshot.Rollback(0)
		require.Nil(t, err)

		a1, err = snapshot.Account(addr1)
		require.Nil(t, err)
		require.Nil(t, a1)

		a2, err = snapshot.Account(addr2)
		require.Nil(t, a2)
	})

	t.Run("rollback error case", func(t *testing.T) {
		// still rollback to state 0, no-op
		err = snapshot.Rollback(0)
		require.Nil(t, err)

		// rollback to state 1
		err = snapshot.Rollback(1)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), ErrorRollbackToHigherNumber.Error())
	})
}

func TestRemoveJournal(t *testing.T) {
	logger := log.NewWithModule("snapshot_test")
	pStateStorage := kv.NewMemory()

	snapshot := NewSnapshot(createMockRepo(t), pStateStorage, logger)

	addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
	addr2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))

	emptyStorageRoot := common.Hash{}
	storageRoot2 := common.Hash{2}

	destructSet := make(map[string]struct{})
	accountSet := make(map[string]*types.InnerAccount)
	storageSet := make(map[string]map[string][]byte)

	account1 := &types.InnerAccount{
		Balance:     big.NewInt(1),
		Nonce:       1,
		StorageRoot: emptyStorageRoot,
	}

	account2 := &types.InnerAccount{
		StorageRoot: storageRoot2,
	}

	accountSet[addr1.String()] = account1
	accountSet[addr2.String()] = account2

	storageSet[addr2.String()] = map[string][]byte{
		"key1": []byte("val1"),
		"key2": []byte("val2"),
	}

	journal1 := &types.SnapshotJournal{}

	journal1.Journals = append(journal1.Journals, &types.SnapshotJournalEntry{
		Address:        addr1,
		PrevAccount:    nil,
		AccountChanged: true,
		PrevStates:     nil,
	})

	journal1.Journals = append(journal1.Journals, &types.SnapshotJournalEntry{
		Address:        addr2,
		PrevAccount:    nil,
		AccountChanged: true,
		PrevStates:     nil,
	})

	_, err := snapshot.Update(1, journal1, destructSet, accountSet, storageSet)
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
	// block1 -> block2

	accountSet2 := make(map[string]*types.InnerAccount)
	storageSet2 := make(map[string]map[string][]byte)
	storageRoot3 := common.Hash{3}

	account11 := &types.InnerAccount{
		Balance:     big.NewInt(11),
		Nonce:       11,
		StorageRoot: emptyStorageRoot,
	}

	account22 := &types.InnerAccount{
		StorageRoot: storageRoot3,
	}

	storageSet2[addr2.String()] = map[string][]byte{
		"key2": []byte("val22"),
		"key3": []byte("val3"),
	}

	accountSet2[addr1.String()] = account11
	accountSet2[addr2.String()] = account22

	journal2 := &types.SnapshotJournal{}

	journal2.Journals = append(journal2.Journals, &types.SnapshotJournalEntry{
		Address:        addr1,
		PrevAccount:    account1,
		AccountChanged: true,
		PrevStates:     nil,
	})

	journal2.Journals = append(journal2.Journals, &types.SnapshotJournalEntry{
		Address:        addr2,
		PrevAccount:    account2,
		AccountChanged: true,
		PrevStates: map[string][]byte{
			"key1": []byte("val1"),
			"key2": []byte("val2"),
		},
	})

	_, err = snapshot.Update(2, journal2, destructSet, accountSet2, storageSet2)
	require.Nil(t, err)

	// remove to higher block
	err = snapshot.RemoveJournalsBeforeBlock(3)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrorRemoveJournalOutOfRange.Error())

	// no-op
	err = snapshot.RemoveJournalsBeforeBlock(0)
	require.Nil(t, err)

	// remove journal, then rollback
	minPreserveJournalNumber = 0
	err = snapshot.RemoveJournalsBeforeBlock(2)
	require.Nil(t, err)
	err = snapshot.Rollback(1)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrorRollbackTooMuch.Error())
}

func TestEmptySnapshot(t *testing.T) {
	logger := log.NewWithModule("snapshot_test")
	pStateStorage := kv.NewMemory()

	snapshot := NewSnapshot(createMockRepo(t), pStateStorage, logger)
	minHeight, maxHeight := snapshot.GetJournalRange()
	assert.Equal(t, uint64(0), minHeight)
	assert.Equal(t, uint64(0), maxHeight)
	assert.Nil(t, snapshot.GetBlockJournal(0))
	assert.Nil(t, snapshot.Rollback(0))
	snapshot.ResetMetrics()
	snapshot.ExportMetrics()
	snapshot = nil
	snapshot.ResetMetrics()
	snapshot.ExportMetrics()
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
	if (a1 == nil && a2 != nil) || (a1 != nil && a2 == nil) {
		return false
	}

	empty := common.Hash{}

	// both are eoa account
	if a1.StorageRoot == empty && a2.StorageRoot == empty {
		return a1.Balance.Int64() == a2.Balance.Int64() && a1.Nonce == a2.Nonce
	}

	// both are contract account
	if a1.StorageRoot != empty && a2.StorageRoot != empty {
		return a1.StorageRoot == a2.StorageRoot
	}

	return false
}

func createMockRepo(t *testing.T) *repo.Repo {
	r := repo.MockRepo(t)
	return r
}
