package memory

import (
	"context"
	"sync"

	"github.com/whynot00/go-telegram-fsm/media"
)

// cacheData holds a concurrent map for user-specific cached data.
type cacheData struct {
	data sync.Map
	mu   sync.Mutex
}

// MemoryStorage implements in-memory Storage using sync.Map.
type MemoryStorage struct {
	storage sync.Map
}

// NewMemoryStorage creates a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// Set stores a key-value pair in the user's local cache.
func (m *MemoryStorage) Set(ctx context.Context, userID int64, key string, value any) {
	cache, _ := m.storage.LoadOrStore(userID, &cacheData{})
	cache.(*cacheData).data.Store(key, value)
}

// Get retrieves a value by key from the user's local cache.
func (m *MemoryStorage) Get(ctx context.Context, userID int64, key string) (any, bool) {
	if cache, ok := m.storage.Load(userID); ok {
		return cache.(*cacheData).data.Load(key)
	}
	return nil, false
}

// SetMedia adds a file to the specified mediaGroupID for a given user.
func (m *MemoryStorage) SetMedia(ctx context.Context, userID int64, mediaGroupID string, file media.File) {
	userVal, _ := m.storage.LoadOrStore(userID, &cacheData{})
	userCache := userVal.(*cacheData)

	mediaVal, _ := userCache.data.LoadOrStore("media", &cacheData{})
	mediaCache := mediaVal.(*cacheData)

	mediaCache.mu.Lock()
	defer mediaCache.mu.Unlock()

	val, _ := mediaCache.data.LoadOrStore(mediaGroupID, &media.MediaData{})
	md := val.(*media.MediaData)
	md.AddFile(file)
	md.Touch()
}

// GetMedia retrieves the MediaData for the given mediaGroupID and user.
func (m *MemoryStorage) GetMedia(ctx context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool) {
	userVal, ok := m.storage.Load(userID)
	if !ok {
		return nil, false
	}
	userCache := userVal.(*cacheData)

	mediaVal, ok := userCache.data.Load("media")
	if !ok {
		return nil, false
	}
	mediaCache := mediaVal.(*cacheData)

	val, ok := mediaCache.data.Load(mediaGroupID)
	if !ok {
		return nil, false
	}
	return val.(*media.MediaData), true
}

// CleanMediaCache removes media cache for the given user and group.
func (m *MemoryStorage) CleanMediaCache(ctx context.Context, userID int64, mediaGroupID string) bool {
	userVal, ok := m.storage.Load(userID)
	if !ok {
		return false
	}
	userCache := userVal.(*cacheData)

	mediaVal, ok := userCache.data.Load("media")
	if !ok {
		return false
	}
	mediaCache := mediaVal.(*cacheData)

	mediaCache.mu.Lock()
	defer mediaCache.mu.Unlock()

	_, ok = mediaCache.data.Load(mediaGroupID)
	if !ok {
		return false
	}

	mediaCache.data.Delete(mediaGroupID)
	return true
}

// CleanCache removes all cached data for the given user.
func (m *MemoryStorage) CleanCache(ctx context.Context, userID int64) {
	m.storage.Delete(userID)
}
