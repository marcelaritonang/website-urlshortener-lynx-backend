package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	JWTSecret     string
	URLPrefix     string
	Host          string
	BaseURL       string

	// SMTP Email Configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Try loading .env.dev in development
		if err := godotenv.Load(".env.dev"); err != nil {
			return nil, err
		}
	}

	config := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "127.0.0.1"), // ✅ UBAH
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "lynx_user"),             // ✅ UBAH
		DBPassword:    getEnv("DB_PASSWORD", "lynx_password_123"), // ✅ UBAH
		DBName:        getEnv("DB_NAME", "lynx_db"),               // ✅ UBAH
		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),          // ✅ UBAH
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		URLPrefix:     getEnv("URL_PREFIX", "http://localhost:8080/"),
		Host:          getEnv("HOST", "localhost"),                 // ← TAMBAHKAN INI
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"), // ← TAMBAHKAN INI

		// SMTP Email Configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM_EMAIL", ""),
	}

	// ✅ Validate and normalize secrets
	if err := config.validateAndNormalizeSecrets(); err != nil {
		return nil, err
	}

	return config, nil
}

// ✅ ENHANCED: Secret validation with auto-generation
func (c *Config) validateAndNormalizeSecrets() error {
	// 1. Validate JWT Secret
	if c.JWTSecret == "" {
		if c.AppEnv == "production" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		// Auto-generate for development
		secret, err := generateSecureSecret(64)
		if err != nil {
			return fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		c.JWTSecret = secret
		fmt.Printf("⚠️  [DEV] Auto-generated JWT_SECRET (save to .env for persistence)\n")
	}

	// 2. Validate JWT Secret strength
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters (current: %d)", len(c.JWTSecret))
	}

	// 3. Validate Database Password (allow empty for postgres superuser)
	// Comment out this check temporarily
	// if c.DBPassword == "" {
	//     return fmt.Errorf("DB_PASSWORD is required")
	// }

	// 4. Validate SMTP credentials for production
	if c.AppEnv == "production" {
		if c.SMTPUsername == "" || c.SMTPPassword == "" {
			return fmt.Errorf("SMTP credentials are required in production")
		}
	}

	return nil
}

// ✅ NEW: Generate cryptographically secure secret
func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
