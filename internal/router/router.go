package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/your-org/grand-canal-guardian/internal/handler"
	"github.com/your-org/grand-canal-guardian/internal/middleware"
	"github.com/your-org/grand-canal-guardian/pkg/auth"
	pkglog "github.com/your-org/grand-canal-guardian/pkg/log"
)

// Dependencies 路由所需的所有依赖（显式注入，无隐式全局状态）
type Dependencies struct {
	AuthHandler      *handler.AuthHandler
	LLMHandler       *handler.LLMHandler
	ShopHandler      *handler.ShopHandler
	AdminShopHandler *handler.AdminShopHandler
	MapHandler       *handler.MapHandler
	AuthMW           *auth.AuthMiddleware
	Logger           *zap.Logger
}

// Setup 注册所有路由和中间件
//
//	中间件顺序: Recovery → RequestID → CORS → Logger → Auth
func Setup(r *gin.Engine, deps Dependencies) {
	r.Use(pkglog.Recovery(deps.Logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(pkglog.GinLogger(deps.Logger))
	r.Use(deps.AuthMW.Handler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 公开路由（Auth 白名单已跳过）
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", deps.AuthHandler.Register)
		authGroup.POST("/login", deps.AuthHandler.Login)
	}

	// 需要鉴权
	api := r.Group("/api/v1")
	{
		api.GET("/users/me", deps.AuthHandler.GetProfile)
	}

	// LLM 路由（无需鉴权，限流由中间件处理）
	llmGroup := r.Group("/api/v1/llm")
	{
		llmGroup.POST("/chat", deps.LLMHandler.Chat)
		llmGroup.POST("/story/generate", deps.LLMHandler.GenerateStory)
		llmGroup.GET("/health", deps.LLMHandler.Health)
	}

	// 兑换商店 — 浏览商品无需登录
	shopGroup := r.Group("/api/v1/shop")
	{
		shopGroup.GET("/items", deps.ShopHandler.ListItems)
	}

	// 兑换商店 — 需登录
	apiShop := r.Group("/api/v1")
	{
		apiShop.POST("/shop/redeem", deps.ShopHandler.Redeem)
		apiShop.GET("/shop/history", deps.ShopHandler.GetHistory)
	}

	// 管理端 — 商品 CRUD
	adminGroup := r.Group("/api/v1/admin/shop")
	{
		adminGroup.POST("/items", deps.AdminShopHandler.CreateItem)
		adminGroup.PUT("/items/:id", deps.AdminShopHandler.UpdateItem)
		adminGroup.DELETE("/items/:id", deps.AdminShopHandler.DeleteItem)
	}

	// 地图 — 无需鉴权
	mapGroup := r.Group("/api/v1/map")
	{
		mapGroup.GET("/layers", deps.MapHandler.ListLayers)
		mapGroup.GET("/layers/:id", deps.MapHandler.GetLayer)
		mapGroup.GET("/pois", deps.MapHandler.ListPOIs)
	}
}
