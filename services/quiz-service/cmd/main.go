package main

import (
	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/app"
	"github.com/grand-canal-guardian/services/quiz-service/internal/config"
	"github.com/grand-canal-guardian/services/quiz-service/internal/handler"
	"github.com/grand-canal-guardian/services/quiz-service/internal/repository"
	"github.com/grand-canal-guardian/services/quiz-service/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
	})
	defer rdb.Close()

	quizRepo := repository.NewQuizRepository(db)
	quizSvc := service.NewQuizService(quizRepo, rdb)
	quizHandler := handler.NewQuizHandler(quizSvc)

	application := app.New(&app.Config{
		Port:    cfg.Server.Port,
		Mode:    cfg.Server.Mode,
		Service: "quiz-service",
	}, logger)

	engine := application.Engine()

	// 从 Gateway 注入的 Header 提取用户信息
	engine.Use(func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		if role := c.GetHeader("X-User-Role"); role != "" {
			c.Set("role", role)
		}
		c.Next()
	})

	api := engine.Group("/api/v1")
	{
		quiz := api.Group("/quiz")
		{
			quiz.GET("/questions", quizHandler.StartSession)
			quiz.POST("/submit", quizHandler.SubmitAnswer)
			quiz.GET("/leaderboard", quizHandler.GetLeaderboard)

		// 同时注册顶层 /leaderboard (供 Gateway 直接代理)
		api.GET("/leaderboard", quizHandler.GetLeaderboard)
			quiz.GET("/users/:id/stats", quizHandler.GetUserStats)
		}
	}
	application.RegisterHealth()

	logger.Info("quiz-service starting...")
	if err := application.Run(); err != nil {
		logger.Fatal("quiz-service failed", zap.Error(err))
	}
}
