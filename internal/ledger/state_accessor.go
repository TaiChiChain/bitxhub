package ledger

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/axiomesh/axiom-kit/jmt"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/snapshot"
	"github.com/axiomesh/axiom-ledger/internal/storagemgr"
)

var _ StateLedger = (*StateLedgerImpl)(nil)

const MinJournalHeight = 10

// GetOrCreateAccount get the account, if not exist, create a new account
func (l *StateLedgerImpl) GetOrCreateAccount(addr *types.Address) IAccount {
	account := l.GetAccount(addr)
	if account == nil {
		account = NewAccount(l.blockHeight, l.cachedDB, l.accountCache, addr, l.changer, l.snapshot)
		l.changer.append(createObjectChange{account: addr})
		l.accounts[addr.String()] = account
		l.logger.Debugf("[GetOrCreateAccount] create account, addr: %v", addr)
	} else {
		l.logger.Debugf("[GetOrCreateAccount] get account, addr: %v", addr)
	}
	account.SetEnableExpensiveMetric(l.enableExpensiveMetric)

	return account
}

// GetAccount get account info using account Address, if not found, create a new account
func (l *StateLedgerImpl) GetAccount(address *types.Address) IAccount {
	addr := address.String()

	value, ok := l.accounts[addr]
	if ok {
		l.logger.Debugf("[GetAccount] cache hit from accounts，addr: %v, account: %v", addr, value)
		return value
	}

	account := NewAccount(l.blockHeight, l.cachedDB, l.accountCache, address, l.changer, l.snapshot)
	account.SetEnableExpensiveMetric(l.enableExpensiveMetric)

	if innerAccount, ok := l.accountCache.getInnerAccount(address); ok {
		account.originAccount = innerAccount
		if !bytes.Equal(innerAccount.CodeHash, nil) {
			code, okCode := l.accountCache.getCode(address)
			if !okCode {
				code = l.cachedDB.Get(compositeCodeKey(account.Addr, account.originAccount.CodeHash))
			}
			account.originCode = code
			account.dirtyCode = code
		}
		l.accounts[addr] = account
		l.logger.Debugf("[GetAccount] cache hit from accountCache，addr: %v, account: %v", addr, account)
		return account
	}

	// try getting account from snapshot first
	if l.snapshot != nil {
		if innerAccount, err := l.snapshot.Account(address); err == nil && innerAccount != nil {
			account.originAccount = innerAccount
			if !bytes.Equal(innerAccount.CodeHash, nil) {
				code, okCode := l.accountCache.getCode(address)
				if !okCode {
					code = l.cachedDB.Get(compositeCodeKey(account.Addr, account.originAccount.CodeHash))
				}
				account.originCode = code
				account.dirtyCode = code
			}
			l.accounts[addr] = account
			l.logger.Debugf("[GetAccount] get account from snapshot, addr: %v, account: %v", addr, account)
			return account
		}
	}

	var rawAccount []byte
	start := time.Now()
	rawAccount, err := l.accountTrie.Get(compositeAccountKey(address))
	if err != nil {
		panic(err)
	}
	if l.enableExpensiveMetric {
		accountReadDuration.Observe(float64(time.Since(start)) / float64(time.Second))
	}

	if rawAccount != nil {
		account.originAccount = &types.InnerAccount{Balance: big.NewInt(0)}
		if err := account.originAccount.Unmarshal(rawAccount); err != nil {
			panic(err)
		}
		if !bytes.Equal(account.originAccount.CodeHash, nil) {
			code := l.cachedDB.Get(compositeCodeKey(account.Addr, account.originAccount.CodeHash))
			account.originCode = code
			account.dirtyCode = code
		}
		l.accounts[addr] = account
		l.logger.Debugf("[GetAccount] get from account trie，addr: %v, account: %v", addr, account)
		return account
	}
	l.logger.Debugf("[GetAccount] account not found，addr: %v", addr)
	return nil
}

// nolint
func (l *StateLedgerImpl) setAccount(account IAccount) {
	l.accounts[account.GetAddress().String()] = account
	l.logger.Debugf("[Revert setAccount] addr: %v, account: %v", account.GetAddress(), account)
}

// GetBalance get account balance using account Address
func (l *StateLedgerImpl) GetBalance(addr *types.Address) *big.Int {
	account := l.GetOrCreateAccount(addr)
	return account.GetBalance()
}

