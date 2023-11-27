package ledger

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	etherTypes "github.com/ethereum/go-ethereum/core/types"
	crypto1 "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/blockfile"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/storage/pebble"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestNew001(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			logger := log.NewWithModule("account_test")
			addr := types.NewAddress(LeftPadBytes([]byte{100}, 20))
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			l, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, blockFile)
			require.Nil(t, err)
			require.NotNil(t, l)
			sl := l.StateLedger.(*StateLedgerImpl)

			sl.blockHeight = 1
			sl.SetNonce(addr, 1)
			rootHash1, err := sl.Commit()
			require.Nil(t, err)

			sl.blockHeight = 2
			rootHash2, err := sl.Commit()
			require.Nil(t, err)

			sl.blockHeight = 3
			sl.SetNonce(addr, 3)
			rootHash3, err := sl.Commit()
			require.Nil(t, err)

			assert.Equal(t, rootHash1, rootHash2)
			assert.NotEqual(t, rootHash1, rootHash3)

			l.Close()
		})
	}
}

func TestNew002(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			tc.blockStorage.Put([]byte(chainMetaKey), []byte{1})
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			l, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, blockFile)
			require.NotNil(t, err)
			require.Nil(t, l)
		})
	}
}

func TestNew003(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			tc.stateStorage.Put(compositeKey(journalKey, maxHeightStr), marshalHeight(1))
			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			l, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, tc.stateStorage, blockFile)
			require.NotNil(t, err)
			require.Nil(t, l)
		})
	}
}

func TestNew004(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			kvdb := tc.stateStorage
			kvdb.Put(compositeKey(journalKey, maxHeightStr), marshalHeight(1))

			journal := &BlockJournal{}
			data, err := json.Marshal(journal)
			assert.Nil(t, err)

			kvdb.Put(compositeKey(journalKey, 1), data)

			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			l, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, kvdb, blockFile)
			require.Nil(t, err)
			require.NotNil(t, l)
		})
	}
}

func TestNew005(t *testing.T) {
	repoRoot := t.TempDir()

	lBlockStorage, err := leveldb.New(filepath.Join(repoRoot, "lStorage"), nil)
	assert.Nil(t, err)
	lStateStorage, err := leveldb.New(filepath.Join(repoRoot, "lLedger"), nil)
	assert.Nil(t, err)
	pBlockStorage, err := pebble.New(filepath.Join(repoRoot, "pStorage"), nil, nil)
	assert.Nil(t, err)
	pStateStorage, err := pebble.New(filepath.Join(repoRoot, "pLedger"), nil, nil)
	assert.Nil(t, err)

	testcase := map[string]struct {
		blockStorage storage.Storage
		stateStorage storage.Storage
	}{
		"leveldb": {blockStorage: lBlockStorage, stateStorage: lStateStorage},
		"pebble":  {blockStorage: pBlockStorage, stateStorage: pStateStorage},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			kvdb := tc.stateStorage
			kvdb.Put(compositeKey(journalKey, maxHeightStr), marshalHeight(5))
			kvdb.Put(compositeKey(journalKey, minHeightStr), marshalHeight(3))

			journal := &BlockJournal{}
			data, err := json.Marshal(journal)
			assert.Nil(t, err)

			kvdb.Put(compositeKey(journalKey, 5), data)

			logger := log.NewWithModule("account_test")
			blockFile, err := blockfile.NewBlockFile(filepath.Join(repoRoot, name), logger)
			assert.Nil(t, err)
			l, err := NewLedgerWithStores(createMockRepo(t), tc.blockStorage, kvdb, blockFile)
			require.NotNil(t, err)
			require.Nil(t, l)
		})
	}
}

func TestChainLedger_PersistBlockData(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			ledger.StateLedger.(*StateLedgerImpl).blockHeight = 1

			// create an account
			account := types.NewAddress(LeftPadBytes([]byte{100}, 20))

			ledger.StateLedger.SetState(account, []byte("a"), []byte("b"))
			stateRoot, err := ledger.StateLedger.Commit()
			assert.Nil(t, err)
			ledger.PersistBlockData(genBlockData(1, stateRoot))
		})
	}
}

