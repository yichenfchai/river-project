package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

// Config 限流配置
type Config struct {
	Enabled        bool
	GlobalRate     int           // 全局每秒请求数
	GlobalBurst    int
	IPRate         int           // 单 IP 每秒请求数
	IPBurst        int
	LLMRate        int           // LLM 接口每分钟请求数（成本控制）
	LLMBurst       int
	WindowDuration time.Duration // 滑动窗口大小，默认 1s
}

// DefaultConfig 返回开发环境默认配置
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		GlobalRate:     1000,
		GlobalBurst:    2000,
		IPRate:         50,
		IPBurst:        100,
		LLMRate:        20,
		LLMBurst:       30,
		WindowDuration: time.Second,
	}
}

// Middleware 返回 Gin 限流中间件
func Middleware(rdb *redis.Client, cfg Config) gin.HandlerFunc {
	if !cfg.Enabled || rdb == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// LLM 接口特殊限流（按分钟）
		if isLLMPath(path) {
			key := fmt.Sprintf("rate_limit:llm:%s", c.ClientIP())
			if !allow(ctx, rdb, key, cfg.LLMRate, time.Minute) {
				response.AbortWithError(c, errors.New(errors.ErrTooManyRequest, "AI 服务繁忙，请稍后再试"))
				return
			}
			c.Next()
			return
		}

		// 单 IP 限流
		ipKey := fmt.Sprintf("rate_limit:ip:%s:%s", c.ClientIP(), path)
		if !allow(ctx, rdb, ipKey, cfg.IPRate, cfg.WindowDuration) {
			response.AbortWithError(c, errors.New(errors.ErrTooManyRequest, "请求过于频繁，请稍后再试"))
			return
		}

		// 全局限流（所有 IP 共享）
		globalKey := fmt.Sprintf("rate_limit:global:%s", path)
		if !allow(ctx, rdb, globalKey, cfg.GlobalRate, cfg.WindowDuration) {
			response.AbortWithError(c, errors.New(errors.ErrTooManyRequest, "服务繁忙，请稍后再试"))
			return
		}

		c.Next()
	}
}

// allow 滑动窗口限流：Redis INCR + EXPIRE
func allow(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) bool {
	pipe := rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		// Redis 不可用时放行（降级）
		return true
	}

	return incr.Val() <= int64(limit)
}

func isLLMPath(path string) bool {
	return path == "/api/v1/llm/chat" || path == "/api/v1/llm/story/generate"
}
