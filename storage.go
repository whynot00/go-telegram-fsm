package fsm

import (
	"github.com/whynot00/go-telegram-fsm/storage"
	"github.com/whynot00/go-telegram-fsm/storage/memory"
	"github.com/whynot00/go-telegram-fsm/storage/redis"
)

func NewRedisStorage(addr, username, password string, db int) storage.Storage {
	return redis.NewRedisStorage(addr, username, password, db)
}

func NewMemoryStorage() storage.Storage {
	return memory.NewMemoryStorage()
}
