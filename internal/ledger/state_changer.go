package ledger

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/blockstm"
)

type stateChange interface {
	// revert undoes the state changes by this entry
	revert(*StateLedgerImpl)

	// dirted returns the address modified by this state entry
	dirtied() *types.Address

	deepcopy() stateChange
}

type stateChanger struct {
	changes []stateChange
	dirties map[types.Address]int // dirty address and the number of changes
}

func (s *stateChanger) stateChangerDeepCopy() *stateChanger {
	// Create a new stateChanger instance
	copyStateChanger := &stateChanger{
		changes: make([]stateChange, len(s.changes)),
		dirties: make(map[types.Address]int),
	}

	for i, change := range s.changes {
		copyStateChanger.changes[i] = change.deepcopy()
	}
	// Copy the dirties map
	for key, value := range s.dirties {
		copyStateChanger.dirties[*types.NewAddress(key.Bytes())] = value
	}

	return copyStateChanger
}

func newChanger() *stateChanger {
	return &stateChanger{
		dirties: make(map[types.Address]int),
	}
}

func (s *stateChanger) append(change stateChange) {
	s.changes = append(s.changes, change)
	if addr := change.dirtied(); addr != nil {
		s.dirties[*addr]++
	}
}

func (s *stateChanger) revert(ledger *StateLedgerImpl, snapshot int) {
	for i := len(s.changes) - 1; i >= snapshot; i-- {
		s.changes[i].revert(ledger)

		if addr := s.changes[i].dirtied(); addr != nil {
			if s.dirties[*addr]--; s.dirties[*addr] == 0 {
				delete(s.dirties, *addr)
			}
		}
	}

	s.changes = s.changes[:snapshot]
}

func (s *stateChanger) dirty(addr types.Address) {
	s.dirties[addr]++
}

func (s *stateChanger) length() int {
	return len(s.changes)
}

func (s *stateChanger) reset() {
	s.changes = []stateChange{}
	s.dirties = make(map[types.Address]int)

}

type (
	createObjectChange struct {
		account *types.Address
	}

	resetObjectChange struct {
		prev IAccount
	}

	suicideChange struct {
		account     *types.Address
		prev        bool
		prevbalance *big.Int
	}

	balanceChange struct {
		account *types.Address
		prev    *big.Int
	}

	nonceChange struct {
		account *types.Address
		prev    uint64
	}

	storageChange struct {
		account       *types.Address
		key, prevalue []byte
	}

	codeChange struct {
		account  *types.Address
		prevcode []byte
	}

	refundChange struct {
		prev uint64
	}

	addLogChange struct {
		txHash *types.Hash
	}

	addPreimageChange struct {
		hash types.Hash
	}

	touchChange struct {
		account *types.Address
	}

	accessListAddAccountChange struct {
		address *types.Address
	}

	accessListAddSlotChange struct {
		address *types.Address
		slot    *types.Hash
	}

	transientStorageChange struct {
		account       *types.Address
		key, prevalue []byte
	}
)

func (ch createObjectChange) revert(l *StateLedgerImpl) {
	delete(l.accounts, ch.account.ETHAddress())
	RevertWrite(l, blockstm.NewAddressKey(*ch.account))
}

func (ch createObjectChange) deepcopy() stateChange {
	copyCh := createObjectChange{
		account: types.NewAddress(ch.account.Bytes()),
	}
	return copyCh
}

func (ch createObjectChange) dirtied() *types.Address {
	return ch.account
}

// nolint
func (ch resetObjectChange) revert(l *StateLedgerImpl) {
	l.setAccount(ch.prev)
	RevertWrite(l, blockstm.NewAddressKey(*ch.prev.GetAddress()))
}

func (ch resetObjectChange) deepcopy() stateChange {
	copyCh := resetObjectChange{
		prev: DeepCopy(ch.prev.(*SimpleAccount)),
	}
	return copyCh
}

// nolint
func (ch resetObjectChange) dirtied() *types.Address {
	return nil
}

func (ch suicideChange) revert(l *StateLedgerImpl) {
	acc := l.GetOrCreateAccount(ch.account)
	account := acc.(*SimpleAccount)
	account.selfDestructed = ch.prev
	account.setBalance(ch.prevbalance)
	RevertWrite(l, blockstm.NewSubpathKey(*ch.account, SuicidePath))
}

func (ch suicideChange) deepcopy() stateChange {
	copyCh := suicideChange{
		account:     types.NewAddress(ch.account.Bytes()),
		prev:        ch.prev,
		prevbalance: new(big.Int).Set(ch.prevbalance),
	}
	return copyCh
}

func (ch suicideChange) dirtied() *types.Address {
	return ch.account
}

// nolint
func (ch touchChange) revert(l *StateLedgerImpl) {}

func (ch touchChange) dirtied() *types.Address {
	return ch.account
}
func (ch touchChange) deepcopy() stateChange {
	copyCh := touchChange{
		account: types.NewAddress(ch.account.Bytes()),
	}
	return copyCh
}

