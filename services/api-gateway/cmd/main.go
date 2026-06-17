package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/auth"
	"github.com/grand-canal-guardian/pkg/app"
	"github.com/grand-canal-guardian/pkg/response"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("API_GATEWAY_PORT", 8080)
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("USER_SERVICE_URL", "http://localhost:8001")
	viper.SetDefault("CONTENT_SERVICE_URL", "http://localhost:8002")
	viper.SetDefault("QUIZ_SERVICE_URL", "http://localhost:8004")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_SECRET", "change-me-in-production")
	viper.SetDefault("JWT_ACCESS_TTL", 900)
	viper.SetDefault("JWT_REFRESH_TTL", 604800)

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 初始化 Redis + TokenManager (网关统一鉴权)
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_HOST") + ":" + viper.GetString("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
	})
	defer rdb.Close()

	tokenMgr := auth.NewTokenManager(
		viper.GetString("JWT_SECRET"),
		rdb,
		time.Duration(viper.GetInt("JWT_ACCESS_TTL"))*time.Second,
		time.Duration(viper.GetInt("JWT_REFRESH_TTL"))*time.Second,
	)

	application := app.New(&app.Config{
		Port:    viper.GetInt("API_GATEWAY_PORT"),
		Mode:    viper.GetString("GIN_MODE"),
		Service: "api-gateway",
	}, logger)

	engine := application.Engine()
	registerProxies(engine, logger, tokenMgr)
	application.RegisterHealth()

	logger.Info("api-gateway starting (with JWT validation)...")
	if err := application.Run(); err != nil {
		logger.Fatal("gateway failed", zap.Error(err))
	}
}

// registerProxies 注册所有代理路由
func registerProxies(r *gin.Engine, logger *zap.Logger, tokenMgr *auth.TokenManager) {
	userURL, _ := url.Parse(viper.GetString("USER_SERVICE_URL"))
	contentURL, _ := url.Parse(viper.GetString("CONTENT_SERVICE_URL"))
	quizURL, _ := url.Parse(viper.GetString("QUIZ_SERVICE_URL"))

	userProxy := newReverseProxy(userURL, logger)
	contentProxy := newReverseProxy(contentURL, logger)
	quizProxy := newReverseProxy(quizURL, logger)

	authMW := auth.NewAuthMiddleware(tokenMgr)

	api := r.Group("/api/v1")
	{
		// === 认证 (无需登录) ===
		authGroup := api.Group("/auth")
		{
			authGroup.Any("/*path", gin.WrapH(userProxy))
		}

		// === 用户 (需要登录) ===
		users := api.Group("/users", authMW.RequireAuth())
		{
			users.Any("/*path", gin.WrapH(injectUserHeaders(userProxy)))
		}

		// === 帖子 (GET 公开, CUD 需登录) ===
		posts := api.Group("/posts")
		{
			// 公开读
			posts.GET("/*path", gin.WrapH(contentProxy))
			// 写操作需登录 + 注入用户头
			writePosts := posts.Group("", authMW.RequireAuth())
			{
				writePosts.POST("/*path", gin.WrapH(injectUserHeaders(contentProxy)))
				writePosts.PUT("/*path", gin.WrapH(injectUserHeaders(contentProxy)))
				writePosts.DELETE("/*path", gin.WrapH(injectUserHeaders(contentProxy)))
			}
		}

		// === 评论 (写需登录) ===
		// /comments 路由不在 posts Group 下，但是 comment CRUD 属于 /posts/:id/comments
		// 已经由上面的 posts Group 覆盖

		// === 问答 (GET 公开, POST 需登录) ===
		quiz := api.Group("/quiz")
		{
			quiz.GET("/*path", gin.WrapH(quizProxy))
			writeQuiz := quiz.Group("", authMW.RequireAuth())
			{
				writeQuiz.POST("/*path", gin.WrapH(injectUserHeaders(quizProxy)))
			}
		}

		// === 排行榜 (公开) ===
		api.GET("/leaderboard", gin.WrapH(quizProxy))

		// === 管理员 (需 admin 角色) ===
		admin := api.Group("/admin", authMW.RequireAuth(), authMW.RequireRoles("admin"))
		{
			admin.Any("/*path", gin.WrapH(injectUserHeaders(userProxy)))
		}
	}

	// Map/LLM/Vision 服务暂未实现，预留路由组
	// api.Group("/map")  → map-service (未来)
	// api.Group("/llm")  → llm-service (未来)
	// api.Group("/vision") → vision-service (未来)

	logger.Info("proxy routes registered",
		zap.String("user", viper.GetString("USER_SERVICE_URL")),
		zap.String("content", viper.GetString("CONTENT_SERVICE_URL")),
		zap.String("quiz", viper.GetString("QUIZ_SERVICE_URL")),
	)
}

// injectUserHeaders 包装 ReverseProxy，从 gin.Context 提取 user_id/role 注入 X- 头
func injectUserHeaders(proxy *httputil.ReverseProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从 gin context 提取用户信息 (由 RequireAuth 中间件注入)
		if gc := ginContextFromRequest(r); gc != nil {
			if userID, exists := gc.Get("user_id"); exists {
				r.Header.Set("X-User-ID", userID.(string))
			}
			if role, exists := gc.Get("role"); exists {
				r.Header.Set("X-User-Role", role.(string))
			}
		}
		proxy.ServeHTTP(w, r)
	})
}

// ginContextFromRequest 从 http.Request 恢复 gin.Context (hack)
func ginContextFromRequest(r *http.Request) *gin.Context {
	// Gin stores context in request's context
	if gc, ok := r.Context().Value("GinContextKey").(*gin.Context); ok {
		return gc
	}
	return nil
}

func newReverseProxy(target *url.URL, logger *zap.Logger) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		logger.Debug("proxy response",
			zap.Int("status", resp.StatusCode),
			zap.String("url", resp.Request.URL.String()),
		)
		return nil
	}
	return proxy
}

// 确保未使用的 import 不报错
func init() {
	_ = context.Background
	_ = strings.TrimSpace
	_ = response.OK
	_ = apperrors.Internal
}
