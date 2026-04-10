package dashboard

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouter wires up the dashboard package and registers routes under /dashboard.
func RegisterRouter(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, db *gorm.DB) {
	repo := NewRepository(db)
	svc := NewService(repo)
	RegisterDashboardRouter(rg, svc, authMiddleware)
}
