package storagemgr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestCachedStorage(t *testing.T) {
	err := Initialize(repo.KVStorageTypePebble, repo.KVStorageCacheSize, repo.KVStorageSync, false)
	require.Nil(t, err)

	s, err := Open(repo.GetStoragePath(t.TempDir()))
	require.Nil(t, err)
	require.NotNil(t, s)

	c := NewCachedStorage(s, 10)

	tests := []struct {
		key   []byte
		value []byte
	}{
		{key: []byte("k1"), value: []byte("v1")},
		{key: []byte("k2"), value: []byte{}},
		{key: []byte{}, value: []byte("v3")},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("non_batch_%d", i), func(t *testing.T) {
			require.Nil(t, c.Get(tt.key))
			require.False(t, c.Has(tt.key))

			c.Put(tt.key, tt.value)
			require.EqualValues(t, tt.value, c.Get(tt.key))
			require.True(t, c.Has(tt.key))

			c.Delete(tt.key)
			require.Nil(t, c.Get(tt.key))
			require.False(t, c.Has(tt.key))
		})
	}

	t.Run("batch", func(t *testing.T) {
		b := c.NewBatch()
		keys := [][]byte{
			[]byte("k1"),
			[]byte("k2"),
			{},
			[]byte("k4"),
		}
		vals := [][]byte{
			[]byte("v1"),
			{},
			[]byte("v3"),
			[]byte("v4"),
		}

		b.Put(keys[0], vals[0])
		c.Put(keys[1], vals[1])
		b.Put(keys[2], vals[2])
		c.Put(keys[3], vals[3])

		require.Nil(t, c.Get(keys[0]))
		require.False(t, c.Has(keys[0]))

		require.EqualValues(t, vals[1], c.Get(keys[1]))
		require.True(t, c.Has(keys[1]))

		require.Nil(t, c.Get(keys[2]))
		require.False(t, c.Has(keys[2]))

		require.EqualValues(t, vals[3], c.Get(keys[3]))
		require.True(t, c.Has(keys[3]))

		b.Delete(keys[0])
		b.Delete(keys[1])
		b.Commit()

		require.Nil(t, c.Get(keys[0]))
		require.False(t, c.Has(keys[0]))

		require.Nil(t, c.Get(keys[1]))
		require.False(t, c.Has(keys[1]))

		require.EqualValues(t, vals[2], c.Get(keys[2]))
		require.True(t, c.Has(keys[2]))

		require.EqualValues(t, vals[3], c.Get(keys[3]))
		require.True(t, c.Has(keys[3]))
	})

	t.Run("batch_reset", func(t *testing.T) {
		kv, err := Open(repo.GetStoragePath(t.TempDir()))
		require.Nil(t, err)
		require.NotNil(t, kv)
		c = NewCachedStorage(kv, 10)
		b := c.NewBatch()
		keys := [][]byte{
			[]byte("k111"),
			[]byte("k2222"),
			{},
			[]byte("k44"),
		}
		vals := [][]byte{
			[]byte("v111"),
			{},
			[]byte("v32"),
			[]byte("v4"),
		}

		b.Put(keys[0], vals[0])
		c.Put(keys[1], vals[1])
		require.Equal(t, b.Size(), len(keys[0])+len(vals[0]))
		b.Commit()
		b.Reset()

		require.EqualValues(t, vals[0], c.Get(keys[0]))
		require.True(t, c.Has(keys[0]))

		b.Put(keys[2], vals[2])
		c.Put(keys[3], vals[3])

		require.EqualValues(t, vals[1], c.Get(keys[1]))
		require.True(t, c.Has(keys[1]))

		require.Nil(t, c.Get(keys[2]))
		require.False(t, c.Has(keys[2]))

		require.EqualValues(t, vals[3], c.Get(keys[3]))
		require.True(t, c.Has(keys[3]))

		b.Delete(keys[0])
		b.Delete(keys[1])
		require.Equal(t, b.Size(), len(keys[2])+len(vals[2])+len(keys[0])+len(keys[1]))
		b.Commit()

		require.Nil(t, c.Get(keys[0]))
		require.False(t, c.Has(keys[0]))

		require.Nil(t, c.Get(keys[1]))
		require.False(t, c.Has(keys[1]))

		require.EqualValues(t, vals[2], c.Get(keys[2]))
		require.True(t, c.Has(keys[2]))

		require.EqualValues(t, vals[3], c.Get(keys[3]))
		require.True(t, c.Has(keys[3]))
	})

}
