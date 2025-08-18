package memory

import (
	"context"
	"sync"
	"time"

	"github.com/whynot00/go-telegram-fsm/media"
)

// cacheData wraps per-user data map.
type cacheData struct {
	data sync.Map
}

// MemoryStorage is an in-memory storage partitioned by userID.
// It maintains last-seen timestamps per user and runs a background
// cleanup worker that evicts inactive users based on TTL.
type MemoryStorage struct {
	storage  sync.Map // userID -> *cacheData
	lastSeen sync.Map // userID -> time.Time
	ttl      time.Duration
	interval time.Duration

	stopOnce sync.Once
	stopFn   context.CancelFunc
}

// NewMemoryStorage creates a MemoryStorage and starts the cleanup worker.
// The worker evicts users that were inactive for longer than ttl,
// scanning with the given interval.
func NewMemoryStorage(ttl, interval time.Duration) *MemoryStorage {
	m := &MemoryStorage{
		ttl:      ttl,
		interval: interval,
	}

	// Start background cleanup worker
	ctx, cancel := context.WithCancel(context.Background())
	m.stopFn = cancel
	go m.cleanupWorker(ctx)

	return m
}

// Close stops the background cleanup worker.
func (m *MemoryStorage) Close() {
	m.stopOnce.Do(func() {
		if m.stopFn != nil {
			m.stopFn()
		}
	})
}

var cacheDataPool = sync.Pool{New: func() any { return &cacheData{} }}

// touch updates last-seen timestamp for the given userID.
func (m *MemoryStorage) touch(userID int64) {
	m.lastSeen.Store(userID, time.Now())
}

// cleanupWorker periodically evicts users that exceeded TTL.
func (m *MemoryStorage) cleanupWorker(ctx context.Context) {
	if m.interval <= 0 || m.ttl <= 0 {
		// nothing to do
		return
	}
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			m.lastSeen.Range(func(k, v any) bool {
				uid, ok := k.(int64)
				if !ok {
					return true
				}
				last, ok := v.(time.Time)
				if !ok {
					return true
				}
				if now.Sub(last) > m.ttl {
					// Evict user: drop data and lastSeen entry
					m.storage.Delete(uid)
					m.lastSeen.Delete(uid)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

// Set stores a key/value pair for the given userID.
func (m *MemoryStorage) Set(_ context.Context, userID int64, key string, value any) {
	// fast path
	if v, ok := m.storage.Load(userID); ok {
		v.(*cacheData).data.Store(key, value)
		m.touch(userID)
		return
	}
	// slow path with pooling
	cd := cacheDataPool.Get().(*cacheData)
	actual, loaded := m.storage.LoadOrStore(userID, cd)
	if loaded {
		cacheDataPool.Put(cd)
	}
	actual.(*cacheData).data.Store(key, value)
	m.touch(userID)
}

// Get retrieves a value by key for the given userID.
func (m *MemoryStorage) Get(_ context.Context, userID int64, key string) (any, bool) {
	if cache, ok := m.storage.Load(userID); ok {
		v, ok := cache.(*cacheData).data.Load(key)
		if ok {
			m.touch(userID)
		}
		return v, ok
	}
	return nil, false
}

// SetMedia appends a media.File into a mediaGroupID for the given user.
// Hierarchy: user → "media" → mediaGroupID → *MediaData
func (m *MemoryStorage) SetMedia(_ context.Context, userID int64, mediaGroupID string, file media.File) {
	// user level
	u, ok := m.storage.Load(userID)
	if !ok {
		cd := cacheDataPool.Get().(*cacheData)
		u, _ = m.storage.LoadOrStore(userID, cd)
		if u != cd {
			cacheDataPool.Put(cd)
		}
	}
	userCache := u.(*cacheData)

	// "media" level
	mv, ok := userCache.data.Load("media")
	var mediaCache *cacheData
	if !ok {
		mc := cacheDataPool.Get().(*cacheData)
		actual, loaded := userCache.data.LoadOrStore("media", mc)
		if loaded {
			cacheDataPool.Put(mc)
		}
		mediaCache = actual.(*cacheData)
	} else {
		mediaCache = mv.(*cacheData)
	}

	// group level
	gv, ok := mediaCache.data.Load(mediaGroupID)
	var md *media.MediaData
	if !ok {
		newMD := &media.MediaData{}
		actual, _ := mediaCache.data.LoadOrStore(mediaGroupID, newMD)
		md = actual.(*media.MediaData) // if raced, discard newMD
	} else {
		md = gv.(*media.MediaData)
	}

	md.AddFile(file)
	md.Touch()
	m.touch(userID)
}

// GetMedia retrieves MediaData for a given user and mediaGroupID.
func (m *MemoryStorage) GetMedia(_ context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool) {
	u, ok := m.storage.Load(userID)
	if !ok {
		return nil, false
	}
	userCache := u.(*cacheData)

	mv, ok := userCache.data.Load("media")
	if !ok {
		return nil, false
	}
	mediaCache := mv.(*cacheData)

	v, ok := mediaCache.data.Load(mediaGroupID)
	if !ok {
		return nil, false
	}
	m.touch(userID)
	return v.(*media.MediaData), true
}

// CleanMediaCache removes MediaData for a given user and mediaGroupID.
func (m *MemoryStorage) CleanMediaCache(_ context.Context, userID int64, mediaGroupID string) bool {
	u, ok := m.storage.Load(userID)
	if !ok {
		return false
	}
	userCache := u.(*cacheData)

	mv, ok := userCache.data.Load("media")
	if !ok {
		return false
	}
	mediaCache := mv.(*cacheData)

	_, existed := mediaCache.data.LoadAndDelete(mediaGroupID)
	if existed {
		m.touch(userID) // consider it an access
	}
	return existed
}

// CleanCache removes all cached data for the given userID.
func (m *MemoryStorage) CleanCache(_ context.Context, userID int64) {
	m.storage.Delete(userID)
	m.lastSeen.Delete(userID)
}
