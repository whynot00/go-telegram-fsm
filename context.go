package fsm

import (
	"context"
)

// ctxKey is a private type used to define context keys uniquely.
// This prevents collisions with keys from other packages.
type ctxKey int

const (
	// FsmKey is the context key for storing/retrieving the FSM instance.
	FsmKey ctxKey = iota

	// UserKey is the context key for storing/retrieving the user ID.
	UserKey
)

// FromContext extracts the FSM instance from the context.
// Returns nil if no FSM is present.
func FromContext(ctx context.Context) *FSM {
	if fsm, ok := ctx.Value(FsmKey).(*FSM); ok {
		return fsm
	}
	return nil
}

// fsmWithContext returns a new context with the FSM instance attached.
func fsmWithContext(ctx context.Context, f *FSM) context.Context {
	return context.WithValue(ctx, FsmKey, f)
}

// userWithContext returns a new context with the user ID attached.
func userWithContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserKey, userID)
}

// userFromContext extracts the user ID from the context.
// Returns 0 if no user ID is present.
func userFromContext(ctx context.Context) int64 {
	if user, ok := ctx.Value(UserKey).(int64); ok {
		return user
	}
	return 0
}
