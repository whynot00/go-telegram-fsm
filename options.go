package fsm

import (
	"time"

	"github.com/whynot00/go-telegram-fsm/storage"
)

// Option configures the FSM instance.
type Option func(*FSM)

// WithStorage sets a custom Storage implementation.
func WithStorage(s storage.Storage) Option {
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

func WithTTL(t time.Duration) Option {

	return func(f *FSM) {
		f.ttl = t
	}
}

func WithCleanupInterval(t time.Duration) Option {
	return func(f *FSM) {
		f.cleanupInterval = t
	}
}
