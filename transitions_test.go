package fsm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	fsm "github.com/whynot00/go-telegram-fsm"
)

func TestFSM_Transition_SetsState(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(42)

	f.Transition(userID, "someState")

	state := f.CurrentState(userID)
	require.Equal(t, fsm.StateFSM("someState"), state)
}

func TestFSM_Transition_StateDefault_CleansCache(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(100)

	// Заносим в localStorage произвольные данные
	f.Set(userID, "key1", "value1")

	// Проверяем, что данные есть
	val, ok := f.Get(userID, "key1")
	require.True(t, ok)
	require.Equal(t, "value1", val)

	// Устанавливаем состояние StateDefault — кэш должен очиститься
	f.Transition(userID, fsm.StateDefault)

	// Проверяем, что кэш удалён
	_, ok = f.Get(userID, "key1")
	require.False(t, ok)

	// Проверяем состояние
	state := f.CurrentState(userID)
	require.Equal(t, fsm.StateDefault, state)
}

func TestFSM_Finish_ResetsStateToDefault(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(7)

	f.Transition(userID, "anyState")
	stateBefore := f.CurrentState(userID)
	require.NotEqual(t, fsm.StateDefault, stateBefore)

	f.Finish(userID)
	stateAfter := f.CurrentState(userID)
	require.Equal(t, fsm.StateDefault, stateAfter)
}

func TestFSM_CurrentState_InitialDefault(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(999)

	state := f.CurrentState(userID)
	require.Equal(t, fsm.StateDefault, state)
}
