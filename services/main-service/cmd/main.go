package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/auth"
	"github.com/grand-canal-guardian/pkg/app"
	"github.com/grand-canal-guardian/services/main-service/internal/config"
	"github.com/grand-canal-guardian/services/main-service/internal/handler"
	"github.com/grand-canal-guardian/services/main-service/internal/repository"
	"github.com/grand-canal-guardian/services/main-service/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	logger.Info("database connected")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
	})
	defer rdb.Close()
	logger.Info("redis connected")

	// TokenManager (JWT 校验 + 黑名单)
	tokenMgr := auth.NewTokenManager(
		cfg.JWT.Secret,
		rdb,
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Second,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Second,
	)
	authMW := auth.NewAuthMiddleware(tokenMgr)

	// === 依赖注入 ===
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, tokenMgr)
	authH := handler.NewAuthHandler(userSvc)
	userH := handler.NewUserHandler(userSvc)

	postRepo := repository.NewPostRepository(db)
	postSvc := service.NewPostService(postRepo)
	postH := handler.NewPostHandler(postSvc)

	quizRepo := repository.NewQuizRepository(db)
	quizSvc := service.NewQuizService(quizRepo, rdb)
	quizH := handler.NewQuizHandler(quizSvc)

	// === App 启动器 ===
	application := app.New(&app.Config{
		Port:    cfg.Server.Port,
		Mode:    cfg.Server.Mode,
		Service: "grand-canal-guardian",
	}, logger)

	engine := application.Engine()
	// 生产模式: 服务前端静态文件 (Vite build 输出到 web/dist/)
	engine.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/" || (!strings.HasPrefix(c.Request.URL.Path, "/api/") && !strings.HasPrefix(c.Request.URL.Path, "/health")) {
			c.File("../../web/dist/index.html")
			c.Abort()
			return
		}
		c.Next()
	})
	engine.Static("/assets", "../../web/dist/assets")

	registerRoutes(engine, authH, userH, postH, quizH, authMW)
	application.RegisterHealth()

	logger.Info("grand-canal-guardian starting (单体模式)...",
		zap.Int("port", cfg.Server.Port),
	)
	if err := application.Run(); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}

// registerRoutes 直注册所有路由 — 无代理、无网络跳转
func registerRoutes(
	r *gin.Engine,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	postH *handler.PostHandler,
	quizH *handler.QuizHandler,
	authMW *auth.AuthMiddleware,
) {
	api := r.Group("/api/v1")

	// === 认证 (公开) ===
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authH.Register)
		authGroup.POST("/login", authH.Login)
		authGroup.POST("/refresh", authH.Refresh)
	}

	// === 用户 (需登录) ===
	users := api.Group("/users", authMW.RequireAuth())
	{
		users.GET("/me", userH.GetProfile)
		users.PUT("/me", userH.UpdateProfile)
		users.GET("/:id", userH.GetUser)
		users.POST("/logout", authH.Logout)
	}

	// === 帖子 (读公开，写需登录) ===
	posts := api.Group("/posts")
	{
		posts.GET("", postH.ListPosts)
		posts.GET("/:id", postH.GetPost)
		posts.GET("/:id/comments", postH.ListComments)

		// 写操作需登录
		postsWrite := posts.Group("", authMW.RequireAuth())
		{
			postsWrite.POST("", postH.CreatePost)
			postsWrite.PUT("/:id", postH.UpdatePost)
			postsWrite.DELETE("/:id", postH.DeletePost)
			postsWrite.POST("/:id/like", postH.ToggleLike)
			postsWrite.POST("/:id/comments", postH.CreateComment)
		}
	}

	// === 删除评论 (需登录) ===
	// 注意: 发评论已在 posts 组内 (/api/v1/posts/:id/comments POST)
	api.DELETE("/comments/:id", authMW.RequireAuth(), postH.DeleteComment)

	// === 问答 (读公开，写需登录) ===
	quiz := api.Group("/quiz")
	{
		quiz.GET("/questions", quizH.StartSession)
		quiz.GET("/leaderboard", quizH.GetLeaderboard)
		quiz.GET("/users/:id/stats", quizH.GetUserStats)

		quizWrite := quiz.Group("", authMW.RequireAuth())
		{
			quizWrite.POST("/submit", quizH.SubmitAnswer)
		}
	}

	// 排行榜（也暴露在顶层）
	api.GET("/leaderboard", quizH.GetLeaderboard)

	// === 管理员 (需 admin 角色) ===
	admin := api.Group("/admin", authMW.RequireAuth(), authMW.RequireRoles("admin"))
	{
		admin.GET("/users", userH.ListUsers)
		admin.PUT("/users/:id/role", userH.ChangeRole)
		admin.POST("/users/:id/ban", userH.BanUser)
	}
}
