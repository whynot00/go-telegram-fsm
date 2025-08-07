package fsm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	fsm "github.com/whynot00/go-telegram-fsm"
)

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	f := fsm.New(ctx)

	ctx = context.WithValue(ctx, fsm.FsmKey, f)
	result := fsm.FromContext(ctx)

	require.NotNil(t, result)
	require.Equal(t, f, result)
}

func TestFromContext_Nil(t *testing.T) {
	ctx := context.Background()
	require.Nil(t, fsm.FromContext(ctx))
}
