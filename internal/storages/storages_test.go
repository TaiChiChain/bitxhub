package storages

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestInitialize")
	require.Nil(t, err)

	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			err = Initialize(dir+tc.kvType, tc.kvType)
			require.Nil(t, err)

			// Initialize twice
			err = Initialize(dir+tc.kvType, tc.kvType)
			require.NotNil(t, err)
			require.Contains(t, err.Error(), "create blockchain storage")
		})
	}
}

func TestGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestGet")
	require.Nil(t, err)

	testcase := map[string]struct {
		kvType string
	}{
		"leveldb": {kvType: "leveldb"},
		"pebble":  {kvType: "pebble"},
	}

	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			err = Initialize(dir+tc.kvType, tc.kvType)
			require.Nil(t, err)

			s, err := Get(BlockChain)
			require.Nil(t, err)
			require.NotNil(t, s)

			s, err = Get("WrongName")
			require.NotNil(t, err)
			require.Nil(t, s)
		})
	}
}
