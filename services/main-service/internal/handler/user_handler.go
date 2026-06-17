package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"github.com/grand-canal-guardian/services/main-service/internal/service"
)

// UserHandler 用户相关 handler
type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile GET /api/v1/users/me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	user, appErr := h.svc.GetProfile(c.Request.Context(), userID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, user)
}

// UpdateProfile PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	user, appErr := h.svc.UpdateProfile(c.Request.Context(), userID, &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, user)
}

// GetUser GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, appErr := h.svc.GetProfile(c.Request.Context(), userID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, user)
}

// ListUsers GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, appErr := h.svc.ListUsers(c.Request.Context(), page, pageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKList(c, users, page, pageSize, total)
}

// BanUser POST /api/v1/admin/users/:id/ban
func (h *UserHandler) BanUser(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Ban bool `json:"ban"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	if appErr := h.svc.BanUser(c.Request.Context(), userID, req.Ban); appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKMessage(c, "操作成功")
}

// ChangeRole PUT /api/v1/admin/users/:id/role
func (h *UserHandler) ChangeRole(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	if appErr := h.svc.ChangeRole(c.Request.Context(), userID, req.Role); appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKMessage(c, "角色修改成功")
}
