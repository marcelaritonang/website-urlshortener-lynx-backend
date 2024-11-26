package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/types"
	"gorm.io/gorm"
)

type URLService struct {
	db               *gorm.DB
	redisClient      *redis.Client
	urlPrefix        string
	shortCodePattern *regexp.Regexp
}

func NewURLService(db *gorm.DB, redisClient *redis.Client, urlPrefix string) *URLService {
	return &URLService{
		db:               db,
		redisClient:      redisClient,
		urlPrefix:        urlPrefix,
		shortCodePattern: regexp.MustCompile("^[a-zA-Z0-9-_]+$"),
	}
}

func (s *URLService) CreateShortURL(ctx context.Context, userID uuid.UUID, longURL string, customShortCode string) (*models.URL, error) {
	shortCode := customShortCode
	if shortCode != "" {
		if !s.shortCodePattern.MatchString(shortCode) {
			return nil, types.ErrInvalidShortCode
		}
		shortCode = strings.ToLower(shortCode)

		exists, err := s.isShortCodeTaken(ctx, shortCode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, types.ErrShortCodeTaken
		}
	} else {
		var err error
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	url := &models.URL{
		ID:        uuid.New(),
		UserID:    userID,
		LongURL:   longURL,
		ShortURL:  shortCode,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(url).Error; err != nil {
			return err
		}

		return s.redisClient.Set(ctx,
			getCacheKey(shortCode),
			longURL,
			24*time.Hour,
		).Err()
	})

	if err != nil {
		return nil, err
	}

	url.ShortURL = fmt.Sprintf("%surls/%s", s.urlPrefix, url.ShortURL)
	return url, nil
}

func (s *URLService) GetURLByID(ctx context.Context, userID, urlID uuid.UUID) (*models.URL, error) {
	var url models.URL
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", urlID, userID).
		First(&url).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrURLNotFound
		}
		return nil, err
	}

	url.ShortURL = fmt.Sprintf("%surls/%s", s.urlPrefix, url.ShortURL)
	return &url, nil
}

func (s *URLService) UpdateURL(ctx context.Context, userID, urlID uuid.UUID, longURL string) (*models.URL, error) {
	var url models.URL
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", urlID, userID).
			First(&url).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return types.ErrURLNotFound
			}
			return err
		}

		url.LongURL = longURL
		url.UpdatedAt = time.Now().UTC()

		if err := tx.Save(&url).Error; err != nil {
			return err
		}

		return s.redisClient.Set(ctx,
			getCacheKey(url.ShortURL),
			longURL,
			24*time.Hour,
		).Err()
	})

	if err != nil {
		return nil, err
	}

	url.ShortURL = fmt.Sprintf("%surls/%s", s.urlPrefix, url.ShortURL)
	return &url, nil
}

func (s *URLService) DeleteURL(ctx context.Context, userID, urlID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var url models.URL
		if err := tx.Where("id = ? AND user_id = ?", urlID, userID).
			First(&url).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return types.ErrURLNotFound
			}
			return err
		}

		if err := tx.Delete(&url).Error; err != nil {
			return err
		}

		// Remove from cache
		pipe := s.redisClient.Pipeline()
		pipe.Del(ctx, getCacheKey(url.ShortURL))
		pipe.Del(ctx, getClicksKey(url.ShortURL))
		_, err := pipe.Exec(ctx)
		return err
	})
}

func getCacheKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}

func getClicksKey(shortCode string) string {
	return fmt.Sprintf("clicks:%s", shortCode)
}

func getStatsKey(shortCode string) string {
	return fmt.Sprintf("stats:%s", shortCode)
}

func (s *URLService) isShortCodeTaken(ctx context.Context, shortCode string) (bool, error) {
	// Check cache first
	exists, err := s.redisClient.Exists(ctx, getCacheKey(shortCode)).Result()
	if err == nil && exists > 0 {
		return true, nil
	}

	// Check database
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.URL{}).
		Where("short_url = ?", shortCode).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *URLService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	// Clean shortCode
	shortCode = strings.TrimPrefix(shortCode, "urls/")

	// Try cache first
	longURL, err := s.redisClient.Get(ctx, getCacheKey(shortCode)).Result()
	if err == nil {
		go s.incrementClickCount(ctx, shortCode)
		return longURL, nil
	}

	// Fallback to database
	var url models.URL
	if err := s.db.WithContext(ctx).
		Where("short_url = ?", shortCode).
		First(&url).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", types.ErrURLNotFound
		}
		return "", err
	}

	// Update cache
	s.redisClient.Set(ctx, getCacheKey(shortCode), url.LongURL, 24*time.Hour)
	go s.incrementClickCount(ctx, shortCode)

	return url.LongURL, nil
}

// GetUserURLsPaginated retrieves paginated URLs for a user
func (s *URLService) GetUserURLsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.URL, int, error) {
	var urls []models.URL
	var total int64

	// Get total count
	err := s.db.WithContext(ctx).Model(&models.URL{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&urls).Error
	if err != nil {
		return nil, 0, err
	}

	// Add prefix to short URLs
	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%surls/%s", s.urlPrefix, urls[i].ShortURL)
	}

	return urls, int(total), nil
}

// GetURLStats retrieves statistics for a URL
func (s *URLService) GetURLStats(ctx context.Context, urlID uuid.UUID) (*models.URLStats, error) {
	var url models.URL
	if err := s.db.WithContext(ctx).First(&url, "id = ?", urlID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrURLNotFound
		}
		return nil, err
	}

	// Get real-time clicks from Redis
	clicks, err := s.redisClient.Get(ctx, getClicksKey(url.ShortURL)).Int64()
	if err != nil {
		// Fallback to database clicks if Redis fails
		clicks = url.Clicks
	}

	stats := &models.URLStats{
		TotalClicks:    clicks,
		LastAccessedAt: url.UpdatedAt,
	}

	return stats, nil
}

func (s *URLService) incrementClickCount(ctx context.Context, shortCode string) {
	pipe := s.redisClient.Pipeline()

	// Increment clicks in Redis
	clicksKey := getClicksKey(shortCode)
	pipe.Incr(ctx, clicksKey)
	pipe.Expire(ctx, clicksKey, 30*24*time.Hour) // 30 days TTL

	if _, err := pipe.Exec(ctx); err != nil {
		return
	}

	// Update database asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s.db.WithContext(ctx).
			Model(&models.URL{}).
			Where("short_url = ?", shortCode).
			UpdateColumn("clicks", gorm.Expr("clicks + ?", 1))
	}()
}

func (s *URLService) generateUniqueShortCode(ctx context.Context) (string, error) {
	for i := 0; i < 3; i++ {
		code, err := generateShortCode()
		if err != nil {
			continue
		}

		exists, err := s.isShortCodeTaken(ctx, code)
		if err != nil || !exists {
			return code, nil
		}
	}
	return "", types.ErrGenerateShortCode
}

func generateShortCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:6], nil
}
