package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/auth"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type QuizHandler struct {
	svc service.QuizService
}

func NewQuizHandler(svc service.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}

func (h *QuizHandler) GetQuestions(c *gin.Context) {
	count := queryInt(c, "count", 10)
	if count > 50 {
		count = 50
	}
	difficulty := c.DefaultQuery("difficulty", "mixed")
	category := c.Query("category")

	result, err := h.svc.GetQuestions(c.Request.Context(), count, difficulty, category)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

func (h *QuizHandler) SubmitAnswer(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	var req model.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请提供 session_id、question_id 和 answer"))
		return
	}

	result, err := h.svc.SubmitAnswer(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

type submitBatchRequest struct {
	SessionID string               `json:"session_id" binding:"required"`
	Answers   []service.AnswerItem `json:"answers" binding:"required,min=1"`
}

func (h *QuizHandler) SubmitBatch(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	var req submitBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请提供 session_id 和 answers"))
		return
	}

	result, err := h.svc.SubmitBatch(c.Request.Context(), userID, req.SessionID, req.Answers)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

func (h *QuizHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.BadRequest("缺少用户 ID"))
		return
	}

	stats, err := h.svc.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, stats)
}

func (h *QuizHandler) GetLeaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "total")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	currentUserID := auth.GetUserID(c)

	result, err := h.svc.GetLeaderboard(c.Request.Context(), period, currentUserID, page, pageSize)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKListKey(c, "leaderboard", result.Leaderboard, page, pageSize, int64(len(result.Leaderboard)))
}