func TestChainLedger_Commit(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			lg, repoRoot := initLedger(t, "", tc.kvType)
			sl := lg.StateLedger.(*StateLedgerImpl)

			// create an account
			account := types.NewAddress(LeftPadBytes([]byte{100}, 20))

			sl.blockHeight = 1
			sl.SetState(account, []byte("a"), []byte("b"))
			sl.Finalise()
			stateRoot1, err := sl.Commit()
			assert.NotNil(t, stateRoot1)
			assert.Nil(t, err)
			sl.GetCommittedState(account, []byte("a"))
			isSuicide := sl.HasSuicide(account)
			assert.Equal(t, isSuicide, false)
			assert.Equal(t, uint64(1), sl.Version())

			sl.blockHeight = 2
			sl.Finalise()
			stateRoot2, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), sl.Version())
			assert.Equal(t, stateRoot1, stateRoot2)

			sl.SetState(account, []byte("a"), []byte("3"))
			sl.SetState(account, []byte("a"), []byte("2"))
			sl.blockHeight = 3
			sl.Finalise()
			stateRoot3, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(3), lg.StateLedger.Version())
			assert.NotEqual(t, stateRoot1, stateRoot3)

			lg.StateLedger.SetBalance(account, new(big.Int).SetInt64(100))
			sl.blockHeight = 4
			stateRoot4, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(4), lg.StateLedger.Version())
			assert.NotEqual(t, stateRoot3, stateRoot4)

			code := RightPadBytes([]byte{100}, 100)
			lg.StateLedger.SetCode(account, code)
			lg.StateLedger.SetState(account, []byte("b"), []byte("3"))
			lg.StateLedger.SetState(account, []byte("c"), []byte("2"))
			sl.blockHeight = 5
			sl.Finalise()
			stateRoot5, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(5), lg.StateLedger.Version())
			assert.NotEqual(t, stateRoot4, stateRoot5)
			// assert.Equal(t, uint64(5), ledger.maxJnlHeight)

			minHeight, maxHeight := getJournalRange(sl.cachedDB)
			journal5 := getBlockJournal(maxHeight, sl.cachedDB)
			assert.Equal(t, uint64(1), minHeight)
			assert.Equal(t, uint64(5), maxHeight)
			assert.Equal(t, 1, len(journal5.Journals))
			entry := journal5.Journals[0]
			assert.Equal(t, account.String(), entry.Address.String())
			assert.True(t, entry.AccountChanged)
			assert.Equal(t, uint64(100), entry.PrevAccount.Balance.Uint64())
			assert.Equal(t, uint64(0), entry.PrevAccount.Nonce)
			assert.Nil(t, entry.PrevAccount.CodeHash)
			assert.Equal(t, 2, len(entry.PrevStates))
			assert.Nil(t, entry.PrevStates[hex.EncodeToString([]byte("b"))])
			assert.Nil(t, entry.PrevStates[hex.EncodeToString([]byte("c"))])
			assert.True(t, entry.CodeChanged)
			assert.Nil(t, entry.PrevCode)
			isExist := sl.Exist(account)
			assert.True(t, isExist)
			isEmpty := sl.Empty(account)
			assert.False(t, isEmpty)
			err = sl.removeJournalsBeforeBlock(10)
			assert.NotNil(t, err)
			err = sl.removeJournalsBeforeBlock(0)
			assert.Nil(t, err)

			// Extra Test
			hash := types.NewHashByStr("0xe9FC370DD36C9BD5f67cCfbc031C909F53A3d8bC7084C01362c55f2D42bA841c")
			revid := lg.StateLedger.(*StateLedgerImpl).Snapshot()
			lg.StateLedger.(*StateLedgerImpl).logs.thash = hash
			lg.StateLedger.(*StateLedgerImpl).AddLog(&types.EvmLog{
				TransactionHash: lg.StateLedger.(*StateLedgerImpl).logs.thash,
			})
			lg.StateLedger.(*StateLedgerImpl).GetLogs(*lg.StateLedger.(*StateLedgerImpl).logs.thash, 1, hash)
			lg.StateLedger.(*StateLedgerImpl).Logs()
			lg.StateLedger.(*StateLedgerImpl).GetCodeHash(account)
			lg.StateLedger.(*StateLedgerImpl).GetCodeSize(account)
			currentAccount := lg.StateLedger.(*StateLedgerImpl).GetAccount(account)
			lg.StateLedger.(*StateLedgerImpl).setAccount(currentAccount)
			lg.StateLedger.(*StateLedgerImpl).AddBalance(account, big.NewInt(1))
			lg.StateLedger.(*StateLedgerImpl).SubBalance(account, big.NewInt(1))
			lg.StateLedger.(*StateLedgerImpl).SetNonce(account, 1)
			lg.StateLedger.(*StateLedgerImpl).AddRefund(1)
			refund := lg.StateLedger.(*StateLedgerImpl).GetRefund()
			assert.Equal(t, refund, uint64(1))
			lg.StateLedger.(*StateLedgerImpl).SubRefund(1)
			refund = lg.StateLedger.(*StateLedgerImpl).GetRefund()
			assert.Equal(t, refund, uint64(0))
			lg.StateLedger.(*StateLedgerImpl).AddAddressToAccessList(*account)
			isInAddressList := lg.StateLedger.(*StateLedgerImpl).AddressInAccessList(*account)
			assert.Equal(t, isInAddressList, true)
			lg.StateLedger.(*StateLedgerImpl).AddSlotToAccessList(*account, *hash)
			isInSlotAddressList, _ := lg.StateLedger.(*StateLedgerImpl).SlotInAccessList(*account, *hash)
			assert.Equal(t, isInSlotAddressList, true)
			lg.StateLedger.(*StateLedgerImpl).AddPreimage(*hash, []byte("11"))
			lg.StateLedger.(*StateLedgerImpl).PrepareAccessList(*account, account, []types.Address{}, AccessTupleList{})
			lg.StateLedger.(*StateLedgerImpl).Suicide(account)
			lg.StateLedger.(*StateLedgerImpl).RevertToSnapshot(revid)
			lg.StateLedger.(*StateLedgerImpl).ClearChangerAndRefund()

			lg.ChainLedger.CloseBlockfile()

			// load ChainLedgerImpl from db, rollback to height 0 since no chain meta stored
			ldg, _ := initLedger(t, repoRoot, tc.kvType)
			stateLedger := ldg.StateLedger.(*StateLedgerImpl)
			assert.Equal(t, uint64(0), stateLedger.maxJnlHeight)
			assert.Equal(t, &types.Hash{}, stateLedger.prevJnlHash)

			ok, _ := ldg.StateLedger.GetState(account, []byte("a"))
			assert.False(t, ok)

			ok, _ = ldg.StateLedger.GetState(account, []byte("b"))
			assert.False(t, ok)

			ok, _ = ldg.StateLedger.GetState(account, []byte("c"))
			assert.False(t, ok)

			assert.Equal(t, uint64(0), ldg.StateLedger.GetBalance(account).Uint64())
			assert.Equal(t, []byte(nil), ldg.StateLedger.GetCode(account))

			ver := ldg.StateLedger.Version()
			assert.Equal(t, uint64(0), ver)
			err = lg.StateLedger.(*StateLedgerImpl).removeJournalsBeforeBlock(4)
			assert.Nil(t, err)
		})
	}
}

