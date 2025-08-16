//go:build !redis

package fsm

// NewRedisStorage is a stub used when the redis build tag is not provided.
// It falls back to in-memory storage.
func NewRedisStorage(addr, username, password string, db int) Storage {
	return NewMemoryStorage()
}
