package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page      int   `json:"page"`
	PerPage   int   `json:"per_page"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"total_page"`
}

type PaginationRequest struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"per_page" binding:"min=1,max=100"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	Logger.Info("Success response",
		"path", c.Request.URL.Path,
		"status_code", statusCode,
		"message", message)

	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, err error) {
	Logger.Error("Error response",
		"path", c.Request.URL.Path,
		"status_code", statusCode,
		"error", err.Error())

	c.JSON(statusCode, Response{
		Success: false,
		Error:   err.Error(),
	})
}

func PaginationResponse(c *gin.Context, statusCode int, message string, data interface{}, meta Meta) {
	Logger.Info("Pagination response",
		"path", c.Request.URL.Path,
		"status_code", statusCode,
		"message", message,
		"meta", meta)

	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

type Analytics struct {
	TotalLinks     int64        `json:"total_links"`
	TotalClicks    int64        `json:"total_clicks"`
	AverageCTR     float64      `json:"average_ctr"`
	TopPerformers  []URLSummary `json:"top_performers"`
	ClicksByPeriod PeriodStats  `json:"clicks_by_period"`
	Growth         GrowthStats  `json:"growth"`
}

type URLAnalytics struct {
	ShortURL       string           `json:"short_url"`
	LongURL        string           `json:"long_url"`
	TotalClicks    int64            `json:"total_clicks"`
	CTR            float64          `json:"ctr"`
	ClicksByPeriod PeriodStats      `json:"clicks_by_period"`
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
	Daily   float64 `json:"daily"`   // Today vs Yesterday
	Weekly  float64 `json:"weekly"`  // This Week vs Last Week
	Monthly float64 `json:"monthly"` // This Month vs Last Month
}