func (ch balanceChange) revert(l *StateLedgerImpl) {
	l.GetOrCreateAccount(ch.account).(*SimpleAccount).setBalance(ch.prev)
}

func (ch balanceChange) dirtied() *types.Address {
	return ch.account
}

func (ch balanceChange) deepcopy() stateChange {
	copyCh := balanceChange{
		account: types.NewAddress(ch.account.Bytes()),
		prev:    new(big.Int).Set(ch.prev),
	}
	return copyCh
}

func (ch nonceChange) revert(l *StateLedgerImpl) {
	l.GetOrCreateAccount(ch.account).(*SimpleAccount).setNonce(ch.prev)
}

func (ch nonceChange) dirtied() *types.Address {
	return ch.account
}

func (ch nonceChange) deepcopy() stateChange {
	copyCh := nonceChange{
		account: types.NewAddress(ch.account.Bytes()),
		prev:    ch.prev,
	}
	return copyCh
}

func (ch codeChange) revert(l *StateLedgerImpl) {
	l.GetOrCreateAccount(ch.account).(*SimpleAccount).setCodeAndHash(ch.prevcode)
	RevertWrite(l, blockstm.NewSubpathKey(*ch.account, CodePath))
}

func (ch codeChange) dirtied() *types.Address {
	return ch.account
}

func (ch codeChange) deepcopy() stateChange {
	copyCh := codeChange{
		account:  types.NewAddress(ch.account.Bytes()),
		prevcode: CopyBytes(ch.prevcode),
	}
	return copyCh
}

func (ch storageChange) revert(l *StateLedgerImpl) {
	l.GetOrCreateAccount(ch.account).(*SimpleAccount).setState(ch.key, ch.prevalue)
	RevertWrite(l, blockstm.NewStateKey(*ch.account, *types.NewHash(ch.key)))
}

func (ch storageChange) dirtied() *types.Address {
	return ch.account
}

func (ch storageChange) deepcopy() stateChange {
	copyCh := storageChange{
		account:  types.NewAddress(ch.account.Bytes()),
		key:      CopyBytes(ch.key),
		prevalue: CopyBytes(ch.prevalue),
	}
	return copyCh
}

func (ch refundChange) revert(l *StateLedgerImpl) {
	l.refund = ch.prev
}

func (ch refundChange) dirtied() *types.Address {
	return nil
}

func (ch refundChange) deepcopy() stateChange {
	copyCh := refundChange{
		prev: ch.prev,
	}
	return copyCh
}

func (ch addPreimageChange) revert(l *StateLedgerImpl) {
	delete(l.preimages, ch.hash)
}

func (ch addPreimageChange) dirtied() *types.Address {
	return nil
}

func (ch addPreimageChange) deepcopy() stateChange {
	copyCh := addPreimageChange{
		hash: *types.NewHash(ch.hash.Bytes()),
	}
	return copyCh
}

func (ch accessListAddAccountChange) revert(l *StateLedgerImpl) {
	l.accessList.DeleteAddress(*ch.address)
}

func (ch accessListAddAccountChange) dirtied() *types.Address {
	return nil
}

func (ch accessListAddAccountChange) deepcopy() stateChange {
	copyCh := accessListAddAccountChange{
		address: types.NewAddress(ch.address.Bytes()),
	}
	return copyCh
}

func (ch accessListAddSlotChange) revert(l *StateLedgerImpl) {
	l.accessList.DeleteSlot(*ch.address, *ch.slot)
}

func (ch accessListAddSlotChange) dirtied() *types.Address {
	return nil
}

func (ch accessListAddSlotChange) deepcopy() stateChange {
	copyCh := accessListAddSlotChange{
		address: types.NewAddress(ch.address.Bytes()),
		slot:    types.NewHash(ch.slot.Bytes()),
	}
	return copyCh
}

func (ch addLogChange) revert(l *StateLedgerImpl) {
	logs := l.logs.Logs[*ch.txHash]
	if len(logs) == 1 {
		delete(l.logs.Logs, *ch.txHash)
	} else {
		l.logs.Logs[*ch.txHash] = logs[:len(logs)-1]
	}
	l.logs.LogSize--
}

func (ch addLogChange) dirtied() *types.Address {
	return nil
}

func (ch addLogChange) deepcopy() stateChange {
	copyCh := addLogChange{
		txHash: types.NewHash(ch.txHash.Bytes()),
	}
	return copyCh
}

func (ch transientStorageChange) revert(l *StateLedgerImpl) {
	l.setTransientState(*ch.account, ch.key, ch.prevalue)
}

func (ch transientStorageChange) dirtied() *types.Address {
	return nil
}

func (ch transientStorageChange) deepcopy() stateChange {
	copyCh := transientStorageChange{
		account:  types.NewAddress(ch.account.Bytes()),
		key:      CopyBytes(ch.key),
		prevalue: CopyBytes(ch.prevalue),
	}
	return copyCh
}
