package storagemgr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitialize(t *testing.T) {
	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			err := Initialize(tc.kvType, repo.KVStorageCacheSize)
			require.Nil(t, err)
		})
	}
}

func TestInitializeWrongType(t *testing.T) {
	err := Initialize("unsupport", repo.KVStorageCacheSize)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unknow kv type unsupport")
}

func TestGet(t *testing.T) {
	dir := t.TempDir()

	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			err := Initialize(tc.kvType, repo.KVStorageCacheSize)
			require.Nil(t, err)

			s, err := Open(repo.GetStoragePath(dir+tc.kvType, BlockChain))
			require.Nil(t, err)
			require.NotNil(t, s)
		})
	}
}
