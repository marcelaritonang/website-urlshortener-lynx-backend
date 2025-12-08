package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
	BlockDuration     time.Duration
}

// RateLimiterMiddleware implements token bucket algorithm for rate limiting
func RateLimiterMiddleware(redisClient *redis.Client, config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := c.Request.Context()

		// Check if IP is blocked
		blockKey := fmt.Sprintf("rate_limit:blocked:%s", ip)
		blocked, err := redisClient.Exists(ctx, blockKey).Result()
		if err == nil && blocked > 0 {
			remaining, _ := redisClient.TTL(ctx, blockKey).Result()
			utils.ErrorResponse(c, http.StatusTooManyRequests,
				fmt.Errorf("IP blocked due to excessive requests. Try again in %d seconds", int(remaining.Seconds())))
			c.Abort()
			return
		}

		// Rate limiting key
		limitKey := fmt.Sprintf("rate_limit:requests:%s", ip)

		// Get current request count
		count, err := redisClient.Get(ctx, limitKey).Int64()
		if err != nil && err != redis.Nil {
			// On Redis error, allow request (fail-open)
			c.Next()
			return
		}

		// First request from this IP
		if err == redis.Nil {
			// Initialize counter
			pipe := redisClient.Pipeline()
			pipe.Set(ctx, limitKey, 1, time.Minute)
			pipe.Exec(ctx)

			// Add headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", config.RequestsPerMinute-1))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= int64(config.RequestsPerMinute) {
			// Increment violation counter
			violationKey := fmt.Sprintf("rate_limit:violations:%s", ip)
			violations, _ := redisClient.Incr(ctx, violationKey).Result()
			redisClient.Expire(ctx, violationKey, 10*time.Minute)

			// Block IP after 3 violations in 10 minutes
			if violations >= 3 {
				redisClient.Set(ctx, blockKey, 1, config.BlockDuration)
				utils.Logger.Warn("IP blocked due to rate limit violations",
					"ip", ip,
					"violations", violations)
			}

			// Add rate limit headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			c.Header("Retry-After", "60")

			utils.ErrorResponse(c, http.StatusTooManyRequests,
				fmt.Errorf("rate limit exceeded: maximum %d requests per minute", config.RequestsPerMinute))
			c.Abort()
			return
		}

		// Increment counter
		newCount, _ := redisClient.Incr(ctx, limitKey).Result()

		// Refresh TTL on first increment
		if newCount == 1 {
			redisClient.Expire(ctx, limitKey, time.Minute)
		}

		// Add rate limit headers
		remaining := config.RequestsPerMinute - int(newCount)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

		c.Next()
	}
}

// AuthRateLimiterMiddleware - Stricter rate limiting for authentication endpoints
func AuthRateLimiterMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath()
		ctx := c.Request.Context()

		// Specific key for auth endpoints
		authKey := fmt.Sprintf("rate_limit:auth:%s:%s", endpoint, ip)

		// Allow only 5 attempts per 15 minutes
		attempts, err := redisClient.Get(ctx, authKey).Int64()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if err == redis.Nil {
			// First attempt
			redisClient.Set(ctx, authKey, 1, 15*time.Minute)
			c.Next()
			return
		}

		// Check limit (5 attempts per 15 minutes)
		if attempts >= 5 {
			ttl, _ := redisClient.TTL(ctx, authKey).Result()

			// Block IP for authentication endpoints
			blockKey := fmt.Sprintf("rate_limit:auth_blocked:%s", ip)
			redisClient.Set(ctx, blockKey, 1, 30*time.Minute)

			utils.Logger.Warn("IP blocked for authentication attempts",
				"ip", ip,
				"endpoint", endpoint,
				"attempts", attempts)

			utils.ErrorResponse(c, http.StatusTooManyRequests,
				fmt.Errorf("too many authentication attempts. Try again in %d minutes", int(ttl.Minutes())))
			c.Abort()
			return
		}

		// Increment attempt counter
		redisClient.Incr(ctx, authKey)

		c.Header("X-Auth-RateLimit-Remaining", fmt.Sprintf("%d", 5-int(attempts)-1))
		c.Next()
	}
}

// ForgotPasswordRateLimiter - Prevent abuse of password reset
func ForgotPasswordRateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email string `json:"email"`
		}

		// Bind request to get email
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Next()
			return
		}

		// Re-set the body for next handlers
		c.Set("email", request.Email)

		ctx := c.Request.Context()
		email := request.Email

		// Rate limit per email (1 request per 5 minutes)
		emailKey := fmt.Sprintf("rate_limit:forgot_password:%s", email)
		exists, _ := redisClient.Exists(ctx, emailKey).Result()

		if exists > 0 {
			ttl, _ := redisClient.TTL(ctx, emailKey).Result()
			utils.ErrorResponse(c, http.StatusTooManyRequests,
				fmt.Errorf("password reset email already sent. Try again in %d seconds", int(ttl.Seconds())))
			c.Abort()
			return
		}

		// Set cooldown
		redisClient.Set(ctx, emailKey, 1, 5*time.Minute)

		c.Next()
	}
}
