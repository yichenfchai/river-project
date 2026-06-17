package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
)

// AuthMiddleware JWT 鉴权中间件
type AuthMiddleware struct {
	tm *TokenManager
}

func NewAuthMiddleware(tm *TokenManager) *AuthMiddleware {
	return &AuthMiddleware{tm: tm}
}

// RequireAuth 要求登录
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
			return
		}

		claims, err := m.tm.ValidateAccess(c.Request.Context(), tokenStr)
		if err != nil {
			if appErr, ok := err.(*errors.AppError); ok {
				response.Error(c, appErr)
			} else {
				response.Error(c, errors.NewDefault(errors.ErrTokenInvalid))
			}
			return
		}

		// 注入用户信息
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("device_id", claims.DeviceID)
		c.Set("claims", claims)
		c.Next()
	}
}

// RequireRoles 要求指定角色
func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
			return
		}

		roleStr := role.(string)
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}
		response.Error(c, errors.NewDefault(errors.ErrForbidden))
	}
}

// RequirePermissions 要求指定权限 (RBAC 位掩码)
// permissions: 所需权限位掩码 OR 组合
// userPermissions 来自 token claims 中的权限字段
func (m *AuthMiddleware) RequirePermissions(requiredMask uint32, getUserMask func(c *gin.Context) uint32) gin.HandlerFunc {
	return func(c *gin.Context) {
		userMask := getUserMask(c)
		if userMask&requiredMask != requiredMask {
			response.Error(c, errors.NewDefault(errors.ErrForbidden))
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Header: Authorization: Bearer <token>
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Query: ?token=<token> (WebSocket fallback)
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}
