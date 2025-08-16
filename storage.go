package fsm

import (
	"github.com/whynot00/go-telegram-fsm/storage"
	"github.com/whynot00/go-telegram-fsm/storage/memory"
	st_redis "github.com/whynot00/go-telegram-fsm/storage/redis"
)

func NewRedisStorage(addr, username, password string, db int) storage.Storage {
	return st_redis.NewRedisStorage(addr, username, password, db)
}

func NewMemoryStorage() storage.Storage {
	return memory.NewMemoryStorage()
}

// Option configures the FSM instance.
type Option func(*FSM)

// WithStorage sets a custom Storage implementation.
func WithStorage(s storage.Storage) Option {
	return func(f *FSM) {
		f.storage = s
	}
}

// WithRedisStorage configures FSM to use Redis as storage.
func WithRedisStorage(addr, username, password string, db int) Option {
	return func(f *FSM) {
		f.storage = NewRedisStorage(addr, username, password, db)
	}
}
