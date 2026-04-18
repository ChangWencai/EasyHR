package todo

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouter 注册 todo 相关路由
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	// 受保护路由需要 auth
	rg.Use(authMiddleware)
	rg.GET("/todos", handler.ListTodos)
	rg.PUT("/todos/:id/pin", handler.PinTodo)
	rg.GET("/todos/export", handler.ExportTodos)
	rg.GET("/carousels", handler.ListCarousels)
}