// SetBalance set account balance
func (l *StateLedgerImpl) SetBalance(addr *types.Address, value *big.Int) {
	account := l.GetOrCreateAccount(addr)
	account.SetBalance(value)
}

func (l *StateLedgerImpl) SubBalance(addr *types.Address, value *big.Int) {
	account := l.GetOrCreateAccount(addr)
	if !account.IsEmpty() {
		account.SubBalance(value)
	}
}

func (l *StateLedgerImpl) AddBalance(addr *types.Address, value *big.Int) {
	account := l.GetOrCreateAccount(addr)
	account.AddBalance(value)
}

// GetState get account state value using account Address and key
func (l *StateLedgerImpl) GetState(addr *types.Address, key []byte) (bool, []byte) {
	account := l.GetOrCreateAccount(addr)
	return account.GetState(key)
}

func (l *StateLedgerImpl) setTransientState(addr types.Address, key, value []byte) {
	l.transientStorage.Set(addr, common.BytesToHash(key), common.BytesToHash(value))
}

func (l *StateLedgerImpl) GetCommittedState(addr *types.Address, key []byte) []byte {
	account := l.GetOrCreateAccount(addr)
	if account.IsEmpty() {
		return (&types.Hash{}).Bytes()
	}
	return account.GetCommittedState(key)
}

// SetState set account state value using account Address and key
func (l *StateLedgerImpl) SetState(addr *types.Address, key []byte, v []byte) {
	account := l.GetOrCreateAccount(addr)
	account.SetState(key, v)
}

// SetCode set contract code
func (l *StateLedgerImpl) SetCode(addr *types.Address, code []byte) {
	account := l.GetOrCreateAccount(addr)
	account.SetCodeAndHash(code)
}

// GetCode get contract code
func (l *StateLedgerImpl) GetCode(addr *types.Address) []byte {
	account := l.GetOrCreateAccount(addr)
	return account.Code()
}

func (l *StateLedgerImpl) GetCodeHash(addr *types.Address) *types.Hash {
	account := l.GetOrCreateAccount(addr)
	if account.IsEmpty() {
		return &types.Hash{}
	}
	return types.NewHash(account.CodeHash())
}

func (l *StateLedgerImpl) GetCodeSize(addr *types.Address) int {
	account := l.GetOrCreateAccount(addr)
	if !account.IsEmpty() {
		if code := account.Code(); code != nil {
			return len(code)
		}
	}
	return 0
}

func (l *StateLedgerImpl) AddRefund(gas uint64) {
	l.changer.append(refundChange{prev: l.refund})
	l.refund += gas
}

func (l *StateLedgerImpl) SubRefund(gas uint64) {
	l.changer.append(refundChange{prev: l.refund})
	if gas > l.refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", gas, l.refund))
	}
	l.refund -= gas
}

func (l *StateLedgerImpl) GetRefund() uint64 {
	return l.refund
}

// GetNonce get account nonce
func (l *StateLedgerImpl) GetNonce(addr *types.Address) uint64 {
	account := l.GetOrCreateAccount(addr)
	return account.GetNonce()
}

// SetNonce set account nonce
func (l *StateLedgerImpl) SetNonce(addr *types.Address, nonce uint64) {
	account := l.GetOrCreateAccount(addr)
	account.SetNonce(nonce)
}

func (l *StateLedgerImpl) Clear() {
	l.accounts = make(map[string]IAccount)
}

// collectDirtyData gets dirty accounts and snapshot journals
func (l *StateLedgerImpl) collectDirtyData() (map[string]IAccount, *snapshot.BlockJournal) {
	dirtyAccounts := make(map[string]IAccount)
	var journals []*snapshot.BlockJournalEntry

	for addr, acc := range l.accounts {
		account := acc.(*SimpleAccount)
		journal := account.getAccountJournal()
		if journal != nil {
			journals = append(journals, journal)
			dirtyAccounts[addr] = account
		}
	}

	blockJournal := &snapshot.BlockJournal{
		Journals: journals,
	}
	l.Clear() // remove accounts that cached during executing current block
	l.accountCache.add(dirtyAccounts)
	return dirtyAccounts, blockJournal
}

