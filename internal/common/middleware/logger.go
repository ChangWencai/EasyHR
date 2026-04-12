package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 记录请求详细信息（Debug 级别）
		zap.L().Debug("request received",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int64("user_id", c.GetInt64("user_id")),
			zap.Int64("org_id", c.GetInt64("org_id")),
		)

		// 读取请求体（用于调试）
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 10240 {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				zap.L().Debug("request body",
					zap.String("path", path),
					zap.String("body", string(bodyBytes)),
				)
			}
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 根据状态码使用不同级别
		if status >= 500 {
			zap.L().Error("request failed",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
				zap.String("client_ip", c.ClientIP()),
				zap.Any("errors", c.Errors.String()),
			)
		} else if status >= 400 {
			zap.L().Warn("request error",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
				zap.String("client_ip", c.ClientIP()),
			)
		} else {
			zap.L().Debug("request completed",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("latency", latency),
			)
		}
	}
}
