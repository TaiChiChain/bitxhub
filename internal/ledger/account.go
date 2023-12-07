package ledger

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/leveldb"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/snapshot"
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

	originCode       []byte
	dirtyCode        []byte
	pendingStateHash *types.Hash
	ldb              storage.Storage
	cache            *AccountCache

	blockHeight uint64
	storageTrie *jmt.JMT

	changer  *stateChanger
	suicided bool

	snapshot *snapshot.Snapshot

	enableExpensiveMetric bool
}

func NewMockAccount(blockHeight uint64, addr *types.Address) *SimpleAccount {
	ldb, err := leveldb.NewMemory()
	if err != nil {
		panic(err)
	}

	ac, err := NewAccountCache(0, true)
	if err != nil {
		panic(err)
	}
	return &SimpleAccount{
		logger:       loggers.Logger(loggers.Storage),
		Addr:         addr,
		originState:  make(map[string][]byte),
		pendingState: make(map[string][]byte),
		dirtyState:   make(map[string][]byte),
		blockHeight:  blockHeight,
		ldb:          ldb,
		cache:        ac,
		changer:      NewChanger(),
		suicided:     false,
	}
}

func NewAccount(blockHeight uint64, ldb storage.Storage, accountCache *AccountCache, addr *types.Address, changer *stateChanger, snapshot *snapshot.Snapshot) *SimpleAccount {
	return &SimpleAccount{
		logger:       loggers.Logger(loggers.Storage),
		Addr:         addr,
		originState:  make(map[string][]byte),
		pendingState: make(map[string][]byte),
		dirtyState:   make(map[string][]byte),
		blockHeight:  blockHeight,
		ldb:          ldb,
		cache:        accountCache,
		changer:      changer,
		suicided:     false,
		snapshot:     snapshot,
	}
}

func (o *SimpleAccount) SetEnableExpensiveMetric(enable bool) {
	o.enableExpensiveMetric = enable
}

func (o *SimpleAccount) String() string {
	return fmt.Sprintf("{origin: %v, dirty: %v}", o.originAccount, o.dirtyAccount)
}

func (o *SimpleAccount) initStorageTrie() {
	if o.storageTrie != nil {
		return
	}
	// new account
	if o.originAccount == nil || o.originAccount.StorageRoot == (common.Hash{}) {
		rootHash := o.Addr.ETHAddress().Hash()
		rootNodeKey := jmt.NodeKey{
			Version: o.blockHeight,
			Path:    []byte{},
			Prefix:  o.Addr.Bytes(),
		}
		nk := rootNodeKey.Encode()
		o.ldb.Put(nk, nil)
		o.ldb.Put(rootHash[:], nk)
		trie, err := jmt.New(rootHash, o.ldb)
		if err != nil {
			panic(err)
		}
		o.storageTrie = trie
		o.logger.Debugf("[initStorageTrie] (new jmt) addr: %v, origin account: %v", o.Addr, o.originAccount)
		return
	}

	trie, err := jmt.New(o.originAccount.StorageRoot, o.ldb)
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
		if value, err := o.snapshot.Storage(o.Addr, key); err == nil && len(value) != 0 {
			o.originState[string(key)] = value
			o.initStorageTrie()
			//fmt.Printf("[GetState] get from snapshot, addr: %v, key: %v, state: %v\n", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			o.logger.Debugf("[GetState] get from snapshot, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			return value != nil, value
		}
	}

	o.initStorageTrie()
	start := time.Now()
	val, err := o.storageTrie.Get(compositeStorageKey(o.Addr, key))
	if err != nil {
		panic(err)
	}
	if o.enableExpensiveMetric {
		stateReadDuration.Observe(float64(time.Since(start)) / float64(time.Second))
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
		if value, err := o.snapshot.Storage(o.Addr, key); err == nil && len(value) != 0 {
			o.originState[string(key)] = value
			o.initStorageTrie()
			//fmt.Printf("[GetCommittedState] get from snapshot, addr: %v, key: %v, state: %v\n", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			o.logger.Debugf("[GetCommittedState] get from snapshot, addr: %v, key: %v, state: %v", o.Addr, &bytesLazyLogger{bytes: key}, &bytesLazyLogger{bytes: value})
			if value == nil {
				return (&types.Hash{}).Bytes()
			}
			return value
		}
	}

	o.initStorageTrie()
	start := time.Now()
	val, err := o.storageTrie.Get(compositeStorageKey(o.Addr, key))
	if err != nil {
		panic(err)
	}
	if o.enableExpensiveMetric {
		stateReadDuration.Observe(float64(time.Since(start)) / float64(time.Second))
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
	o.logger.Debugf("[SetState] addr: %v, key: %v, before state: %v, after state: %v", o.Addr, string(key), &bytesLazyLogger{bytes: prev}, &bytesLazyLogger{bytes: value})
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
		return o.dirtyCode
	}

	if o.originCode != nil {
		return o.originCode
	}

	if bytes.Equal(o.CodeHash(), nil) {
		return nil
	}

	code, ok := o.cache.getCode(o.Addr)
	if !ok {
		start := time.Now()
		code = o.ldb.Get(compositeKey(codeKey, o.Addr))
		if o.enableExpensiveMetric {
			codeReadDuration.Observe(float64(time.Since(start)) / float64(time.Second))
		}
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

		// collect all stoage key of the account
		keys2Preload = append(keys2Preload, compositeStorageKey(o.Addr, []byte(key)))
	}
	o.dirtyState = make(map[string][]byte)
	return keys2Preload
}

