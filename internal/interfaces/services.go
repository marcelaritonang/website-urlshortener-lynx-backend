package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
)

type AuthService interface {
	Register(ctx context.Context, user *models.User) error
	Login(ctx context.Context, email, password string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
	RequestPasswordReset(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type URLService interface {
	CreateShortURL(ctx context.Context, userID uuid.UUID, longURL string, customShortCode string) (*models.URL, error)
	CreateAnonymousURL(ctx context.Context, longURL string, customShortCode string, expiryHours int) (*models.URL, error) // ← TAMBAHKAN INI
	GetLongURL(ctx context.Context, shortCode string) (string, error)
	GetURLByID(ctx context.Context, userID, urlID uuid.UUID) (*models.URL, error)
	GetUserURLsPaginated(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.URL, int64, error) // ← UBAH int menjadi int64
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

type EmailService interface {
	SendResetPasswordEmail(toEmail, toName, resetToken string) error
}
