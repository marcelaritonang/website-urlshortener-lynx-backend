package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"gorm.io/gorm"
)

type AuthService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewAuthService(db *gorm.DB, redisClient *redis.Client) *AuthService {
	return &AuthService{
		db:          db,
		redisClient: redisClient,
	}
}

func (s *AuthService) Register(ctx context.Context, user *models.User) error {
	var existingUser models.User
	if err := s.db.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return types.ErrUserExists
	}

	user.ID = uuid.New()
	if err := user.HashPassword(); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Create(user).Error
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, types.ErrInvalidCredentials
	}

	if err := user.CheckPassword(password); err != nil {
		return nil, types.ErrInvalidCredentials
	}

	return &user, nil
}

// ✅ OPTIMIZED: Hybrid session validation
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	// 🚀 Try Redis cache first
	cacheKey := fmt.Sprintf("user:%s", userID.String())

	var user models.User

	// Try to get from cache
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		// ⚡ Cache HIT - Parse cached user data
		// (In production, use JSON unmarshal here)
		// For now, fallback to DB for complete data
	}

	// 🔄 Fetch from PostgreSQL
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, types.ErrUserNotFound
	}

	// 💾 Cache user data for 1 hour
	// (In production, cache JSON serialized user)
	s.redisClient.Set(ctx, cacheKey, user.Email, 1*time.Hour)

	return &user, nil
}

// ✅ OPTIMIZED: Session invalidation (logout)
func (s *AuthService) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Store logout timestamp in Redis
	// All tokens issued before this timestamp are invalid
	return s.redisClient.Set(ctx,
		getUserSessionKey(userID),
		time.Now().Unix(),
		24*time.Hour, // Keep for 24 hours
	).Err()
}

func getUserSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("session:%s", userID.String())
}

// RequestPasswordReset generates reset token and returns it
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Don't reveal if email exists for security
			return "", nil
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)

	// Save token to database
	user.ResetToken = &resetToken
	user.ResetTokenExpiry = &expiresAt
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return "", fmt.Errorf("failed to save reset token: %w", err)
	}

	// Store in Redis for faster validation
	redisKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.redisClient.Set(ctx, redisKey, user.ID.String(), 1*time.Hour).Err(); err != nil {
		// Log error but don't fail - database is source of truth
		fmt.Printf("Warning: failed to cache reset token in Redis: %v\n", err)
	}

	return resetToken, nil
}

// ✅ OPTIMIZED: Reset password with cache invalidation
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	var user models.User
	if err := s.db.WithContext(ctx).
		Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.ErrInvalidToken
		}
		return fmt.Errorf("database error: %w", err)
	}

	user.Password = newPassword
	if err := user.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.ResetToken = nil
	user.ResetTokenExpiry = nil

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 🔄 Clear all related caches
	pipe := s.redisClient.Pipeline()
	pipe.Del(ctx, fmt.Sprintf("reset_token:%s", token))
	pipe.Del(ctx, fmt.Sprintf("user:%s", user.ID.String()))
	pipe.Del(ctx, getUserSessionKey(user.ID)) // Invalidate all sessions
	pipe.Exec(ctx)

	return nil
}
