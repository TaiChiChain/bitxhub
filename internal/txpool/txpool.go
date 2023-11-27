package txpool

import (
	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/axiomesh/axiom-kit/types"
)

func NewTxPool[T any, Constraint types.TXConstraint[T]](config Config) (txpool.TxPool[T, Constraint], error) {
	return newTxPoolImpl[T, Constraint](config)
}
