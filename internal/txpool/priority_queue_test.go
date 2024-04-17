package txpool

import (
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
)

func TestPriorityQueue_Push(t *testing.T) {
	q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, log.NewWithModule("priorityQueue"))
	s, _ := types.GenerateSigner()
	tx := constructTx(s, 0)
	poolTx := &internalTransaction[types.Transaction, *types.Transaction]{
		rawTx: tx,
		local: true,
	}
	q.push(poolTx)
	require.Equal(t, poolTx, q.peek())
	require.Equal(t, uint64(1), q.size())

	actualTx := q.pop()
	require.Equal(t, poolTx.getNonce(), actualTx.getNonce())
	require.Equal(t, uint64(0), q.size())

	poolTxs := constructPoolTxListByGas(s, 5, big.NewInt(100))
	// lack tx2
	newTxs := append(poolTxs[1:2], poolTxs[3:]...)
	for _, pt := range newTxs {
		q.push(pt)
	}
	require.Equal(t, uint64(1), q.peek().getNonce())
	require.Equal(t, uint64(1), q.size(), "insert tx1")

	q.push(poolTxs[0])
	require.Equal(t, uint64(0), q.peek().getNonce())
	require.Equal(t, uint64(2), q.size(), "exist tx0 and tx1")
}

func TestPriorityQueue_Pop(t *testing.T) {
	q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, log.NewWithModule("priorityQueue"))
	s1, _ := types.GenerateSigner()
	from1 := s1.Addr.String()
	price := big.NewInt(100)

	s2, _ := types.GenerateSigner()
	from2 := s2.Addr.String()
	tx10 := constructPoolTxByGas(s1, 0, price)
	highPrice := new(big.Int).Add(price, big.NewInt(10))
	tx20 := constructPoolTxByGas(s2, 0, highPrice)
	q.push(tx10)
	require.Equal(t, from1, q.peek().getAccount())
	q.push(tx20)
	require.Equal(t, from2, q.pop().getAccount())
	lowPrice := new(big.Int).Sub(price, big.NewInt(10))
	tx21 := constructPoolTxByGas(s2, 1, lowPrice)
	q.push(tx21)
	require.Equal(t, from1, q.peek().getAccount())
}

func TestPriorityQueue_PushByMulti(t *testing.T) {
	q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, log.NewWithModule("priorityQueue"))
	s1, _ := types.GenerateSigner()
	from1 := s1.Addr.String()
	price := big.NewInt(100)

	s2, _ := types.GenerateSigner()
	from2 := s2.Addr.String()

	lowPrice := new(big.Int).Sub(price, big.NewInt(10))
	highPrice := new(big.Int).Add(price, big.NewInt(10))

	tx10 := constructPoolTxByGas(s1, 0, price)
	q.push(tx10)
	tx := q.peek()
	require.Equal(t, tx.getNonce(), uint64(0))
	require.Equal(t, tx.getAccount(), from1)

	tx11 := constructPoolTxByGas(s1, 1, price)
	q.push(tx11)
	tx = q.peek()
	require.Equal(t, tx.getNonce(), uint64(0))

	tx20 := constructPoolTxByGas(s2, 0, highPrice)
	q.push(tx20)
	tx = q.peek()
	require.Equal(t, tx.getAccount(), from2)
	tx21 := constructPoolTxByGas(s2, 1, highPrice)
	q.push(tx21)
	tx = q.pop()
	require.Equal(t, tx.getAccount(), from2, "remove tx20")
	require.Equal(t, tx.getNonce(), uint64(0))

	tx = q.peek()
	require.Equal(t, tx.getAccount(), from2)
	require.Equal(t, tx.getNonce(), uint64(1))
	require.Equal(t, 2, q.txsByPrice.length(), "tx21,tx10")

	tx12 := constructPoolTxByGas(s1, 2, price)
	tx13 := constructPoolTxByGas(s1, 3, price)
	tx22 := constructPoolTxByGas(s2, 2, lowPrice)
	q.push(tx12)
	q.push(tx13)
	q.push(tx22)
	require.Equal(t, q.txsByPrice.length(), 2, "tx21,tx10")
	require.Equal(t, q.accountsM[from1].length(), uint64(4), "tx10-tx13")
	require.Equal(t, q.accountsM[from2].length(), uint64(2), "tx21,tx22")

	txList := make([]*internalTransaction[types.Transaction, *types.Transaction], 0)
	expectTxList := []*internalTransaction[types.Transaction, *types.Transaction]{
		tx21, // max price
		tx10, // same price, sort by time
		tx11,
		tx12,
		tx13,
		tx22, // min price
	}
	for q.txsByPrice.length() > 0 {
		actualTx := q.pop()
		txList = append(txList, actualTx)

		if q.txsByPrice.length() > 0 {
			require.Equal(t, expectTxList[len(txList)], q.peek())
		}
	}

	require.Equal(t, 6, len(txList))
	require.Equal(t, expectTxList, txList)
}

