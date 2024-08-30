package txpool

import (
	"encoding/binary"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
)

func TestHandleRemoveTimeoutEvent(t *testing.T) {
	t.Parallel()
	t.Run("remove timeout txs, expect priority and batched txs", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.toleranceRemoveTime = 10 * time.Millisecond
		pool.chainState.EpochInfo.ConsensusParams.BlockMaxTxNum = 4
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

		s, err := types.GenerateSigner()
		ast.Nil(err)
		txs := constructTxs(s, 5) // 4 batched txs + 1 priority tx
		pool.AddRemoteTxs(txs)
		<-ch

		// wait to ensure that tx had been timeout
		time.Sleep(12 * time.Millisecond)
		ast.Equal(uint64(5), pool.GetTotalPendingTxCount())
		ast.Equal(4, len(pool.txStore.batchedTxs))
		ast.Equal(1, len(pool.txStore.batchesCache))
		if pool.enablePricePriority {
			ast.Equal(1, getPrioritySize(pool))
		} else {
			ast.Equal(5, getPrioritySize(pool))
		}
		ast.Equal(uint64(1), pool.txStore.priorityNonBatchSize)

		pool.handleRemoveTimeout(RemoveTx)
		ast.Equal(uint64(5), pool.GetTotalPendingTxCount(), "ignore batched and priority txs")
		pool.Stop()
	})

	t.Run("remove Event successful", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.toleranceRemoveTime = 5 * time.Millisecond
		err := pool.Start()
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
			ast.Equal(0, getPrioritySize(pool))
			ast.Equal(4, pool.txStore.parkingLotIndex.size())
			ast.Equal(uint64(0), pool.txStore.priorityNonBatchSize)

			// sleep a while to trigger the remove timeout event
			time.Sleep(6 * time.Millisecond)
			pool.handleRemoveTimeout(RemoveTx)
			assert.Equal(t, uint64(0), pool.GetTotalPendingTxCount())

			assert.Equal(t, 1, len(pool.txStore.nonceCache.commitNonces))
			assert.Equal(t, 0, len(pool.txStore.nonceCache.pendingNonces))
			assert.Equal(t, 1, len(pool.txStore.allTxs))
		}

		pool.cleanEmptyAccountTime = 5 * time.Millisecond
		// sleep a while to trigger the clean empty account timeout event
		time.Sleep(6 * time.Millisecond)
		pool.handleRemoveTimeout(CleanEmptyAccount)
		assert.Equal(t, uint64(0), pool.GetTotalPendingTxCount())
		assert.Equal(t, 0, len(pool.txStore.nonceCache.commitNonces))
		assert.Equal(t, 0, len(pool.txStore.nonceCache.pendingNonces))
		assert.Equal(t, 0, len(pool.txStore.allTxs))
		pool.Stop()
	})

	t.Run("rotate tx locals", func(t *testing.T) {
		ast := assert.New(t)
		pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
		pool.rotateTxLocalsInterval = 1 * time.Millisecond
		err := pool.Start()
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
		pool.handleRemoveTimeout(RotateTxLocals)
		ast.Equal(uint64(0), pool.GetTotalPendingTxCount())
		records2, err := GetAllTxRecords(pool.txRecordsFile)
		ast.True(len(records2) == 0)

		// case pool has tx
		_, err = pool.addTx(tx, true)
		ast.Nil(err)
		removeErr := os.Remove(pool.txRecordsFile)
		ast.Nil(removeErr)
		pool.handleRemoveTimeout(RotateTxLocals)
		time.Sleep(2 * time.Millisecond)
		ast.Equal(uint64(1), pool.GetTotalPendingTxCount())
		records3, err := GetAllTxRecords(pool.txRecordsFile)
		ast.True(len(records3) == 1)
		pool.Stop()
	})
}
