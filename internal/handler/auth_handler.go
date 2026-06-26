package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/auth"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

// toAppError 将 error 统一转为 *errors.AppError（Service 层返回类型即为 *AppError）
func toAppError(err error) *errors.AppError {
	if e, ok := err.(*errors.AppError); ok {
		return e
	}
	return errors.Internal(err)
}

// AuthHandler HTTP 层 — 只做绑定/校验/响应，业务委托给 AuthService
type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请输入用户名和密码"))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"token_type":    result.TokenType,
		"expires_in":    result.ExpiresIn,
		"user":          result.User,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写注册信息"))
		return
	}

	result, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: req.Nickname,
	})
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.Created(c, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"token_type":    result.TokenType,
		"expires_in":    result.ExpiresIn,
		"user":          result.User,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, user)
}

type updateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写资料"))
		return
	}

	user, err := h.svc.UpdateProfile(c.Request.Context(), userID, service.UpdateProfileInput{
		Nickname:  req.Nickname,
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, user)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请提供 refresh_token"))
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    tokens.TokenType,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := auth.GetUserID(c)
	jti := auth.GetJTI(c)
	expiresAt := auth.GetExpiresAt(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	if err := h.svc.Logout(c.Request.Context(), userID, jti, expiresAt); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.NoContent(c)
}

// GetUserProfile 获取用户公开信息
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.BadRequest("缺少用户 ID"))
		return
	}

	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"nickname":   user.Nickname,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
		"role":       user.Role,
		"points":     user.Points,
		"rank_title": user.RankTitle,
		"created_at": user.CreatedAt,
	})
}
