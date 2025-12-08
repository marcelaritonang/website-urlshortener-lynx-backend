package utils

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ✅ Define custom type untuk context key (fix SA1029 warning)
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

var Logger *slog.Logger

func InitLogger(env string) {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	if env == "production" {
		opts.Level = slog.LevelWarn
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
}

type LoggerMiddleware struct {
	logger *slog.Logger
}

func NewLoggerMiddleware(logger *slog.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

func (l *LoggerMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Set request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = GenerateRequestID()
		}
		c.Set(string(RequestIDKey), requestID) // ✅ Convert contextKey to string

		// ✅ Add request ID to context with custom type
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		method := c.Request.Method // ✅ FIX: Remove () - Method is a string field, not a function
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		l.logger.LogAttrs(context.Background(),
			getLogLevel(statusCode),
			"Request completed",
			slog.String("request_id", requestID),
			slog.String("client_ip", clientIP),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("user_agent", userAgent),
			slog.Int("status_code", statusCode),
			slog.Duration("latency", latency),
			slog.String("error", errorMessage),
		)
	}
}

func getLogLevel(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError
	case statusCode >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// ✅ Helper functions untuk mengambil value dari context dengan type safety
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// ✅ Helper untuk set user ID di context (untuk middleware auth)
func SetUserIDInContext(c *gin.Context, userID string) {
	c.Set(string(UserIDKey), userID)
	ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
	c.Request = c.Request.WithContext(ctx)
}
