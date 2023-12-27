package txpool

import (
	"bytes"
	"os"
	"testing"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/stretchr/testify/assert"
)

func TestTxRecords(t *testing.T) {
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	assert.Nil(t, err)
	records := pool.txRecords
	defer func() {
		cleanupErr := os.Remove(records.filePath)
		if cleanupErr != nil {
			t.Errorf("Failed to clean up file: %v", cleanupErr)
		}
	}()

	// test rotate
	err = records.rotate(pool.txStore.allTxs)
	assert.Nil(t, err)

	// test insert
	tx := constructTx(s, 0)
	err = records.insert(tx)
	assert.Nil(t, err)

	// test get
	all, err := GetAllTxRecords(records.filePath)
	assert.Nil(t, err)
	marshal, err := tx.RbftMarshal()
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(all[0], marshal))

	tx2 := &types.Transaction{}
	err = tx2.RbftUnmarshal(all[0])
	assert.Nil(t, err)
	assert.Equal(t, tx.RbftGetTxHash(), tx2.RbftGetTxHash())

	// test load
	err = records.load(pool.addTxs)
	assert.Nil(t, err)
}

func TestDevNull(t *testing.T) {
	devNull := devNull{}
	n, err := devNull.Write([]byte("test"))
	assert.Nil(t, err)
	assert.Equal(t, len("test"), n)

	err = devNull.Close()
	assert.Nil(t, err)
}

func TestTxRecords_load(t *testing.T) {
	// test not exist file
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	records := pool.txRecords
	err := records.load(pool.addTxs)
	assert.Nil(t, err) // if not exist, no error

	// test wrong read
	err = pool.Start()
	assert.Nil(t, err)
	defer pool.Stop()
	records.writer.Write([]byte{1, 0, 0, 0, 0, 0, 0, 0})
	err = records.load(pool.addTxs)
	assert.NotNil(t, err)

	// test wrong tx
	err = records.rotate(nil) // reset pb file
	assert.Nil(t, err)
	records.writer.Write([]byte{1, 0, 0, 0, 0, 0, 0, 0, 1})
	err = records.load(pool.addTxs)
	assert.Nil(t, err) // if not right tx, no error

	// test 1000+ txs load
	err = records.rotate(nil) // reset pb file
	assert.Nil(t, err)
	s, err := types.GenerateSigner()
	txs := constructTxs(s, TxRecordsBatchSize+1)
	for _, tx := range txs {
		records.insert(tx)
	}
	records.load(pool.addTxs)
}

func TestTxRecords_LoadMoreThanOneBatch(t *testing.T) {
	// init pool and txs
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	pool.Start() // used for init txrecord
	records := pool.txRecords
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	txs := constructTxs(s, TxRecordsBatchSize+1)
	for _, tx := range txs {
		records.insert(tx)
	}
	pool.Start() // used for test txrecord load
	allTxs := pool.txStore.allTxs
	var nums = 0
	for _, txmap := range allTxs {
		for _, item := range txmap.items {
			if item.local {
				nums++
			}
		}
	}
	assert.Equal(t, TxRecordsBatchSize+1, nums)
	pool.Stop()
}

func TestTxRecords_Rotate(t *testing.T) {
	// init pool and txs
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	pool.Start()
	defer pool.Stop()
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	tx1 := constructTx(s, 0)
	tx2 := constructTx(s, 1)
	pool.addTxs([]*types.Transaction{tx1, tx2}, true)
	s2, err := types.GenerateSigner()
	assert.Nil(t, err)
	tx3 := constructTx(s2, 0)
	pool.addTxs([]*types.Transaction{tx3}, false)

	// test rotate for locals and unlocals
	txs := pool.txStore.allTxs
	err = pool.txRecords.rotate(txs)
	assert.Nil(t, err)
}

func TestTxRecords_RotateMoreThanOneBatch(t *testing.T) {
	// init pool and txs
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	pool.Start()
	defer pool.Stop()
	s, err := types.GenerateSigner()
	txs := constructTxs(s, TxRecordsBatchSize+1)
	errs := pool.addTxs(txs, true)
	for _, e := range errs {
		assert.Nil(t, e)
	}
	err = os.Remove(pool.txRecords.filePath)
	err = pool.txRecords.rotate(pool.txStore.allTxs)
	records, err := GetAllTxRecords(pool.txRecords.filePath)
	assert.Nil(t, err)
	assert.True(t, len(records) == TxRecordsBatchSize+1)
}

func TestTxRecords_Insert(t *testing.T) {
	// init pool and txs
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	records := pool.txRecords

	// test insert no writer
	s, err := types.GenerateSigner()
	assert.Nil(t, err)
	tx1 := constructTx(s, 0)
	err = records.insert(tx1)
	assert.NotNil(t, err) // if not start pool first, records writer is null
}
