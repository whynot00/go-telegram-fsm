package fsm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	fsm "github.com/whynot00/go-telegram-fsm"
)

func TestSetAndGet(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(123)

	// Сохраняем данные
	f.Set(userID, "key1", "value1")
	f.Set(userID, "key2", 42)

	// Получаем и проверяем данные
	val1, ok1 := f.Get(userID, "key1")
	require.True(t, ok1)
	require.Equal(t, "value1", val1)

	val2, ok2 := f.Get(userID, "key2")
	require.True(t, ok2)
	require.Equal(t, 42, val2)

	// Запрос несуществующего ключа
	_, ok3 := f.Get(userID, "nonexistent")
	require.False(t, ok3)

	// Запрос несуществующего userID
	_, ok4 := f.Get(9999, "key1")
	require.False(t, ok4)
}

func TestCleanCache(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(555)

	f.Set(userID, "key", "val")

	val, ok := f.Get(userID, "key")
	require.True(t, ok)
	require.Equal(t, "val", val)

	// Очищаем кэш
	f.CleanCache(userID)

	_, ok = f.Get(userID, "key")
	require.False(t, ok)
}
