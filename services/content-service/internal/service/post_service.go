package service

import (
	"context"
	"encoding/json"

	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"gorm.io/gorm"
	"github.com/grand-canal-guardian/services/content-service/internal/model"
	"github.com/grand-canal-guardian/services/content-service/internal/repository"
)

type PostService struct {
	repo *repository.PostRepository
}

func NewPostService(repo *repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

// CreatePost 发帖
func (s *PostService) CreatePost(ctx context.Context, authorID string, req *model.CreatePostRequest) (*model.Post, *apperrors.AppError) {
	imagesJSON, _ := json.Marshal(req.Images)
	tagsJSON, _ := json.Marshal(req.Tags)

	post := &model.Post{
		AuthorID: authorID,
		Title:    req.Title,
		Content:  req.Content,
		Images:   string(imagesJSON),
		Tags:     string(tagsJSON),
		Topic:    req.Topic,
		Status:   "published",
	}
	if post.Topic == "" {
		post.Topic = "share"
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return post, nil
}

// GetPost 获取帖子
func (s *PostService) GetPost(ctx context.Context, postID string) (*model.Post, *apperrors.AppError) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if post == nil {
		return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	return post, nil
}

// ListPosts 帖子列表
func (s *PostService) ListPosts(ctx context.Context, page, pageSize int, topic, keyword, sort string) ([]*model.Post, int64, *apperrors.AppError) {
	posts, total, err := s.repo.List(ctx, page, pageSize, topic, keyword, sort)
	if err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return posts, total, nil
}

// UpdatePost 编辑帖子
func (s *PostService) UpdatePost(ctx context.Context, postID, authorID string, req *model.UpdatePostRequest) (*model.Post, *apperrors.AppError) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if post == nil {
		return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	if post.AuthorID != authorID {
		return nil, apperrors.NewDefault(apperrors.ErrPostNotOwner)
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Images != nil {
		imagesJSON, _ := json.Marshal(req.Images)
		post.Images = string(imagesJSON)
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		post.Tags = string(tagsJSON)
	}
	if req.Topic != "" {
		post.Topic = req.Topic
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return post, nil
}

// DeletePost 删除帖子
func (s *PostService) DeletePost(ctx context.Context, postID, authorID string) *apperrors.AppError {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if post == nil {
		return apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	if post.AuthorID != authorID {
		return apperrors.NewDefault(apperrors.ErrPostNotOwner)
	}

	if err := s.repo.Delete(ctx, postID); err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

// CreateComment 评论
func (s *PostService) CreateComment(ctx context.Context, postID, authorID string, req *model.CreateCommentRequest) (*model.Comment, *apperrors.AppError) {
	post, repoErr := s.repo.FindByID(ctx, postID)
	if repoErr != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, repoErr)
	}
	if post == nil {
		return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
	}

	comment := &model.Comment{
		PostID:   postID,
		AuthorID: authorID,
		Content:  req.Content,
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	// 更新评论数
	_ = s.repo.IncrCommentCount(ctx, postID)

	return comment, nil
}

// ListComments 评论列表
func (s *PostService) ListComments(ctx context.Context, postID string, page, pageSize int) ([]*model.Comment, int64, *apperrors.AppError) {
	comments, total, err := s.repo.ListComments(ctx, postID, page, pageSize)
	if err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return comments, total, nil
}

// DeleteComment 删除评论
func (s *PostService) DeleteComment(ctx context.Context, commentID, userID string) *apperrors.AppError {
	if err := s.repo.DeleteCommentByID(ctx, commentID, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewDefault(apperrors.ErrCommentNotFound)
		}
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}


// ToggleLike 点赞/取消
func (s *PostService) ToggleLike(ctx context.Context, postID, userID string) (bool, *apperrors.AppError) {
	liked, err := s.repo.ToggleLike(ctx, postID, userID)
	if err != nil {
		return false, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return liked, nil
}
