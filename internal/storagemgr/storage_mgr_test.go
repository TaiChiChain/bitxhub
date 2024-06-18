package storagemgr

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-ledger/pkg/repo"
)

func TestInitializeWrongType(t *testing.T) {
	repoConfig := &repo.Config{Storage: repo.Storage{
		KvType:      "unsupport",
		Sync:        false,
		KVCacheSize: repo.KVStorageCacheSize,
		Pebble:      repo.Pebble{},
	}, Monitor: repo.Monitor{Enable: false}}
	err := Initialize(repoConfig)
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
			repoConfig := &repo.Config{Storage: repo.Storage{
				KvType:      tc.kvType,
				Sync:        false,
				KVCacheSize: repo.KVStorageCacheSize,
				Pebble:      repo.Pebble{},
			}, Monitor: repo.Monitor{Enable: false}}
			err := Initialize(repoConfig)
			require.Nil(t, err)

			s, err := Open(repo.GetStoragePath(dir+tc.kvType, BlockChain))
			require.Nil(t, err)
			require.NotNil(t, s)
		})
	}
}
