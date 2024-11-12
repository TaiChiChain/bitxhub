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

func (sp *StateDbPool) GetStateDb(original ledger.StateLedger) ledger.StateLedger {
	instance := sp.pool.Get()
	var resInstance ledger.StateLedger
	if instance == nil {
		resInstance = original.Copy()
	} else {
		resInstance = instance.(ledger.StateLedger)
		resInstance.SetFromOrigin(original)
	}
	return resInstance
}

func (sp *StateDbPool) PutBackToPool(instance ledger.StateLedger) {
	instance.Reset()
	sp.pool.Put(instance)
}
