package fsm

import "time"

// MediaData stores information about a specific media group.
// FileIDs    – a list of file identifiers belonging to this group.
// LastUpdate – the timestamp of the last modification to this group.
type MediaData struct {
	files      []File
	lastUpdate time.Time
}

type File struct {
	Type   string // File type (e.g., "photo", "video")
	FileID string // Telegram file identifier
}

// Files returns a copy of the stored files to preserve encapsulation.
func (md *MediaData) Files() []File {
	out := make([]File, len(md.files))
	copy(out, md.files)
	return out
}

// Elapsed reports whether more than t has passed since lastUpdate.
func (md *MediaData) Elapsed(t time.Duration) bool {
	return time.Since(md.lastUpdate) > t
}

// touch updates lastUpdate to the current time.
func (md *MediaData) touch() {
	md.lastUpdate = time.Now()
}

// addFile appends a file to the internal list without updating lastUpdate.
func (md *MediaData) addFile(file File) {
	md.files = append(md.files, file)
}
