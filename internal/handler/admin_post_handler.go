package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type AdminPostHandler struct {
	svc service.PostService
}

func NewAdminPostHandler(svc service.PostService) *AdminPostHandler {
	return &AdminPostHandler{svc: svc}
}

// ─── 待审核帖子列表 ───

func (h *AdminPostHandler) ListPending(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	result, err := h.svc.ListPending(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKListKey(c, "posts", result.Posts, result.Page, result.PageSize, result.Total)
}

// ─── 审核帖子 ───

type reviewPostRequest struct {
	Action string  `json:"action" binding:"required,oneof=approve reject"`
	Reason *string `json:"reason"`
}

func (h *AdminPostHandler) Review(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	var req reviewPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请选择审核操作"))
		return
	}

	reason := ""
	if req.Reason != nil {
		reason = *req.Reason
	}

	result, err := h.svc.Review(c.Request.Context(), id, req.Action, reason)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}
