package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/response"
	"github.com/wencai/easyhr/pkg/jwt"
)

func Auth(jwtSecret string, rdb *redis.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.SugarLogger.Debugw("Auth: 缺少Authorization头",
				"path", c.Request.URL.Path,
				"client_ip", c.ClientIP(),
			)
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			logger.SugarLogger.Debugw("Auth: Authorization格式无效",
				"path", c.Request.URL.Path,
				"client_ip", c.ClientIP(),
			)
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(tokenStr, jwtSecret)
		if err != nil {
			logger.SugarLogger.Debugw("Auth: Token解析失败",
				"path", c.Request.URL.Path,
				"error", err.Error(),
				"client_ip", c.ClientIP(),
			)
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		logger.SugarLogger.Debugw("Auth: Token解析成功",
			"user_id", claims.UserID,
			"org_id", claims.OrgID,
			"role", claims.Role,
		)

		ctx := c.Request.Context()
		exists, err := rdb.Exists(ctx, "token:blacklist:"+claims.ID).Result()
		if err != nil {
			logger.SugarLogger.Debugw("Auth: Redis查询失败",
				"path", c.Request.URL.Path,
				"error", err.Error(),
			)
			response.Unauthorized(c, "token verification failed")
			c.Abort()
			return
		}
		if exists > 0 {
			logger.SugarLogger.Debugw("Auth: Token已被撤销",
				"path", c.Request.URL.Path,
				"token_id", claims.ID,
			)
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
