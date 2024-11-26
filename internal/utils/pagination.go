package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPaginationFromContext(c *gin.Context) PaginationRequest {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	return PaginationRequest{
		Page:    page,
		PerPage: perPage,
	}
}
