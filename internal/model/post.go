package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID           string         `gorm:"primaryKey;size:36" json:"id"`
	AuthorID     string         `gorm:"size:36;not null;index:idx_posts_author_id" json:"author_id"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Images       string         `gorm:"type:text;default:'[]'" json:"-"`
	Tags         string         `gorm:"type:text;default:'[]'" json:"-"`
	Topic        string         `gorm:"size:32;default:share;index:idx_posts_topic" json:"topic"`
	LikeCount    int            `gorm:"default:0" json:"like_count"`
	CommentCount int            `gorm:"default:0" json:"comment_count"`
	Status       string         `gorm:"size:16;default:pending;index:idx_posts_status" json:"status"`
	CreatedAt    time.Time      `gorm:"index:idx_posts_created_at,sort:desc" json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Post) TableName() string { return "posts" }

type PostLike struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	PostID    string    `gorm:"size:36;not null;uniqueIndex:idx_post_like_unique" json:"post_id"`
	UserID    string    `gorm:"size:36;not null;uniqueIndex:idx_post_like_unique" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (PostLike) TableName() string { return "post_likes" }

type Comment struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	PostID    string         `gorm:"size:36;not null;index:idx_comments_post_id" json:"post_id"`
	AuthorID  string         `gorm:"size:36;not null;index:idx_comments_author_id" json:"author_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	ParentID  *string        `gorm:"size:36" json:"parent_id,omitempty"`
	ReplyToID *string        `gorm:"size:36" json:"reply_to_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Comment) TableName() string { return "comments" }

// ─── DTOs ───

type PostResponse struct {
	ID           string    `json:"id"`
	Author       UserJSON  `json:"author"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Images       []string  `json:"images"`
	Tags         []string  `json:"tags"`
	Topic        string    `json:"topic"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	IsLiked      bool      `json:"is_liked"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserJSON struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
}

type CommentResponse struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	Author    UserJSON   `json:"author"`
	Content   string     `json:"content"`
	ParentID  *string    `json:"parent_id,omitempty"`
	ReplyTo   *UserJSON  `json:"reply_to,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
