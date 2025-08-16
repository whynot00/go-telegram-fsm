package fsm

import "github.com/whynot00/go-telegram-fsm/media"

// Set stores a key-value pair for the user using configured storage.
func (f *FSM) Set(userID int64, key string, value any) {
	f.storage.Set(userID, key, value)
}

// Get retrieves a cached value by key for the user.
func (f *FSM) Get(userID int64, key string) (any, bool) {
	return f.storage.Get(userID, key)
}

// SetMedia stores a media file for the specified media group.
func (f *FSM) SetMedia(userID int64, mediaGroupID string, file media.File) {
	f.storage.SetMedia(userID, mediaGroupID, file)
}

// GetMedia returns media data for the specified media group.
func (f *FSM) GetMedia(userID int64, mediaGroupID string) (*MediaData, bool) {
	return f.storage.GetMedia(userID, mediaGroupID)
}

// CleanMediaCache removes cached media for the user and group.
func (f *FSM) CleanMediaCache(userID int64, mediaGroupID string) bool {
	return f.storage.CleanMediaCache(userID, mediaGroupID)
}

// CleanCache removes all cached data for the user.
func (f *FSM) CleanCache(userID int64) {
	f.storage.CleanCache(userID)
}
