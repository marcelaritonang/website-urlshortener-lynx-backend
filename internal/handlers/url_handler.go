package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/interfaces"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/models"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
)

type URLHandler struct {
	urlService interfaces.URLService
	baseURL    string
}

// Constructor function for initializing URLHandler
func NewURLHandler(urlService interfaces.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		baseURL:    strings.TrimSuffix(baseURL, "/"), // Removes trailing slash
	}
}

// CreateShortURL creates a short URL for a given long URL (authenticated users)
func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, types.NewValidationError(err.Error()))
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	ctx := c.Request.Context()
	url, err := h.urlService.CreateShortURL(ctx, userID, req.LongURL, req.ShortCode)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Short URL created successfully", url)
}

// ✅ NEW: CreateAnonymousURL creates a short URL without authentication
func (h *URLHandler) CreateAnonymousURL(c *gin.Context) {
	var req models.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, types.NewValidationError(err.Error()))
		return
	}

	ctx := c.Request.Context()

	// Create anonymous URL with default 7 days expiry (168 hours)
	url, err := h.urlService.CreateAnonymousURL(ctx, req.LongURL, req.ShortCode, 168)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Short URL created successfully", url)
}

// GetUserURLs retrieves paginated short URLs created by the user
func (h *URLHandler) GetUserURLs(c *gin.Context) {
	var pagination utils.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	if pagination.Page == 0 {
		pagination.Page = 1
	}
	if pagination.PerPage == 0 {
		pagination.PerPage = 10
	}

	ctx := c.Request.Context()
	urls, total, err := h.urlService.GetUserURLsPaginated(ctx, userID, pagination.Page, pagination.PerPage)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	urlResponses := make([]types.URLResponse, len(urls))
	for i, url := range urls {
		shortCode := strings.TrimPrefix(url.ShortURL, h.baseURL+"/urls/")

		urlResponses[i] = types.URLResponse{
			URL: &url,
			QRCodes: types.QRCodeURLs{
				PNG:    fmt.Sprintf("%s/qr/%s", h.baseURL, shortCode),
				Base64: fmt.Sprintf("%s/qr/%s/base64", h.baseURL, shortCode),
			},
		}
	}

	// ✅ FIX: Cast int to int64 untuk perhitungan
	totalPages := (total + int64(pagination.PerPage) - 1) / int64(pagination.PerPage)

	utils.PaginationResponse(c, http.StatusOK, "URLs retrieved successfully", urlResponses, utils.Meta{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Total:     total,      // int64
		TotalPage: totalPages, // int64
	})
}

// GetURL fetches details of a specific short URL
func (h *URLHandler) GetURL(c *gin.Context) {
	urlID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidUUID)
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	ctx := c.Request.Context()
	url, err := h.urlService.GetURLByID(ctx, userID, urlID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	shortCode := strings.TrimPrefix(url.ShortURL, h.baseURL+"/urls/")

	response := types.URLResponse{
		URL: url,
		QRCodes: types.QRCodeURLs{
			PNG:    fmt.Sprintf("%s/qr/%s", h.baseURL, shortCode),
			Base64: fmt.Sprintf("%s/qr/%s/base64", h.baseURL, shortCode),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "URL retrieved successfully", response)
}

// DeleteURL deletes a specific short URL
func (h *URLHandler) DeleteURL(c *gin.Context) {
	urlID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidUUID)
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, types.ErrInvalidUUID)
		return
	}

	ctx := c.Request.Context()
	if err := h.urlService.DeleteURL(ctx, userID, urlID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "URL deleted successfully", nil)
}

// RedirectToLongURL redirects a short URL to the original long URL
func (h *URLHandler) RedirectToLongURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// ✅ ADD: Debug logs
	fmt.Printf("🌐 [HANDLER] Redirect requested for: %s\n", shortCode)
	fmt.Printf("🌐 [HANDLER] Full path: %s\n", c.Request.URL.Path)
	fmt.Printf("🌐 [HANDLER] Method: %s\n", c.Request.Method)

	if shortCode == "" {
		fmt.Printf("❌ [HANDLER] Empty short code!\n")
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidShortCode)
		return
	}

	ctx := c.Request.Context()
	longURL, err := h.urlService.GetLongURL(ctx, shortCode)
	if err != nil {
		fmt.Printf("❌ [HANDLER] Error getting long URL: %v\n", err)
		switch err {
		case types.ErrURLNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, err)
		case types.ErrInvalidShortCode:
			utils.ErrorResponse(c, http.StatusBadRequest, err)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, err)
		}
		return
	}

	fmt.Printf("✅ [HANDLER] Redirecting to: %s\n", longURL)

	utils.Logger.Info("Redirecting to URL",
		"short_code", shortCode,
		"long_url", longURL,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"referer", c.Request.Referer())

	c.Redirect(http.StatusMovedPermanently, longURL)
}
