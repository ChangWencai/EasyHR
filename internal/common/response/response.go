package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": msg,
		"data":    nil,
	})
}

func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    list,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, 40100, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, 40300, msg)
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, 40000, msg)
}
