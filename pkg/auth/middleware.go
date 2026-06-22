package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/grand-canal-guardian/pkg/errors"
	"github.com/your-org/grand-canal-guardian/pkg/response"
)

type AuthMiddleware struct {
	tm      *TokenManager
	skipPaths []string
}

func NewAuthMiddleware(tm *TokenManager, skipPaths []string) *AuthMiddleware {
	return &AuthMiddleware{tm: tm, skipPaths: skipPaths}
}

func (a *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, p := range a.skipPaths {
			if strings.HasPrefix(c.Request.URL.Path, p) {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.AbortWithError(c, errors.NewDefault(errors.ErrUnauthorized))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := a.tm.ParseToken(tokenStr)
		if err != nil {
			response.AbortWithError(c, errors.NewDefault(errors.ErrTokenInvalid))
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("role", claims.Role)
		c.Set("device_id", claims.DeviceID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	uid, _ := c.Get("user_id")
	if s, ok := uid.(string); ok {
		return s
	}
	return ""
}

func GetRole(c *gin.Context) string {
	role, _ := c.Get("role")
	if s, ok := role.(string); ok {
		return s
	}
	return ""
}
