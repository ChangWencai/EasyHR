package wxmp

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterWXMPRouter 注册微信小程序路由到 gin.RouterGroup
// 将 wxmp 路由注册到 /api/v1/wxmp 下
func RegisterWXMPRouter(rg *gin.RouterGroup, db *gorm.DB, jwtSecret string, jwtAccessTTL, jwtRefreshTTL time.Duration, rdb *redis.Client, cryptoKey string) {
	repo := NewRepository(db, cryptoKey)
	svc := NewWXMPService(repo, jwtSecret, jwtAccessTTL, jwtRefreshTTL, rdb, cryptoKey)
	handler := NewHandler(svc)

	// 创建带有 jwtSecret 的中间件实例
	memberAuth := WXMPMemberAuth([]byte(jwtSecret))

	// 微信小程序路由组
	wxmp := rg.Group("/wxmp")

	// 认证路由（无需登录）
	auth := wxmp.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/wechat/bind", handler.BindWechat)
	}

	// 会员路由（MEMBER JWT 认证）
	member := wxmp.Group("")
	member.Use(memberAuth)
	{
		// 工资条
		member.GET("/payslips", handler.ListPayslips)
		member.POST("/payslips/:id/verify", handler.VerifyPayslip)
		member.GET("/payslips/:id", handler.GetPayslipDetail)
		member.POST("/payslips/:id/sign", handler.SignPayslip)

		// 合同
		member.GET("/contracts", handler.ListContracts)
		member.GET("/contracts/:id/pdf", handler.GetContractPDF)

		// 社保
		member.GET("/social-insurance", handler.GetSocialInsurance)

		// 报销
		member.POST("/expenses", handler.CreateExpense)
		member.GET("/expenses", handler.ListExpenses)
		member.GET("/expenses/:id", handler.GetExpenseDetail)

		// OSS
		member.POST("/oss/upload-url", handler.GetOssUploadURL)
	}
}
