package fsm

import (
	"context"
)

// ctxKey defines a custom type for context keys to avoid collisions.
type ctxKey int

const (
	// FsmKey is the key used to store/retrieve FSM instance in context.
	FsmKey ctxKey = iota
	UserKey
)

// WithContext returns a new context with the FSM instance stored.
func FSMWithContext(ctx context.Context, f *FSM) context.Context {
	return context.WithValue(ctx, FsmKey, f)
}

// FromContext retrieves the FSM instance from context if present, otherwise returns nil.
func FSMFromContext(ctx context.Context) *FSM {
	if fsm, ok := ctx.Value(FsmKey).(*FSM); ok {
		return fsm
	}
	return nil
}

func UserWithContext(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserKey, userID)
}

func UserFromContext(ctx context.Context) int64 {
	if user, ok := ctx.Value(UserKey).(int64); ok {
		return user
	}

	return -1
}
