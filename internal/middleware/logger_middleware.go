package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/pkg/logger"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		logger.Info("API Request", logger.Fields{
			"path":   path,
			"method": method,
			"ip":     c.ClientIP(),
		})

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("API Response Time", logger.Fields{
			"path":    path,
			"method":  method,
			"status":  status,
			"latency": latency.String(),
			"ip":      c.ClientIP(),
		})
	}
}
