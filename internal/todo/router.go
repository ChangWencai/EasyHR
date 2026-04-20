package todo

import (
	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// RegisterRouter 注册 todo 相关路由
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	// 公开路由（不需要 authMiddleware）
	// 注意：这些路由必须在受保护路由之前注册，避免被 auth 中间件拦截
	rg.GET("/todos/invite/:token", handler.VerifyInviteToken)
	rg.GET("/todos/invite/:token/detail", handler.GetInviteTodo)
	rg.POST("/todos/invite/:token/submit", handler.SubmitInvite)

	// 受保护路由需要 auth
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)
	authGroup.GET("/todos", handler.ListTodos)
	authGroup.PUT("/todos/:id/pin", handler.PinTodo)
	authGroup.PUT("/todos/:id/terminate", handler.TerminateTodo)
	authGroup.POST("/todos/:id/invite", handler.InviteTodo)
	authGroup.GET("/todos/export", handler.ExportTodos)
	authGroup.GET("/carousels", handler.ListCarousels)

	// 轮播图管理 CRUD
	authGroup.GET("/carousels/admin", handler.ListAllCarousels)
	authGroup.POST("/carousels", handler.CreateCarousel)
	authGroup.PUT("/carousels/:id", handler.UpdateCarousel)
	authGroup.DELETE("/carousels/:id", handler.DeleteCarousel)
}
