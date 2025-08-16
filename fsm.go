package fsm

import (
	"context"
	"sync"
	"time"

	"github.com/whynot00/go-telegram-fsm/storage"
)

// FSM implements a finite state machine for users, maintaining states and local cache.
type FSM struct {
	current sync.Map        // current state and last usage time keyed by user ID.
	storage storage.Storage // pluggable storage backend.
}

// stateData holds the FSM state and the timestamp of last update.
type stateData struct {
	state   StateFSM  // current FSM state
	lastUse time.Time // last update time of the state
}

// New creates a new FSM instance and starts a background worker
// to periodically clean up expired states.
// Storage backend can be customised via options.
func New(ctx context.Context, opts ...Option) *FSM {
	fsm := &FSM{
		current: sync.Map{},
		storage: NewMemoryStorage(),
	}

	for _, opt := range opts {
		opt(fsm)
	}

	// Start a goroutine for periodic cleanup of inactive states
	go fsm.startCleanupWorker(ctx, 30*time.Minute)

	return fsm
}
