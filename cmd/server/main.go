package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	appconfig "github.com/your-org/grand-canal-guardian/internal/config"
	"github.com/your-org/grand-canal-guardian/internal/handler"
	"github.com/your-org/grand-canal-guardian/internal/model"
	"github.com/your-org/grand-canal-guardian/internal/repository"
	"github.com/your-org/grand-canal-guardian/internal/router"
	"github.com/your-org/grand-canal-guardian/internal/service"
	"github.com/your-org/grand-canal-guardian/pkg/app"
	"github.com/your-org/grand-canal-guardian/pkg/auth"
	"github.com/your-org/grand-canal-guardian/pkg/database"
	"github.com/your-org/grand-canal-guardian/pkg/errors"
	"github.com/your-org/grand-canal-guardian/pkg/secrets"
)

func main() {
	sec := secrets.New("")
	cfg := appconfig.Load(sec)

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
		if err := db.AutoMigrate(&model.User{}, &model.ShopItem{}, &model.Redemption{}, &model.MapLayer{}, &model.MapPOI{}); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
		log.Println("[DB] 迁移完成")
	}

	// ── JWT ──
	tm := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

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

	// ── 依赖注入 (显式构造，无反射/无框架) ──
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

	authMW := auth.NewAuthMiddleware(tm, []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/llm/chat",
		"/api/v1/llm/story",
		"/api/v1/llm/health",
		"/api/v1/shop/items",
		"/api/v1/map/layers",
		"/api/v1/map/pois",
	})

	router.Setup(a.Engine, router.Dependencies{
		AuthHandler:      authHandler,
		LLMHandler:       llmHandler,
		ShopHandler:      shopHandler,
		AdminShopHandler: adminShopHandler,
		MapHandler:       mapHandler,
		AuthMW:           authMW,
		Logger:           a.Logger,
	})

	a.Logger.Info("框架初始化完成",
		zap.String("listen", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
	)

	if err := a.Run(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal(err)
	}
}

// runNoDB 无数据库降级运行
func runNoDB(a *app.App) {
	a.Logger.Warn("数据库未连接，仅提供 /health")
	a.Engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 10003, "message": "资源不存在"})
	})
	if err := a.Run(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal(err)
	}
}
