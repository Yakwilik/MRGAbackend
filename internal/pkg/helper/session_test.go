package helper

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSessionIDFromContextMD(t *testing.T) {
	_, got := SessionIDFromContextMD(context.Background())

	require.False(t, got)
}
