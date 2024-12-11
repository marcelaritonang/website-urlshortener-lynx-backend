package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/naufalrafianto/lynx-api/internal/interfaces"
	"github.com/naufalrafianto/lynx-api/internal/models"
	"github.com/naufalrafianto/lynx-api/internal/types"
	"github.com/naufalrafianto/lynx-api/internal/utils"
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

// CreateShortURL creates a short URL for a given long URL
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

	totalPages := (total + pagination.PerPage - 1) / pagination.PerPage
	utils.PaginationResponse(c, http.StatusOK, "URLs retrieved successfully", urlResponses, utils.Meta{
		Page:      pagination.Page,
		PerPage:   pagination.PerPage,
		Total:     total,
		TotalPage: totalPages,
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
	if shortCode == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidShortCode)
		return
	}

	ctx := c.Request.Context()
	longURL, err := h.urlService.GetLongURL(ctx, shortCode)
	if err != nil {
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

	utils.Logger.Info("Redirecting to URL",
		"short_code", shortCode,
		"long_url", longURL,
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"referer", c.Request.Referer())

	c.Redirect(http.StatusMovedPermanently, longURL)
}