func TestChainLedger_EVMAccessor(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)

			hash := common.HexToHash("0xe9FC370DD36C9BD5f67cCfbc031C909F53A3d8bC7084C01362c55f2D42bA841c")
			// create an account
			account := common.BytesToAddress(LeftPadBytes([]byte{100}, 20))

			ledger.StateLedger.(*StateLedgerImpl).CreateEVMAccount(account)
			ledger.StateLedger.(*StateLedgerImpl).AddEVMBalance(account, big.NewInt(2))
			balance := ledger.StateLedger.(*StateLedgerImpl).GetEVMBalance(account)
			assert.Equal(t, balance, big.NewInt(2))
			ledger.StateLedger.(*StateLedgerImpl).SubEVMBalance(account, big.NewInt(1))
			balance = ledger.StateLedger.(*StateLedgerImpl).GetEVMBalance(account)
			assert.Equal(t, balance, big.NewInt(1))
			ledger.StateLedger.(*StateLedgerImpl).SetEVMNonce(account, 10)
			nonce := ledger.StateLedger.(*StateLedgerImpl).GetEVMNonce(account)
			assert.Equal(t, nonce, uint64(10))
			ledger.StateLedger.(*StateLedgerImpl).GetEVMCodeHash(account)
			ledger.StateLedger.(*StateLedgerImpl).SetEVMCode(account, []byte("111"))
			code := ledger.StateLedger.(*StateLedgerImpl).GetEVMCode(account)
			assert.Equal(t, code, []byte("111"))
			codeSize := ledger.StateLedger.(*StateLedgerImpl).GetEVMCodeSize(account)
			assert.Equal(t, codeSize, 3)
			ledger.StateLedger.(*StateLedgerImpl).AddEVMRefund(2)
			refund := ledger.StateLedger.(*StateLedgerImpl).GetEVMRefund()
			assert.Equal(t, refund, uint64(2))
			ledger.StateLedger.(*StateLedgerImpl).SubEVMRefund(1)
			refund = ledger.StateLedger.(*StateLedgerImpl).GetEVMRefund()
			assert.Equal(t, refund, uint64(1))
			ledger.StateLedger.(*StateLedgerImpl).GetEVMCommittedState(account, hash)
			ledger.StateLedger.(*StateLedgerImpl).SetEVMState(account, hash, hash)
			value := ledger.StateLedger.(*StateLedgerImpl).GetEVMState(account, hash)
			assert.Equal(t, value, hash)
			ledger.StateLedger.(*StateLedgerImpl).SuicideEVM(account)
			isSuicide := ledger.StateLedger.(*StateLedgerImpl).HasSuicideEVM(account)
			assert.Equal(t, isSuicide, false)
			isExist := ledger.StateLedger.(*StateLedgerImpl).ExistEVM(account)
			assert.Equal(t, isExist, true)
			isEmpty := ledger.StateLedger.(*StateLedgerImpl).EmptyEVM(account)
			assert.Equal(t, isEmpty, false)
			ledger.StateLedger.(*StateLedgerImpl).PrepareEVMAccessList(account, &account, []common.Address{}, etherTypes.AccessList{})
			ledger.StateLedger.(*StateLedgerImpl).AddAddressToEVMAccessList(account)
			isIn := ledger.StateLedger.(*StateLedgerImpl).AddressInEVMAccessList(account)
			assert.Equal(t, isIn, true)
			ledger.StateLedger.(*StateLedgerImpl).AddSlotToEVMAccessList(account, hash)
			isSlotIn, _ := ledger.StateLedger.(*StateLedgerImpl).SlotInEVMAceessList(account, hash)
			assert.Equal(t, isSlotIn, true)
			ledger.StateLedger.(*StateLedgerImpl).AddEVMPreimage(hash, []byte("1111"))
			// ledger.StateLedgerImpl.(*SimpleLedger).PrepareEVM(hash, 1)
			ledger.StateLedger.(*StateLedgerImpl).StateDB()
			ledger.StateLedger.SetTxContext(types.NewHash(hash.Bytes()), 1)
			ledger.StateLedger.(*StateLedgerImpl).AddEVMLog(&etherTypes.Log{})

			addr := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
			ledger.StateLedger.(*StateLedgerImpl).transientStorage = newTransientStorage()
			ledger.StateLedger.(*StateLedgerImpl).SetEVMTransientState(addr, hash, hash)
			ledger.StateLedger.(*StateLedgerImpl).GetEVMTransientState(addr, hash)
			_ = ledger.StateLedger.(*StateLedgerImpl).transientStorage.Copy()
			ledger.StateLedger.PrepareEVM(params.Rules{IsBerlin: true}, addr, addr, &addr, []common.Address{addr}, nil)
		})
	}
}

func TestChainLedger_Rollback(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, repoRoot := initLedger(t, "", tc.kvType)
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)

			// create an addr0
			addr0 := types.NewAddress(LeftPadBytes([]byte{100}, 20))
			addr1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))

			hash0 := types.Hash{}
			assert.Equal(t, &hash0, stateLedger.prevJnlHash)

			ledger.StateLedger.PrepareBlock(nil, nil, 1)
			ledger.StateLedger.SetBalance(addr0, new(big.Int).SetInt64(1))
			stateLedger.Finalise()
			stateRoot1, err := stateLedger.Commit()
			assert.Nil(t, err)
			assert.NotNil(t, stateRoot1)
			ledger.PersistBlockData(genBlockData(1, stateRoot1))

			ledger.StateLedger.PrepareBlock(stateRoot1, nil, 2)
			ledger.StateLedger.SetBalance(addr0, new(big.Int).SetInt64(2))
			ledger.StateLedger.SetState(addr0, []byte("a"), []byte("2"))

			code := sha256.Sum256([]byte("code"))
			ret := crypto1.Keccak256Hash(code[:])
			codeHash := ret.Bytes()
			ledger.StateLedger.SetCode(addr0, code[:])
			ledger.StateLedger.Finalise()

			stateRoot2, err := stateLedger.Commit()
			assert.Nil(t, err)
			ledger.PersistBlockData(genBlockData(2, stateRoot2))

			ledger.StateLedger.PrepareBlock(stateRoot2, nil, 3)
			account0 := ledger.StateLedger.GetAccount(addr0)
			assert.Equal(t, uint64(2), account0.GetBalance().Uint64())

			ledger.StateLedger.SetBalance(addr1, new(big.Int).SetInt64(3))
			ledger.StateLedger.SetBalance(addr0, new(big.Int).SetInt64(4))
			ledger.StateLedger.SetState(addr0, []byte("a"), []byte("3"))
			ledger.StateLedger.SetState(addr0, []byte("b"), []byte("4"))

			code1 := sha256.Sum256([]byte("code1"))
			ret1 := crypto1.Keccak256Hash(code1[:])
			codeHash1 := ret1.Bytes()
			ledger.StateLedger.SetCode(addr0, code1[:])
			ledger.StateLedger.Finalise()

			stateRoot3, err := stateLedger.Commit()
			assert.Nil(t, err)
			ledger.PersistBlockData(genBlockData(3, stateRoot3))

			block, err := ledger.ChainLedger.GetBlock(3)
			assert.Nil(t, err)
			assert.NotNil(t, block)
			assert.Equal(t, uint64(3), ledger.ChainLedger.GetChainMeta().Height)

			account0 = ledger.StateLedger.GetAccount(addr0)
			assert.Equal(t, uint64(4), account0.GetBalance().Uint64())

			err = ledger.Rollback(4)
			assert.True(t, errors.Is(err, ErrNotFound))

			hash := ledger.ChainLedger.GetBlockHash(3)
			assert.NotNil(t, hash)

			hash = ledger.ChainLedger.GetBlockHash(100)
			assert.NotNil(t, hash)

			num, err := ledger.ChainLedger.GetTransactionCount(0)
			assert.NotNil(t, err)
			num, err = ledger.ChainLedger.GetTransactionCount(3)
			assert.Nil(t, err)
			assert.NotNil(t, num)

			meta, err := ledger.ChainLedger.LoadChainMeta()
			assert.Nil(t, err)
			assert.NotNil(t, meta)
			assert.Equal(t, uint64(3), meta.Height)

			//
			// err = ledger.Rollback(0)
			// assert.Equal(t, ErrorRollbackTooMuch, err)
			//
			// err = ledger.Rollback(1)
			// assert.Equal(t, ErrorRollbackTooMuch, err)
			// assert.Equal(t, uint64(3), ledger.GetChainMeta().Height)

			err = ledger.Rollback(3)
			assert.Nil(t, err)
			block3, err := ledger.ChainLedger.GetBlock(3)
			assert.Nil(t, err)
			assert.NotNil(t, block3)
			assert.Equal(t, stateRoot3, block3.BlockHeader.StateRoot)
			assert.Equal(t, uint64(3), ledger.ChainLedger.GetChainMeta().Height)
			assert.Equal(t, codeHash1, account0.CodeHash())
			assert.Equal(t, code1[:], account0.Code())

			err = ledger.Rollback(2)
			assert.Nil(t, err)
			block, err = ledger.ChainLedger.GetBlock(3)
			assert.Equal(t, ErrNotFound, err)
			assert.Nil(t, block)
			block2, err := ledger.ChainLedger.GetBlock(2)
			assert.Nil(t, err)
			assert.NotNil(t, block2)
			assert.Equal(t, uint64(2), ledger.ChainLedger.GetChainMeta().Height)
			assert.Equal(t, stateRoot2.String(), block2.BlockHeader.StateRoot.String())
			assert.Equal(t, uint64(1), stateLedger.minJnlHeight)
			assert.Equal(t, uint64(2), stateLedger.maxJnlHeight)

			account0 = ledger.StateLedger.GetAccount(addr0)
			assert.Equal(t, uint64(2), account0.GetBalance().Uint64())
			assert.Equal(t, uint64(0), account0.GetNonce())
			assert.Equal(t, codeHash[:], account0.CodeHash())
			assert.Equal(t, code[:], account0.Code())
			ok, val := account0.GetState([]byte("a"))
			assert.True(t, ok)
			assert.Equal(t, []byte("2"), val)

			account1 := ledger.StateLedger.GetAccount(addr1)
			assert.Nil(t, account1)

			ledger.ChainLedger.GetChainMeta()
			ledger.ChainLedger.CloseBlockfile()

			ledger, _ = initLedger(t, repoRoot, tc.kvType)
			assert.Equal(t, uint64(1), stateLedger.minJnlHeight)
			assert.Equal(t, uint64(2), stateLedger.maxJnlHeight)

			err = ledger.Rollback(1)
			assert.Nil(t, err)
			err = ledger.Rollback(0)
			assert.Nil(t, err)

			err = ledger.Rollback(100)
			assert.NotNil(t, err)
		})
	}
}

