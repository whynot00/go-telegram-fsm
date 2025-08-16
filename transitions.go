package fsm

import "time"

// Transition sets the current FSM state for a user and updates the last usage timestamp.
// If the new state is StateDefault, it also clears the user's local cache.
func (f *FSM) Transition(userID int64, state StateFSM) {
	f.current.Store(userID, stateData{
		state:   state,
		lastUse: time.Now(),
	})

	if state == StateDefault {
		f.CleanCache(userID)
	}
}

// Finish resets the user's FSM state to the default state.
func (f *FSM) Finish(userID int64) {
	f.Transition(userID, StateDefault)
}

// CurrentState returns the current FSM state for the user.
// If no state exists, it initializes and returns StateDefault.
func (f *FSM) CurrentState(userID int64) StateFSM {
	actRaw, _ := f.current.LoadOrStore(userID, stateData{
		state:   StateDefault,
		lastUse: time.Now(),
	})
	act := actRaw.(stateData)
	act.lastUse = time.Now()

	return act.state
}
