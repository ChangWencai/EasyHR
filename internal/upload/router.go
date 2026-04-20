package upload

import (
	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
)

// RegisterRouter 注册 upload 路由
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, uploadDir, baseURL string) {
	handler := NewHandler(uploadDir, baseURL)
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)
	authGroup.POST("/upload/image", handler.UploadImage)
}
