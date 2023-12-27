package txpool

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
)

func TestHandleRemoveTimeoutEvent(t *testing.T) {
	t.Parallel()
	t.Run("remove timeout txs, expect priority and batched txs", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.toleranceRemoveTime = 10 * time.Millisecond
		pool.batchSize = 4
		ch := make(chan struct{}, 1)
		pool.notifyGenerateBatchFn = func(typ int) {
			go func() {
				_, err := pool.GenerateRequestBatch(typ)
				ast.Nil(err)
			}()
			ch <- struct{}{}
		}
		err := pool.Start()
		ast.Nil(err)
		defer pool.Stop()

		s, err := types.GenerateSigner()
		ast.Nil(err)
		txs := constructTxs(s, 5) // 4 batched txs + 1 priority tx
		pool.AddRemoteTxs(txs)
		<-ch
		// sleep a while to trigger the remove timeout event
		time.Sleep(10 * time.Millisecond)
		ast.Equal(uint64(5), pool.GetTotalPendingTxCount())
		ast.Equal(4, len(pool.txStore.batchedTxs))
		ast.Equal(1, len(pool.txStore.batchesCache))
		ast.Equal(5, pool.txStore.priorityIndex.size())
		ast.Equal(uint64(1), pool.txStore.priorityNonBatchSize)

		pool.handleRemoveTimeout(timer.RemoveTx)
		res := pool.GetTotalPendingTxCount()
		fmt.Println(res)
		ast.Equal(uint64(5), pool.GetTotalPendingTxCount(), "ignore batched and priority txs")
	})

	t.Run("remove Event successful", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.toleranceRemoveTime = 1 * time.Millisecond
		err := pool.Start()
		defer pool.Stop()
		ast.Nil(err)

		// test wrong event
		pool.handleRemoveTimeout("wrongEventType")

		s, err := types.GenerateSigner()
		ast.Nil(err)
		txs := constructTxs(s, 5)
		// remove tx0
		txs = txs[1:]

		// retry 2 times to test account empty flag change(empty -> not empty -> empty)
		for i := 0; i < 2; i++ {
			pool.AddRemoteTxs(txs)
			assert.Equal(t, uint64(4), pool.GetTotalPendingTxCount())
			ast.Equal(0, pool.txStore.priorityIndex.size())
			ast.Equal(4, pool.txStore.parkingLotIndex.size())
			ast.Equal(uint64(0), pool.txStore.priorityNonBatchSize)

			// sleep a while to trigger the remove timeout event
			time.Sleep(2 * time.Millisecond)
			pool.handleRemoveTimeout(timer.RemoveTx)
			assert.Equal(t, uint64(0), pool.GetTotalPendingTxCount())

			assert.Equal(t, 1, len(pool.txStore.nonceCache.commitNonces))
			assert.Equal(t, 1, len(pool.txStore.nonceCache.pendingNonces))
			assert.Equal(t, 1, len(pool.txStore.allTxs))
		}

		pool.cleanEmptyAccountTime = 1 * time.Millisecond
		// sleep a while to trigger the clean empty account timeout event
		time.Sleep(2 * time.Millisecond)
		pool.handleRemoveTimeout(timer.CleanEmptyAccount)
		assert.Equal(t, uint64(0), pool.GetTotalPendingTxCount())
		assert.Equal(t, 0, len(pool.txStore.nonceCache.commitNonces))
		assert.Equal(t, 0, len(pool.txStore.nonceCache.pendingNonces))
		assert.Equal(t, 0, len(pool.txStore.allTxs))
	})

	t.Run("rotate tx locals", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.rotateTxLocalsInterval = 1 * time.Millisecond
		err := pool.Start()
		defer pool.Stop()
		ast.Nil(err)
		s, err := types.GenerateSigner()
		ast.Nil(err)
		tx := constructTx(s, 0)
		// case pool has no tx
		b, err := tx.RbftMarshal()
		ast.Nil(err)
		length := uint64(len(b))
		var lengthBytes [TxRecordPrefixLength]byte
		binary.LittleEndian.PutUint64(lengthBytes[:], length)
		_, err = pool.txRecords.writer.Write(lengthBytes[:])
		ast.Nil(err)
		_, err = pool.txRecords.writer.Write(b)
		ast.Nil(err)
		records, err := GetAllTxRecords(pool.txRecordsFile)
		ast.True(len(records) == 1)
		pool.handleRemoveTimeout(timer.RotateTxLocals)
		time.Sleep(2 * time.Millisecond)
		ast.Equal(uint64(0), pool.GetTotalPendingTxCount())
		records2, err := GetAllTxRecords(pool.txRecordsFile)
		ast.True(len(records2) == 0)

		// case pool has tx
		pool.addTxs([]*types.Transaction{tx}, true)
		removeErr := os.Remove(pool.txRecordsFile)
		ast.Nil(removeErr)
		pool.handleRemoveTimeout(timer.RotateTxLocals)
		time.Sleep(2 * time.Millisecond)
		ast.Equal(uint64(1), pool.GetTotalPendingTxCount())
		records3, err := GetAllTxRecords(pool.txRecordsFile)
		ast.True(len(records3) == 1)
	})
}
