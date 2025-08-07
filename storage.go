package fsm

import "sync"

// cacheData holds a concurrent map for user-specific cached data.
type cacheData struct {
	data sync.Map
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

// CleanCache removes all cached data for the given user.
func (f *FSM) CleanCache(userID int64) {
	f.localStorage.Delete(userID)
}
