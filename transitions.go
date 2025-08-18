package fsm

import (
	"context"
	"time"
)

// Create ensures a user entry exists with StateDefault.
// If the user already exists, it leaves the entry unchanged.
// This method does not modify existing state, only initializes missing entries.
func (f *FSM) Create(ctx context.Context) {
	userID := userFromContext(ctx)

	f.current.LoadOrStore(userID, stateData{
		state:   StateDefault,
		lastUse: time.Now(),
	})
}

// Transition sets the user's FSM state and updates the last-use timestamp to now.
// If the new state is StateDefault, it also clears the user's local cache via CleanCache.
// This method overwrites any existing state for the user.
func (f *FSM) Transition(ctx context.Context, state StateFSM) {
	userID := userFromContext(ctx)

	f.current.Store(userID, stateData{
		state:   state,
		lastUse: time.Now(),
	})

	if state == StateDefault {
		f.CleanCache(ctx, userID)
	}
}

// Finish resets the user's state to StateDefault.
// This is a convenience wrapper around Transition(ctx, StateDefault).
func (f *FSM) Finish(ctx context.Context) {
	f.Transition(ctx, StateDefault)
}

// CurrentState returns the current FSM state for the user and a boolean flag.
// It does NOT create an entry if absent.
//   - On hit: updates the last-use timestamp and returns (state, true).
//   - On miss: returns (StateNil, false).
func (f *FSM) CurrentState(ctx context.Context) (StateFSM, bool) {
	userID := userFromContext(ctx)

	v, ok := f.current.Load(userID)
	if !ok {
		return StateNil, false
	}

	sd := v.(stateData)
	sd.lastUse = time.Now()
	f.current.Store(userID, sd)

	return sd.state, true
}
