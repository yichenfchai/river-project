package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/auth"
	"github.com/grand-canal-guardian/pkg/app"
	"github.com/grand-canal-guardian/services/user-service/internal/config"
	"github.com/grand-canal-guardian/services/user-service/internal/handler"
	"github.com/grand-canal-guardian/services/user-service/internal/repository"
	"github.com/grand-canal-guardian/services/user-service/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. 初始化数据库
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	logger.Info("database connected")

	// 4. 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
	})
	defer rdb.Close()
	logger.Info("redis connected")

	// 5. 初始化 TokenManager
	tokenMgr := auth.NewTokenManager(
		cfg.JWT.Secret,
		rdb,
		time.Duration(cfg.JWT.AccessTokenTTL)*time.Second,
		time.Duration(cfg.JWT.RefreshTokenTTL)*time.Second,
	)

	// 6. 依赖注入: Repository → Service → Handler
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, tokenMgr)
	authHandler := handler.NewAuthHandler(userSvc)
	userHandler := handler.NewUserHandler(userSvc)

	// 7. 创建 App + 注册中间件
	application := app.New(&app.Config{
		Port:    cfg.Server.Port,
		Mode:    cfg.Server.Mode,
		Service: "user-service",
	}, logger)

	engine := application.Engine()
	registerRoutes(engine, authHandler, userHandler, tokenMgr)
	application.RegisterHealth()

	// 8. 启动
	logger.Info("user-service starting...")
	if err := application.Run(); err != nil {
		logger.Fatal("user-service failed", zap.Error(err))
	}
}

func registerRoutes(
	r *gin.Engine,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	tokenMgr *auth.TokenManager,
) {
	authMiddleware := auth.NewAuthMiddleware(tokenMgr)

	api := r.Group("/api/v1")
	{
		// 认证 (无需登录)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authH.Register)
			authGroup.POST("/login", authH.Login)
			authGroup.POST("/refresh", authH.Refresh)
		}

		// 用户 (需要登录)
		users := api.Group("/users", authMiddleware.RequireAuth())
		{
			users.GET("/me", userH.GetProfile)
			users.PUT("/me", userH.UpdateProfile)
			users.GET("/:id", userH.GetUser)
			users.POST("/logout", authH.Logout)
		}

		// 管理员 (需要 admin 角色)
		admin := api.Group("/admin", authMiddleware.RequireAuth(),
			authMiddleware.RequireRoles("admin"))
		{
			admin.GET("/users", userH.ListUsers)
			admin.PUT("/users/:id/role", userH.ChangeRole)
			admin.POST("/users/:id/ban", userH.BanUser)
		}
	}
}