// Commit the state, and get account trie root hash
func (l *StateLedgerImpl) Commit() (*types.Hash, error) {
	l.logger.Debugf("==================[Commit-Start]==================")
	defer l.logger.Debugf("==================[Commit-End]==================")

	l.accountCache.exportMetrics()
	storagemgr.ExportCachedStorageMetrics()
	defer ExportTriePreloaderMetrics()
	if l.triePreloader != nil {
		defer l.triePreloader.close()
	}

	accounts, journals := l.collectDirtyData()
	height := l.blockHeight
	destructSet := make(map[string]struct{})
	accountSet := make(map[string]*types.InnerAccount)
	storageSet := make(map[string]map[string][]byte)

	ldbBatch := l.cachedDB.NewBatch()

	accSize := 0
	for _, acc := range accounts {
		account := acc.(*SimpleAccount)
		if account.Suicided() {
			accSize++
			data, err := l.accountTrie.Get(compositeAccountKey(account.Addr))
			if err != nil {
				return nil, err
			}
			if data != nil {
				err = l.accountTrie.Update(height, compositeAccountKey(account.Addr), nil)
				if err != nil {
					return nil, err
				}
			}
			destructSet[account.Addr.String()] = struct{}{}
			continue
		}

		if !bytes.Equal(account.originCode, account.dirtyCode) && account.dirtyCode != nil {
			ldbBatch.Put(compositeCodeKey(account.Addr, account.dirtyAccount.CodeHash), account.dirtyCode)
		}

		l.logger.Debugf("[Commit-Before] committing storage trie begin, addr: %v,account.dirtyAccount.StorageRoot: %v", account.Addr, account.dirtyAccount.StorageRoot)

		stateSize := 0
		addr := account.Addr.String()
		storageSet[addr] = make(map[string][]byte)
		for key, valBytes := range account.pendingState {
			origValBytes := account.originState[key]

			if !bytes.Equal(origValBytes, valBytes) {
				if err := account.storageTrie.Update(height, compositeStorageKey(account.Addr, []byte(key)), valBytes); err != nil {
					panic(err)
				}
				storageSet[addr][key] = valBytes
				if account.storageTrie.Root() != nil {
					l.logger.Debugf("[Commit-Update-After][%v] after updating storage trie, addr: %v, key: %v, origin state: %v, "+
						"dirty state: %v, root node: %v", stateSize, account.Addr, &bytesLazyLogger{bytes: compositeStorageKey(account.Addr, []byte(key))},
						&bytesLazyLogger{bytes: origValBytes}, &bytesLazyLogger{bytes: valBytes}, account.storageTrie.Root().Print())
				}
				stateSize++
			}
		}
		// commit account's storage trie
		if account.storageTrie != nil {
			account.dirtyAccount.StorageRoot = account.storageTrie.Commit()
			l.logger.Debugf("[Commit-After] committing storage trie end, addr: %v,account.dirtyAccount.StorageRoot: %v", account.Addr, account.dirtyAccount.StorageRoot)
		}
		if l.enableExpensiveMetric {
			stateFlushSize.Set(float64(stateSize))
		}
		// update account trie if needed
		if account.originAccount.InnerAccountChanged(account.dirtyAccount) {
			accSize++
			data, err := account.dirtyAccount.Marshal()
			if err != nil {
				panic(err)
			}
			if err := l.accountTrie.Update(height, compositeAccountKey(account.Addr), data); err != nil {
				panic(err)
			}
			accountSet[addr] = account.dirtyAccount
			l.logger.Debugf("[Commit] update account trie, addr: %v, origin account: %v, dirty account: %v", account.Addr, account.originAccount, account.dirtyAccount)
		}
	}
	ldbBatch.Commit()

	// Commit world state trie.
	// If world state is not changed in current block (which is very rarely), this is no-op.
	stateRoot := l.accountTrie.Commit()
	if l.enableExpensiveMetric {
		accountFlushSize.Set(float64(accSize))
	}
	l.logger.Debugf("[Commit] after committed world state trie, StateRoot: %v", stateRoot)

	// update snapshot
	err := l.snapshot.Update(stateRoot, destructSet, accountSet, storageSet)
	if err != nil {
		return nil, fmt.Errorf("update snapshot error: %w", err)
	}

	// persist snapshot journals
	err = l.snapshot.UpdateJournal(height, journals)
	if err != nil {
		return nil, fmt.Errorf("update snapshot journal error: %w", err)
	}

	if height > l.getJnlHeightSize() {
		if err := l.snapshot.RemoveJournalsBeforeBlock(height - l.getJnlHeightSize()); err != nil {
			return nil, fmt.Errorf("remove journals before block %d failed: %w", height-l.getJnlHeightSize(), err)
		}
	}

	return types.NewHash(stateRoot.Bytes()), nil
}

