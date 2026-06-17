package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:32;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Email        string         `gorm:"uniqueIndex;size:128;not null" json:"email"`
	Nickname     string         `gorm:"size:64" json:"nickname"`
	AvatarURL    string         `gorm:"size:512" json:"avatar_url"`
	Bio          string         `gorm:"size:500" json:"bio"`
	Role         string         `gorm:"size:16;default:user" json:"role"`
	Points       int            `gorm:"default:0" json:"points"`
	RankTitle    string         `gorm:"size:32;default:青铜守护者" json:"rank_title"`
	Status       string         `gorm:"size:16;default:active" json:"status"` // active | banned
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"max=64"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"` // 可选: 设备指纹
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user"`
}

// RefreshRequest 刷新 Token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"max=64"`
	AvatarURL string `json:"avatar_url" binding:"max=512"`
	Bio       string `json:"bio" binding:"max=500"`
}
