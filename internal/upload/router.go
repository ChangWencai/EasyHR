package upload

import (
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册 upload 路由
func RegisterRouter(rg *gin.RouterGroup, uploadDir, baseURL string) {
	handler := NewHandler(uploadDir, baseURL)
	// 图片上传接口，放在 auth group 外（由调用方决定是否需要 auth）
	rg.POST("/upload/image", handler.UploadImage)
}
