package fsm

import (
	"sync"
	"time"

	"github.com/whynot00/go-telegram-fsm/media"
)

// MediaData stores information about a specific media group.
// FileIDs    – a list of file identifiers belonging to this group.
// LastUpdate – the timestamp of the last modification to this group.
type MediaData struct {
	mu         sync.RWMutex
	files      []media.File
	lastUpdate time.Time
}

// Files returns a copy of the stored files to preserve encapsulation.
func (md *MediaData) Files() []media.File {
	md.mu.RLock()
	defer md.mu.RUnlock()

	out := make([]media.File, len(md.files))
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
func (md *MediaData) touch() {
	md.mu.Lock()
	md.lastUpdate = time.Now()
	md.mu.Unlock()
}

// addFile appends a file to the internal list without updating lastUpdate.
func (md *MediaData) addFile(file media.File) {
	md.mu.Lock()
	md.files = append(md.files, file)
	md.mu.Unlock()
}
