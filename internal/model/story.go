package model

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	Title     string         `gorm:"size:200;not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Topic     string         `gorm:"size:32;index:idx_stories_topic" json:"topic"`
	Emoji     string         `gorm:"size:16" json:"emoji"`
	AgeGroup  string         `gorm:"size:32;default:青少年" json:"age_group"`
	Likes     int            `gorm:"default:0" json:"likes"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Story) TableName() string {
	return "stories"
}
