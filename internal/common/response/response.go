package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	// 记录错误详情（Debug 级别，包含调用栈上下文）
	zap.L().Debug("api error",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("http_status", httpStatus),
		zap.Int("error_code", code),
		zap.String("error_message", msg),
		zap.String("client_ip", c.ClientIP()),
		zap.Int64("user_id", c.GetInt64("user_id")),
		zap.Int64("org_id", c.GetInt64("org_id")),
	)
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