func (l *StateLedgerImpl) getJnlHeightSize() uint64 {
	if l.repo.EpochInfo.ConsensusParams.CheckpointPeriod < MinJournalHeight {
		return MinJournalHeight
	}
	return l.repo.EpochInfo.ConsensusParams.CheckpointPeriod
}

// Version returns the current version
func (l *StateLedgerImpl) Version() uint64 {
	return l.blockHeight
}

// RollbackState does not delete the state data that has been persisted in KV.
// This manner will not affect the correctness of ledger,
// todo but maybe need to optimize to free allocated space in KV.
func (l *StateLedgerImpl) RollbackState(height uint64, stateRoot *types.Hash) error {
	// clean cache account
	l.Clear()
	l.accountCache.clear()

	// rollback snapshots
	if err := l.snapshot.Rollback(height); err != nil {
		return err
	}

	// rollback world state trie
	if height != 0 {
		l.refreshAccountTrie(stateRoot)
	}

	return nil
}

func (l *StateLedgerImpl) Suicide(addr *types.Address) bool {
	account := l.GetOrCreateAccount(addr)
	l.changer.append(suicideChange{
		account:     addr,
		prev:        account.Suicided(),
		prevbalance: new(big.Int).Set(account.GetBalance()),
	})
	l.logger.Debugf("[Suicide] addr: %v, before balance: %v", addr, account.GetBalance())
	account.SetSuicided(true)
	account.SetBalance(new(big.Int))

	return true
}

func (l *StateLedgerImpl) HasSuicide(addr *types.Address) bool {
	account := l.GetOrCreateAccount(addr)
	if account.IsEmpty() {
		l.logger.Debugf("[HasSuicide] addr: %v, is empty, suicide: false", addr)
		return false
	}
	l.logger.Debugf("[HasSuicide] addr: %v, suicide: %v", addr, account.Suicided())
	return account.Suicided()
}

func (l *StateLedgerImpl) Exist(addr *types.Address) bool {
	exist := !l.GetOrCreateAccount(addr).IsEmpty()
	l.logger.Debugf("[Exist] addr: %v, exist: %v", addr, exist)
	return exist
}

func (l *StateLedgerImpl) Empty(addr *types.Address) bool {
	empty := l.GetOrCreateAccount(addr).IsEmpty()
	l.logger.Debugf("[Empty] addr: %v, empty: %v", addr, empty)
	return empty
}

func (l *StateLedgerImpl) Snapshot() int {
	l.logger.Debugf("-------------------------- [Snapshot] --------------------------")
	id := l.nextRevisionId
	l.nextRevisionId++
	l.validRevisions = append(l.validRevisions, revision{id: id, changerIndex: l.changer.length()})
	return id
}

func (l *StateLedgerImpl) RevertToSnapshot(revid int) {
	idx := sort.Search(len(l.validRevisions), func(i int) bool {
		return l.validRevisions[i].id >= revid
	})
	if idx == len(l.validRevisions) || l.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannod be reverted", revid))
	}
	snapshot := l.validRevisions[idx].changerIndex

	l.changer.revert(l, snapshot)
	l.validRevisions = l.validRevisions[:idx]
}

func (l *StateLedgerImpl) ClearChangerAndRefund() {
	if len(l.changer.changes) > 0 {
		l.changer = NewChanger()
		l.refund = 0
	}
	l.validRevisions = l.validRevisions[:0]
	l.nextRevisionId = 0
}

func (l *StateLedgerImpl) AddAddressToAccessList(addr types.Address) {
	if l.accessList.AddAddress(addr) {
		l.changer.append(accessListAddAccountChange{address: &addr})
	}
}

