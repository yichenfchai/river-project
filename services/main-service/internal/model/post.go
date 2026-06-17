package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 帖子模型
type Post struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	AuthorID     string         `gorm:"index;not null" json:"author_id"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Images       string         `gorm:"type:text" json:"images"`       // JSON array stored as string
	Tags         string         `gorm:"type:text" json:"tags"`         // JSON array stored as string
	Topic        string         `gorm:"size:32;index;default:share" json:"topic"`
	LikeCount    int            `gorm:"default:0" json:"like_count"`
	CommentCount int            `gorm:"default:0" json:"comment_count"`
	Status       string         `gorm:"size:16;default:draft" json:"status"` // draft | published | rejected
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Post) TableName() string { return "posts" }

// Comment 评论模型
type Comment struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PostID    string         `gorm:"index;not null" json:"post_id"`
	AuthorID  string         `gorm:"index;not null" json:"author_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Comment) TableName() string { return "comments" }

// PostLike 点赞记录
type PostLike struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	PostID    string    `gorm:"uniqueIndex:idx_post_user;not null" json:"post_id"`
	UserID    string    `gorm:"uniqueIndex:idx_post_user;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (PostLike) TableName() string { return "post_likes" }


// CreatePostRequest 发帖请求
type CreatePostRequest struct {
	Title   string   `json:"title" binding:"required,max=200"`
	Content string   `json:"content" binding:"required"`
	Images  []string `json:"images"`
	Tags    []string `json:"tags"`
	Topic   string   `json:"topic"`
}

// UpdatePostRequest 编辑帖子请求
type UpdatePostRequest struct {
	Title   string   `json:"title" binding:"max=200"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Tags    []string `json:"tags"`
	Topic   string   `json:"topic"`
}

// CreateCommentRequest 评论请求
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}
