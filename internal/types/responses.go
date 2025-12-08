package types

import (
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
)

type RegisterResponse struct {
	User *models.User `json:"user"`
}
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type URLResponse struct {
	URL     *models.URL `json:"url"`
	QRCodes QRCodeURLs  `json:"qr_codes"`
}

type QRCodeURLs struct {
	PNG    string `json:"png"`
	Base64 string `json:"base64"`
}

type URLListResponse struct {
	URLs      []URLResponse `json:"urls"`
	Meta      *Meta         `json:"meta,omitempty"`
	Analytics Analytics     `json:"analytics"`
}

type Meta struct {
	Page      int `json:"page"`
	PerPage   int `json:"per_page"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

type Analytics struct {
	TotalLinks     int64        `json:"total_links"`
	TotalClicks    int64        `json:"total_clicks"`
	AverageCTR     float64      `json:"average_ctr"`
	TopPerformers  []URLSummary `json:"top_performers"`
	ClicksByPeriod *PeriodStats `json:"clicks_by_period"`
	Growth         GrowthStats  `json:"growth"`
}

type URLAnalytics struct {
	ShortURL       string           `json:"short_url"`
	LongURL        string           `json:"long_url"`
	TotalClicks    int64            `json:"total_clicks"`
	ClicksByPeriod *PeriodStats     `json:"clicks_by_period"`
	Growth         GrowthStats      `json:"growth"`
	TopReferrers   map[string]int64 `json:"top_referrers"`
	Browsers       map[string]int64 `json:"browsers"`
	Devices        map[string]int64 `json:"devices"`
	Countries      map[string]int64 `json:"countries"`
}

type URLSummary struct {
	ShortURL    string  `json:"short_url"`
	LongURL     string  `json:"long_url"`
	TotalClicks int64   `json:"total_clicks"`
	CTR         float64 `json:"ctr"`
}

type PeriodStats struct {
	Today     int64 `json:"today"`
	Yesterday int64 `json:"yesterday"`
	ThisWeek  int64 `json:"this_week"`
	LastWeek  int64 `json:"last_week"`
	ThisMonth int64 `json:"this_month"`
	LastMonth int64 `json:"last_month"`
	Total     int64 `json:"total"`
}

type GrowthStats struct {
	Daily   float64 `json:"daily"`
	Weekly  float64 `json:"weekly"`
	Monthly float64 `json:"monthly"`
}