func (l *StateLedgerImpl) AddSlotToAccessList(addr types.Address, slot types.Hash) {
	addrMod, slotMod := l.accessList.AddSlot(addr, slot)
	if addrMod {
		l.changer.append(accessListAddAccountChange{address: &addr})
	}
	if slotMod {
		l.changer.append(accessListAddSlotChange{
			address: &addr,
			slot:    &slot,
		})
	}
}

func (l *StateLedgerImpl) PrepareAccessList(sender types.Address, dst *types.Address, precompiles []types.Address, list AccessTupleList) {
	l.AddAddressToAccessList(sender)

	if dst != nil {
		l.AddAddressToAccessList(*dst)
	}

	for _, addr := range precompiles {
		l.AddAddressToAccessList(addr)
	}
	for _, el := range list {
		l.AddAddressToAccessList(el.Address)
		for _, key := range el.StorageKeys {
			l.AddSlotToAccessList(el.Address, key)
		}
	}
}

func (l *StateLedgerImpl) AddressInAccessList(addr types.Address) bool {
	return l.accessList.ContainsAddress(addr)
}

func (l *StateLedgerImpl) SlotInAccessList(addr types.Address, slot types.Hash) (bool, bool) {
	return l.accessList.Contains(addr, slot)
}

func (l *StateLedgerImpl) AddPreimage(hash types.Hash, preimage []byte) {
	if _, ok := l.preimages[hash]; !ok {
		l.changer.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		l.preimages[hash] = pi
	}
}

func (l *StateLedgerImpl) PrepareBlock(lastStateRoot *types.Hash, hash *types.Hash, currentExecutingHeight uint64) {
	l.logs = NewEvmLogs()
	l.logs.bhash = hash
	l.blockHeight = currentExecutingHeight
	l.refreshAccountTrie(lastStateRoot)
	l.accountCache.resetMetrics()
	storagemgr.ResetCachedStorageMetrics()
	ResetTriePreloaderMetrics()
	l.logger.Debugf("[PrepareBlock] height: %v, hash: %v", currentExecutingHeight, hash)
}

func (l *StateLedgerImpl) refreshAccountTrie(lastStateRoot *types.Hash) {
	if lastStateRoot == nil {
		// dummy state
		rootHash := common.Hash{}
		rootNodeKey := jmt.NodeKey{
			Version: 0,
			Path:    []byte{},
			Prefix:  []byte{},
		}
		nk := rootNodeKey.Encode()
		l.cachedDB.Put(nk, nil)
		l.cachedDB.Put(rootHash[:], nk)
		trie, _ := jmt.New(rootHash, l.cachedDB)
		l.accountTrie = trie
		l.triePreloader = newTriePreloader(l.logger, l.cachedDB, rootHash)
		return
	}

	trie, err := jmt.New(lastStateRoot.ETHHash(), l.cachedDB)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"lastStateRoot": lastStateRoot,
			"currentHeight": l.blockHeight,
			"err":           err.Error(),
		}).Errorf("load account trie from db error")
		return
	}
	l.accountTrie = trie
	l.triePreloader = newTriePreloader(l.logger, l.cachedDB, lastStateRoot.ETHHash())
}

func (l *StateLedgerImpl) AddLog(log *types.EvmLog) {
	if log.TransactionHash == nil {
		log.TransactionHash = l.thash
	}

	log.TransactionIndex = uint64(l.txIndex)

	l.changer.append(addLogChange{txHash: log.TransactionHash})

	log.BlockHash = l.logs.bhash
	log.LogIndex = uint64(l.logs.logSize)
	if _, ok := l.logs.logs[*log.TransactionHash]; !ok {
		l.logs.logs[*log.TransactionHash] = make([]*types.EvmLog, 0)
	}

	l.logs.logs[*log.TransactionHash] = append(l.logs.logs[*log.TransactionHash], log)
	l.logs.logSize++
}

func (l *StateLedgerImpl) GetLogs(hash types.Hash, height uint64, blockHash *types.Hash) []*types.EvmLog {
	logs := l.logs.logs[hash]
	for _, l := range logs {
		l.BlockNumber = height
		l.BlockHash = blockHash
	}
	return logs
}

func (l *StateLedgerImpl) Logs() []*types.EvmLog {
	var logs []*types.EvmLog
	for _, lgs := range l.logs.logs {
		logs = append(logs, lgs...)
	}
	return logs
}
