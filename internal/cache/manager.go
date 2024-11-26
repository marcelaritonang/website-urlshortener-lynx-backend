package cache

import (
	"context"
	"encoding/json"
	"time"
)

type CacheManager struct {
	cache CacheService
}

func NewCacheManager(cache CacheService) *CacheManager {
	return &CacheManager{cache: cache}
}

// GetObject retrieves and unmarshals a cached object
func (m *CacheManager) GetObject(ctx context.Context, key string, dest interface{}) error {
	data, err := m.cache.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// SetObject marshals and caches an object
func (m *CacheManager) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.cache.Set(ctx, key, value, expiration)
}
