package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/interfaces"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/utils"
)

type QRHandler struct {
	qrService  interfaces.QRService
	urlService interfaces.URLService
}

func NewQRHandler(qrService interfaces.QRService, urlService interfaces.URLService) *QRHandler {
	return &QRHandler{
		qrService:  qrService,
		urlService: urlService,
	}
}

// GetQRCode returns the QR code as an image
func (h *QRHandler) GetQRCode(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidInput)
		return
	}

	// Verify URL exists
	ctx := c.Request.Context()
	_, err := h.urlService.GetLongURL(ctx, shortCode)
	if err != nil {
		if err == types.ErrURLNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	// Generate QR code
	qrCode, err := h.qrService.GenerateQRCode(ctx, shortCode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	c.Data(http.StatusOK, "image/png", qrCode)
}

// GetQRCodeBase64 returns the QR code as a base64 encoded string
func (h *QRHandler) GetQRCodeBase64(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, types.ErrInvalidInput)
		return
	}

	ctx := c.Request.Context()
	base64QR, err := h.qrService.GetQRCodeAsBase64(ctx, shortCode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "QR code generated successfully", gin.H{
		"qr_code": fmt.Sprintf("data:image/png;base64,%s", base64QR),
	})
}