func TestChainLedger_GetAccount(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)

			addr := types.NewAddress(LeftPadBytes([]byte{1}, 20))
			code := LeftPadBytes([]byte{1}, 120)
			key0 := []byte{100, 100}
			key1 := []byte{100, 101}

			account := ledger.StateLedger.GetOrCreateAccount(addr)
			account.SetBalance(new(big.Int).SetInt64(1))
			account.SetNonce(2)
			account.SetCodeAndHash(code)

			account.SetState(key0, key1)
			account.SetState(key1, key0)

			stateLedger.blockHeight = 1
			stateLedger.Finalise()
			stateRoot, err := stateLedger.Commit()
			assert.Nil(t, err)
			assert.NotNil(t, stateRoot)

			account1 := ledger.StateLedger.GetAccount(addr)

			assert.Equal(t, account.GetBalance(), ledger.StateLedger.GetBalance(addr))
			assert.Equal(t, account.GetBalance(), account1.GetBalance())
			assert.Equal(t, account.GetNonce(), account1.GetNonce())
			assert.Equal(t, account.CodeHash(), account1.CodeHash())
			assert.Equal(t, account.Code(), account1.Code())
			ok0, val0 := account.GetState(key0)
			ok1, val1 := account.GetState(key1)
			assert.Equal(t, ok0, ok1)
			assert.Equal(t, val0, key1)
			assert.Equal(t, val1, key0)

			key2 := []byte{100, 102}
			val2 := []byte{111}
			ledger.StateLedger.SetState(addr, key0, val0)
			ledger.StateLedger.SetState(addr, key2, val2)
			ledger.StateLedger.SetState(addr, key0, val1)
			stateLedger.blockHeight = 2
			stateLedger.Finalise()
			stateRoot, err = stateLedger.Commit()
			assert.Nil(t, err)
			assert.NotNil(t, stateRoot)

			ledger.StateLedger.SetState(addr, key0, val0)
			ledger.StateLedger.SetState(addr, key0, val1)
			ledger.StateLedger.SetState(addr, key2, nil)
			stateLedger.blockHeight = 3
			stateLedger.Finalise()
			stateRoot, err = stateLedger.Commit()
			assert.Nil(t, err)
			assert.NotNil(t, stateRoot)

			ok, val := ledger.StateLedger.GetState(addr, key0)
			assert.True(t, ok)
			assert.Equal(t, val1, val)

			ok, val2 = ledger.StateLedger.GetState(addr, key2)
			assert.False(t, ok)
			assert.Nil(t, val2)
		})
	}
}

func TestChainLedger_GetCode(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			stateLedger := ledger.StateLedger.(*StateLedgerImpl)

			addr := types.NewAddress(LeftPadBytes([]byte{1}, 20))
			code := LeftPadBytes([]byte{10}, 120)

			code0 := ledger.StateLedger.GetCode(addr)
			assert.Nil(t, code0)

			ledger.StateLedger.SetCode(addr, code)

			stateLedger.blockHeight = 1
			stateRoot, err := stateLedger.Commit()
			assert.Nil(t, err)
			assert.NotNil(t, stateRoot)

			vals := ledger.StateLedger.GetCode(addr)
			assert.Equal(t, code, vals)
		})
	}
}

func TestChainLedger_AddState(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)

			account := types.NewAddress(LeftPadBytes([]byte{100}, 20))
			key0 := "100"
			value0 := []byte{100}
			ledger.StateLedger.(*StateLedgerImpl).blockHeight = 1
			ledger.StateLedger.SetState(account, []byte(key0), value0)
			ledger.StateLedger.Finalise()
			rootHash, err := ledger.StateLedger.Commit()
			assert.Nil(t, err)

			ledger.PersistBlockData(genBlockData(1, rootHash))
			require.Equal(t, uint64(1), ledger.StateLedger.Version())

			ok, val := ledger.StateLedger.GetState(account, []byte(key0))
			assert.True(t, ok)
			assert.Equal(t, value0, val)

			key1 := "101"
			value0 = []byte{99}
			value1 := []byte{101}
			ledger.StateLedger.(*StateLedgerImpl).blockHeight = 2
			ledger.StateLedger.SetState(account, []byte(key0), value0)
			ledger.StateLedger.SetState(account, []byte(key1), value1)
			ledger.StateLedger.Finalise()
			rootHash, err = ledger.StateLedger.Commit()
			assert.Nil(t, err)

			ledger.PersistBlockData(genBlockData(2, rootHash))
			require.Equal(t, uint64(2), ledger.StateLedger.Version())

			ok, val = ledger.StateLedger.GetState(account, []byte(key0))
			assert.True(t, ok)
			assert.Equal(t, value0, val)

			ok, val = ledger.StateLedger.GetState(account, []byte(key1))
			assert.True(t, ok)
			assert.Equal(t, value1, val)
		})
	}
}

func TestGetBlockSign(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetBlockSign(uint64(0))
			assert.NotNil(t, err)
		})
	}
}

