package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marcelaritonang/website-urlshortener-lynx-backend/internal/types"
)

func HandleError(c *gin.Context, err error) {
	switch err {
	case types.ErrShortCodeTaken:
		ErrorResponse(c, http.StatusConflict, err)
	case types.ErrInvalidShortCode:
		ErrorResponse(c, http.StatusBadRequest, err)
	case types.ErrURLNotFound:
		ErrorResponse(c, http.StatusNotFound, err)
	case types.ErrUnauthorized:
		ErrorResponse(c, http.StatusForbidden, err)
	case types.ErrInvalidUUID:
		ErrorResponse(c, http.StatusBadRequest, err)
	case types.ErrGenerateShortCode:
		ErrorResponse(c, http.StatusInternalServerError, err)
	default:
		ErrorResponse(c, http.StatusInternalServerError, err)
	}
}
