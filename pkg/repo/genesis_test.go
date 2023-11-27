package repo

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisConfig(t *testing.T) {
	repoPath := t.TempDir()
	cnf, err := LoadGenesisConfig(repoPath)
	require.Nil(t, err)
	require.Equal(t, uint64(0x54c), cnf.ChainID)
	cnf.ChainID = 157
	err = writeConfigWithEnv(path.Join(repoPath, genesisCfgFileName), cnf)
	require.Nil(t, err)
	cnf2, err := LoadGenesisConfig(repoPath)
	require.Nil(t, err)
	require.Equal(t, uint64(0x9d), cnf2.ChainID)
}
