package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/your-org/grand-canal-guardian/internal/model"
	"github.com/your-org/grand-canal-guardian/internal/repository"
	apperrors "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// ShopService 兑换商店业务逻辑接口
type ShopService interface {
	ListItems(ctx context.Context) ([]model.ShopItem, error)
	Redeem(ctx context.Context, userID, role, itemID string) (*RedemptionResult, error)
	GetHistory(ctx context.Context, userID string, page, pageSize int) (*RedemptionList, error)

	// Admin
	CreateItem(ctx context.Context, item *model.ShopItem) error
	UpdateItem(ctx context.Context, item *model.ShopItem) error
	DeleteItem(ctx context.Context, id string) error
}

type RedemptionResult struct {
	Redemption model.Redemption `json:"redemption"`
	UserPoints int              `json:"user_points"`
}

type RedemptionList struct {
	Items []model.Redemption `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
}

type shopService struct {
	shopRepo repository.ShopRepository
	userRepo repository.UserRepository
	log      *zap.Logger
}

func NewShopService(shopRepo repository.ShopRepository, userRepo repository.UserRepository, log *zap.Logger) ShopService {
	return &shopService{shopRepo: shopRepo, userRepo: userRepo, log: log}
}

func (s *shopService) ListItems(ctx context.Context) ([]model.ShopItem, error) {
	return s.shopRepo.ListItems(ctx)
}

func (s *shopService) Redeem(ctx context.Context, userID, role, itemID string) (*RedemptionResult, error) {
	if role == "guest" {
		return nil, apperrors.Forbidden("游客无法兑换商品，请先登录")
	}

	// 先查用户积分（事务外读取，降低持锁时间）
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 先查商品
	item, err := s.shopRepo.GetItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if !item.IsActive {
		return nil, apperrors.NewDefault(apperrors.ErrItemNotFound)
	}
	if item.Stock == 0 {
		return nil, apperrors.NewDefault(apperrors.ErrOutOfStock)
	}
	if user.Points < item.PointsCost {
		return nil, apperrors.NewDefault(apperrors.ErrInsufficientPoints)
	}

	redemption := model.Redemption{
		ID:          uuid.New().String(),
		UserID:      userID,
		ItemID:      itemID,
		ItemName:    item.Name,
		PointsSpent: item.PointsCost,
		Status:      "completed",
	}

	err = s.shopRepo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.shopRepo.DeductPoints(ctx, tx, userID, item.PointsCost); err != nil {
			return err
		}
		if item.Stock > 0 {
			if err := s.shopRepo.DecrementStock(ctx, tx, itemID); err != nil {
				return err
			}
		}
		return s.shopRepo.CreateRedemption(ctx, tx, &redemption)
	})
	if err != nil {
		return nil, err
	}

	s.log.Info("兑换成功",
		zap.String("user_id", userID),
		zap.String("item", item.Name),
		zap.Int("points", item.PointsCost),
	)

	return &RedemptionResult{
		Redemption: redemption,
		UserPoints: user.Points - item.PointsCost,
	}, nil
}

func (s *shopService) GetHistory(ctx context.Context, userID string, page, pageSize int) (*RedemptionList, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	records, total, err := s.shopRepo.ListRedemptions(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	return &RedemptionList{
		Items: records,
		Total: total,
		Page:  page,
	}, nil
}

func (s *shopService) CreateItem(ctx context.Context, item *model.ShopItem) error {
	item.ID = uuid.New().String()
	return s.shopRepo.CreateItem(ctx, item)
}

func (s *shopService) UpdateItem(ctx context.Context, item *model.ShopItem) error {
	existing, err := s.shopRepo.GetItem(ctx, item.ID)
	if err != nil {
		return err
	}
	existing.Name = item.Name
	existing.Description = item.Description
	existing.ImageURL = item.ImageURL
	existing.PointsCost = item.PointsCost
	existing.Stock = item.Stock
	existing.IsActive = item.IsActive
	return s.shopRepo.UpdateItem(ctx, existing)
}

func (s *shopService) DeleteItem(ctx context.Context, id string) error {
	if _, err := s.shopRepo.GetItem(ctx, id); err != nil {
		return err
	}
	return s.shopRepo.DeleteItem(ctx, id)
}
