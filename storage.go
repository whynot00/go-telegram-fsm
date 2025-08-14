package fsm

import (
	"sync"

	"github.com/whynot00/go-telegram-fsm/media"
)

// cacheData holds a concurrent map for user-specific cached data.
type cacheData struct {
	data sync.Map
	mu   sync.Mutex
}

// Set stores a key-value pair in the user's local cache.
func (f *FSM) Set(userID int64, key string, value any) {
	cache, _ := f.localStorage.LoadOrStore(userID, &cacheData{})
	cache.(*cacheData).data.Store(key, value)
}

// Get retrieves a value by key from the user's local cache.
// Returns the value and true if found, or nil and false otherwise.
func (f *FSM) Get(userID int64, key string) (any, bool) {
	if cache, ok := f.localStorage.Load(userID); ok {
		return cache.(*cacheData).data.Load(key)
	}
	return nil, false
}

// SetMedia adds a fileID to the specified mediaGroupID for a given user.
// Creates nested structures in localStorage if they don't exist yet.
func (f *FSM) SetMedia(userID int64, mediaGroupID string, file media.File) {
	userVal, _ := f.localStorage.LoadOrStore(userID, &cacheData{})
	userCache := userVal.(*cacheData)

	mediaVal, _ := userCache.data.LoadOrStore("media", &cacheData{})
	mediaCache := mediaVal.(*cacheData)

	mediaCache.mu.Lock()
	defer mediaCache.mu.Unlock()

	val, _ := mediaCache.data.LoadOrStore(mediaGroupID, &MediaData{})
	md := val.(*MediaData)
	md.addFile(file)
	md.touch()
	// mediaCache.data.Store(mediaGroupID, md) // не обязательно, md — тот же указатель
}

// GetMedia retrieves the MediaData for the given mediaGroupID and user.
// Returns nil and false if no data exists.
func (f *FSM) GetMedia(userID int64, mediaGroupID string) (*MediaData, bool) {
	userVal, ok := f.localStorage.Load(userID)
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
	return val.(*MediaData), true
}

func (f *FSM) CleanMediaCache(userID int64, mediaGroupID string) bool {
	userVal, ok := f.localStorage.Load(userID)
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
func (f *FSM) CleanCache(userID int64) {
	f.localStorage.Delete(userID)
}
