package txpool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeConfig(t *testing.T) {
	c := Config{
		PoolSize:               0,
		ToleranceTime:          0,
		ToleranceRemoveTime:    0,
		CleanEmptyAccountTime:  0,
		RotateTxLocalsInterval: 0,
	}
	c.sanitize()
	require.Equal(t, uint64(DefaultPoolSize), c.PoolSize)
	require.Equal(t, DefaultToleranceTime, c.ToleranceTime)
	require.Equal(t, DefaultToleranceRemoveTime, c.ToleranceRemoveTime)
	require.Equal(t, DefaultCleanEmptyAccountTime, c.CleanEmptyAccountTime)
	require.Equal(t, DefaultRotateTxLocalsInterval, c.RotateTxLocalsInterval)
}
