package fsm

import (
	"context"

	"github.com/whynot00/go-telegram-fsm/media"
)

// Set stores a key-value pair for the user using configured storage.
func (f *FSM) Set(ctx context.Context, userID int64, key string, value any) {
	f.storage.Set(ctx, userID, key, value)
}

// Get retrieves a cached value by key for the user.
func (f *FSM) Get(ctx context.Context, userID int64, key string) (any, bool) {
	return f.storage.Get(ctx, userID, key)
}

// SetMedia stores a media file for the specified media group.
func (f *FSM) SetMedia(ctx context.Context, userID int64, mediaGroupID string, file media.File) {
	f.storage.SetMedia(ctx, userID, mediaGroupID, file)
}

// GetMedia returns media data for the specified media group.
func (f *FSM) GetMedia(ctx context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool) {
	return f.storage.GetMedia(ctx, userID, mediaGroupID)
}

// CleanMediaCache removes cached media for the user and group.
func (f *FSM) CleanMediaCache(ctx context.Context, userID int64, mediaGroupID string) bool {
	return f.storage.CleanMediaCache(ctx, userID, mediaGroupID)
}

// CleanCache removes all cached data for the user.
func (f *FSM) CleanCache(ctx context.Context, userID int64) {
	f.storage.CleanCache(ctx, userID)
}
