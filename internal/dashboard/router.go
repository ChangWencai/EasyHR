package dashboard

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouter wires up the dashboard package and registers the GET /dashboard route.
// The caller passes an already-prefixed router group (e.g. v1.Group("/dashboard")).
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	rg.Use(authMiddleware)
	rg.GET("", handler.GetDashboard)
	rg.GET("/todo-stats", handler.GetTodoStats)
	rg.GET("/time-limited-stats", handler.GetTimeLimitedStats)
}
