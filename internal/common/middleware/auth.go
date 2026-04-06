package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/response"
	"github.com/wencai/easyhr/pkg/jwt"
)

func Auth(jwtSecret string, rdb *redis.Client) func(c *gin.Context) {
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

		claims, err := jwt.ParseToken(tokenStr, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		exists, err := rdb.Exists(ctx, "token:blacklist:"+claims.ID).Result()
		if err != nil {
			response.Unauthorized(c, "token verification failed")
			c.Abort()
			return
		}
		if exists > 0 {
			response.Unauthorized(c, "token has been revoked")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("org_id", claims.OrgID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
