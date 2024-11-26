package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

func GenerateRequestID() string {
	return uuid.New().String()
}
