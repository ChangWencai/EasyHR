package wxmp

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wencai/easyhr/internal/common/response"
)

const (
	RoleMEMBER = "MEMBER"
)

// WXMPMemberAuth 小程序会员认证中间件
// 所有 /api/v1/wxmp/* 路由（除 /auth/* 外）均需此中间件保护
// 解析 JWT，验证 MEMBER 角色，将 user_id/employee_id/org_id 注入 context
func WXMPMemberAuth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &WXMPClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*WXMPClaims)
		if !ok || !token.Valid {
			response.Unauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		// 强制 MEMBER 角色
		if claims.Role != RoleMEMBER {
			response.Forbidden(c, "小程序端仅支持会员账号")
			c.Abort()
			return
		}

		// 注入 JWT claims 到 context
		c.Set("user_id", claims.UserID)
		c.Set("employee_id", claims.EmployeeID)
		c.Set("org_id", claims.OrgID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
