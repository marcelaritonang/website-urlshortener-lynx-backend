package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"gorm.io/gorm"
)

type CacheWarmer struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewCacheWarmer(db *gorm.DB, redisClient *redis.Client) *CacheWarmer {
	return &CacheWarmer{
		db:          db,
		redisClient: redisClient,
	}
}

// WarmTopURLs preloads most accessed URLs into Redis cache
func (cw *CacheWarmer) WarmTopURLs(ctx context.Context) error {
	// Get top 1000 most clicked URLs
	var urls []models.URL
	if err := cw.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("clicks DESC").
		Limit(1000).
		Find(&urls).Error; err != nil {
		return err
	}

	// Warm cache with top URLs
	pipe := cw.redisClient.Pipeline()
	for _, url := range urls {
		cacheKey := fmt.Sprintf("url:%s", url.ShortCode)

		if url.ExpiresAt != nil {
			cacheDuration := time.Until(*url.ExpiresAt)
			pipe.Set(ctx, cacheKey, url.LongURL, cacheDuration)
		} else {
			pipe.Set(ctx, cacheKey, url.LongURL, 24*time.Hour)
		}
	}

	_, err := pipe.Exec(ctx)
	fmt.Printf("✅ Cache warmed with %d top URLs\n", len(urls))
	return err
}

// StartCacheWarmer runs cache warming every 1 hour
func (cw *CacheWarmer) StartCacheWarmer() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		// Initial warm on startup
		ctx := context.Background()
		cw.WarmTopURLs(ctx)

		// Periodic warming
		for range ticker.C {
			cw.WarmTopURLs(ctx)
		}
	}()
}