func TestGetBlockByHash(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetBlockByHash(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(blockHashKey, types.NewHash([]byte("1")).String()), []byte("1"))
			_, err = ledger.ChainLedger.GetBlockByHash(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestGetTransaction(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(transactionMetaKey, types.NewHash([]byte("1")).String()), []byte("1"))
			_, err = ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
			err = ledger.ChainLedger.(*ChainLedgerImpl).bf.AppendBlock(0, []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			require.Nil(t, err)
			_, err = ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestGetTransaction1(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			meta := types.TransactionMeta{
				BlockHeight: 0,
			}
			metaBytes, err := meta.Marshal()
			require.Nil(t, err)
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(transactionMetaKey, types.NewHash([]byte("1")).String()), metaBytes)
			_, err = ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
			err = ledger.ChainLedger.(*ChainLedgerImpl).bf.AppendBlock(0, []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			require.Nil(t, err)
			_, err = ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestGetTransactionMeta(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetTransactionMeta(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(transactionMetaKey, types.NewHash([]byte("1")).String()), []byte("1"))
			_, err = ledger.ChainLedger.GetTransactionMeta(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
			err = ledger.ChainLedger.(*ChainLedgerImpl).bf.AppendBlock(0, []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			require.Nil(t, err)
			_, err = ledger.ChainLedger.GetTransactionMeta(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestGetReceipt(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetReceipt(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(transactionMetaKey, types.NewHash([]byte("1")).String()), []byte("0"))
			_, err = ledger.ChainLedger.GetReceipt(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
			err = ledger.ChainLedger.(*ChainLedgerImpl).bf.AppendBlock(0, []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			require.Nil(t, err)
			_, err = ledger.ChainLedger.GetReceipt(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestGetReceipt1(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			_, err := ledger.ChainLedger.GetTransaction(types.NewHash([]byte("1")))
			assert.True(t, errors.Is(err, ErrNotFound))
			meta := types.TransactionMeta{
				BlockHeight: 0,
			}
			metaBytes, err := meta.Marshal()
			require.Nil(t, err)
			ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.Put(compositeKey(transactionMetaKey, types.NewHash([]byte("1")).String()), metaBytes)
			_, err = ledger.ChainLedger.GetReceipt(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
			err = ledger.ChainLedger.(*ChainLedgerImpl).bf.AppendBlock(0, []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			require.Nil(t, err)
			_, err = ledger.ChainLedger.GetReceipt(types.NewHash([]byte("1")))
			assert.NotNil(t, err)
		})
	}
}

func TestPrepare(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			ledger, _ := initLedger(t, "", tc.kvType)
			batch := ledger.ChainLedger.(*ChainLedgerImpl).blockchainStore.NewBatch()
			var transactions []*types.Transaction
			transaction, err := types.GenerateEmptyTransactionAndSigner()
			require.Nil(t, err)
			transactions = append(transactions, transaction)
			block := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number: uint64(0),
				},
				BlockHash:    types.NewHash([]byte{1}),
				Transactions: transactions,
			}
			_, err = ledger.ChainLedger.(*ChainLedgerImpl).prepareBlock(batch, block)
			require.Nil(t, err)
			var receipts []*types.Receipt
			receipt := &types.Receipt{
				TxHash: types.NewHash([]byte("1")),
			}
			receipts = append(receipts, receipt)
			_, err = ledger.ChainLedger.(*ChainLedgerImpl).prepareReceipts(batch, block, receipts)
			require.Nil(t, err)
			_, err = ledger.ChainLedger.(*ChainLedgerImpl).prepareTransactions(batch, block)
			require.Nil(t, err)

			bloomRes := CreateBloom(receipts)
			require.NotNil(t, bloomRes)
		})
	}
}

// =========================== Test History Ledger ===========================

func TestStateLedger_EOAHistory(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			lg, _ := initLedger(t, "", tc.kvType)
			sl := lg.StateLedger.(*StateLedgerImpl)

			// create an account
			account1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
			account2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
			account3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))

			// set EOA account data in block 1
			// account1: balance=101, nonce=0
			// account2: balance=201, nonce=0
			// account3: balance=301, nonce=0
			sl.blockHeight = 1
			sl.SetBalance(account1, new(big.Int).SetInt64(101))
			sl.SetBalance(account2, new(big.Int).SetInt64(201))
			sl.SetBalance(account3, new(big.Int).SetInt64(301))
			stateRoot1, err := sl.Commit()
			assert.NotNil(t, stateRoot1)
			assert.Nil(t, err)
			isSuicide := sl.HasSuicide(account1)
			assert.Equal(t, isSuicide, false)
			assert.Equal(t, uint64(1), sl.Version())

			// set EOA account data in block 2
			// account1: balance=102, nonce=12
			// account2: balance=201, nonce=22
			// account3: balance=302, nonce=32
			sl.blockHeight = 2
			sl.SetBalance(account1, new(big.Int).SetInt64(102))
			sl.SetBalance(account3, new(big.Int).SetInt64(302))
			sl.SetNonce(account1, 12)
			sl.SetNonce(account2, 22)
			sl.SetNonce(account3, 32)
			stateRoot2, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), sl.Version())
			assert.NotEqual(t, stateRoot1, stateRoot2)

			// set EOA account data in block 3
			// account1: balance=103, nonce=13
			// account2: balance=203, nonce=23
			// account3: balance=302, nonce=32
			sl.blockHeight = 3
			sl.SetBalance(account1, new(big.Int).SetInt64(103))
			sl.SetBalance(account2, new(big.Int).SetInt64(203))
			sl.SetNonce(account1, 13)
			sl.SetNonce(account2, 23)
			stateRoot3, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(3), sl.Version())
			assert.NotEqual(t, stateRoot2, stateRoot3)

			// set EOA account data in block 4 (same with block 3)
			// account1: balance=103, nonce=13
			// account2: balance=203, nonce=23
			// account3: balance=302, nonce=32
			sl.blockHeight = 4
			stateRoot4, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(4), sl.Version())
			assert.Equal(t, stateRoot3, stateRoot4)

			// set EOA account data in block 5
			// account1: balance=103, nonce=15
			// account2: balance=203, nonce=25
			// account3: balance=305, nonce=35
			sl.blockHeight = 5
			sl.SetBalance(account1, new(big.Int).SetInt64(103))
			sl.SetBalance(account2, new(big.Int).SetInt64(203))
			sl.SetBalance(account3, new(big.Int).SetInt64(305))
			sl.SetNonce(account1, 15)
			sl.SetNonce(account2, 25)
			sl.SetNonce(account3, 35)
			stateRoot5, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(5), sl.Version())
			assert.NotEqual(t, stateRoot4, stateRoot5)

			// check state ledger in block 1
			block1 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    1,
					StateRoot: stateRoot1,
				},
			}
			lg1 := sl.NewViewWithoutCache(block1)
			assert.Equal(t, uint64(101), lg1.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(201), lg1.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(301), lg1.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(0), lg1.GetNonce(account1))
			assert.Equal(t, uint64(0), lg1.GetNonce(account2))
			assert.Equal(t, uint64(0), lg1.GetNonce(account3))

			// check state ledger in block 2
			block2 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    2,
					StateRoot: stateRoot2,
				},
			}
			lg2 := sl.NewViewWithoutCache(block2)
			assert.Equal(t, uint64(102), lg2.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(201), lg2.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(302), lg2.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(12), lg2.GetNonce(account1))
			assert.Equal(t, uint64(22), lg2.GetNonce(account2))
			assert.Equal(t, uint64(32), lg2.GetNonce(account3))

			// check state ledger in block 3
			block3 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    3,
					StateRoot: stateRoot3,
				},
			}
			lg3 := sl.NewViewWithoutCache(block3)
			assert.Equal(t, uint64(103), lg3.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(203), lg3.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(302), lg3.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(13), lg3.GetNonce(account1))
			assert.Equal(t, uint64(23), lg3.GetNonce(account2))
			assert.Equal(t, uint64(32), lg3.GetNonce(account3))

			// check state ledger in block 4
			block4 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    4,
					StateRoot: stateRoot4,
				},
			}
			lg4 := sl.NewViewWithoutCache(block4)
			assert.Equal(t, uint64(103), lg4.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(203), lg4.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(302), lg4.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(13), lg4.GetNonce(account1))
			assert.Equal(t, uint64(23), lg4.GetNonce(account2))
			assert.Equal(t, uint64(32), lg4.GetNonce(account3))

			// check state ledger in block 5
			block5 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    5,
					StateRoot: stateRoot5,
				},
			}
			lg5 := sl.NewViewWithoutCache(block5)
			assert.Equal(t, uint64(103), lg5.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(203), lg5.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(305), lg5.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(15), lg5.GetNonce(account1))
			assert.Equal(t, uint64(25), lg5.GetNonce(account2))
			assert.Equal(t, uint64(35), lg5.GetNonce(account3))
		})
	}
}

