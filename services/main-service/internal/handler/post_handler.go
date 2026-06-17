package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/pkg/response"
	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"github.com/grand-canal-guardian/services/main-service/internal/service"
)

type PostHandler struct {
	svc *service.PostService
}

func NewPostHandler(svc *service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// CreatePost POST /api/v1/posts
func (h *PostHandler) CreatePost(c *gin.Context) {
	authorID := c.GetString("user_id")
	var req model.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	post, appErr := h.svc.CreatePost(c.Request.Context(), authorID, &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.Created(c, post)
}

// GetPost GET /api/v1/posts/:id
func (h *PostHandler) GetPost(c *gin.Context) {
	post, appErr := h.svc.GetPost(c.Request.Context(), c.Param("id"))
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, post)
}

// ListPosts GET /api/v1/posts
func (h *PostHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }

	topic := c.DefaultQuery("topic", "all")
	keyword := c.Query("keyword")
	sort := c.DefaultQuery("sort", "latest")

	posts, total, appErr := h.svc.ListPosts(c.Request.Context(), page, pageSize, topic, keyword, sort)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKList(c, posts, page, pageSize, total)
}

// UpdatePost PUT /api/v1/posts/:id
func (h *PostHandler) UpdatePost(c *gin.Context) {
	authorID := c.GetString("user_id")
	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	post, appErr := h.svc.UpdatePost(c.Request.Context(), c.Param("id"), authorID, &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OK(c, post)
}

// DeletePost DELETE /api/v1/posts/:id
func (h *PostHandler) DeletePost(c *gin.Context) {
	authorID := c.GetString("user_id")
	if appErr := h.svc.DeletePost(c.Request.Context(), c.Param("id"), authorID); appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.NoContent(c)
}

// CreateComment POST /api/v1/posts/:id/comments
func (h *PostHandler) CreateComment(c *gin.Context) {
	authorID := c.GetString("user_id")
	var req model.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	comment, appErr := h.svc.CreateComment(c.Request.Context(), c.Param("id"), authorID, &req)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.Created(c, comment)
}

// ListComments GET /api/v1/posts/:id/comments
func (h *PostHandler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	comments, total, appErr := h.svc.ListComments(c.Request.Context(), c.Param("id"), page, pageSize)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.OKList(c, comments, page, pageSize, total)


// ToggleLike POST /api/v1/posts/:id/like
func (h *PostHandler) ToggleLike(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, apperrors.NewDefault(apperrors.ErrUnauthorized))
		return
	}

	liked, appErr := h.svc.ToggleLike(c.Request.Context(), c.Param("id"), userID)
	if appErr != nil {
		response.Error(c, appErr)
		return
	}

	response.OK(c, gin.H{"liked": liked})
}

// DeleteComment DELETE /api/v1/comments/:id
func (h *PostHandler) DeleteComment(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, apperrors.NewDefault(apperrors.ErrUnauthorized))
		return
	}

	if appErr := h.svc.DeleteComment(c.Request.Context(), c.Param("id"), userID); appErr != nil {
		response.Error(c, appErr)
		return
	}
	response.NoContent(c)
}

}