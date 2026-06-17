package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
	"go.uber.org/zap"
)

// Recovery 捕获 panic，返回统一错误 (不泄露原始信息)
// 必须放在中间件链最内层 (第一个注册)
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				// 日志记录完整信息 (受访问控制保护)
				logger.Error("panic recovered",
					zap.String("request_id", GetRequestID(c)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Any("panic", r),
					zap.String("stack", stack),
				)

				// HTTP 响应: 只返回通用内部错误
				response.Error(c, &errors.AppError{
					Code:    errors.ErrInternal,
					Message: "服务器内部错误",
					Safe: &errors.SafeDetail{
						Type:    "panic",
						Summary: "internal_server_error",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func init() { _ = fmt.Sprintf("") }
