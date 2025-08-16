package storage

import (
	"context"

	"github.com/whynot00/go-telegram-fsm/media"
)

// Storage defines the behaviour for caching user data and media.
type Storage interface {
	Set(ctx context.Context, userID int64, key string, value any)
	Get(ctx context.Context, userID int64, key string) (any, bool)
	SetMedia(ctx context.Context, userID int64, mediaGroupID string, file media.File)
	GetMedia(ctx context.Context, userID int64, mediaGroupID string) (*media.MediaData, bool)
	CleanMediaCache(ctx context.Context, userID int64, mediaGroupID string) bool
	CleanCache(ctx context.Context, userID int64)
}