func TestStateLedger_ContractStateHistory(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			lg, _ := initLedger(t, "", tc.kvType)
			sl := lg.StateLedger.(*StateLedgerImpl)

			// create an account
			account1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
			account2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
			account3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))

			// set contract account data in block 1
			// account1: key1=val101, key2=val102
			// account3: key1=val301, key2=val302
			sl.blockHeight = 1
			sl.SetState(account1, []byte("key1"), []byte("val101"))
			sl.SetState(account1, []byte("key2"), []byte("val102"))
			sl.SetState(account3, []byte("key1"), []byte("val301"))
			sl.SetState(account3, []byte("key2"), []byte("val302"))
			sl.Finalise()
			stateRoot1, err := sl.Commit()
			assert.NotNil(t, stateRoot1)
			assert.Nil(t, err)
			isSuicide := sl.HasSuicide(account1)
			assert.Equal(t, isSuicide, false)
			assert.Equal(t, uint64(1), sl.Version())

			// set contract account data in block 2
			// account1: key1=val1011, key2=val102
			// account2: key1=val201, key2=val202
			// account3: key1=val3011, key2=val3021
			sl.blockHeight = 2
			sl.SetState(account1, []byte("key1"), []byte("val1011"))
			sl.SetState(account2, []byte("key1"), []byte("val201"))
			sl.SetState(account2, []byte("key2"), []byte("val202"))
			sl.SetState(account3, []byte("key1"), []byte("val3011"))
			sl.SetState(account3, []byte("key2"), []byte("val3021"))
			sl.Finalise()
			stateRoot2, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), sl.Version())
			assert.NotEqual(t, stateRoot1, stateRoot2)

			// set contract account data in block 3
			// account1: key1=val1013, key2=val102
			// account2: key1=val2011, key2=nil
			// account3: key1=nil, key2=nil
			sl.blockHeight = 3
			sl.SetState(account1, []byte("key1"), []byte("val1013"))
			sl.SetState(account2, []byte("key1"), []byte("val2011"))
			sl.SetState(account2, []byte("key2"), nil)
			sl.SetState(account3, []byte("key1"), nil)
			sl.SetState(account3, []byte("key2"), nil)
			sl.Finalise()
			stateRoot3, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(3), sl.Version())
			assert.NotEqual(t, stateRoot2, stateRoot3)

			// set contract account data in block 4 (same with block 3)
			// account1: key1=val1013, key2=val102
			// account2: key1=val2011, key2=nil
			// account3: key1=nil, key2=nil
			sl.blockHeight = 4
			sl.Finalise()
			stateRoot4, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(4), sl.Version())
			assert.Equal(t, stateRoot3, stateRoot4)

			// set contract account data in block 5
			// account1: key1=val1015, key2=val1025
			// account2: key1=val2015, key2=val2025
			// account3: key1=val3015, key2=val3025
			sl.blockHeight = 5
			sl.SetState(account1, []byte("key1"), []byte("val1015"))
			sl.SetState(account1, []byte("key2"), []byte("val1025"))
			sl.SetState(account2, []byte("key1"), []byte("val2015"))
			sl.SetState(account2, []byte("key2"), []byte("val2025"))
			sl.SetState(account3, []byte("key1"), []byte("val3015"))
			sl.SetState(account3, []byte("key2"), []byte("val3025"))
			sl.Finalise()
			stateRoot5, err := sl.Commit()
			assert.Nil(t, err)
			assert.Equal(t, uint64(5), sl.Version())
			assert.NotEqual(t, stateRoot4, stateRoot5)

			// check state ledger in block 1
			block1 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    1,
					StateRoot: stateRoot1,
				},
			}
			lg1 := sl.NewViewWithoutCache(block1)
			exist, a1k1 := lg1.GetState(account1, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val101"), a1k1)
			exist, a1k2 := lg1.GetState(account1, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val102"), a1k2)
			exist, _ = lg1.GetState(account2, []byte("key1"))
			assert.False(t, exist)
			exist, a3k1 := lg1.GetState(account3, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val301"), a3k1)
			exist, a3k2 := lg1.GetState(account3, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val302"), a3k2)

			// check state ledger in block 2
			block2 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    2,
					StateRoot: stateRoot2,
				},
			}
			lg2 := sl.NewViewWithoutCache(block2)
			exist, a1k1 = lg2.GetState(account1, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1011"), a1k1)
			exist, a1k2 = lg2.GetState(account1, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val102"), a1k2)
			exist, a2k1 := lg2.GetState(account2, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val201"), a2k1)
			exist, a2k2 := lg2.GetState(account2, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val202"), a2k2)
			exist, a3k1 = lg2.GetState(account3, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val3011"), a3k1)
			exist, a3k2 = lg2.GetState(account3, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val3021"), a3k2)

			// check state ledger in block 3
			block3 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    3,
					StateRoot: stateRoot3,
				},
			}
			lg3 := sl.NewViewWithoutCache(block3)
			exist, a1k1 = lg3.GetState(account1, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1013"), a1k1)
			exist, a1k2 = lg3.GetState(account1, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val102"), a1k2)
			exist, a2k1 = lg3.GetState(account2, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val2011"), a2k1)
			exist, a2k2 = lg3.GetState(account2, []byte("key2"))
			assert.False(t, exist)
			exist, a3k1 = lg3.GetState(account3, []byte("key1"))
			assert.False(t, exist)
			exist, a3k2 = lg3.GetState(account3, []byte("key2"))
			assert.False(t, exist)

			// check state ledger in block 4
			block4 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    4,
					StateRoot: stateRoot4,
				},
			}
			lg4 := sl.NewViewWithoutCache(block4)
			exist, a1k1 = lg4.GetState(account1, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1013"), a1k1)
			exist, a1k2 = lg4.GetState(account1, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val102"), a1k2)
			exist, a2k1 = lg4.GetState(account2, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val2011"), a2k1)
			exist, a2k2 = lg4.GetState(account2, []byte("key2"))
			assert.False(t, exist)
			exist, a3k1 = lg4.GetState(account3, []byte("key1"))
			assert.False(t, exist)
			exist, a3k2 = lg4.GetState(account3, []byte("key2"))
			assert.False(t, exist)

			// check state ledger in block 5
			block5 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    5,
					StateRoot: stateRoot5,
				},
			}
			lg5 := sl.NewViewWithoutCache(block5)
			exist, a1k1 = lg5.GetState(account1, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1015"), a1k1)
			exist, a1k2 = lg5.GetState(account1, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1025"), a1k2)
			exist, a2k1 = lg5.GetState(account2, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val2015"), a2k1)
			exist, a2k2 = lg5.GetState(account2, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val2025"), a2k2)
			exist, a3k1 = lg5.GetState(account3, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val3015"), a3k1)
			exist, a3k2 = lg5.GetState(account3, []byte("key2"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val3025"), a3k2)
		})
	}
}

func TestStateLedger_ContractCodeHistory(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			lg, _ := initLedger(t, "", tc.kvType)
			sl := lg.StateLedger.(*StateLedgerImpl)

			// create an account
			eoaAccount := types.NewAddress(LeftPadBytes([]byte{101}, 20))
			contractAccount := types.NewAddress(LeftPadBytes([]byte{102}, 20))

			// set account data in block 1
			// account1: nonce=1
			// account2: key1=val1,code=code1
			sl.blockHeight = 1
			sl.SetNonce(eoaAccount, 1)
			code1 := sha256.Sum256([]byte("code1"))
			sl.SetState(contractAccount, []byte("key1"), []byte("val1"))
			sl.SetCode(contractAccount, code1[:])
			sl.Finalise()
			stateRoot1, err := sl.Commit()
			assert.NotNil(t, stateRoot1)
			assert.Nil(t, err)
			assert.Equal(t, uint64(1), sl.Version())

			// set account data in block 2
			// account1: nonce=1
			// account2: key1=val1,code=code2
			sl.blockHeight = 2
			code2 := sha256.Sum256([]byte("code2"))
			sl.SetCode(contractAccount, code2[:])
			sl.Finalise()
			stateRoot2, err := sl.Commit()
			assert.NotNil(t, stateRoot2)
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), sl.Version())
			assert.NotEqual(t, stateRoot1, stateRoot2)

			// set account data in block 3
			// account1: nonce=2
			// account2: key1=val1,code=code2
			sl.blockHeight = 3
			sl.SetNonce(eoaAccount, 2)
			sl.Finalise()
			stateRoot3, err := sl.Commit()
			assert.NotNil(t, stateRoot3)
			assert.Nil(t, err)
			assert.Equal(t, uint64(3), sl.Version())
			assert.NotEqual(t, stateRoot2, stateRoot3)

			// check state ledger in block 1
			block1 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    1,
					StateRoot: stateRoot1,
				},
			}
			lg1 := sl.NewViewWithoutCache(block1)
			assert.Equal(t, uint64(1), lg1.GetNonce(eoaAccount))
			exist, a2k1 := lg1.GetState(contractAccount, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1"), a2k1)
			assert.Equal(t, code1[:], lg1.GetCode(contractAccount))

			// check state ledger in block 2
			block2 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    2,
					StateRoot: stateRoot2,
				},
			}
			lg2 := sl.NewViewWithoutCache(block2)
			assert.Equal(t, uint64(1), lg2.GetNonce(eoaAccount))
			exist, a2k1 = lg2.GetState(contractAccount, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1"), a2k1)
			assert.Equal(t, code2[:], lg2.GetCode(contractAccount))

			// check state ledger in block 3
			block3 := &types.Block{
				BlockHeader: &types.BlockHeader{
					Number:    3,
					StateRoot: stateRoot3,
				},
			}
			lg3 := sl.NewViewWithoutCache(block3)
			assert.Equal(t, uint64(2), lg3.GetNonce(eoaAccount))
			exist, a2k1 = lg3.GetState(contractAccount, []byte("key1"))
			assert.True(t, exist)
			assert.Equal(t, []byte("val1"), a2k1)
			assert.Equal(t, code2[:], lg3.GetCode(contractAccount))
		})
	}
}

func TestStateLedger_RollbackToHistoryVersion(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			lg, _ := initLedger(t, "", tc.kvType)
			sl := lg.StateLedger.(*StateLedgerImpl)

			// create an account
			account1 := types.NewAddress(LeftPadBytes([]byte{101}, 20))
			account2 := types.NewAddress(LeftPadBytes([]byte{102}, 20))
			account3 := types.NewAddress(LeftPadBytes([]byte{103}, 20))

			// set EOA account data in block 1
			// account1: balance=101, nonce=0
			// account2: balance=201, nonce=0
			// account3: balance=301, nonce=0
			sl.blockHeight = 1
			sl.SetBalance(account1, new(big.Int).SetInt64(101))
			sl.SetBalance(account2, new(big.Int).SetInt64(201))
			sl.SetBalance(account3, new(big.Int).SetInt64(301))
			sl.Finalise()
			stateRoot1, err := sl.Commit()
			assert.NotNil(t, stateRoot1)
			assert.Nil(t, err)
			lg.PersistBlockData(genBlockData(1, stateRoot1))
			isSuicide := sl.HasSuicide(account1)
			assert.Equal(t, isSuicide, false)
			assert.Equal(t, uint64(1), sl.Version())

			// set EOA account data in block 2
			// account1: balance=102, nonce=12
			// account2: balance=201, nonce=22
			// account3: balance=302, nonce=32
			sl.blockHeight = 2
			sl.SetBalance(account1, new(big.Int).SetInt64(102))
			sl.SetBalance(account3, new(big.Int).SetInt64(302))
			sl.SetNonce(account1, 12)
			sl.SetNonce(account2, 22)
			sl.SetNonce(account3, 32)
			sl.Finalise()
			stateRoot2, err := sl.Commit()
			assert.Nil(t, err)
			lg.PersistBlockData(genBlockData(2, stateRoot2))
			assert.Equal(t, uint64(2), sl.Version())
			assert.NotEqual(t, stateRoot1, stateRoot2)

			// set EOA account data in block 3
			// account1: balance=103, nonce=13
			// account2: balance=203, nonce=23
			// account3: balance=302, nonce=32
			sl.blockHeight = 3
			sl.SetBalance(account1, new(big.Int).SetInt64(103))
			sl.SetBalance(account2, new(big.Int).SetInt64(203))
			sl.SetNonce(account1, 13)
			sl.SetNonce(account2, 23)
			sl.Finalise()
			stateRoot3, err := sl.Commit()
			assert.Nil(t, err)
			lg.PersistBlockData(genBlockData(3, stateRoot3))
			assert.Equal(t, uint64(3), sl.Version())
			assert.NotEqual(t, stateRoot2, stateRoot3)

			// revert from block 3 to block 2
			err = lg.Rollback(2)
			assert.Nil(t, err)

			// check state ledger in block 2
			lg = lg.NewViewWithoutCache()
			assert.Equal(t, uint64(102), lg.StateLedger.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(201), lg.StateLedger.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(302), lg.StateLedger.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(12), lg.StateLedger.GetNonce(account1))
			assert.Equal(t, uint64(22), lg.StateLedger.GetNonce(account2))
			assert.Equal(t, uint64(32), lg.StateLedger.GetNonce(account3))

			// revert from block 2 to block 1
			err = lg.Rollback(1)
			assert.Nil(t, err)

			// check state ledger in block 1
			assert.Equal(t, uint64(101), lg.StateLedger.GetBalance(account1).Uint64())
			assert.Equal(t, uint64(201), lg.StateLedger.GetBalance(account2).Uint64())
			assert.Equal(t, uint64(301), lg.StateLedger.GetBalance(account3).Uint64())
			assert.Equal(t, uint64(0), lg.StateLedger.GetNonce(account1))
			assert.Equal(t, uint64(0), lg.StateLedger.GetNonce(account2))
			assert.Equal(t, uint64(0), lg.StateLedger.GetNonce(account3))

			// revert from block 1 to block 3
			err = lg.Rollback(3)
			assert.NotNil(t, err)
		})
	}
}

func genBlockData(height uint64, stateRoot *types.Hash) *BlockData {
	block := &types.Block{
		BlockHeader: &types.BlockHeader{
			Number:    height,
			StateRoot: stateRoot,
		},
		Transactions: []*types.Transaction{},
	}

	block.BlockHash = block.Hash()
	return &BlockData{
		Block: &types.Block{
			BlockHeader: &types.BlockHeader{
				Number:    height,
				StateRoot: stateRoot,
			},
			BlockHash:    types.NewHash([]byte{1}),
			Transactions: []*types.Transaction{},
		},
		Receipts: nil,
	}
}

func createMockRepo(t *testing.T) *repo.Repo {
	r, err := repo.Default(t.TempDir())
	require.Nil(t, err)
	return r
}

func initLedger(t *testing.T, repoRoot string, kv string) (*Ledger, string) {
	rep := createMockRepo(t)
	if repoRoot != "" {
		rep.RepoRoot = repoRoot
	}

	err := storagemgr.Initialize(kv, repo.KVStorageCacheSize, repo.KVStorageSync)
	require.Nil(t, err)
	rep.Config.Monitor.EnableExpensive = true
	l, err := NewLedger(rep)
	require.Nil(t, err)

	return l, rep.RepoRoot
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

func RightPadBytes(slice []byte, l int) []byte {
	if l <= len(slice) {
		return slice
	}

	padded := make([]byte, l)
	copy(padded, slice)

	return padded
}

func TestEvmLogs(t *testing.T) {
	logs := NewEvmLogs()
	hash := types.NewHashByStr("0xe9FC370DD36C9BD5f67cCfbc031C909F53A3d8bC7084C01362c55f2D42bA841c")
	logs.SetBHash(hash)
	logs.SetTHash(hash)
	logs.SetIndex(1)
}

func BenchmarkStateLedgerWrite(b *testing.B) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		b.Run(name, func(b *testing.B) {
			r, _ := repo.Default(b.TempDir())
			storagemgr.Initialize(tc.kvType, 256, repo.KVStorageSync)
			l, _ := NewLedger(r)
			benchStateLedgerWrite(b, l.StateLedger)
		})
	}
}

