package fsm

import (
	"github.com/whynot00/go-telegram-fsm/media"
)

// Storage defines the behaviour for caching user data and media.
type Storage interface {
	Set(userID int64, key string, value any)
	Get(userID int64, key string) (any, bool)
	SetMedia(userID int64, mediaGroupID string, file media.File)
	GetMedia(userID int64, mediaGroupID string) (*MediaData, bool)
	CleanMediaCache(userID int64, mediaGroupID string) bool
	CleanCache(userID int64)
}

// Option configures the FSM instance.
type Option func(*FSM)

// WithStorage sets a custom Storage implementation.
func WithStorage(s Storage) Option {
	return func(f *FSM) {
		f.storage = s
	}
}

// WithRedisStorage configures FSM to use Redis as storage.
func WithRedisStorage(addr, username, password string, db int) Option {
	return func(f *FSM) {
		f.storage = NewRedisStorage(addr, username, password, db)
	}
}
