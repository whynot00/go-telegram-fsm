package fsm

import (
	"context"
	"testing"
	"time"
)

func TestNew_DefaultsAndWithers(t *testing.T) {
	ctx := context.Background()

	// дефолты
	f := New(ctx)
	if f.ttl != 30*time.Minute || f.cleanupInterval != 30*time.Second {
		t.Fatalf("defaults broken: ttl=%v, interval=%v", f.ttl, f.cleanupInterval)
	}

	// кастомные ttl/interval
	f2 := New(ctx, WithTTL(time.Minute), WithCleanupInterval(2*time.Second))
	if f2.ttl != time.Minute || f2.cleanupInterval != 2*time.Second {
		t.Fatalf("withers not applied: ttl=%v, interval=%v", f2.ttl, f2.cleanupInterval)
	}

	// кастомный storage — FSM не должен считать его «своим»
	ss := &stubStorage{}
	f3 := New(ctx, WithStorage(ss))
	if f3.storage != ss || f3.ownsStorage {
		t.Fatalf("WithStorage not applied correctly (ownsStorage=%v)", f3.ownsStorage)
	}
}
