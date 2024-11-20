package ledger

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/ledger/blockstm"
	"github.com/ethereum/go-ethereum/common"
)

type RustStateChange interface {
	// revert undoes the state changes by this entry
	Revert(ledger *RustStateLedger)

	// dirted returns the address modified by this state entry
	dirtied() *types.Address
}

type RustStateChanger struct {
	changes []RustStateChange
	dirties map[types.Address]int // dirty address and the number of changes
}

func NewChanger() *RustStateChanger {
	return &RustStateChanger{
		dirties: make(map[types.Address]int),
	}
}

func (s *RustStateChanger) Append(change RustStateChange) {
	s.changes = append(s.changes, change)
	if addr := change.dirtied(); addr != nil {
		s.dirties[*addr]++
	}
}

func (s *RustStateChanger) Revert(ledger *RustStateLedger, snapshot int) {
	for i := len(s.changes) - 1; i >= snapshot; i-- {
		s.changes[i].Revert(ledger)

		if addr := s.changes[i].dirtied(); addr != nil {
			if s.dirties[*addr]--; s.dirties[*addr] == 0 {
				delete(s.dirties, *addr)
			}
		}
	}

	s.changes = s.changes[:snapshot]
}

func (s *RustStateChanger) dirty(addr types.Address) {
	s.dirties[addr]++
}

func (s *RustStateChanger) length() int {
	return len(s.changes)
}

func (s *RustStateChanger) Reset() {
	s.changes = []RustStateChange{}
	s.dirties = make(map[types.Address]int)
}

type (
	RustCreateObjectChange struct {
		account *types.Address
	}

	RustResetObjectChange struct {
		prev IAccount
	}

	RustSuicideChange struct {
		account     *types.Address
		prev        bool
		prevbalance *big.Int
	}

	RustBalanceChange struct {
		account *types.Address
		prev    *big.Int
	}

	RustNonceChange struct {
		account *types.Address
		prev    uint64
	}

	RustStorageChange struct {
		account       *types.Address
		key, prevalue []byte
	}

	RustCodeChange struct {
		account  *types.Address
		prevcode []byte
	}

	RustRefundChange struct {
		prev uint64
	}

	RustAddLogChange struct {
		txHash *types.Hash
	}

	RustAddPreimageChange struct {
		hash types.Hash
	}

	RustAccessListAddAccountChange struct {
		address *types.Address
	}

	RustAccessListAddSlotChange struct {
		address *types.Address
		slot    *types.Hash
	}

	RustTransientStorageChange struct {
		account       *types.Address
		key, prevalue []byte
	}
)

func (ch RustCreateObjectChange) Revert(r *RustStateLedger) {
	delete(r.Accounts, ch.account.String())
	RustRevertWrite(r, blockstm.NewAddressKey(ch.account))
}

func (ch RustCreateObjectChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustResetObjectChange) Revert(r *RustStateLedger) {
	r.setAccount(ch.prev)
	RustRevertWrite(r, blockstm.NewAddressKey(ch.prev.GetAddress()))
}

// nolint
func (ch RustResetObjectChange) dirtied() *types.Address {
	return nil
}

func (ch RustSuicideChange) Revert(r *RustStateLedger) {
	acc := r.GetOrCreateAccount(ch.account)
	account := acc.(*RustAccount)
	account.selfDestructed = ch.prev
	account.setBalance(ch.prevbalance)
	RustRevertWrite(r, blockstm.NewSubpathKey(ch.account, SuicidePath))
}

func (ch RustSuicideChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustBalanceChange) Revert(r *RustStateLedger) {
	r.GetOrCreateAccount(ch.account).(*RustAccount).setBalance(ch.prev)
}

func (ch RustBalanceChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustNonceChange) Revert(r *RustStateLedger) {
	r.GetOrCreateAccount(ch.account).(*RustAccount).setNonce(ch.prev)
}

func (ch RustNonceChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustCodeChange) Revert(r *RustStateLedger) {
	r.GetOrCreateAccount(ch.account).(*RustAccount).setCodeAndHash(ch.prevcode)
	RustRevertWrite(r, blockstm.NewSubpathKey(ch.account, CodePath))
}

func (ch RustCodeChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustStorageChange) Revert(r *RustStateLedger) {
	r.GetOrCreateAccount(ch.account).(*RustAccount).setState(ch.key, ch.prevalue)
	RustRevertWrite(r, blockstm.NewStateKey(ch.account, *types.NewHash(ch.key)))
}

func (ch RustStorageChange) dirtied() *types.Address {
	return ch.account
}

func (ch RustRefundChange) Revert(r *RustStateLedger) {
	r.Refund = ch.prev
}

func (ch RustRefundChange) dirtied() *types.Address {
	return nil
}

func (ch RustAddPreimageChange) Revert(r *RustStateLedger) {
	delete(r.preimages, ch.hash)
}

func (ch RustAddPreimageChange) dirtied() *types.Address {
	return nil
}

func (ch RustAccessListAddAccountChange) Revert(l *RustStateLedger) {
	l.AccessList.DeleteAddress(*ch.address)
}

func (ch RustAccessListAddAccountChange) dirtied() *types.Address {
	return nil
}

func (ch RustAccessListAddSlotChange) Revert(l *RustStateLedger) {
	l.AccessList.DeleteSlot(*ch.address, *ch.slot)
}

func (ch RustAccessListAddSlotChange) dirtied() *types.Address {
	return nil
}

func (ch RustAddLogChange) Revert(l *RustStateLedger) {
	logs := l.Logs.Logs[*ch.txHash]
	if len(logs) == 1 {
		delete(l.Logs.Logs, *ch.txHash)
	} else {
		l.Logs.Logs[*ch.txHash] = logs[:len(logs)-1]
	}
	l.Logs.LogSize--
}

func (ch RustAddLogChange) dirtied() *types.Address {
	return nil
}

func (ch RustTransientStorageChange) Revert(r *RustStateLedger) {
	r.setTransientState(*ch.account, common.BytesToHash(ch.key), common.BytesToHash(ch.prevalue))
}

func (ch RustTransientStorageChange) dirtied() *types.Address {
	return nil
}
