package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type AdminQuizHandler struct {
	svc service.QuizService
}

func NewAdminQuizHandler(svc service.QuizService) *AdminQuizHandler {
	return &AdminQuizHandler{svc: svc}
}

type createQuestionRequest struct {
	Question    string   `json:"question" binding:"required"`
	Options     []string `json:"options" binding:"required,min=2,max=6"`
	Answer      string   `json:"answer" binding:"required"`
	Explanation string   `json:"explanation"`
	Difficulty  string   `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Category    string   `json:"category" binding:"required,oneof=history ecology culture geography water_conservancy"`
}

func (h *AdminQuizHandler) CreateQuestion(c *gin.Context) {
	var req createQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写题目信息"))
		return
	}

	if err := h.svc.CreateQuestion(c.Request.Context(), req.Question, req.Category, req.Difficulty, req.Options, req.Answer, req.Explanation); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.Created(c, nil)
}
