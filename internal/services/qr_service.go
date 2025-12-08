package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/color"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type QRService struct {
	db          *gorm.DB
	redisClient *redis.Client
	urlPrefix   string
}

func NewQRService(db *gorm.DB, redisClient *redis.Client, urlPrefix string) *QRService {
	return &QRService{
		db:          db,
		redisClient: redisClient,
		urlPrefix:   urlPrefix,
	}
}

func (s *QRService) GenerateQRCode(ctx context.Context, shortCode string) ([]byte, error) {
	// Check cache first
	qrKey := getQRCodeKey(shortCode)
	cachedQR, err := s.redisClient.Get(ctx, qrKey).Bytes()
	if err == nil {
		return cachedQR, nil
	}

	// Generate QR code
	fullURL := fmt.Sprintf("%surls/%s", s.urlPrefix, shortCode)
	qr, err := qrcode.New(fullURL, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Set QR code options
	qr.BackgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White background
	qr.ForegroundColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}

	// Get PNG bytes
	var buf bytes.Buffer
	err = qr.Write(256, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	// Cache the QR code
	if err := s.redisClient.Set(ctx, qrKey, buf.Bytes(), 24*time.Hour).Err(); err != nil {
		// Log error but don't fail the request
		utils.Logger.Error("Failed to cache QR code", "error", err)
	}

	return buf.Bytes(), nil
}

func (s *QRService) GetQRCodeAsBase64(ctx context.Context, shortCode string) (string, error) {
	qrBytes, err := s.GenerateQRCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(qrBytes), nil
}

func getQRCodeKey(shortCode string) string {
	return fmt.Sprintf("qr:%s", shortCode)
}
