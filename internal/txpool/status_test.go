package txpool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTurnOnStatus(t *testing.T) {
	st := newPoolStatusMgr()
	require.False(t, st.InOne(ReadyGenerateBatch, PoolFull, PoolEmpty))
	st.On(ReadyGenerateBatch)
	require.True(t, st.InOne(ReadyGenerateBatch, PoolFull, PoolEmpty))
	require.False(t, st.InOne(PoolFull))
	require.True(t, st.In(ReadyGenerateBatch))
	st.Off(ReadyGenerateBatch)
	require.False(t, st.InOne(ReadyGenerateBatch, PoolFull, PoolEmpty))

	st.On(HasPendingRequest)
	require.True(t, st.In(HasPendingRequest))

	st.On(HasPendingRequest)
	require.True(t, st.In(HasPendingRequest))

	st.Off(HasPendingRequest)
	require.False(t, st.InOne(HasPendingRequest))
}
