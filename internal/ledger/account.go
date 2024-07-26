package ledger

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage/kv"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/snapshot"
	"github.com/axiomesh/axiom-ledger/internal/ledger/utils"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
	"github.com/axiomesh/axiom-ledger/pkg/loggers"
)

var _ IAccount = (*SimpleAccount)(nil)

type bytesLazyLogger struct {
	bytes []byte
}

func (l *bytesLazyLogger) String() string {
	return hexutil.Encode(l.bytes)
}

type SimpleAccount struct {
	logger        logrus.FieldLogger
	Addr          *types.Address
	originAccount *types.InnerAccount
	dirtyAccount  *types.InnerAccount

	// TODO: use []byte as key
	// The confirmed state of the previous block
	originState map[string][]byte

	// Modified state of previous transactions in the current block
	pendingState map[string][]byte

	// The latest state of the current transaction
	dirtyState map[string][]byte

	originCode []byte
	dirtyCode  []byte

	backend          kv.Storage
	storageTrieCache *storagemgr.CacheWrapper
	pruneCache       jmt.PruneCache

	blockHeight uint64
	storageTrie *jmt.JMT

	changer        *stateChanger
	selfDestructed bool
	created        bool // Flag whether the account was created in the current transaction

	snapshot *snapshot.Snapshot
}

func NewMockAccount(blockHeight uint64, addr *types.Address) *SimpleAccount {
	ldb := kv.NewMemory()

	return &SimpleAccount{
		logger:           loggers.Logger(loggers.Storage),
		Addr:             addr,
		originState:      make(map[string][]byte),
		pendingState:     make(map[string][]byte),
		dirtyState:       make(map[string][]byte),
		blockHeight:      blockHeight,
		backend:          ldb,
		storageTrieCache: storagemgr.NewCacheWrapper(32, false),
		changer:          newChanger(),
		selfDestructed:   false,
		created:          false,
	}
}

func NewAccount(blockHeight uint64, backend kv.Storage, storageTrieCache *storagemgr.CacheWrapper, pruneCache jmt.PruneCache, addr *types.Address, changer *stateChanger, snapshot *snapshot.Snapshot) *SimpleAccount {
	return &SimpleAccount{
		logger:           loggers.Logger(loggers.Storage),
		Addr:             addr,
		originState:      make(map[string][]byte),
		pendingState:     make(map[string][]byte),
		dirtyState:       make(map[string][]byte),
		blockHeight:      blockHeight,
		backend:          backend,
		storageTrieCache: storageTrieCache,
		pruneCache:       pruneCache,
		changer:          changer,
		selfDestructed:   false,
		created:          false,
		snapshot:         snapshot,
	}
}

func (o *SimpleAccount) String() string {
	return fmt.Sprintf("{origin: %v, dirty: %v, code length: %v}", o.originAccount, o.dirtyAccount, len(o.Code()))
}

func (o *SimpleAccount) initStorageTrie() {
	if o.storageTrie != nil {
		return
	}
	// new account
	if o.originAccount == nil || o.originAccount.StorageRoot == (common.Hash{}) {
		rootHash := crypto.Keccak256Hash(o.Addr.ETHAddress().Bytes())
		rootNodeKey := &types.NodeKey{
			Version: o.blockHeight,
			Path:    []byte{},
			Type:    o.Addr.Bytes(),
		}
		nk := rootNodeKey.Encode()
		batch := o.backend.NewBatch()
		batch.Put(nk, nil)
		batch.Put(rootHash[:], nk)
		batch.Commit()
		trie, err := jmt.New(rootHash, o.backend, o.storageTrieCache, o.pruneCache, o.logger)
		if err != nil {
			panic(err)
		}
		o.storageTrie = trie
		o.logger.Debugf("[initStorageTrie] (new jmt) addr: %v, origin account: %v", o.Addr, o.originAccount)
		return
	}

	trie, err := jmt.New(o.originAccount.StorageRoot, o.backend, o.storageTrieCache, o.pruneCache, o.logger)
	if err != nil {
		panic(err)
	}
	o.storageTrie = trie
	o.logger.Debugf("[initStorageTrie] addr: %v, origin account: %v", o.Addr, o.originAccount)
}

func (o *SimpleAccount) GetAddress() *types.Address {
	return o.Addr
}

