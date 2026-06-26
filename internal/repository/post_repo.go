package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// PostRepository 帖子数据访问接口
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id string) (*model.Post, error)
	List(ctx context.Context, opts PostListOptions) ([]model.Post, int64, error)
	Update(ctx context.Context, post *model.Post) error
	SoftDelete(ctx context.Context, id string) error

	ToggleLike(ctx context.Context, postID, userID string) (isLiked bool, likeCount int, err error)
	IsLikedByUser(ctx context.Context, postID, userID string) (bool, error)

	CreateComment(ctx context.Context, comment *model.Comment) error
	ListComments(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error)
	SoftDeleteComment(ctx context.Context, id string) error

	UpdateStatus(ctx context.Context, id, status string) error

	// 批量查询作者信息（用于组装响应）
	FindUsersByIDs(ctx context.Context, ids []string) (map[string]model.UserJSON, error)

	// 统计
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

type PostListOptions struct {
	Page     int
	PageSize int
	Topic    string
	Tag      string
	Keyword  string
	Status   string
	Sort     string // "created_at" | "like_count"
}

type postRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(ctx context.Context, post *model.Post) error {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *postRepo) FindByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &post, nil
}

func (r *postRepo) List(ctx context.Context, opts PostListOptions) ([]model.Post, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Post{})

	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}
	if opts.Topic != "" {
		q = q.Where("topic = ?", opts.Topic)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", like, like)
	}
	if opts.Tag != "" {
		q = q.Where("tags LIKE ?", "%"+opts.Tag+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	order := "created_at DESC"
	switch opts.Sort {
	case "like_count":
		order = "like_count DESC"
	case "created_at_asc":
		order = "created_at ASC"
	}

	offset := (opts.Page - 1) * opts.PageSize
	var posts []model.Post
	if err := q.Order(order).Offset(offset).Limit(opts.PageSize).Find(&posts).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return posts, total, nil
}

func (r *postRepo) Update(ctx context.Context, post *model.Post) error {
	if err := r.db.WithContext(ctx).Save(post).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *postRepo) SoftDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Post{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	return nil
}

func (r *postRepo) ToggleLike(ctx context.Context, postID, userID string) (bool, int, error) {
	var like model.PostLike
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postID, userID).
		First(&like).Error

	if err == nil {
		if delErr := r.db.WithContext(ctx).Delete(&like).Error; delErr != nil {
			return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, delErr)
		}
		if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; updErr != nil {
			return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
		}
		var post model.Post
		_ = r.db.WithContext(ctx).Select("like_count").Where("id = ?", postID).First(&post).Error
		return false, post.LikeCount, nil
	}

	if err != gorm.ErrRecordNotFound {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	like = model.PostLike{PostID: postID, UserID: userID}
	if createErr := r.db.WithContext(ctx).Create(&like).Error; createErr != nil {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, createErr)
	}
	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; updErr != nil {
		return false, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}
	var post model.Post
	_ = r.db.WithContext(ctx).Select("like_count").Where("id = ?", postID).First(&post).Error
	return true, post.LikeCount, nil
}

func (r *postRepo) IsLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PostLike{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count > 0, nil
}

func (r *postRepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", comment.PostID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; updErr != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}
	return nil
}

func (r *postRepo) ListComments(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	var comments []model.Comment
	if err := q.Order("created_at ASC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return comments, total, nil
}

func (r *postRepo) SoftDeleteComment(ctx context.Context, id string) error {
	var comment model.Comment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewDefault(apperrors.ErrCommentNotFound)
		}
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	result := r.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}

	if updErr := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", comment.PostID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)")).Error; updErr != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, updErr)
	}

	return nil
}

func (r *postRepo) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrPostNotFound)
	}
	return nil
}

func (r *postRepo) FindUsersByIDs(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
	if len(ids) == 0 {
		return map[string]model.UserJSON{}, nil
	}
	dedup := make([]string, 0, len(ids))
	seen := make(map[string]bool)
	for _, id := range ids {
		if !seen[id] {
			dedup = append(dedup, id)
			seen[id] = true
		}
	}

	var users []model.User
	if err := r.db.WithContext(ctx).
		Select("id, username, nickname, avatar_url, role").
		Where("id IN ?", dedup).
		Find(&users).Error; err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	result := make(map[string]model.UserJSON, len(users))
	for _, u := range users {
		result[u.ID] = model.UserJSON{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			AvatarURL: u.AvatarURL,
			Role:      u.Role,
		}
	}
	return result, nil
}

func parseImages(raw string) []string {
	if raw == "" || raw == "null" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []string{}
	}
	if arr == nil {
		return []string{}
	}
	return arr
}

func parseTags(raw string) []string {
	if raw == "" || raw == "null" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []string{}
	}
	if arr == nil {
		return []string{}
	}
	return arr
}

func stringifyJSON(arr []string) string {
	if arr == nil {
		return "[]"
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

func (r *postRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func (r *postRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Post{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}
