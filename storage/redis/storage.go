package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/whynot00/go-telegram-fsm/media"
	"github.com/whynot00/go-telegram-fsm/storage"
)

// RedisStorage implements Storage backed by Redis.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a RedisStorage instance.
func NewRedisStorage(addr, username, password string, db int) storage.Storage {
	opts := &redis.Options{Addr: addr, Username: username, Password: password, DB: db}
	return &RedisStorage{client: redis.NewClient(opts)}
}

func userKey(userID int64, key string) string {
	return fmt.Sprintf("user:%d:kv:%s", userID, key)
}

func mediaKey(userID int64, group string) string {
	return fmt.Sprintf("user:%d:media:%s", userID, group)
}

// Set stores a key-value pair in Redis.
func (r *RedisStorage) Set(ctx context.Context, userID int64, key string, value any) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return
	}
	r.client.Set(ctx, userKey(userID, key), buf.Bytes(), 0)
}

// Get retrieves a value from Redis.
func (r *RedisStorage) Get(ctx context.Context, userID int64, key string) (any, bool) {
	data, err := r.client.Get(ctx, userKey(userID, key)).Bytes()
	if err != nil {
		return nil, false
	}
	var v any
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&v); err != nil {
		return nil, false
	}
	return v, true
}

// helper struct for media data serialization
// exported fields to allow JSON marshal

type redisMedia struct {
	Files      []media.File `json:"files"`
	LastUpdate time.Time    `json:"last_update"`
}

// SetMedia appends a file to a media group in Redis.
func (r *RedisStorage) SetMedia(ctx context.Context, userID int64, mediaGroupID string, file media.File) {

	key := mediaKey(userID, mediaGroupID)
	var md redisMedia
	if data, err := r.client.Get(ctx, key).Bytes(); err == nil {
		_ = json.Unmarshal(data, &md)
	}
	md.Files = append(md.Files, file)
	md.LastUpdate = time.Now()
	b, err := json.Marshal(md)
	if err != nil {
		return
	}
	r.client.Set(ctx, key, b, 0)
}

// GetMedia retrieves media data from Redis.
func (r *RedisStorage) GetMedia(ctx context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool) {

	key := mediaKey(userID, mediaGroupID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var md redisMedia
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, false
	}
	var out *media.MediaData
	for _, f := range md.Files {
		out.AddFile(f)
	}
	out.Touch()
	return out, true
}

// CleanMediaCache removes media group data from Redis.
func (r *RedisStorage) CleanMediaCache(ctx context.Context, userID int64, mediaGroupID string) bool {

	key := mediaKey(userID, mediaGroupID)
	res, err := r.client.Del(ctx, key).Result()
	return err == nil && res > 0
}

// CleanCache removes all cached data for the given user.
func (r *RedisStorage) CleanCache(ctx context.Context, userID int64) {

	pattern := fmt.Sprintf("user:%d:*", userID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}
	if len(keys) > 0 {
		r.client.Del(ctx, keys...)
	}
}
