package fsm

// StateFSM represents a user's state in the FSM.
type StateFSM string

const (
	// StateDefault is the initial state for all users.
	StateDefault StateFSM = "default"

	// StateAny matches any state.
	StateAny StateFSM = "any"
)
