package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type AdminHandler struct {
	svc service.AdminService
}

func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}
	response.OK(c, stats)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	role := c.Query("role")
	keyword := c.Query("keyword")

	result, err := h.svc.ListUsers(c.Request.Context(), page, pageSize, role, keyword)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKListKey(c, "users", result.Users, result.Page, result.PageSize, result.Total)
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user monitor admin"`
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.BadRequest("缺少用户 ID"))
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请选择有效角色"))
		return
	}

	user, err := h.svc.UpdateUserRole(c.Request.Context(), userID, req.Role)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, user)
}

type banUserRequest struct {
	Banned bool    `json:"banned"`
	Reason *string `json:"reason"`
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.BadRequest("缺少用户 ID"))
		return
	}

	var req banUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请指定封禁状态"))
		return
	}

	if err := h.svc.BanUser(c.Request.Context(), userID, req.Banned); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKMessage(c, "操作成功")
}
