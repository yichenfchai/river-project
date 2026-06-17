package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/middleware"
	"go.uber.org/zap"
)

// App 应用启动器
type App struct {
	engine *gin.Engine
	logger *zap.Logger
	config *Config
}

// Config 应用配置
type Config struct {
	Port    int
	Mode    string // debug | release
	Service string // 服务名
}

// New 创建新的 App 实例
func New(cfg *Config, logger *zap.Logger) *App {
	// 设置 Gin 模式
	gin.SetMode(cfg.Mode)

	// 设置错误脱敏模式
	if cfg.Mode == "release" {
		errors.SetMode("release")
	}

	engine := gin.New()

	// 中间件链 (顺序很重要!)
	engine.Use(middleware.Recovery(logger))
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS())

	return &App{
		engine: engine,
		logger: logger,
		config: cfg,
	}
}

// Engine 返回 gin.Engine (供 handler 注册路由)
func (a *App) Engine() *gin.Engine {
	return a.engine
}

// Logger 返回 zap.Logger
func (a *App) Logger() *zap.Logger {
	return a.logger
}

// RegisterHealth 注册健康检查路由
func (a *App) RegisterHealth() {
	a.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": a.config.Service,
		})
	})
	a.engine.GET("/ready", func(c *gin.Context) {
		// TODO: 检查 DB/Redis 连接
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

// Run 启动并等待信号优雅关闭
func (a *App) Run() error {
	addr := fmt.Sprintf(":%d", a.config.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      a.engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		a.logger.Info("server starting",
			zap.String("service", a.config.Service),
			zap.String("addr", addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("shutting down server...", zap.String("service", a.config.Service))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Fatal("server forced to shutdown", zap.Error(err))
		return err
	}

	a.logger.Info("server exited", zap.String("service", a.config.Service))
	return nil
}
