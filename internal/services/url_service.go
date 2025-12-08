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
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
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

// ✅ UPDATED: CreateShortURL for authenticated users
func (s *URLService) CreateShortURL(ctx context.Context, userID uuid.UUID, longURL string, customShortCode string) (*models.URL, error) {
	// Validate long URL
	if longURL == "" {
		return nil, types.NewValidationError("long URL is required")
	}

	// Generate or validate short code
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

	// Create URL model
	url := &models.URL{
		ID:          uuid.New(),
		UserID:      &userID, // ✅ Changed to pointer
		LongURL:     longURL,
		ShortCode:   shortCode, // ✅ Added
		ShortURL:    fmt.Sprintf("%surls/%s", s.urlPrefix, shortCode),
		Clicks:      0,
		IsAnonymous: false, // ✅ Added
		ExpiresAt:   nil,   // ✅ Added (no expiry for auth users)
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Save to database with transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(url).Error; err != nil {
			return err
		}

		// Cache the URL
		return s.redisClient.Set(ctx,
			getCacheKey(shortCode),
			longURL,
			24*time.Hour,
		).Err()
	})

	if err != nil {
		return nil, err
	}

	return url, nil
}

// ✅ NEW: CreateAnonymousURL for unauthenticated users
func (s *URLService) CreateAnonymousURL(ctx context.Context, longURL string, customShortCode string, expiryHours int) (*models.URL, error) {
	// Validate long URL
	if longURL == "" {
		return nil, types.NewValidationError("long URL is required")
	}

	// Generate or validate short code
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

	// Calculate expiry time (default: 7 days)
	var expiresAt *time.Time
	if expiryHours > 0 {
		expiry := time.Now().UTC().Add(time.Duration(expiryHours) * time.Hour)
		expiresAt = &expiry
	} else {
		// Default: 7 days (168 hours)
		expiry := time.Now().UTC().Add(168 * time.Hour)
		expiresAt = &expiry
	}

	// Create URL model
	url := &models.URL{
		ID:          uuid.New(),
		UserID:      nil, // No user (anonymous)
		LongURL:     longURL,
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("%surls/%s", s.urlPrefix, shortCode),
		Clicks:      0,
		IsAnonymous: true, // Anonymous URL
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Save to database with transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(url).Error; err != nil {
			return err
		}

		// Cache with expiry
		cacheDuration := time.Until(*expiresAt)
		return s.redisClient.Set(ctx,
			getCacheKey(shortCode),
			longURL,
			cacheDuration,
		).Err()
	})

	if err != nil {
		return nil, err
	}

	return url, nil
}

// ✅ UPDATED: GetURLByID handles nullable UserID
func (s *URLService) GetURLByID(ctx context.Context, userID, urlID uuid.UUID) (*models.URL, error) {
	var url models.URL
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", urlID, userID).
		First(&url).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrURLNotFound
		}
		return nil, err
	}

	return &url, nil
}

// UpdateURL updates an existing URL
func (s *URLService) UpdateURL(ctx context.Context, userID, urlID uuid.UUID, longURL string) (*models.URL, error) {
	var url models.URL
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL", urlID, userID).
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
			getCacheKey(url.ShortCode),
			longURL,
			24*time.Hour,
		).Err()
	})

	if err != nil {
		return nil, err
	}

	return &url, nil
}

// ✅ UPDATED: DeleteURL with HARD delete (permanently remove from database)
func (s *URLService) DeleteURL(ctx context.Context, userID, urlID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var url models.URL
		if err := tx.Where("id = ? AND user_id = ? AND deleted_at IS NULL", urlID, userID).
			First(&url).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return types.ErrURLNotFound
			}
			return err
		}

		// ✅ HARD DELETE: Permanently remove from database
		if err := tx.Unscoped().Delete(&url).Error; err != nil {
			return err
		}

		// Remove from cache
		pipe := s.redisClient.Pipeline()
		pipe.Del(ctx, getCacheKey(url.ShortCode))
		pipe.Del(ctx, getClicksKey(url.ShortCode))
		_, err := pipe.Exec(ctx)
		return err
	})
}

// ✅ OPTIMIZED: Hybrid cache strategy
func (s *URLService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	shortCode = strings.TrimPrefix(shortCode, "urls/")

	fmt.Printf("🔍 [DEBUG] GetLongURL called with shortCode: %s\n", shortCode) // ✅ ADD

	// Try Redis cache first
	longURL, err := s.redisClient.Get(ctx, getCacheKey(shortCode)).Result()
	if err == nil {
		fmt.Printf("✅ [DEBUG] Cache HIT for: %s\n", shortCode) // ✅ ADD
		// ✅ SYNCHRONOUS: Increment immediately before return
		s.incrementClickCount(ctx, shortCode)
		return longURL, nil
	}

	fmt.Printf("⚠️  [DEBUG] Cache MISS for: %s, fetching from DB...\n", shortCode) // ✅ ADD

	// Cache MISS - Fetch from PostgreSQL
	var url models.URL
	if err := s.db.WithContext(ctx).
		Where("short_code = ? AND deleted_at IS NULL", shortCode).
		First(&url).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("❌ [DEBUG] URL not found in DB: %s\n", shortCode) // ✅ ADD
			s.redisClient.Set(ctx, getCacheKey(shortCode), "NOT_FOUND", 5*time.Minute)
			return "", types.ErrURLNotFound
		}
		return "", err
	}

	fmt.Printf("✅ [DEBUG] URL found in DB: %s → %s\n", shortCode, url.LongURL) // ✅ ADD

	// Check expiry
	if url.IsExpired() {
		go s.deleteExpiredURL(context.Background(), url.ID)
		s.redisClient.Set(ctx, getCacheKey(shortCode), "EXPIRED", 5*time.Minute)
		return "", types.ErrURLNotFound
	}

	// Write-through cache
	if url.ExpiresAt != nil {
		cacheDuration := time.Until(*url.ExpiresAt)
		s.redisClient.Set(ctx, getCacheKey(shortCode), url.LongURL, cacheDuration)
	} else {
		s.redisClient.Set(ctx, getCacheKey(shortCode), url.LongURL, 24*time.Hour)
	}

	// ✅ SYNCHRONOUS: Increment before return
	s.incrementClickCount(ctx, shortCode)
	return url.LongURL, nil
}

