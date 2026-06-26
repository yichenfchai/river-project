package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/auth"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type PostHandler struct {
	svc service.PostService
}

func NewPostHandler(svc service.PostService) *PostHandler {
	return &PostHandler{svc: svc}
}

// ─── 发帖 ───

type createPostRequest struct {
	Title   string   `json:"title" binding:"required,min=1,max=200"`
	Content string   `json:"content" binding:"required,min=1,max=10000"`
	Topic   string   `json:"topic" binding:"omitempty,oneof=share ecology culture question other"`
	Images  []string `json:"images" binding:"omitempty,max=9"`
	Tags    []string `json:"tags" binding:"omitempty,max=5"`
}

func (h *PostHandler) Create(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写帖子信息"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	if req.Topic == "" {
		req.Topic = "share"
	}

	post, err := h.svc.Create(c.Request.Context(), userID, req.Title, req.Content, req.Topic, req.Images, req.Tags)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.Created(c, post)
}

// ─── 帖子列表 ───

func (h *PostHandler) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)
	topic := c.Query("topic")
	tag := c.Query("tag")
	keyword := c.Query("keyword")
	sort := c.DefaultQuery("sort", "created_at")

	result, err := h.svc.List(c.Request.Context(), service.PostListQuery{
		Page:     page,
		PageSize: pageSize,
		Topic:    topic,
		Tag:      tag,
		Keyword:  keyword,
		Sort:     sort,
	})
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKListKey(c, "posts", result.Posts, result.Page, result.PageSize, result.Total)
}

// ─── 帖子详情 ───

func (h *PostHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	currentUserID := auth.GetUserID(c)

	post, err := h.svc.GetByID(c.Request.Context(), id, currentUserID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, post)
}

// ─── 编辑帖子 ───

type updatePostRequest struct {
	Title   string   `json:"title" binding:"omitempty,max=200"`
	Content string   `json:"content" binding:"omitempty,max=10000"`
	Tags    []string `json:"tags" binding:"omitempty,max=5"`
}

func (h *PostHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写帖子信息"))
		return
	}

	post, err := h.svc.Update(c.Request.Context(), id, userID, req.Title, req.Content, req.Tags)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, post)
}

// ─── 删除帖子 ───

func (h *PostHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.NoContent(c)
}

// ─── 点赞 / 取消点赞 ───

func (h *PostHandler) ToggleLike(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	result, err := h.svc.ToggleLike(c.Request.Context(), id, userID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

// ─── 评论列表 ───

func (h *PostHandler) ListComments(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	result, err := h.svc.ListComments(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKListKey(c, "comments", result.Comments, result.Page, result.PageSize, result.Total)
}

// ─── 发表评论 ───

type createCommentRequest struct {
	Content  string  `json:"content" binding:"required,min=1,max=2000"`
	ParentID *string `json:"parent_id"`
}

func (h *PostHandler) CreateComment(c *gin.Context) {
	postID := c.Param("id")
	if postID == "" {
		response.Error(c, errors.BadRequest("缺少帖子 ID"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请填写评论内容"))
		return
	}

	comment, err := h.svc.CreateComment(c.Request.Context(), postID, userID, req.Content, req.ParentID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.Created(c, comment)
}

// ─── 删除评论 ───

func (h *PostHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		response.Error(c, errors.BadRequest("缺少评论 ID"))
		return
	}

	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	if err := h.svc.DeleteComment(c.Request.Context(), commentID, userID); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.NoContent(c)
}
