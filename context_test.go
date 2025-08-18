package fsm

import (
	"context"
	"testing"
)

func TestContextHelpers(t *testing.T) {
	f := &FSM{}
	ctx := context.Background()

	// FSM
	ctx = fsmWithContext(ctx, f)
	if FromContext(ctx) != f {
		t.Fatalf("FromContext failed to retrieve FSM")
	}

	// user
	ctx = userWithContext(ctx, 555)
	if got := userFromContext(ctx); got != 555 {
		t.Fatalf("userFromContext = %d, want 555", got)
	}
	// no user
	if got := userFromContext(context.Background()); got != 0 {
		t.Fatalf("userFromContext on empty ctx = %d, want 0", got)
	}
}
