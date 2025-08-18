package fsm

// StateFSM represents a user's state in the FSM.
// It is used as a symbolic identifier for transitions and matching.
type StateFSM string

const (
	// StateDefault is the baseline state automatically assigned to new users.
	// Transitioning to this state also clears the user's local cache.
	StateDefault StateFSM = "default"

	// StateAny is a wildcard state that matches regardless of the user's current state.
	// Useful for handlers that should be executed in any context.
	StateAny StateFSM = "any"

	// StateNil indicates that no state has been set for the user.
	// Returned by FSM.CurrentState when the user entry does not exist.
	StateNil StateFSM = "nil"
)
