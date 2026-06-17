package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/grand-canal-guardian/pkg/auth"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"github.com/grand-canal-guardian/services/main-service/internal/service"
)

// AuthHandler 认证相关 handler
type AuthHandler struct {
	svc *service.UserService
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	user, appErr := h.svc.Register(c.Request.Context(), &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.Created(c, user)
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	result, appErr := h.svc.Login(c.Request.Context(), &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, result)
}

// Refresh POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	result, appErr := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, result)
}

// Logout POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		response.Error(c, apperrors.NewDefault(apperrors.ErrUnauthorized))
		return
	}
	claims := claimsVal.(*auth.Claims)
	if appErr := h.svc.Logout(c.Request.Context(), claims); appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKMessage(c, "已退出登录")
}
