package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// RequireOrg 检查用户是否已完善企业信息（org_id > 0）
// 未完善时返回 40301，引导用户前往企业设置
func RequireOrg(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	if orgID == 0 {
		response.OrgNotSetup(c)
		c.Abort()
		return
	}
	c.Next()
}
