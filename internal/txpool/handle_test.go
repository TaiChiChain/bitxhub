package txpool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/types"
	"github.com/axiomesh/axiom-ledger/internal/components/timer"
)

func TestHandleRemoveTimeoutEvent(t *testing.T) {
	ast := assert.New(t)
	pool := mockTxPoolImpl[types.Transaction, *types.Transaction](t)
	pool.toleranceRemoveTime = 1 * time.Millisecond
	err := pool.Start()
	defer pool.Stop()
	ast.Nil(err)

	s, err := types.GenerateSigner()
	ast.Nil(err)
	txs := constructTxs(s, 4)

	// retry 2 times to test account empty flag change(empty -> not empty -> empty)
	for i := 0; i < 2; i++ {
		pool.AddRemoteTxs(txs)
		assert.Equal(t, uint64(4), pool.GetTotalPendingTxCount())

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
}
