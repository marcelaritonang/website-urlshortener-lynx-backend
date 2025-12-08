package types

import (
	"time"

	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
)

type URLStats struct {
	TotalClicks    int64     `json:"total_clicks"`
	LastAccessedAt time.Time `json:"last_accessed_at,omitempty"`
	TodayClicks    int64     `json:"today_clicks,omitempty"`
	WeeklyClicks   int64     `json:"weekly_clicks,omitempty"`
	MonthlyClicks  int64     `json:"monthly_clicks,omitempty"`
}

func ConvertURLStats(stats *models.URLStats) *URLStats {
	if stats == nil {
		return nil
	}
	return &URLStats{
		TotalClicks:    stats.TotalClicks,
		LastAccessedAt: stats.LastAccessedAt,
	}
}
