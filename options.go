package fsm

import (
	"time"

	"github.com/whynot00/go-telegram-fsm/storage"
)

// Option defines a configuration function for FSM.
// Options are applied when creating a new FSM instance.
type Option func(*FSM)

// WithStorage replaces the default in-memory storage with a custom Storage implementation.
// If set, FSM will not manage (own/close) the lifecycle of the provided storage.
func WithStorage(s storage.Storage) Option {
	return func(f *FSM) {
		f.storage = s
		f.ownsStorage = false
	}
}

// WithTTL sets the time-to-live (TTL) for FSM states.
// States that exceed this duration without usage will be removed.
func WithTTL(t time.Duration) Option {
	return func(f *FSM) {
		f.ttl = t
	}
}

// WithCleanupInterval sets how often the background cleanup worker runs
// to remove expired states from the FSM.
func WithCleanupInterval(t time.Duration) Option {
	return func(f *FSM) {
		f.cleanupInterval = t
	}
}
