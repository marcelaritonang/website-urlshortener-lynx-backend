package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrMissingToken)
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, types.ErrInvalidSigningMethod
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidToken)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidClaims)
			c.Abort()
			return
		}

		// Get user_id from claims as string
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUserID)
			c.Abort()
			return
		}

		// Parse UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
			c.Abort()
			return
		}

		// Set UUID in context
		c.Set("user_id", userID.String())
		c.Next()
	}
}
