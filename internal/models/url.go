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
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	LongURL   string    `json:"long_url" gorm:"not null"`
	ShortURL  string    `json:"short_url" gorm:"uniqueIndex;not null"`
	Clicks    int64     `json:"clicks" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type CreateURLRequest struct {
	LongURL   string `json:"long_url" binding:"required,url"`
	ShortCode string `json:"short_code" binding:"omitempty,min=3,max=20,alphanum"`
}

type UpdateURLRequest struct {
	LongURL string `json:"long_url" binding:"required,url"`
}
