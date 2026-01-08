package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrackedBuffer(t *testing.T) {
	tb := newTrackedBuffer()

	// Test initial size
	require.Equal(t, int64(0), tb.size())

	// Write some bytes
	data := []byte("hello")
	err := tb.write(data)
	require.NoError(t, err)
	require.Equal(t, int64(len(data)), tb.size())

	// Write more bytes
	more := []byte("world")
	err = tb.write(more)
	require.NoError(t, err)
	require.Equal(t, int64(len(data)+len(more)), tb.size())

	// Get and reset
	out := tb.getAndReset()
	require.Equal(t, append(data, more...), out)
	require.Equal(t, int64(0), tb.size())
}