// GetState Get state from local cache, if not found, then get it from DB
func (o *SimpleAccount) GetState(key []byte) (bool, []byte) {
	if value, exist := o.dirtyState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from dirty, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	if value, exist := o.pendingState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	if value, exist := o.originState[string(key)]; exist {
		o.logger.Debugf("[GetState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return value != nil, value
	}

	if o.snapshot != nil {
		if value, err := o.snapshot.Storage(o.Addr, key); err == nil {
			o.originState[string(key)] = value
			o.initStorageTrie()
			o.logger.Debugf("[GetState] get from snapshot, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			return value != nil, value
		}
	}

	o.initStorageTrie()
	val, err := o.storageTrie.Get(utils.CompositeStorageKey(o.Addr, key))
	if err != nil {
		panic(err)
	}
	o.logger.Debugf("[GetState] get from storage trie, addr: %v, key: %v, state: %v", o.Addr, string(key), &bytesLazyLogger{bytes: val})

	o.originState[string(key)] = val

	return val != nil, val
}

func (o *SimpleAccount) GetCommittedState(key []byte) []byte {
	if value, exist := o.pendingState[string(key)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			return (&types.Hash{}).Bytes()
		}
		o.logger.Debugf("[GetCommittedState] get from pending, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return value
	}

	if value, exist := o.originState[string(key)]; exist {
		if value == nil {
			o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			return (&types.Hash{}).Bytes()
		}
		o.logger.Debugf("[GetCommittedState] get from origin, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
		return value
	}

	if o.snapshot != nil {
		if value, err := o.snapshot.Storage(o.Addr, key); err == nil {
			o.originState[string(key)] = value
			o.initStorageTrie()
			o.logger.Debugf("[GetCommittedState] get from snapshot, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			if value == nil {
				return (&types.Hash{}).Bytes()
			}
			return value
		}
	}

	o.initStorageTrie()
	val, err := o.storageTrie.Get(utils.CompositeStorageKey(o.Addr, key))
	if err != nil {
		panic(err)
	}
	o.logger.Debugf("[GetCommittedState] get from storage trie, addr: %v, key: %v, state: %v", o.Addr, string(key), &bytesLazyLogger{bytes: val})

	o.originState[string(key)] = val

	if val == nil {
		return (&types.Hash{}).Bytes()
	}
	return val
}

// SetState Set account state
func (o *SimpleAccount) SetState(key []byte, value []byte) {
	_, prev := o.GetState(key)
	o.changer.append(storageChange{
		account:  o.Addr,
		key:      key,
		prevalue: prev,
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.logger.Debugf("[SetState] addr: %v, key: %v, before state: %v, after state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: prev}, &bytesLazyLogger{bytes: value})
	o.setState(key, value)
}

func (o *SimpleAccount) setState(key []byte, value []byte) {
	o.dirtyState[string(key)] = value
}

// SetCodeAndHash Set the contract code and hash
func (o *SimpleAccount) SetCodeAndHash(code []byte) {
	ret := crypto.Keccak256Hash(code)
	o.changer.append(codeChange{
		account:  o.Addr,
		prevcode: o.Code(),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.logger.Debugf("[SetCodeAndHash] addr: %v, before code hash: %v, after code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.CodeHash()}, &bytesLazyLogger{bytes: ret.Bytes()})
	o.dirtyAccount.CodeHash = ret.Bytes()
	o.dirtyCode = code
}

func (o *SimpleAccount) setCodeAndHash(code []byte) {
	ret := crypto.Keccak256Hash(code)
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.CodeHash = ret.Bytes()
	o.dirtyCode = code
}

// Code return the contract code
func (o *SimpleAccount) Code() []byte {
	if o.dirtyCode != nil {
		o.logger.Debugf("[Code] get from dirty, addr: %v, code: %v", o.Addr, &bytesLazyLogger{bytes: o.dirtyCode[:4]})
		return o.dirtyCode
	}

	if o.originCode != nil {
		o.logger.Debugf("[Code] get from origin, addr: %v, code: %v", o.Addr, &bytesLazyLogger{bytes: o.originCode[:4]})
		return o.originCode
	}

	if bytes.Equal(o.CodeHash(), nil) {
		return nil
	}

	code := o.backend.Get(utils.CompositeCodeKey(o.Addr, o.CodeHash()))
	if code != nil {
		o.logger.Debugf("[Code] get from storage, addr: %v, code: %v", o.Addr, &bytesLazyLogger{bytes: code[:4]})
	}

	o.originCode = code
	o.dirtyCode = code

	return code
}

func (o *SimpleAccount) CodeHash() []byte {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[CodeHash] get from dirty, addr: %v, code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.dirtyAccount.CodeHash})
		return o.dirtyAccount.CodeHash
	}
	if o.originAccount != nil {
		o.logger.Debugf("[CodeHash] get from origin, addr: %v, code hash: %v", o.Addr, &bytesLazyLogger{bytes: o.originAccount.CodeHash})
		return o.originAccount.CodeHash
	}
	o.logger.Debugf("[CodeHash] not found, addr: %v", o.Addr)
	return nil
}

// SetNonce Set the nonce which indicates the contract number
func (o *SimpleAccount) SetNonce(nonce uint64) {
	o.changer.append(nonceChange{
		account: o.Addr,
		prev:    o.GetNonce(),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	if o.originAccount == nil {
		o.logger.Debugf("[SetNonce] addr: %v, before origin nonce: %v, before dirty nonce: %v, after dirty nonce: %v", o.Addr, 0, o.dirtyAccount.Nonce, nonce)
	} else {
		o.logger.Debugf("[SetNonce] addr: %v, before origin nonce: %v, before dirty nonce: %v, after dirty nonce: %v", o.Addr, o.originAccount.Nonce, o.dirtyAccount.Nonce, nonce)
	}
	o.dirtyAccount.Nonce = nonce
}

func (o *SimpleAccount) setNonce(nonce uint64) {
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.Nonce = nonce
}

// GetNonce Get the nonce from user account
func (o *SimpleAccount) GetNonce() uint64 {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[GetNonce] get from dirty, addr: %v, nonce: %v", o.Addr, o.dirtyAccount.Nonce)
		return o.dirtyAccount.Nonce
	}
	if o.originAccount != nil {
		o.logger.Debugf("[GetNonce] get from origin, addr: %v, nonce: %v", o.Addr, o.originAccount.Nonce)
		return o.originAccount.Nonce
	}
	o.logger.Debugf("[GetNonce] not found, addr: %v, nonce: 0", o.Addr)
	return 0
}

// GetBalance Get the balance from the account
func (o *SimpleAccount) GetBalance() *big.Int {
	if o.dirtyAccount != nil {
		o.logger.Debugf("[GetBalance] get from dirty, addr: %v, balance: %v", o.Addr, o.dirtyAccount.Balance)
		return o.dirtyAccount.Balance
	}
	if o.originAccount != nil {
		o.logger.Debugf("[GetBalance] get from origin, addr: %v, balance: %v", o.Addr, o.originAccount.Balance)
		return o.originAccount.Balance
	}
	o.logger.Debugf("[GetBalance] not found, addr: %v, balance: 0", o.Addr)
	return new(big.Int).SetInt64(0)
}

// SetBalance Set the balance to the account
func (o *SimpleAccount) SetBalance(balance *big.Int) {
	o.changer.append(balanceChange{
		account: o.Addr,
		prev:    new(big.Int).Set(o.GetBalance()),
	})
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	if o.originAccount == nil {
		o.logger.Debugf("[SetBalance] addr: %v, before origin balance: %v, before dirty balance: %v, after dirty balance: %v", o.Addr, 0, o.dirtyAccount.Balance, balance)
	} else {
		o.logger.Debugf("[SetBalance] addr: %v, before origin balance: %v, before dirty balance: %v, after dirty balance: %v", o.Addr, o.originAccount.Balance, o.dirtyAccount.Balance, balance)
	}
	o.dirtyAccount.Balance = balance
}

func (o *SimpleAccount) setBalance(balance *big.Int) {
	if o.dirtyAccount == nil {
		o.dirtyAccount = o.originAccount.CopyOrNewIfEmpty()
	}
	o.dirtyAccount.Balance = balance
}

func (o *SimpleAccount) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	o.logger.Debugf("[SubBalance] addr: %v, sub amount: %v", o.Addr, amount)
	o.SetBalance(new(big.Int).Sub(o.GetBalance(), amount))
}

func (o *SimpleAccount) AddBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	o.logger.Debugf("[AddBalance] addr: %v, add amount: %v", o.Addr, amount)
	o.SetBalance(new(big.Int).Add(o.GetBalance(), amount))
}

// Finalise moves all dirty states into the pending states.
// Return all dirty state keys
func (o *SimpleAccount) Finalise() [][]byte {
	keys2Preload := make([][]byte, 0, len(o.dirtyState))
	for key, value := range o.dirtyState {
		o.pendingState[key] = value

		// collect all storage key of the account
		keys2Preload = append(keys2Preload, utils.CompositeStorageKey(o.Addr, []byte(key)))
	}
	o.dirtyState = make(map[string][]byte)
	return keys2Preload
}

func (o *SimpleAccount) getAccountJournal() *types.SnapshotJournalEntry {
	entry := &types.SnapshotJournalEntry{Address: o.Addr}
	prevStates := o.getStateJournal()

	if o.originAccount.InnerAccountChanged(o.dirtyAccount) || len(prevStates) != 0 || o.selfDestructed {
		entry.AccountChanged = true
		entry.PrevAccount = o.originAccount
		entry.PrevStates = prevStates
	}

	if o.originCode == nil && o.originAccount != nil && o.originAccount.CodeHash != nil {
		o.originCode = o.backend.Get(utils.CompositeCodeKey(o.Addr, o.originAccount.CodeHash))
	}

	if entry.AccountChanged {
		return entry
	}

	return nil
}

func (o *SimpleAccount) getStateJournal() map[string][]byte {
	prevStates := make(map[string][]byte)

	for key, value := range o.pendingState {
		origVal := o.originState[key]
		if !bytes.Equal(origVal, value) {
			prevStates[key] = origVal
		}
	}
	return prevStates
}

func (o *SimpleAccount) SetSelfDestructed(selfDestructed bool) {
	o.selfDestructed = selfDestructed
}

func (o *SimpleAccount) IsEmpty() bool {
	return o.GetBalance().Sign() == 0 && o.GetNonce() == 0 && o.Code() == nil && !o.selfDestructed
}

func (o *SimpleAccount) SelfDestructed() bool {
	return o.selfDestructed
}

func (o *SimpleAccount) GetStorageRoot() common.Hash {
	if o.originAccount == nil {
		return common.Hash{}
	}
	return o.originAccount.StorageRoot
}

func (o *SimpleAccount) IsCreated() bool {
	return o.created
}

func (o *SimpleAccount) SetCreated(created bool) {
	o.created = created
}
