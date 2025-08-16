package fsm

import (
	"context"
	"sync"
	"time"
)

// FSM implements a finite state machine for users, maintaining states and local cache.
type FSM struct {
	current      sync.Map // current state and last usage time keyed by user ID.
	localStorage sync.Map // user-specific cached data keyed by user ID.
}

// stateData holds the FSM state and the timestamp of last update.
type stateData struct {
	state   StateFSM  // state is the current FSM state.
	lastUse time.Time // lastUse records when the state was last updated.
}

// New creates a new FSM instance and starts a background worker
// to periodically clean up expired states.
func New(ctx context.Context) *FSM {
	fsm := &FSM{
		current:      sync.Map{},
		localStorage: sync.Map{},
	}

	// Start a goroutine for periodic cleanup of inactive states
	go fsm.startCleanupWorker(ctx, 30*time.Minute)

	return fsm
}
