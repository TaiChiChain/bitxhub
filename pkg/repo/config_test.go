package repo

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	repoPath := t.TempDir()
	cnf, err := LoadConfig(repoPath)
	require.Nil(t, err)
	require.Equal(t, "rbft", cnf.Consensus.Type)
	require.Equal(t, int64(10000), cnf.JsonRPC.WriteLimiter.Capacity)
	require.Equal(t, int64(10000), cnf.JsonRPC.ReadLimiter.Capacity)
	require.Equal(t, false, cnf.JsonRPC.WriteLimiter.Enable)
	require.Equal(t, false, cnf.JsonRPC.ReadLimiter.Enable)
	cnf.Consensus.Type = "solo"
	cnf.JsonRPC.WriteLimiter.Capacity = 100
	cnf.JsonRPC.ReadLimiter.Capacity = 100
	cnf.JsonRPC.WriteLimiter.Enable = true
	cnf.JsonRPC.ReadLimiter.Enable = true
	err = writeConfigWithEnv(path.Join(repoPath, CfgFileName), cnf)
	require.Nil(t, err)
	cnf2, err := LoadConfig(repoPath)
	require.Nil(t, err)
	require.Equal(t, "solo", cnf2.Consensus.Type)
	require.Equal(t, int64(100), cnf2.JsonRPC.WriteLimiter.Capacity)
	require.Equal(t, int64(100), cnf2.JsonRPC.ReadLimiter.Capacity)
	require.Equal(t, true, cnf.JsonRPC.WriteLimiter.Enable)
	require.Equal(t, true, cnf.JsonRPC.ReadLimiter.Enable)
}
