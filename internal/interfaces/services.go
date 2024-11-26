package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/types"
)

type AuthService interface {
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, email, password string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
}
type URLService interface {
	CreateShortURL(ctx context.Context, userID uuid.UUID, longURL string, customShortCode string) (*models.URL, error)
	GetLongURL(ctx context.Context, shortCode string) (string, error)
	GetURLByID(ctx context.Context, userID, urlID uuid.UUID) (*models.URL, error)
	GetUserURLsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.URL, int, error)
	UpdateURL(ctx context.Context, userID, urlID uuid.UUID, longURL string) (*models.URL, error)
	DeleteURL(ctx context.Context, userID, urlID uuid.UUID) error
	GetURLStats(ctx context.Context, urlID uuid.UUID) (*models.URLStats, error)
}

type AnalyticsService interface {
	GetUserAnalytics(ctx context.Context, userID uint) (*types.Analytics, error)
	GetURLAnalytics(ctx context.Context, userID, urlID uint) (*types.URLAnalytics, error)
}
type QRService interface {
	GenerateQRCode(ctx context.Context, shortCode string) ([]byte, error)
	GetQRCodeAsBase64(ctx context.Context, shortCode string) (string, error)
}
