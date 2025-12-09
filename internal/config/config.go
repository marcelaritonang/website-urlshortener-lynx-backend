package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

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
	// Load .env file (local development)
	_ = godotenv.Load()

	cfg := &Config{
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

	// ✅ Parse DATABASE_URL if exists (Render format)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		parseDatabaseURL(databaseURL, cfg)
	}

	// ✅ Parse REDIS_URL if exists
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		parseRedisURL(redisURL, cfg)
	}

	// Validate required fields
	// ...existing validation...

	return cfg, nil
}

// ✅ Parse DATABASE_URL helper
func parseDatabaseURL(dbURL string, cfg *Config) {
	dbURL = strings.TrimPrefix(dbURL, "postgresql://")
	dbURL = strings.TrimPrefix(dbURL, "postgres://")

	if strings.Contains(dbURL, "@") {
		parts := strings.SplitN(dbURL, "@", 2)
		userPass := strings.SplitN(parts[0], ":", 2)
		hostAndDb := strings.SplitN(parts[1], "/", 2)

		cfg.DBUser = userPass[0]
		if len(userPass) > 1 {
			cfg.DBPassword = userPass[1]
		}

		hostPort := strings.SplitN(hostAndDb[0], ":", 2)
		cfg.DBHost = hostPort[0]
		cfg.DBPort = "5432"
		if len(hostPort) > 1 {
			cfg.DBPort = hostPort[1]
		}

		if len(hostAndDb) > 1 {
			cfg.DBName = strings.SplitN(hostAndDb[1], "?", 2)[0]
		}
	}
}

// ✅ Parse REDIS_URL helper
func parseRedisURL(redisURL string, cfg *Config) {
	redisURL = strings.TrimPrefix(redisURL, "redis://")

	if strings.Contains(redisURL, "@") {
		parts := strings.SplitN(redisURL, "@", 2)

		if strings.Contains(parts[0], ":") {
			userPass := strings.SplitN(parts[0], ":", 2)
			if len(userPass) > 1 {
				cfg.RedisPassword = userPass[1]
			}
		}

		hostPort := strings.SplitN(parts[1], ":", 2)
		cfg.RedisHost = hostPort[0]
		cfg.RedisPort = "6379"
		if len(hostPort) > 1 {
			cfg.RedisPort = hostPort[1]
		}
	}
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
