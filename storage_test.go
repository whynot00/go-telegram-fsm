package fsm_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	fsm "github.com/whynot00/go-telegram-fsm"
	"github.com/whynot00/go-telegram-fsm/media"
)

func TestSetAndGet(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(123)

	f.Set(userID, "key1", "value1")
	f.Set(userID, "key2", 42)

	val1, ok1 := f.Get(userID, "key1")
	require.True(t, ok1)
	require.Equal(t, "value1", val1)

	val2, ok2 := f.Get(userID, "key2")
	require.True(t, ok2)
	require.Equal(t, 42, val2)

	_, ok3 := f.Get(userID, "nonexistent")
	require.False(t, ok3)

	_, ok4 := f.Get(9999, "key1")
	require.False(t, ok4)
}

func TestGetMedia_Empty(t *testing.T) {
	f := fsm.New(context.Background())

	require.False(t, f.CleanMediaCache(1, "grp"))

	_, ok := f.GetMedia(1, "grp")
	require.False(t, ok)
}

func TestIsolation_UserAndGroup(t *testing.T) {
	f := fsm.New(context.Background())

	f.SetMedia(1, "A", media.File{FileID: "x1"})
	f.SetMedia(1, "B", media.File{FileID: "x2"})
	f.SetMedia(2, "A", media.File{FileID: "y1"})

	md1A, ok := f.GetMedia(1, "A")
	require.True(t, ok)
	require.Equal(t, []string{"x1"}, fileIDs(md1A.Files()))

	md1B, ok := f.GetMedia(1, "B")
	require.True(t, ok)
	require.Equal(t, []string{"x2"}, fileIDs(md1B.Files()))

	md2A, ok := f.GetMedia(2, "A")
	require.True(t, ok)
	require.Equal(t, []string{"y1"}, fileIDs(md2A.Files()))

	require.True(t, f.CleanMediaCache(1, "A"))

	_, ok = f.GetMedia(1, "A")
	require.False(t, ok)

	md1B, ok = f.GetMedia(1, "B")
	require.True(t, ok)
	require.Equal(t, []string{"x2"}, fileIDs(md1B.Files()))

	md2A, ok = f.GetMedia(2, "A")
	require.True(t, ok)
	require.Equal(t, []string{"y1"}, fileIDs(md2A.Files()))

	require.False(t, f.CleanMediaCache(1, "A"))
}

func TestLastUpdate_ResetOnSet(t *testing.T) {
	f := fsm.New(context.Background())
	uid, gid := int64(42), "g"

	f.SetMedia(uid, gid, media.File{FileID: "a"})
	md1, ok := f.GetMedia(uid, gid)
	require.True(t, ok)

	const d = 5 * time.Millisecond
	time.Sleep(10 * time.Millisecond)
	require.True(t, md1.Elapsed(d), "elapsed should be true after waiting")

	f.SetMedia(uid, gid, media.File{FileID: "b"})
	md2, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.Same(t, md1, md2)
	require.False(t, md2.Elapsed(d))
	time.Sleep(d + time.Millisecond)
	require.True(t, md2.Elapsed(d))
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
			f.SetMedia(uid, gid, media.File{FileID: fmt.Sprintf("id-%d", i)})
		}()
	}
	wg.Wait()

	md, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.Len(t, md.Files(), N)

	got := make(map[string]int, N)
	for _, f := range md.Files() {
		got[f.FileID]++
	}
	for i := 0; i < N; i++ {
		require.Equal(t, 1, got[fmt.Sprintf("id-%d", i)])
	}

	require.True(t, f.CleanMediaCache(uid, gid))
	_, ok = f.GetMedia(uid, gid)
	require.False(t, ok)
}

func TestPointerSemantics_NoExtraStoreNeeded(t *testing.T) {
	f := fsm.New(context.Background())
	uid, gid := int64(10), "g"

	f.SetMedia(uid, gid, media.File{FileID: "x"})
	md1, ok := f.GetMedia(uid, gid)
	require.True(t, ok)

	f.SetMedia(uid, gid, media.File{FileID: "y"})

	md2, ok := f.GetMedia(uid, gid)
	require.True(t, ok)
	require.Same(t, md1, md2)
	require.ElementsMatch(t, []string{"x", "y"}, fileIDs(md2.Files()))

	require.True(t, f.CleanMediaCache(uid, gid))
	_, ok = f.GetMedia(uid, gid)
	require.False(t, ok)
}

func TestCleanCache(t *testing.T) {
	f := fsm.New(context.Background())
	userID := int64(555)

	f.Set(userID, "key", "val")
	f.SetMedia(userID, "md_gr_1", media.File{FileID: "asd"})

	val, ok := f.Get(userID, "key")
	require.True(t, ok)
	require.Equal(t, "val", val)

	f.CleanCache(userID)

	_, ok = f.Get(userID, "key")
	require.False(t, ok)
}

func fileIDs(files []media.File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.FileID
	}
	return out
}
