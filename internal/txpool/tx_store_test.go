package txpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	log2 "github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
)

func TestInsertPoolTx(t *testing.T) {
	ast := assert.New(t)
	nonceFn := func(address string) uint64 {
		return 0
	}
	log := log2.NewWithModule("txpool")
	store := newTransactionStore[types.Transaction, *types.Transaction](nonceFn, log)

	s, err := types.GenerateSigner()
	ast.Nil(err)
	from := s.Addr.String()

	tx := constructTx(s, 0)
	pointer := &txPointer{
		account: from,
		nonce:   0,
	}
	store.insertPoolTxPointer(tx.RbftGetTxHash(), pointer)
	store.insertPoolTxPointer(tx.RbftGetTxHash(), pointer) // test duplicate insert
	ast.Equal(1, len(store.txHashMap))

	now := time.Now().UnixNano()
	txItem := &internalTransaction[types.Transaction, *types.Transaction]{
		rawTx:       tx,
		local:       true,
		lifeTime:    tx.RbftGetTimeStamp(),
		arrivedTime: now,
	}

	store.insertPoolTx(from, txItem)
	ast.Equal(1, len(store.txHashMap))
	ast.Equal(1, len(store.allTxs))
	ast.NotNil(store.allTxs[from])
	ast.Equal(1, len(store.allTxs[from].items))

	oldItemTime := txItem.arrivedTime
	time.Sleep(1 * time.Millisecond)
	dupItem := &internalTransaction[types.Transaction, *types.Transaction]{
		rawTx:       tx,
		local:       false,
		lifeTime:    tx.RbftGetTimeStamp(),
		arrivedTime: time.Now().UnixNano(),
	}
	newItemTime := dupItem.arrivedTime
	ast.False(oldItemTime == newItemTime)

	store.insertPoolTx(from, dupItem)
	list := store.allTxs[from]
	ast.Equal(1, len(list.items))
	item := list.items[0]
	ast.Equal(newItemTime, item.arrivedTime, "new item replace old item")

	store.deletePoolTx(from, 1000) // delete pool tx which not exist in allTxs
	store.deletePoolTx(from, 0)
	list = store.allTxs[from]
	ast.Equal(0, len(list.items), "delete pool tx successful")
}
