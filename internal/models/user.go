package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Email            string         `gorm:"uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"not null" json:"-"`
	FirstName        string         `gorm:"not null" json:"first_name"`
	LastName         string         `gorm:"not null" json:"last_name"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	ResetToken       *string        `gorm:"index" json:"-"`
	ResetTokenExpiry *time.Time     `json:"-"`
	URLs             []URL          `json:"urls,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (r *RegisterRequest) Validate() error {
	if !isValidEmail(r.Email) {
		return errors.New("invalid email format")
	}

	if !isValidPassword(r.Password) {
		return errors.New("password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}

const (
	Argon2Time      uint32 = 1         // Iterations
	Argon2Memory    uint32 = 64 * 1024 // 64MB RAM
	Argon2Threads   uint8  = 4         // Parallel threads
	Argon2KeyLength uint32 = 32        // Hash length
	SaltLength      int    = 16        // Random salt length
)

func generateSalt() (string, error) {
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(salt), nil
}

func (u *User) HashPassword() error {
	// 1️⃣ Generate random salt (16 bytes)
	salt, err := generateSalt()
	if err != nil {
		return err
	}

	// 2️⃣ Generate Argon2 hash
	// Argon2(password, salt, time, memory, threads, keyLength)
	hash := argon2.IDKey(
		[]byte(u.Password),
		[]byte(salt),
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLength,
	)

	// 3️⃣ Store: salt$hash
	u.Password = fmt.Sprintf("%s$%s",
		salt,
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return nil
}

func (u *User) CheckPassword(password string) error {
	// Split the stored password into salt and hash
	parts := splitPassword(u.Password)
	if len(parts) != 2 {
		return fmt.Errorf("invalid password format")
	}

	salt := parts[0]
	storedHash := parts[1]

	// Generate Argon2 hash using the provided password and extracted salt
	hash := argon2.IDKey([]byte(password), []byte(salt), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)

	// Compare the hashes
	if base64.RawStdEncoding.EncodeToString(hash) != storedHash {
		return fmt.Errorf("incorrect password")
	}

	return nil
}

func splitPassword(password string) []string {
	return split(password, '$')
}

func split(s string, delim rune) []string {
	var result []string
	current := ""
	for _, char := range s {
		if char == delim {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// GenerateResetToken creates a secure random token for password reset
func (u *User) GenerateResetToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	expiry := time.Now().Add(1 * time.Hour)
	u.ResetToken = &token
	u.ResetTokenExpiry = &expiry

	return token, nil
}

// IsResetTokenValid checks if the reset token is valid and not expired
func (u *User) IsResetTokenValid() bool {
	if u.ResetToken == nil || u.ResetTokenExpiry == nil {
		return false
	}
	return time.Now().Before(*u.ResetTokenExpiry)
}

// ClearResetToken removes the reset token after use
func (u *User) ClearResetToken() {
	u.ResetToken = nil
	u.ResetTokenExpiry = nil
}
