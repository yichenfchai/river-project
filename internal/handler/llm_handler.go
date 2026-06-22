package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type LLMHandler struct {
	svc service.LLMService
}

func NewLLMHandler(svc service.LLMService) *LLMHandler {
	return &LLMHandler{svc: svc}
}

// Chat SSE 流式对话
func (h *LLMHandler) Chat(c *gin.Context) {
	var req struct {
		Message   string `json:"message" binding:"required,min=1,max=2000"`
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请输入消息内容"))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, errors.BadRequest("不支持 SSE"))
		return
	}

	ctx := c.Request.Context()
	index := 0

	err := h.svc.Chat(ctx, service.ChatInput{
		Message:   req.Message,
		SessionID: req.SessionID,
	}, func(token string) error {
		event := map[string]interface{}{
			"type":  "token",
			"token": token,
			"index": index,
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
		index++
		return nil
	})

	if err != nil {
		event := map[string]interface{}{
			"type":  "error",
			"error": "对话服务暂不可用",
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	// 发送完成事件
	done := map[string]interface{}{"type": "done", "done": true}
	doneData, _ := json.Marshal(done)
	fmt.Fprintf(c.Writer, "data: %s\n\n", doneData)
	flusher.Flush()
}

// GenerateStory 生成科普故事（非流式）
func (h *LLMHandler) GenerateStory(c *gin.Context) {
	var req struct {
		Topic   string `json:"topic" binding:"required,oneof=history ecology culture legend technology"`
		Keyword string `json:"keyword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请选择主题方向"))
		return
	}

	result, err := h.svc.GenerateStory(c.Request.Context(), service.StoryInput{
		Topic:   req.Topic,
		Keyword: req.Keyword,
	})
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

// Health LLM 服务健康检查
func (h *LLMHandler) Health(c *gin.Context) {
	ok := h.svc.Health(c.Request.Context())
	status := "unavailable"
	if ok {
		status = "ok"
	}
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"llm":    ok,
	})
}
