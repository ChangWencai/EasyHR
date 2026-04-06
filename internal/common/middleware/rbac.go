package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		for _, r := range roles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}
