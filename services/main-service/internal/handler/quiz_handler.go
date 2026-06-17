package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"github.com/grand-canal-guardian/services/main-service/internal/service"
)

type QuizHandler struct {
	svc *service.QuizService
}

func NewQuizHandler(svc *service.QuizService) *QuizHandler {
	return &QuizHandler{svc: svc}
}

// StartSession GET /api/v1/quiz/questions
func (h *QuizHandler) StartSession(c *gin.Context) {
	userID := c.GetString("user_id")
	difficulty := c.DefaultQuery("difficulty", "all")
	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))
	if count < 1 || count > 30 {
		count = 10
	}

	session, questions, appErr := h.svc.StartSession(c.Request.Context(), userID, difficulty, count)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	type sessionResponse struct {
		SessionID string           `json:"session_id"`
		Questions []*model.Question `json:"questions"`
	}
	response.OK(c, sessionResponse{
		SessionID: session.ID,
		Questions: questions,
	})
}

// SubmitAnswer POST /api/v1/quiz/submit
func (h *QuizHandler) SubmitAnswer(c *gin.Context) {
	userID := c.GetString("user_id")
	var req model.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	result, appErr := h.svc.SubmitAnswer(c.Request.Context(), userID, req.SessionID, req.QuestionID, req.SelectedOption)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, result)
}

// GetLeaderboard GET /api/v1/quiz/leaderboard
func (h *QuizHandler) GetLeaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "total")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 { limit = 20 }

	entries, appErr := h.svc.GetLeaderboard(c.Request.Context(), period, limit)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, entries)
}

// GetUserStats GET /api/v1/quiz/users/:id/stats
func (h *QuizHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" || userID == "me" {
		userID = c.GetString("user_id")
	}
	stats, appErr := h.svc.GetUserStats(c.Request.Context(), userID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, stats)
}
