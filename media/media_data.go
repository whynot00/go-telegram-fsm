package media

import (
	"sync"
	"time"
)

// MediaData stores files belonging to a specific media group and records the time of the last update.
type MediaData struct {
	mu         sync.RWMutex
	files      []File
	lastUpdate time.Time
}

// Files returns a copy of the stored files to preserve encapsulation.
func (md *MediaData) Files() []File {
	md.mu.RLock()
	defer md.mu.RUnlock()

	out := make([]File, len(md.files))
	copy(out, md.files)
	return out
}

// Elapsed reports whether more than t has passed since lastUpdate.
func (md *MediaData) Elapsed(t time.Duration) bool {
	md.mu.RLock()
	defer md.mu.RUnlock()

	return time.Since(md.lastUpdate) > t
}

// touch updates lastUpdate to the current time.
func (md *MediaData) Touch() {
	md.mu.Lock()
	md.lastUpdate = time.Now()
	md.mu.Unlock()
}

// addFile appends a file to the internal list without updating lastUpdate.
func (md *MediaData) AddFile(file File) {
	md.mu.Lock()
	md.files = append(md.files, file)
	md.mu.Unlock()
}
