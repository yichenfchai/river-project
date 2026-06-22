package model

import (
	"time"

	"gorm.io/gorm"
)

type ShopItem struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	ImageURL    string         `gorm:"size:512" json:"image_url"`
	PointsCost  int            `gorm:"not null;default:0" json:"points_cost"`
	Stock       int            `gorm:"not null;default:-1" json:"stock"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ShopItem) TableName() string { return "shop_items" }

type Redemption struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	UserID      string    `gorm:"size:36;not null;index" json:"user_id"`
	ItemID      string    `gorm:"size:36;not null" json:"item_id"`
	ItemName    string    `gorm:"size:128" json:"item_name"`
	PointsSpent int       `gorm:"not null" json:"points_spent"`
	Status      string    `gorm:"size:32;default:completed" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Redemption) TableName() string { return "redemptions" }
