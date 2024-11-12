package ledger

import (
	"github.com/axiomesh/axiom-kit/types"
)

type RustStateChange interface {
	// revert undoes the state changes by this entry
	Revert(ledger *RustStateLedger)

	// dirted returns the address modified by this state entry
	dirtied() *types.Address

	deepcopy() RustStateChange
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

func (s *RustStateChanger) ruststateChangerDeepCopy() *RustStateChanger {
	// Create a new stateChanger instance
	copyStateChanger := &RustStateChanger{
		changes: make([]RustStateChange, len(s.changes)),
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

	//RustResetObjectChange struct {
	//	prev IAccount
	//}
	//
	//RustRefundChange struct {
	//	prev uint64
	//}

	RustAddLogChange struct {
		txHash *types.Hash
	}

	//RustAddPreimageChange struct {
	//	hash types.Hash
	//}
	//
	RustAccessListAddAccountChange struct {
		address *types.Address
	}

	RustAccessListAddSlotChange struct {
		address *types.Address
		slot    *types.Hash
	}
	//
	//RustTransientStorageChange struct {
	//	account       *types.Address
	//	key, prevalue []byte
	//}
)

//func (ch RustCreateObjectChange) Revert(l *RustStateLedger) {
//	delete(l.Accounts, ch.account.String())
//}
//
//func (ch RustCreateObjectChange) dirtied() *types.Address {
//	return ch.account
//}

//	func (ch RustRefundChange) Revert(l *RustStateLedger) {
//		l.Refund = ch.prev
//	}
//
//	func (ch RustRefundChange) dirtied() *types.Address {
//		return nil
//	}
//
//	func (ch RustAddPreimageChange) Revert(l *RustStateLedger) {
//		delete(l.Preimages, ch.hash)
//	}
//
//	func (ch RustAddPreimageChange) dirtied() *types.Address {
//		return nil
//	}
func (ch RustAccessListAddAccountChange) Revert(l *RustStateLedger) {
	l.AccessList.DeleteAddress(*ch.address)
}

func (ch RustAccessListAddAccountChange) deepcopy() RustStateChange {
	copyCh := RustAccessListAddAccountChange{
		address: types.NewAddress(ch.address.Bytes()),
	}
	return copyCh
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

func (ch RustAccessListAddSlotChange) deepcopy() RustStateChange {
	copyCh := RustAccessListAddSlotChange{
		address: types.NewAddress(ch.address.Bytes()),
		slot:    types.NewHash(ch.slot.Bytes()),
	}
	return copyCh
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

func (ch RustAddLogChange) deepcopy() RustStateChange {
	copyCh := RustAddLogChange{
		txHash: types.NewHash(ch.txHash.Bytes()),
	}
	return copyCh
}

//func (ch RustTransientStorageChange) Revert(l *RustStateLedger) {
//	l.setTransientState(*ch.account, ch.key, ch.prevalue)
//}
//
//func (ch RustTransientStorageChange) dirtied() *types.Address {
//	return nil
//}
