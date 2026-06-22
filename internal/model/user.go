package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string         `gorm:"size:256;not null" json:"-"`
	Nickname  string         `gorm:"size:128" json:"nickname"`
	Email     string         `gorm:"size:256" json:"email"`
	AvatarURL string         `gorm:"size:512" json:"avatar_url"`
	Role      string         `gorm:"size:32;default:user" json:"role"`
	Bio       string         `gorm:"size:512" json:"bio"`
	Points    int            `gorm:"default:0" json:"points"`
	RankTitle string         `gorm:"size:64" json:"rank_title"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