func BenchmarkStateLedgerRead(b *testing.B) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		b.Run(name, func(b *testing.B) {
			r, _ := repo.Default(b.TempDir())
			storagemgr.Initialize(tc.kvType, 256, repo.KVStorageSync)
			l, _ := NewLedger(r)
			benchStateLedgerRead(b, l.StateLedger)
		})
	}
}

func benchStateLedgerWrite(b *testing.B, sl StateLedger) {
	var (
		keys, vals = makeDataset(5_000_000, 32, 32, false)
	)

	b.Run("Write", func(b *testing.B) {
		stateLedger := sl.(*StateLedgerImpl)
		addr := types.NewAddress(LeftPadBytes([]byte{1}, 20))
		for i := 0; i < len(keys); i++ {
			stateLedger.SetState(addr, keys[i], vals[i])
		}
		b.ResetTimer()
		b.ReportAllocs()
		stateLedger.blockHeight = 1
		stateLedger.Finalise()
		stateLedger.Commit()
	})
}

func benchStateLedgerRead(b *testing.B, sl StateLedger) {
	var (
		keys, vals = makeDataset(10_000_000, 32, 32, false)
	)
	stateLedger := sl.(*StateLedgerImpl)
	addr := types.NewAddress(LeftPadBytes([]byte{1}, 20))
	for i := 0; i < len(keys); i++ {
		stateLedger.SetState(addr, keys[i], vals[i])
	}
	stateLedger.blockHeight = 1
	stateLedger.Finalise()
	stateLedger.Commit()

	b.Run("Read", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < len(keys); i++ {
			stateLedger.GetState(addr, keys[i])
		}
	})
}

func makeDataset(size, ksize, vsize int, order bool) ([][]byte, [][]byte) {
	var keys [][]byte
	var vals [][]byte
	for i := 0; i < size; i++ {
		keys = append(keys, randBytes(ksize))
		vals = append(vals, randBytes(vsize))
	}

	// order generated slice according to bytes order
	if order {
		sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i], keys[j]) < 0 })
	}
	return keys, vals
}

// randomHash generates a random blob of data and returns it as a hash.
func randBytes(len int) []byte {
	buf := make([]byte, len)
	if n, err := rand.Read(buf); n != len || err != nil {
		panic(err)
	}
	return buf
}
