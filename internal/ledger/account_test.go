package ledger

import (
	"math/big"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/blockfile"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
)

func TestAccount_GetState(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	lSnapshotStorage, err := leveldb.New(filepath.Join(repoRoot, "lSnapshot"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)
	pSnapshotStorage, err := pebble.New(filepath.Join(repoRoot, "pSnapshot"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage    storage.Storage
		stateStorage    storage.Storage
		snapshotStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage, snapshotStorage: lSnapshotStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage, snapshotStorage: pSnapshotStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			ledger, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, tc.snapshotStorage, blockFile)
			assert.Nil(t, err)

			addr := types.NewAddressByStr("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)
			account := NewAccount(1, stateLedger.cachedDB, stateLedger.accountCache, addr, NewChanger(), stateLedger.snapshot)

			addr1 := account.GetAddress()
			assert.Equal(t, addr, addr1)

			account.SetState([]byte("a"), []byte("b"))
			ok, v := account.GetState([]byte("a"))
			assert.True(t, ok)
			assert.Equal(t, []byte("b"), v)

			account.pendingState["a"] = []byte("v1")
			v = account.GetCommittedState([]byte("a"))
			assert.Equal(t, []byte("v1"), v)

			account.pendingState["a"] = nil
			v = account.GetCommittedState([]byte("a"))
			assert.Equal(t, (&types.Hash{}).Bytes(), v)

			delete(account.pendingState, "a")
			account.originState["a"] = []byte("v2")
			v = account.GetCommittedState([]byte("a"))
			assert.Equal(t, []byte("v2"), v)

			account.originState["a"] = nil
			v = account.GetCommittedState([]byte("a"))
			assert.Equal(t, (&types.Hash{}).Bytes(), v)

			delete(account.originState, "a")
			account.snapshot = nil
			v = account.GetCommittedState([]byte("a"))
			assert.Equal(t, (&types.Hash{}).Bytes(), v)
		})
	}
}

func TestAccount_AccountBalance(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	lSnapshotStorage, err := leveldb.New(filepath.Join(repoRoot, "lSnapshot"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)
	pSnapshotStorage, err := pebble.New(filepath.Join(repoRoot, "pSnapshot"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage    storage.Storage
		stateStorage    storage.Storage
		snapshotStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage, snapshotStorage: lSnapshotStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage, snapshotStorage: pSnapshotStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			ledger, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, tc.snapshotStorage, blockFile)
			assert.Nil(t, err)

			addr := types.NewAddressByStr("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)
			account := NewAccount(1, stateLedger.cachedDB, stateLedger.accountCache, addr, NewChanger(), stateLedger.snapshot)

			account.AddBalance(big.NewInt(1))
			account.SubBalance(big.NewInt(1))

			account.SubBalance(big.NewInt(0))

			account.setCodeAndHash([]byte{'1'})
			account.dirtyAccount = nil
			account.setBalance(big.NewInt(1))
		})
	}
}

func TestByteLogger_String(t *testing.T) {
	logger := &bytesLazyLogger{
		bytes: []byte("1111"),
	}
	assert.Equal(t, logger.String(), "0x31313131")
}

func TestAccount_setNonce(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	lSnapshotStorage, err := leveldb.New(filepath.Join(repoRoot, "lSnapshot"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)
	pSnapshotStorage, err := pebble.New(filepath.Join(repoRoot, "pSnapshot"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage    storage.Storage
		stateStorage    storage.Storage
		snapshotStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage, snapshotStorage: lSnapshotStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage, snapshotStorage: pSnapshotStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			ledger, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, tc.snapshotStorage, blockFile)
			assert.Nil(t, err)

			addr := types.NewAddressByStr("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)
			account := NewAccount(1, stateLedger.cachedDB, stateLedger.accountCache, addr, NewChanger(), stateLedger.snapshot)

			account.setNonce(1)

			ledger.Close()
		})
	}
}

func TestAccount_InitJMTError(t *testing.T) {
	lStateStorage, err := leveldb.NewMemory()
	assert.Nil(t, err)
	ac, err := NewAccountCache(0, true)
	assert.Nil(t, err)

	addr := types.NewAddressByStr("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	account := NewAccount(1, lStateStorage, ac, addr, NewChanger(), nil)

	defer func() {
		if r := recover(); r != nil {
			assert.NotNil(t, r)
		}
	}()
	account.storageTrie = nil
	account.originAccount = &types.InnerAccount{
		StorageRoot: common.Hash{1},
	}
	_ = account.originAccount.String()

	account.initStorageTrie()
}

func TestAccount_getAccountJournal(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	lSnapshotStorage, err := leveldb.New(filepath.Join(repoRoot, "lSnapshot"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)
	pSnapshotStorage, err := pebble.New(filepath.Join(repoRoot, "pSnapshot"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage    storage.Storage
		stateStorage    storage.Storage
		snapshotStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage, snapshotStorage: lSnapshotStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage, snapshotStorage: pSnapshotStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			ledger, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, tc.snapshotStorage, blockFile)
			assert.Nil(t, err)

			addr := types.NewAddressByStr("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)
			account := NewAccount(1, stateLedger.cachedDB, stateLedger.accountCache, addr, NewChanger(), stateLedger.snapshot)

			code := []byte("code")
			codeHash := []byte("codeHash")
			account.originCode = nil
			account.originAccount = &types.InnerAccount{CodeHash: codeHash}
			account.dirtyAccount = &types.InnerAccount{CodeHash: nil}
			account.ldb.Put(utils.CompositeCodeKey(account.Addr, account.originAccount.CodeHash), code)

			jnl := account.getAccountJournal()
			assert.NotNil(t, jnl)
			assert.Equal(t, account.originCode, code)
		})
	}
}
