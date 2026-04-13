package wxmp

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/user"
	"gorm.io/gorm"
)

// RegisterWXMPRouter 注册微信小程序路由到 gin.RouterGroup
// 将 wxmp 路由注册到 /api/v1/wxmp 下
func RegisterWXMPRouter(rg *gin.RouterGroup, db *gorm.DB, jwtSecret string, jwtAccessTTL, jwtRefreshTTL time.Duration, rdb *redis.Client, cryptoKey string, userSvc *user.Service) {
	repo := NewRepository(db, cryptoKey)
	svc := NewWXMPService(repo, jwtSecret, jwtAccessTTL, jwtRefreshTTL, rdb, cryptoKey)
	handler := NewHandler(svc, userSvc)

	// 微信小程序路由组
	wxmp := rg.Group("/wxmp")
	handler.RegisterRoutes(wxmp)
}
