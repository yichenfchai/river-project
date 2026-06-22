package model

import (
	"time"

	"gorm.io/gorm"
)

// MapLayer 地图图层（朝代运河路线）
type MapLayer struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Era         string         `gorm:"size:64;not null;index" json:"era"`
	Description string         `gorm:"size:512" json:"description"`
	Color       string         `gorm:"size:32" json:"color"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	GeoJSON     string         `gorm:"type:text" json:"geojson"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MapLayer) TableName() string { return "map_layers" }

// MapPOI 地图兴趣点
type MapPOI struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"`
	LayerID     string         `gorm:"size:36;index" json:"layer_id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	Category    string         `gorm:"size:64;index" json:"category"`
	Lat         float64        `gorm:"not null;index" json:"lat"`
	Lng         float64        `gorm:"not null;index" json:"lng"`
	IconURL     string         `gorm:"size:512" json:"icon_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MapPOI) TableName() string { return "map_pois" }
