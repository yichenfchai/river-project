package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	appconfig "github.com/yichenfchai/river-project/internal/config"
	"github.com/yichenfchai/river-project/internal/handler"
	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	"github.com/yichenfchai/river-project/internal/router"
	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/app"
	"github.com/yichenfchai/river-project/pkg/auth"
	"github.com/yichenfchai/river-project/pkg/database"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/ratelimit"
	"github.com/yichenfchai/river-project/pkg/secrets"
)

func main() {
	sec := secrets.New("")
	cfg := appconfig.Load(sec)

	if cfg.JWT.Secret == "" {
		log.Fatal("FATAL: JWT_SECRET is empty. Set it via environment variable or Docker secret.")
	}
	if cfg.LLM.APIKey == "" {
		log.Println("WARN: LLM_API_KEY is empty - LLM features will use offline fallback")
	}

	for _, line := range sec.Summary() {
		log.Println(line)
	}
	errors.SetMode(cfg.Server.Mode)

	// ── Database ──
	db, err := database.NewPostgres(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	})
	if err != nil {
		log.Printf("[WARN] 数据库连接失败: %v — 以无 DB 模式运行", err)
		db = nil
	}

	if db != nil {
		if err := db.AutoMigrate(
			&model.User{}, &model.ShopItem{}, &model.Redemption{},
			&model.MapLayer{}, &model.MapPOI{},
			&model.Post{}, &model.PostLike{}, &model.Comment{},
			&model.Question{}, &model.QuizRecord{},
		); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		log.Println("[DB] 迁移完成")
	}

	// ── Redis ──
	rdb, err := database.NewRedis(database.RedisConfig{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Printf("[WARN] Redis 连接失败: %v — 黑名单/限流降级为内存模式", err)
		rdb = nil
	}

	// ── JWT (Redis-backed blacklist) ──
	tm := auth.NewTokenManagerWithRedis(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, rdb)

	// ── App 引擎 ──
	a, err := app.New(app.Config{
		Name: "grand-canal-guardian",
		Host: cfg.Server.Host,
		Port: cfg.Server.Port,
		Mode: cfg.Server.Mode,
	}, db)
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}

	// 无数据库降级模式
	if db == nil {
		runNoDB(a)
		return
	}

	// ── 依赖注入 ──
	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, tm, a.Logger)
	authHandler := handler.NewAuthHandler(authSvc)

	llmSvc := service.NewLLMService(service.LLMConfig{
		BaseURL:     cfg.LLM.BaseURL,
		APIKey:      cfg.LLM.APIKey,
		Model:       cfg.LLM.Model,
		Timeout:     cfg.LLM.Timeout,
		MaxTokens:   cfg.LLM.MaxTokens,
		Temperature: cfg.LLM.Temperature,
	}, a.Logger)
	llmHandler := handler.NewLLMHandler(llmSvc)

	shopRepo := repository.NewShopRepo(db)
	if err := shopRepo.Seed(context.Background()); err != nil {
		log.Printf("[WARN] 种子数据写入失败: %v", err)
	} else {
		log.Println("[DB] 商店种子数据就绪")
	}
	shopSvc := service.NewShopService(shopRepo, userRepo, a.Logger)
	shopHandler := handler.NewShopHandler(shopSvc)
	adminShopHandler := handler.NewAdminShopHandler(shopSvc)

	mapRepo := repository.NewMapRepo(db)
	if err := mapRepo.Seed(context.Background()); err != nil {
		log.Printf("[WARN] 地图种子数据写入失败: %v", err)
	} else {
		log.Println("[DB] 地图种子数据就绪")
	}
	mapSvc := service.NewMapService(mapRepo)
	mapHandler := handler.NewMapHandler(mapSvc)

	postRepo := repository.NewPostRepo(db)
	postSvc := service.NewPostService(postRepo, userRepo, a.Logger)
	postHandler := handler.NewPostHandler(postSvc)
	adminPostHandler := handler.NewAdminPostHandler(postSvc)

	log.Println("[DB] 帖子/评论/点赞模块就绪")

	quizRepo := repository.NewQuizRepo(db)
	quizSvc := service.NewQuizService(quizRepo, userRepo, a.Logger)
	quizHandler := handler.NewQuizHandler(quizSvc)
	adminQuizHandler := handler.NewAdminQuizHandler(quizSvc)

	log.Println("[DB] 问答/排行榜模块就绪")

	adminSvc := service.NewAdminService(userRepo, postRepo, quizRepo, a.Logger)
	adminHandler := handler.NewAdminHandler(adminSvc)

	log.Println("[DB] 管理后台模块就绪")

	authMW := auth.NewAuthMiddleware(tm, []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/llm/chat",
		"/api/v1/llm/story",
		"/api/v1/llm/health",
		"/api/v1/llm/chat/sessions",
		"/api/v1/shop/items",
		"/api/v1/map/layers",
		"/api/v1/map/pois",
		"/api/v1/map/search",
		"/api/v1/map/timeline",
		"/api/v1/users",
		"/api/v1/posts",
		"/api/v1/quiz/questions",
		"/api/v1/leaderboard",
	})

	// ── Rate Limiting (before routes are registered) ──
	rlCfg := ratelimit.DefaultConfig()
	a.Engine.Use(ratelimit.Middleware(rdb, rlCfg))

	// ── 静态文件中间件 (必须在 auth 中间件之前) ──
	// SPA 模式：非 API 路径尝试返回静态文件，不存在则回退到 index.html
	webDist := "/opt/gcg/web/dist"
	a.Engine.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || path == "/health" {
			c.Next()
			return
		}
		// 尝试匹配静态文件
		fsPath := filepath.Join(webDist, path)
		if path == "/" {
			fsPath = filepath.Join(webDist, "index.html")
		}
		if info, err := os.Stat(fsPath); err == nil && !info.IsDir() {
			c.File(fsPath)
		} else {
			// SPA fallback：返回 index.html 让前端路由接管
			c.File(filepath.Join(webDist, "index.html"))
		}
		c.Abort()
	})

	router.Setup(a.Engine, router.Dependencies{
		AuthHandler:      authHandler,
		LLMHandler:       llmHandler,
		ShopHandler:      shopHandler,
		AdminShopHandler: adminShopHandler,
		MapHandler:       mapHandler,
		PostHandler:      postHandler,
		AdminPostHandler: adminPostHandler,
		QuizHandler:      quizHandler,
		AdminQuizHandler: adminQuizHandler,
		AdminHandler:     adminHandler,
		AuthMW:           authMW,
		Logger:           a.Logger,
	})

	a.Logger.Info("框架初始化完成",
		zap.String("listen", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
	)

	if err := a.Run(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal(err)
	}

	// ── 关闭 Redis ──
	if rdb != nil {
		_ = rdb.Close()
	}
}

func runNoDB(a *app.App) {
	a.Logger.Warn("数据库未连接，仅提供 /health")
	a.Engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 10003, "message": "资源不存在"})
	})
	if err := a.Run(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal(err)
	}
}
