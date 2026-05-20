package app

import (
	"github.com/gin-gonic/gin"
	"k8soperation/pkg/utils"
)

func GetPage(ctx *gin.Context) int {
	page := utils.StrTo(ctx.Query("page")).MustInt()
	if page <= 0 {
		return 1
	}
	return page
}

func GetPageSize(ctx *gin.Context) int {
	pageSize := utils.StrTo(ctx.Query("page_size")).MustInt()
	if pageSize <= 0 {
		a := FromContext(ctx)
		if a != nil {
			return a.DefaultPageSize
		}
		return 10
	}
	a := FromContext(ctx)
	if a != nil && pageSize > a.MaxPageSize {
		return a.MaxPageSize
	}
	return pageSize
}

func GetPageOffSet(page, pageSize int) int {
	if page > 0 {
		return (page - 1) * pageSize
	}
	return 0
}
