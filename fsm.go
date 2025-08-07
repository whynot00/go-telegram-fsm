package fsm

import (
	"context"
	"sync"
	"time"
)

// FSM implements a finite state machine for users, maintaining states and local cache.
type FSM struct {
	current      sync.Map // map[userID]stateData holding current state and last usage time
	localStorage sync.Map // map[userID]cacheData storing user-specific cached data
}

// stateData holds the FSM state and the timestamp of last update.
type stateData struct {
	state   StateFSM  // current FSM state
	lastUse time.Time // last update time of the state
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
