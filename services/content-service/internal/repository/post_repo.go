package repository

import (
	"context"
	"errors"

	"github.com/grand-canal-guardian/services/content-service/internal/model"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepository) FindByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &post, err
}

func (r *PostRepository) Update(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Post{}).Error
}

func (r *PostRepository) List(ctx context.Context, page, pageSize int, topic, keyword, sort string) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Post{}).Where("status = ?", "published")

	if topic != "" && topic != "all" {
		query = query.Where("topic = ?", topic)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	orderClause := "created_at DESC"
	if sort == "popular" {
		orderClause = "like_count DESC, created_at DESC"
	}

	err := query.Order(orderClause).Offset(offset).Limit(pageSize).Find(&posts).Error
	return posts, total, err
}

// ---- Comments ----

func (r *PostRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *PostRepository) ListComments(ctx context.Context, postID string, page, pageSize int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID)
	_ = query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&comments).Error
	return comments, total, err
}

func (r *PostRepository) DeleteComment(ctx context.Context, commentID, authorID string) error {
	return r.db.WithContext(ctx).Where("id = ? AND author_id = ?", commentID, authorID).
		Delete(&model.Comment{}).Error
}

func (r *PostRepository) IncrCommentCount(ctx context.Context, postID string) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// ToggleLike 点赞/取消 → 返回当前是否已点赞
func (r *PostRepository) ToggleLike(ctx context.Context, postID, userID string) (bool, error) {
	// 检查是否已点赞
	var existing model.PostLike
	err := r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&existing).Error

	if err == nil {
		// 已点赞 → 取消
		if delErr := r.db.WithContext(ctx).Delete(&existing).Error; delErr != nil {
			return false, delErr
		}
		r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
		return false, nil
	}

	// 未点赞 → 点赞
	like := &model.PostLike{PostID: postID, UserID: userID}
	if createErr := r.db.WithContext(ctx).Create(like).Error; createErr != nil {
		return false, createErr
	}
	r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1"))
	return true, nil
}

// DeleteCommentByID 通过 comment ID 删除（需验证作者）
func (r *PostRepository) DeleteCommentByID(ctx context.Context, commentID, userID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND author_id = ?", commentID, userID).
		Delete(&model.Comment{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
