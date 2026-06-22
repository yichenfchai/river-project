package log

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		requestID, _ := c.Get("X-Request-ID")

		auth := c.GetHeader("Authorization")
		if len(auth) > 20 {
			auth = auth[:14] + "***REDACTED***"
		}

		fields := []zap.Field{
			zap.Any("request_id", requestID),
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("auth", auth),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("gin_errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			logger.Error("请求完成 [5xx]", fields...)
		case status >= 400:
			logger.Warn("请求完成 [4xx]", fields...)
		case latency > 500*time.Millisecond:
			logger.Warn("请求完成 [慢请求]", fields...)
		default:
			logger.Info("请求完成", fields...)
		}
	}
}

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("request_id", c.Value("X-Request-ID")),
					zap.Any("panic", r),
					zap.Stack("stack"),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(500, gin.H{
					"code":    10001,
					"message": "服务器内部错误",
					"data":    gin.H{"error_type": "panic"},
				})
			}
		}()
		c.Next()
	}
}
