package txpool

import (
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

	err = records.rotate(pool.txStore.allTxs)
	assert.Nil(t, err)

	tx := constructTx(s, 0)
	err = records.insert(tx)
	assert.Nil(t, err)

	err = records.load(pool.addTxs)
	assert.Nil(t, err)
}
