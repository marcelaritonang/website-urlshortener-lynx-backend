package interfaces

import (
	"context"
	"time"

	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	FindByShortURL(ctx context.Context, shortURL string) (*models.URL, error)
	FindByUserID(ctx context.Context, userID uint) ([]models.URL, error)
	Delete(ctx context.Context, id uint, userID uint) error
	Update(ctx context.Context, url *models.URL) error
}

type TokenRepository interface {
	StoreToken(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetToken(ctx context.Context, key string) (string, error)
	DeleteToken(ctx context.Context, key string) error
	InvalidateUserTokens(ctx context.Context, userID uint) error
}
