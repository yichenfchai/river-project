package main

import (
	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/app"
	"github.com/grand-canal-guardian/services/content-service/internal/config"
	"github.com/grand-canal-guardian/services/content-service/internal/handler"
	"github.com/grand-canal-guardian/services/content-service/internal/repository"
	"github.com/grand-canal-guardian/services/content-service/internal/service"
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

	postRepo := repository.NewPostRepository(db)
	postSvc := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postSvc)

	application := app.New(&app.Config{
		Port:    cfg.Server.Port,
		Mode:    cfg.Server.Mode,
		Service: "content-service",
	}, logger)

	engine := application.Engine()

	// 注册提取 X-User-ID 的中间件 (由 Gateway 注入)
	engine.Use(extractUserFromHeader())

	api := engine.Group("/api/v1")
	{
		posts := api.Group("/posts")
		{
			posts.GET("", postHandler.ListPosts)
			posts.GET("/:id", postHandler.GetPost)
			posts.POST("", postHandler.CreatePost)
			posts.PUT("/:id", postHandler.UpdatePost)
			posts.DELETE("/:id", postHandler.DeletePost)
			posts.POST("/:id/like", postHandler.ToggleLike)
			posts.GET("/:id/comments", postHandler.ListComments)
			posts.POST("/:id/comments", postHandler.CreateComment)
		}

		// 删除评论 (独立路由)
		comments := api.Group("/comments")
		{
			comments.DELETE("/:id", postHandler.DeleteComment)
		}
	}
	application.RegisterHealth()

	logger.Info("content-service starting (auth from gateway headers)...")
	if err := application.Run(); err != nil {
		logger.Fatal("content-service failed", zap.Error(err))
	}
}

// extractUserFromHeader 从 Gateway 注入的 X-User-ID / X-User-Role 头提取用户信息
// Gateway 已在代理前完成 JWT 校验，服务只信任内网头
func extractUserFromHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		if role := c.GetHeader("X-User-Role"); role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
}
