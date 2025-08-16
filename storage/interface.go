package storage

import "github.com/whynot00/go-telegram-fsm/media"

// Storage defines the behaviour for caching user data and media.
type Storage interface {
	Set(userID int64, key string, value any)
	Get(userID int64, key string) (any, bool)
	SetMedia(userID int64, mediaGroupID string, file media.File)
	GetMedia(userID int64, mediaGroupID string) (*media.MediaData, bool)
	CleanMediaCache(userID int64, mediaGroupID string) bool
	CleanCache(userID int64)
}
