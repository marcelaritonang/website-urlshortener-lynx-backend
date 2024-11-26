package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/types"
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

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, types.ErrUserNotFound
	}
	return &user, nil
}

func (s *AuthService) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.redisClient.Set(ctx,
		getUserSessionKey(userID),
		time.Now().Unix(),
		24*time.Hour,
	).Err()
}

func getUserSessionKey(userID uuid.UUID) string {
	return fmt.Sprintf("session:%s", userID.String())
}
