package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/interfaces"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
)

type AnalyticsHandler struct {
	analyticsService interfaces.AnalyticsService
}

func NewAnalyticsHandler(analyticsService interfaces.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetUserAnalytics retrieves analytics for all user's URLs
func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	analytics, err := h.analyticsService.GetUserAnalytics(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Analytics retrieved successfully", analytics)
}

// GetURLAnalytics retrieves analytics for a specific URL
func (h *AnalyticsHandler) GetURLAnalytics(c *gin.Context) {
	urlID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidURLID)
		return
	}

	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	analytics, err := h.analyticsService.GetURLAnalytics(ctx, userID, uint(urlID))
	if err != nil {
		switch err {
		case types.ErrURLNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case types.ErrUnauthorized:
			utils.ErrorResponse(c, http.StatusForbidden, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "URL analytics retrieved successfully", analytics)
}
