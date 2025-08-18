package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/whynot00/go-telegram-fsm/media"
)

func f(tpe, id string) media.File {
	return media.File{Type: tpe, FileID: id}
}

func TestSetAndGet(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)

	store.Set(ctx, userID, "key1", "value1")

	v, ok := store.Get(ctx, userID, "key1")
	if !ok || v.(string) != "value1" {
		t.Errorf("expected value1, got %#v", v)
	}

	_, ok = store.Get(ctx, userID, "nonexistent")
	if ok {
		t.Error("expected not found for nonexistent key")
	}
}

func TestSetMediaAndGetMedia(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(42)
	groupID := "grp1"

	// add files
	store.SetMedia(ctx, userID, groupID, f("photo", "id1"))
	store.SetMedia(ctx, userID, groupID, f("video", "id2"))

	// retrieve
	md, ok := store.GetMedia(ctx, userID, groupID)
	if !ok {
		t.Fatal("expected media group to exist")
	}

	files := md.Files()
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestCleanMediaCache(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(99)
	groupID := "g"

	// add file
	store.SetMedia(ctx, userID, groupID, f("photo", "id"))
	if _, ok := store.GetMedia(ctx, userID, groupID); !ok {
		t.Fatal("media not found after set")
	}

	// clean it
	ok := store.CleanMediaCache(ctx, userID, groupID)
	if !ok {
		t.Error("expected CleanMediaCache to succeed")
	}
	if _, ok := store.GetMedia(ctx, userID, groupID); ok {
		t.Error("expected media to be gone after CleanMediaCache")
	}

	// try cleaning non-existing
	ok = store.CleanMediaCache(ctx, userID, "nonexistent")
	if ok {
		t.Error("expected false when cleaning nonexistent")
	}
}

func TestCleanCache(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(100)

	store.Set(ctx, userID, "k", "v")
	store.CleanCache(ctx, userID)

	_, ok := store.Get(ctx, userID, "k")
	if ok {
		t.Error("expected cache to be cleaned")
	}
}

func TestConcurrencySafety(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(1)
	groupID := "grp"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.Set(ctx, userID, "key", i)
			store.SetMedia(ctx, userID, groupID, f("photo", string(rune('A'+i%26))))
			_, _ = store.Get(ctx, userID, "key")
			_, _ = store.GetMedia(ctx, userID, groupID)
		}(i)
	}
	wg.Wait()

	// At least last key should exist
	if v, ok := store.Get(ctx, userID, "key"); !ok {
		t.Error("expected key to exist after concurrency")
	} else {
		_ = v
	}
}

func TestMediaTouchAndElapsed(t *testing.T) {
	store := NewMemoryStorage(30*time.Second, 30*time.Minute)
	ctx := context.Background()
	userID := int64(777)
	groupID := "media"

	store.SetMedia(ctx, userID, groupID, f("photo", "123"))
	md, ok := store.GetMedia(ctx, userID, groupID)
	if !ok {
		t.Fatal("media not found")
	}
	if md.Elapsed(time.Hour) {
		t.Error("just touched media should not be elapsed for 1h")
	}
}