func (o *SimpleAccount) getJournalIfModified() *snapshot.BlockJournalEntry {
	entry := &snapshot.BlockJournalEntry{Address: o.Addr}

	if o.originAccount.InnerAccountChanged(o.dirtyAccount) {
		entry.AccountChanged = true
		entry.PrevAccount = o.originAccount
	}

	if o.originCode == nil && o.originAccount != nil && o.originAccount.CodeHash != nil {
		o.originCode = o.ldb.Get(compositeCodeKey(o.Addr, o.originAccount.CodeHash))
	}

	prevStates := o.getStateJournalAndComputeHash()
	if len(prevStates) != 0 {
		entry.PrevStates = prevStates
	}

	if entry.AccountChanged || len(entry.PrevStates) != 0 {
		return entry
	}

	return nil
}

func (o *SimpleAccount) getStateJournalAndComputeHash() map[string][]byte {
	prevStates := make(map[string][]byte)
	//var pendingStateKeys []string
	//var pendingStateData []byte

	for key, value := range o.pendingState {
		origVal := o.originState[key]
		if !bytes.Equal(origVal, value) {
			prevStates[key] = origVal
			//pendingStateKeys = append(pendingStateKeys, key)
		}
	}

	//sort.Strings(pendingStateKeys)
	//
	//for _, key := range pendingStateKeys {
	//	pendingStateData = append(pendingStateData, key...)
	//	pendingVal := o.pendingState[key]
	//	pendingStateData = append(pendingStateData, pendingVal...)
	//}
	//hash := sha256.Sum256(pendingStateData)
	//o.pendingStateHash = types.NewHash(hash[:])

	return prevStates
}

//func (o *SimpleAccount) getDirtyData() []byte {
//	var dirtyData []byte
//
//	dirtyData = append(dirtyData, o.Addr.Bytes()...)
//
//	if o.dirtyAccount != nil {
//		data, err := o.dirtyAccount.Marshal()
//		if err != nil {
//			panic(err)
//		}
//		dirtyData = append(dirtyData, data...)
//	}
//
//	return append(dirtyData, o.pendingStateHash.Bytes()...)
//}

func (o *SimpleAccount) SetSuicided(suicided bool) {
	o.suicided = suicided
}

func (o *SimpleAccount) IsEmpty() bool {
	return o.GetBalance().Sign() == 0 && o.GetNonce() == 0 && o.Code() == nil && !o.suicided
}

func (o *SimpleAccount) Suicided() bool {
	return false
}

func (o *SimpleAccount) GetStorageRootHash() common.Hash {
	if o.originAccount == nil || o.originAccount.StorageRoot == (common.Hash{}) {
		return o.Addr.ETHAddress().Hash()
	}
	return o.originAccount.StorageRoot
}
