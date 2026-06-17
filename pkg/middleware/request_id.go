package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID 为每个请求生成唯一 ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// GetRequestID 从 context 获取请求 ID
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
