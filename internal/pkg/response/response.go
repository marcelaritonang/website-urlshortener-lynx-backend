package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/pkg/logger"
)

type Response struct {
	Status     bool        `json:"status"`
	Message    string      `json:"message"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

type Pagination struct {
	CurrentPage  int `json:"current_page"`
	PerPage      int `json:"per_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func Success(c *gin.Context, message string, data interface{}) {
	logger.Info("API Response", logger.Fields{
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"status":  true,
		"message": message,
	})

	c.JSON(http.StatusOK, Response{
		Status:    true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}

func SuccessWithPagination(c *gin.Context, message string, data interface{}, pagination *Pagination) {
	logger.Info("API Response", logger.Fields{
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"status":     true,
		"message":    message,
		"pagination": pagination,
	})

	c.JSON(http.StatusOK, Response{
		Status:     true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Timestamp:  time.Now().UTC(),
	})
}

func Error(c *gin.Context, statusCode int, message string, err error) {
	logger.Error("API Error", err, logger.Fields{
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"status":  false,
		"message": message,
	})

	c.JSON(statusCode, Response{
		Status:    false,
		Message:   message,
		Error:     err.Error(),
		Timestamp: time.Now().UTC(),
	})
}

func ValidationError(c *gin.Context, message string, errors []ErrorDetail) {
	logger.Error("Validation Error", nil, logger.Fields{
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"status":  false,
		"message": message,
		"errors":  errors,
	})

	c.JSON(http.StatusBadRequest, Response{
		Status:    false,
		Message:   message,
		Data:      errors,
		Timestamp: time.Now().UTC(),
	})
}
