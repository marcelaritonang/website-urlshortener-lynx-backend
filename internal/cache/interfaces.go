package cache

import (
	"context"
	"time"
)

// CacheService defines the interface for cache operations
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	GetMultiple(ctx context.Context, keys []string) (map[string]string, error)
	SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
	DeleteMultiple(ctx context.Context, keys []string) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}
