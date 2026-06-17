package middleware

import (
	"github.com/gin-gonic/gin"
)

// Chain 中间件注册顺序
func Chain(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	r.Use(middlewares...)
}
