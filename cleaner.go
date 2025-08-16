package fsm

import (
	"context"
	"time"
)

// startCleanupWorker runs a background goroutine that periodically
// calls cleanup to remove expired states based on the given TTL.
// It stops when the context is canceled.
func (f *FSM) startCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(f.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.cleanup(ctx, f.ttl)
		case <-ctx.Done():
			return
		}
	}
}

// cleanup removes FSM states and caches for users
// whose last usage time exceeds the TTL.
func (f *FSM) cleanup(ctx context.Context, ttl time.Duration) {
	now := time.Now()
	f.current.Range(func(key, value any) bool {
		userID := key.(int64)
		data := value.(stateData)

		if now.Sub(data.lastUse) > ttl {
			f.current.Delete(userID)
			f.CleanCache(ctx, userID)
		}

		return true // continue iteration
	})
}