func TestPriorityQueue_RemoveTx(t *testing.T) {
	q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, log.NewWithModule("priorityQueue"))
	s1, _ := types.GenerateSigner()
	from1 := s1.Addr.String()
	s2, _ := types.GenerateSigner()
	from2 := s2.Addr.String()
	price := big.NewInt(100)

	// tx0-tx4 with from1
	txListByFrom1 := constructPoolTxListByGas(s1, 5, price)
	txListByFrom2 := constructPoolTxListByGas(s2, 5, price)
	for _, tx := range txListByFrom1 {
		q.push(tx)
	}
	for _, tx := range txListByFrom2 {
		q.push(tx)
	}

	require.Equal(t, uint64(10), q.nonBatchSize)
	require.Equal(t, from1, q.peek().getAccount())
	require.Equal(t, uint64(0), q.peek().getNonce())

	q.removeTxBeforeNonce(from1, 0)
	require.Equal(t, from1, q.peek().getAccount(), "becasue tx1 with from1 is arrived earlier than tx0 with from2")
	require.Equal(t, 2, len(q.txsByPrice.dirtyAccounts))
	require.Equal(t, uint64(1), q.peek().getNonce())
	require.Equal(t, uint64(9), q.nonBatchSize)
	require.Equal(t, uint64(4), q.accountsM[from1].length())

	q.removeTxBeforeNonce(from1, 3)
	require.Equal(t, from1, q.peek().getAccount())
	require.Equal(t, uint64(4), q.peek().getNonce(), "tx0-tx3 had been removed")
	require.Equal(t, uint64(6), q.nonBatchSize)
	require.Equal(t, uint64(1), q.accountsM[from1].length())

	q.removeTxBehindNonce(txListByFrom2[3])
	require.Equal(t, uint64(4), q.nonBatchSize, "tx4_from1, tx0_from2, tx1_from2, tx2_from2")
	require.Equal(t, uint64(3), q.accountsM[from2].length())
}

func TestPriorityQueue_ReplaceTx(t *testing.T) {
	q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, log.NewWithModule("priorityQueue"))
	s1, _ := types.GenerateSigner()
	oldTx := constructPoolTxByGas(s1, 0, big.NewInt(100))
	newTx := constructPoolTxByGas(s1, 0, big.NewInt(200))
	q.push(oldTx)
	require.Equal(t, uint64(1), q.size())

	q.replaceTx(newTx)
	require.True(t, q.peek().getGasPrice().Cmp(newTx.getGasPrice()) == 0)
	require.Equal(t, uint64(1), q.size())
}

func Benchmark_PriorityQueue(t *testing.B) {
	testTable := []struct {
		name        string
		unsortedTxs []*internalTransaction[types.Transaction, *types.Transaction]
	}{
		// with 1 account
		{
			name:        "1000 transactions",
			unsortedTxs: generateTxs(1, 1000),
		},
		{
			name:        "10000 transactions",
			unsortedTxs: generateTxs(1, 10000),
		},

		// with 10 account
		{
			name:        "1000 transactions",
			unsortedTxs: generateTxs(10, 100),
		},
		{
			name:        "10000 transactions",
			unsortedTxs: generateTxs(10, 1000),
		},

		// with 100 account
		{
			name:        "1000 transactions",
			unsortedTxs: generateTxs(100, 10),
		},
		{
			name:        "10000 transactions",
			unsortedTxs: generateTxs(100, 100),
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(b *testing.B) {
			for i := 0; i < t.N; i++ {
				logger := log.NewWithModule("priorityQueue")
				logger.Logger.SetLevel(logrus.ErrorLevel)
				q := newPriorityQueue[types.Transaction, *types.Transaction](getNonce, logger)
				lo.ForEach(tt.unsortedTxs, func(tx *internalTransaction[types.Transaction, *types.Transaction], _ int) {
					q.push(tx)
				})
				preNonceM := make(map[string]uint64)
				prePrice := q.peek().getGasPrice()
				preNonceM[q.peek().getAccount()] = q.peek().getNonce()

				for q.size() > 0 {
					tx := q.pop()
					if nonce, ok := preNonceM[tx.getAccount()]; !ok || nonce == 0 {
						prePrice = tx.getGasPrice()
						preNonceM[tx.getAccount()] = tx.getNonce()
						continue
					}
					if prePrice.Cmp(tx.getGasPrice()) < 0 {
						b.Errorf("pre price %v is less than current %v", prePrice, tx.getGasPrice())
						b.Fail()
					}
					if preNonceM[tx.getAccount()]+1 != tx.getNonce() {
						b.Errorf("pre nonce %v is less than current %v", preNonceM[tx.getAccount()], tx.getNonce())
						b.Fail()
					}
					prePrice = tx.getGasPrice()
					preNonceM[tx.getAccount()] = tx.getNonce()
				}
			}
		})
	}
}

func generateTxs(accountCount, perAccountTxCount int) []*internalTransaction[types.Transaction, *types.Transaction] {
	lock := new(sync.Mutex)
	txs := make([]*internalTransaction[types.Transaction, *types.Transaction], 0)
	wg := new(sync.WaitGroup)
	wg.Add(accountCount)
	for i := 0; i < accountCount; i++ {
		time.Sleep(1 * time.Millisecond)
		go func() {
			defer wg.Done()
			s, _ := types.GenerateSigner()
			source := rand.NewSource(int64(i))
			r := rand.New(source)
			accountTxs := make([]*internalTransaction[types.Transaction, *types.Transaction], perAccountTxCount)
			randomGasPrice := big.NewInt(int64(r.Intn(10000000)))
			for j := 0; j < perAccountTxCount; j++ {
				accountTxs[j] = constructPoolTxByGas(s, uint64(j), randomGasPrice)
			}
			lock.Lock()
			txs = append(txs, accountTxs...)
			lock.Unlock()
		}()
	}
	wg.Wait()
	return txs
}

func getNonce(_ string) uint64 {
	return 0
}
