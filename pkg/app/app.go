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
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/grand-canal-guardian/pkg/log"
)

type App struct {
	Engine *gin.Engine
	Logger *zap.Logger
	DB     *gorm.DB
	Config *Config
	srv    *http.Server
}

type Config struct {
	Name    string
	Host    string
	Port    int
	Mode    string
	Log     log.Config
}

func DefaultConfig() Config {
	return Config{
		Name: "grand-canal-guardian",
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "release",
	}
}

func New(cfg Config, db *gorm.DB) (*App, error) {
	logCfg := cfg.Log
	if logCfg.Service == "" {
		logCfg = log.DefaultConfig(cfg.Name)
	}
	logger, err := log.New(logCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 logger 失败: %w", err)
	}

	gin.SetMode(cfg.Mode)
	engine := gin.New()

	app := &App{
		Engine: engine,
		Logger: logger,
		DB:     db,
		Config: &cfg,
	}

	return app, nil
}

func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.Config.Host, a.Config.Port)
	a.srv = &http.Server{
		Addr:    addr,
		Handler: a.Engine,
	}

	a.Logger.Info("服务启动", zap.String("addr", addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("服务异常退出", zap.Error(err))
		}
	}()

	<-quit
	a.Logger.Info("收到关闭信号，开始优雅关闭...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		a.Logger.Error("优雅关闭失败", zap.Error(err))
		return err
	}

	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	a.Logger.Info("服务已关闭")
	return nil
}