// ✅ FIXED: Synchronous click counter with proper error handling
func (s *URLService) incrementClickCount(ctx context.Context, shortCode string) {
	clicksKey := getClicksKey(shortCode)

	fmt.Printf("📊 [SYNC] Incrementing click count for: %s (key: %s)\n", shortCode, clicksKey)

	// ✅ Check if Redis client is available
	if s.redisClient == nil {
		fmt.Printf("❌ [SYNC] Redis client is nil!\n")
		return
	}

	// ✅ Test Redis connection first
	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("❌ [SYNC] Redis ping failed: %v\n", err)
		return
	}

	// ✅ SYNCHRONOUS: Increment Redis immediately
	newClicks, err := s.redisClient.Incr(ctx, clicksKey).Result()
	if err != nil {
		fmt.Printf("❌ [SYNC] Redis increment error: %v\n", err)
		fmt.Printf("❌ [SYNC] Context error: %v\n", ctx.Err())
		return
	}

	// Set expiry (30 days)
	if err := s.redisClient.Expire(ctx, clicksKey, 30*24*time.Hour).Err(); err != nil {
		fmt.Printf("⚠️  [SYNC] Failed to set expiry: %v\n", err)
	}

	fmt.Printf("✅ [SYNC] Current clicks in Redis: %d\n", newClicks)

	// Batch sync to DB every 10 clicks (async)
	if newClicks%10 == 0 {
		fmt.Printf("📝 [ASYNC] Syncing %d clicks to database for: %s\n", 10, shortCode)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result := s.db.WithContext(bgCtx).
				Model(&models.URL{}).
				Where("short_code = ?", shortCode).
				UpdateColumn("clicks", gorm.Expr("clicks + ?", 10))

			if result.Error != nil {
				fmt.Printf("❌ [ASYNC] DB sync error: %v\n", result.Error)
			} else {
				fmt.Printf("✅ [ASYNC] Synced 10 clicks to DB (rows: %d)\n", result.RowsAffected)
			}
		}()
	}
}

// ✅ UPDATED: GetUserURLsPaginated dengan real-time clicks
func (s *URLService) GetUserURLsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.URL, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	var urls []models.URL
	var total int64

	err := s.db.WithContext(ctx).Model(&models.URL{}).
		Where("user_id = ? AND is_anonymous = false AND deleted_at IS NULL", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.WithContext(ctx).
		Where("user_id = ? AND is_anonymous = false AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&urls).Error
	if err != nil {
		return nil, 0, err
	}

	// Sync real-time clicks from Redis
	for i := range urls {
		clicksKey := getClicksKey(urls[i].ShortCode)
		redisClicks, err := s.redisClient.Get(ctx, clicksKey).Int64()

		if err == nil && redisClicks > 0 {
			urls[i].Clicks += redisClicks
			fmt.Printf("📊 URL %s: DB clicks=%d, Redis clicks=%d, Total=%d\n",
				urls[i].ShortCode, urls[i].Clicks-redisClicks, redisClicks, urls[i].Clicks)
		}
	}

	return urls, total, nil
}

// GetURLStats retrieves statistics for a URL
func (s *URLService) GetURLStats(ctx context.Context, urlID uuid.UUID) (*models.URLStats, error) {
	var url models.URL
	if err := s.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", urlID).
		First(&url).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrURLNotFound
		}
		return nil, err
	}

	// Get real-time clicks from Redis
	clicks, err := s.redisClient.Get(ctx, getClicksKey(url.ShortCode)).Int64()
	if err != nil {
		clicks = url.Clicks
	}

	stats := &models.URLStats{
		TotalClicks:    clicks,
		LastAccessedAt: url.UpdatedAt,
	}

	return stats, nil
}

// Helper functions
func (s *URLService) isShortCodeTaken(ctx context.Context, shortCode string) (bool, error) {
	exists, err := s.redisClient.Exists(ctx, getCacheKey(shortCode)).Result()
	if err == nil && exists > 0 {
		return true, nil
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.URL{}).
		Where("short_code = ? AND deleted_at IS NULL", shortCode).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ✅ NEW: Delete expired URL (hard delete)
func (s *URLService) deleteExpiredURL(ctx context.Context, urlID uuid.UUID) {
	s.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", urlID).
		Delete(&models.URL{})
}

func (s *URLService) generateUniqueShortCode(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
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
	code := base64.URLEncoding.EncodeToString(bytes)[:6]
	code = strings.ReplaceAll(code, "+", "")
	code = strings.ReplaceAll(code, "/", "")
	code = strings.ReplaceAll(code, "=", "")
	return code, nil
}

// Cache key helpers
func getCacheKey(shortCode string) string {
	return fmt.Sprintf("url:%s", shortCode)
}

func getClicksKey(shortCode string) string {
	return fmt.Sprintf("clicks:%s", shortCode)
}
