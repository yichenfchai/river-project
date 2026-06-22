package repository

import (
	"context"
	"crypto/sha1"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// ShopRepository 兑换商店数据访问接口
type ShopRepository interface {
	ListItems(ctx context.Context) ([]model.ShopItem, error)
	GetItem(ctx context.Context, id string) (*model.ShopItem, error)

	// 原子操作（事务内调用）
	DeductPoints(ctx context.Context, tx *gorm.DB, userID string, amount int) error
	DecrementStock(ctx context.Context, tx *gorm.DB, itemID string) error
	CreateRedemption(ctx context.Context, tx *gorm.DB, r *model.Redemption) error

	// 查询
	ListRedemptions(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error)

	// 管理
	CreateItem(ctx context.Context, item *model.ShopItem) error
	UpdateItem(ctx context.Context, item *model.ShopItem) error
	DeleteItem(ctx context.Context, id string) error

	// 事务
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// Seed 幂等种子数据
	Seed(ctx context.Context) error
}

type shopRepo struct {
	db *gorm.DB
}

func NewShopRepo(db *gorm.DB) ShopRepository {
	return &shopRepo{db: db}
}

func uuid5(namespace, name string) string {
	h := sha1.New()
	h.Write([]byte(namespace + ":" + name))
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}

func (r *shopRepo) ListItems(ctx context.Context) ([]model.ShopItem, error) {
	var items []model.ShopItem
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("points_cost ASC").Find(&items).Error
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return items, nil
}

func (r *shopRepo) GetItem(ctx context.Context, id string) (*model.ShopItem, error) {
	var item model.ShopItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrItemNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &item, nil
}

func (r *shopRepo) DeductPoints(ctx context.Context, tx *gorm.DB, userID string, amount int) error {
	result := tx.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND points >= ?", userID, amount).
		UpdateColumn("points", gorm.Expr("points - ?", amount))
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrInsufficientPoints)
	}
	return nil
}

func (r *shopRepo) DecrementStock(ctx context.Context, tx *gorm.DB, itemID string) error {
	result := tx.WithContext(ctx).
		Model(&model.ShopItem{}).
		Where("id = ? AND stock != 0", itemID).
		UpdateColumn("stock", gorm.Expr("stock - 1"))
	if result.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrOutOfStock)
	}
	return nil
}

func (r *shopRepo) CreateRedemption(ctx context.Context, tx *gorm.DB, rd *model.Redemption) error {
	if err := tx.WithContext(ctx).Create(rd).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *shopRepo) ListRedemptions(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Redemption{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	var records []model.Redemption
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return records, total, nil
}

func (r *shopRepo) CreateItem(ctx context.Context, item *model.ShopItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *shopRepo) UpdateItem(ctx context.Context, item *model.ShopItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *shopRepo) DeleteItem(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.ShopItem{}, "id = ?", id).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *shopRepo) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *shopRepo) Seed(ctx context.Context) error {
	seeds := []model.ShopItem{
		{
			ID: uuid5("grand-canal-shop", "运河守护者称号"),
			Name: "🏅 运河守护者称号", Description: "解锁专属称号，在排行榜和个人主页中展示",
			PointsCost: 100, Stock: -1, IsActive: true,
		},
		{
			ID: uuid5("grand-canal-shop", "专属头像框"),
			Name: "🎨 专属头像框", Description: "限定时节头像框，彰显运河文化品味",
			PointsCost: 200, Stock: -1, IsActive: true,
		},
		{
			ID: uuid5("grand-canal-shop", "运河知识手册"),
			Name: "📖 运河知识手册", Description: "解锁隐藏的运河科普故事与冷知识",
			PointsCost: 50, Stock: -1, IsActive: true,
		},
		{
			ID: uuid5("grand-canal-shop", "抽奖券"),
			Name: "🎫 抽奖券", Description: "使用抽奖券获得随机积分奖励（100~500分）",
			PointsCost: 30, Stock: -1, IsActive: true,
		},
	}

	for _, item := range seeds {
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&item).Error; err != nil {
			return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
		}
	}
	return nil
}
