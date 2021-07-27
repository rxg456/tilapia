package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetPage(c *gin.Context) int {
	result := 0
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		zap.L().Error("GetPage strconv.Atoi faild", zap.Error(err))
	}
	if page > 0 {
		result = (page - 1) * GetPageSize(c)
	}

	return result
}

func GetPageSize(c *gin.Context) int {
	result, err := strconv.Atoi(c.Query("pagesize"))
	if err != nil {
		zap.L().Error("GetPageSize strconv.Atoi faild", zap.Error(err))
	}
	if result == 0 {
		return 10
	}
	return result
}
