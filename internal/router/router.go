package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yichenfchai/river-project/internal/handler"
	"github.com/yichenfchai/river-project/internal/middleware"
	"github.com/yichenfchai/river-project/pkg/auth"
	pkglog "github.com/yichenfchai/river-project/pkg/log"
)

// Dependencies 路由所需的所有依赖（显式注入，无隐式全局状态）
type Dependencies struct {
	AuthHandler       *handler.AuthHandler
	LLMHandler        *handler.LLMHandler
	ShopHandler       *handler.ShopHandler
	AdminShopHandler  *handler.AdminShopHandler
	MapHandler        *handler.MapHandler
	PostHandler       *handler.PostHandler
	AdminPostHandler  *handler.AdminPostHandler
	QuizHandler       *handler.QuizHandler
	AdminQuizHandler  *handler.AdminQuizHandler
	AdminHandler      *handler.AdminHandler
	AuthMW            *auth.AuthMiddleware
	Logger            *zap.Logger
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

	// Token 刷新 — 无需 access token（access 过期时使用 refresh token）
	r.POST("/api/v1/auth/refresh", deps.AuthHandler.Refresh)

	// 需要鉴权
	api := r.Group("/api/v1")
	{
		api.GET("/users/me", deps.AuthHandler.GetProfile)
		api.PUT("/users/me", deps.AuthHandler.UpdateProfile)
	}

	// 公开用户资料
	r.GET("/api/v1/users/:user_id", deps.AuthHandler.GetUserProfile)

	// 登出 — 需要鉴权
	apiLogout := r.Group("/api/v1/auth")
	{
		apiLogout.POST("/logout", deps.AuthHandler.Logout)
	}

	// LLM 路由（无需鉴权，限流由中间件处理）
	llmGroup := r.Group("/api/v1/llm")
	{
		llmGroup.POST("/chat", deps.LLMHandler.Chat)
		llmGroup.POST("/story/generate", deps.LLMHandler.GenerateStory)
		llmGroup.GET("/health", deps.LLMHandler.Health)
	}

	// LLM 会话管理
	r.GET("/api/v1/llm/chat/sessions", deps.LLMHandler.ListSessions)
	r.GET("/api/v1/llm/chat/sessions/:session_id", deps.LLMHandler.GetSessionMessages)
	r.DELETE("/api/v1/llm/chat/sessions/:session_id", deps.LLMHandler.DeleteSession)

	// LLM 故事
	r.GET("/api/v1/llm/stories", deps.LLMHandler.ListStories)
	r.GET("/api/v1/llm/story/:story_id", deps.LLMHandler.GetStory)

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
	adminGroup.Use(auth.RequireRole("admin"))
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
	r.GET("/api/v1/map/pois/:poi_id", deps.MapHandler.GetPOI)
	r.GET("/api/v1/map/search", deps.MapHandler.SearchPOIs)
	r.GET("/api/v1/map/timeline", deps.MapHandler.GetTimeline)

	// 帖子 — 浏览/点赞/评论列表无需登录
	postsGroup := r.Group("/api/v1/posts")
	{
		postsGroup.GET("", deps.PostHandler.List)
		postsGroup.GET("/:id", deps.PostHandler.GetByID)
	}

	// 帖子 — 需登录
	apiPosts := r.Group("/api/v1")
	{
		apiPosts.POST("/posts", deps.PostHandler.Create)
		apiPosts.PUT("/posts/:id", deps.PostHandler.Update)
		apiPosts.DELETE("/posts/:id", deps.PostHandler.Delete)
		apiPosts.POST("/posts/:id/like", deps.PostHandler.ToggleLike)
	}

	// 评论 — 查询无需登录，发表/删除需登录
	r.GET("/api/v1/posts/:id/comments", deps.PostHandler.ListComments)
	apiComments := r.Group("/api/v1")
	{
		apiComments.POST("/posts/:id/comments", deps.PostHandler.CreateComment)
		apiComments.DELETE("/comments/:comment_id", deps.PostHandler.DeleteComment)
	}

	// 管理端 — 帖子审核
	adminPostGroup := r.Group("/api/v1/admin/posts")
	adminPostGroup.Use(auth.RequireRole("admin"))
	{
		adminPostGroup.GET("/pending", deps.AdminPostHandler.ListPending)
		adminPostGroup.POST("/:id/review", deps.AdminPostHandler.Review)
	}

	// 问答 — 获取题目无需登录；提交/统计/排行榜需登录
	r.GET("/api/v1/quiz/questions", deps.QuizHandler.GetQuestions)
	quizAuth := r.Group("/api/v1/quiz")
	{
		quizAuth.POST("/submit", deps.QuizHandler.SubmitAnswer)
		quizAuth.POST("/submit-batch", deps.QuizHandler.SubmitBatch)
		quizAuth.GET("/users/:user_id/stats", deps.QuizHandler.GetUserStats)
	}
	r.GET("/api/v1/leaderboard", deps.QuizHandler.GetLeaderboard)

	// 管理端 — 题目 CRUD
	adminQuestionGroup := r.Group("/api/v1/admin/questions")
	adminQuestionGroup.Use(auth.RequireRole("admin"))
	{
		adminQuestionGroup.POST("", deps.AdminQuizHandler.CreateQuestion)
	}

	// 管理端 — 用户管理 & 数据看板
	adminUserGroup := r.Group("/api/v1/admin")
	adminUserGroup.Use(auth.RequireRole("admin"))
	{
		adminUserGroup.GET("/dashboard", deps.AdminHandler.Dashboard)
		adminUserGroup.GET("/users", deps.AdminHandler.ListUsers)
		adminUserGroup.PUT("/users/:user_id/role", deps.AdminHandler.UpdateUserRole)
		adminUserGroup.POST("/users/:user_id/ban", deps.AdminHandler.BanUser)
	}
}
