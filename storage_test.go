package fsm_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

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

func TestGetMedia_Empty(t *testing.T) {
	f := fsm.New(context.Background())
	_, ok := f.GetMedia(1, "grp")
	require.False(t, ok)
}

func TestIsolation_UserAndGroup(t *testing.T) {
	f := fsm.New(context.Background())

	f.SetMedia(1, "A", "x1")
	f.SetMedia(1, "B", "x2")
	f.SetMedia(2, "A", "y1")

	md1A, ok := f.GetMedia(1, "A")
	require.True(t, ok)
	require.Equal(t, []string{"x1"}, md1A.FileIDs)

	md1B, ok := f.GetMedia(1, "B")
	require.True(t, ok)
	require.Equal(t, []string{"x2"}, md1B.FileIDs)

	md2A, ok := f.GetMedia(2, "A")
	require.True(t, ok)
	require.Equal(t, []string{"y1"}, md2A.FileIDs)
}

func TestLastUpdate_Increases(t *testing.T) {
	f := fsm.New(context.Background())
	uid, gid := int64(42), "g"

	f.SetMedia(uid, gid, "a")
	md1, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	t1 := md1.LastUpdate

	time.Sleep(10 * time.Millisecond)
	f.SetMedia(uid, gid, "b")
	md2, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.True(t, md2.LastUpdate.After(t1), "LastUpdate must increase")
}

func TestConcurrentAppend(t *testing.T) {
	f := fsm.New(context.Background())
	uid, gid := int64(7), "grp"
	const N = 100

	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			f.SetMedia(uid, gid, fmt.Sprintf("id-%d", i))
		}()
	}
	wg.Wait()

	md, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.Len(t, md.FileIDs, N)

	// Проверим, что все значения присутствуют (порядок не гарантирован).
	got := make(map[string]int, N)
	for _, v := range md.FileIDs {
		got[v]++
	}
	for i := 0; i < N; i++ {
		require.Equal(t, 1, got[fmt.Sprintf("id-%d", i)])
	}
}

func TestPointerSemantics_NoExtraStoreNeeded(t *testing.T) {
	f := fsm.New(context.Background())
	uid, gid := int64(10), "g"

	f.SetMedia(uid, gid, "x")
	md1, ok := f.GetMedia(uid, gid)
	require.True(t, ok)

	// Вновь добавляем — мутация того же *MediaData
	f.SetMedia(uid, gid, "y")

	md2, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.Same(t, md1, md2) // тот же указатель
	require.ElementsMatch(t, []string{"x", "y"}, md2.FileIDs)
}

func TestCleanCache(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(555)

	f.Set(userID, "key", "val")
	f.SetMedia(userID, "md_gr_1", "asd")

	val, ok := f.Get(userID, "key")

	require.True(t, ok)
	require.Equal(t, "val", val)

	// Очищаем кэш
	f.CleanCache(userID)

	_, ok = f.Get(userID, "key")
	require.False(t, ok)
}
