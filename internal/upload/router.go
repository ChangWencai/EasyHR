package upload

import (
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册 upload 路由
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, uploadDir, baseURL string) {
	handler := NewHandler(uploadDir, baseURL)
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)
	authGroup.POST("/upload/image", handler.UploadImage)
}
