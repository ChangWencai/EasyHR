package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, limit int, window time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := "ratelimit:" + keyFunc(c)
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			c.AbortWithStatusJSON(429, gin.H{"code": 42900, "message": "请求过于频繁，请稍后重试"})
			return
		}
		c.Next()
	}
}
