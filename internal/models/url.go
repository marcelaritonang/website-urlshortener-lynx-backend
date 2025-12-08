package models

import (
	"time"

	"github.com/google/uuid"
)

type URLStats struct {
	TotalClicks    int64     `json:"total_clicks"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

type URL struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	LongURL     string     `json:"long_url" gorm:"not null"`
	ShortURL    string     `json:"short_url" gorm:"uniqueIndex;not null"`
	ShortCode   string     `json:"short_code" gorm:"uniqueIndex;not null;size:10"` // ← ADD THIS
	Clicks      int64      `json:"clicks" gorm:"default:0"`
	IsAnonymous bool       `json:"is_anonymous" gorm:"default:false;index"` // ← Fix default
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`                    // ← Uppercase!
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"` // ← ADD (optional)
	User        *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type CreateURLRequest struct {
	LongURL   string `json:"long_url" binding:"required,url"`
	ShortCode string `json:"short_code" binding:"omitempty,min=3,max=20,alphanum"`
}

type UpdateURLRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}

// Helper: Check if URL is owned by user
func (u *URL) IsOwnedBy(userID uuid.UUID) bool {
	return u.UserID != nil && *u.UserID == userID
}

// Helper: Check if URL is expired
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// Helper: Check if URL can be edited by user
func (u *URL) CanBeEditedBy(userID uuid.UUID) bool {
	return !u.IsAnonymous && u.IsOwnedBy(userID)
}

// Helper: Check if URL can be deleted by user
func (u *URL) CanBeDeletedBy(userID uuid.UUID) bool {
	return u.IsOwnedBy(userID)
}
