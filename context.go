package fsm

import (
	"context"
)

// ctxKey defines a custom type for context keys to avoid collisions.
type ctxKey string

const (
	// FsmKey is the key used to store/retrieve FSM instance in context.
	FsmKey ctxKey = "fsm"
)

// WithContext returns a new context with the FSM instance stored.
func WithContext(ctx context.Context, f *FSM) context.Context {
	return context.WithValue(ctx, FsmKey, f)
}

// FromContext retrieves the FSM instance from context if present, otherwise returns nil.
func FromContext(ctx context.Context) *FSM {
	if fsm, ok := ctx.Value(FsmKey).(*FSM); ok {
		return fsm
	}
	return nil
}
