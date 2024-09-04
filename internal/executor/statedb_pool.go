package executor

import (
	"sync"

	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type StateDbPool struct {
	pool sync.Pool
}

func NewStateDbPool() *StateDbPool {
	return &StateDbPool{
		pool: sync.Pool{
			New: func() interface{} {
				return nil
			},
		},
	}

}

func (sp *StateDbPool) GetStateDb(original *ledger.StateLedgerImpl) *ledger.StateLedgerImpl {
	instance := sp.pool.Get()
	var resInstance *ledger.StateLedgerImpl
	if instance == nil {
		resInstance = original.Copy().(*ledger.StateLedgerImpl)
	} else {
		resInstance = instance.(*ledger.StateLedgerImpl)
		resInstance.SetFromOrigin(original)
	}
	return resInstance
}

func (sp *StateDbPool) PutBackToPool(instance *ledger.StateLedgerImpl) {
	instance.Reset()
	sp.pool.Put(instance)
}
