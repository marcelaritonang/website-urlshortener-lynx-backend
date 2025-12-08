package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// SecretManager handles JWT secret rotation
type SecretManager struct {
	CurrentSecret  string
	PreviousSecret string
	RotatedAt      time.Time
}

// NewSecretManager creates a new secret manager
func NewSecretManager(currentSecret string) *SecretManager {
	return &SecretManager{
		CurrentSecret:  currentSecret,
		PreviousSecret: "",
		RotatedAt:      time.Now(),
	}
}

// RotateSecret generates a new secret and keeps the old one for grace period
func (sm *SecretManager) RotateSecret() error {
	newSecret, err := GenerateSecureSecret(64)
	if err != nil {
		return fmt.Errorf("failed to rotate secret: %w", err)
	}

	sm.PreviousSecret = sm.CurrentSecret
	sm.CurrentSecret = newSecret
	sm.RotatedAt = time.Now()

	return nil
}

// GetValidSecrets returns all valid secrets (current + previous for grace period)
func (sm *SecretManager) GetValidSecrets() []string {
	secrets := []string{sm.CurrentSecret}

	// Keep previous secret valid for 7 days after rotation
	if sm.PreviousSecret != "" && time.Since(sm.RotatedAt) < 7*24*time.Hour {
		secrets = append(secrets, sm.PreviousSecret)
	}

	return secrets
}

// GenerateSecureSecret creates a cryptographically secure random string
func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Use URL-safe base64 encoding
	encoded := base64.URLEncoding.EncodeToString(bytes)

	// Ensure exact length
	if len(encoded) > length {
		return encoded[:length], nil
	}

	return encoded, nil
}

// ValidateSecretStrength checks if secret meets security requirements
func ValidateSecretStrength(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("secret too short: minimum 32 characters required")
	}

	if len(secret) < 64 {
		return fmt.Errorf("warning: secret length %d is below recommended 64 characters", len(secret))
	}

	// Check for base64 characters
	if _, err := base64.URLEncoding.DecodeString(secret); err != nil {
		// Not base64, check if it's at least alphanumeric
		for _, c := range secret {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				return fmt.Errorf("secret contains invalid characters")
			}
		}
	}

	return nil
}
