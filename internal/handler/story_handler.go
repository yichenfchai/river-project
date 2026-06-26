package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type StoryHandler struct {
	svc service.StoryService
}

func NewStoryHandler(svc service.StoryService) *StoryHandler {
	return &StoryHandler{svc: svc}
}

func (h *StoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	stories, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, gin.H{
		"stories": stories,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *StoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, apperrors.NewDefault(apperrors.ErrBadRequest))
		return
	}

	story, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	if story == nil {
		response.Error(c, apperrors.NewDefault(apperrors.ErrStoryNotFound))
		return
	}

	response.OK(c, story)
}
