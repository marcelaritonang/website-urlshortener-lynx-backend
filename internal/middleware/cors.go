package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allowed origins
		allowedOrigins := map[string]bool{
			"https://shorteny.vercel.app": true,
			"https://shorteny.my.id":      true,
			"http://localhost:3000":       true,
			"http://localhost:3001":       true,
			"https://www.shorteny.site/":  true,
		}

		// Set CORS headers if origin is allowed
		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods",
				"POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Expose-Headers",
				"Content-Length, Content-Type")
			c.Writer.Header().Set("Access-Control-Max-Age", "43200")
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
