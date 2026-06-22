package service

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// ─── Mock Shop Repository ───

type mockShopRepo struct {
	items          map[string]*model.ShopItem
	listItemsFn    func(ctx context.Context) ([]model.ShopItem, error)
	getItemFn      func(ctx context.Context, id string) (*model.ShopItem, error)
	transactionFn  func(ctx context.Context, fn func(tx *gorm.DB) error) error
	createItemFn   func(ctx context.Context, item *model.ShopItem) error
	updateItemFn   func(ctx context.Context, item *model.ShopItem) error
	deleteItemFn   func(ctx context.Context, id string) error
	listRedemptionsFn func(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error)
}

func (m *mockShopRepo) ListItems(ctx context.Context) ([]model.ShopItem, error) {
	if m.listItemsFn != nil {
		return m.listItemsFn(ctx)
	}
	return nil, nil
}
func (m *mockShopRepo) GetItem(ctx context.Context, id string) (*model.ShopItem, error) {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, id)
	}
	return nil, apperrors.NewDefault(apperrors.ErrItemNotFound)
}
func (m *mockShopRepo) DeductPoints(ctx context.Context, tx *gorm.DB, userID string, amount int) error {
	return nil
}
func (m *mockShopRepo) DecrementStock(ctx context.Context, tx *gorm.DB, itemID string) error {
	return nil
}
func (m *mockShopRepo) CreateRedemption(ctx context.Context, tx *gorm.DB, r *model.Redemption) error {
	return nil
}
func (m *mockShopRepo) ListRedemptions(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error) {
	if m.listRedemptionsFn != nil {
		return m.listRedemptionsFn(ctx, userID, offset, limit)
	}
	return nil, 0, nil
}
func (m *mockShopRepo) CreateItem(ctx context.Context, item *model.ShopItem) error {
	if m.createItemFn != nil {
		return m.createItemFn(ctx, item)
	}
	return nil
}
func (m *mockShopRepo) UpdateItem(ctx context.Context, item *model.ShopItem) error {
	if m.updateItemFn != nil {
		return m.updateItemFn(ctx, item)
	}
	return nil
}
func (m *mockShopRepo) DeleteItem(ctx context.Context, id string) error {
	if m.deleteItemFn != nil {
		return m.deleteItemFn(ctx, id)
	}
	return nil
}
func (m *mockShopRepo) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.transactionFn != nil {
		return m.transactionFn(ctx, fn)
	}
	// Default: execute fn with nil tx (tests just need the business logic)
	return fn(nil)
}
func (m *mockShopRepo) Seed(ctx context.Context) error { return nil }

// ─── Tests ───

func TestShopService_Redeem_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "u1", Points: 100, Role: "user"}, nil
		},
	}
	shopRepo := &mockShopRepo{
		getItemFn: func(ctx context.Context, id string) (*model.ShopItem, error) {
			return &model.ShopItem{
				ID: "i1", Name: "运河文创",
				PointsCost: 50, Stock: 10, IsActive: true,
			}, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	result, err := svc.Redeem(ctx, "u1", "user", "i1")
	if err != nil {
		t.Fatalf("Redeem error: %v", err)
	}
	if result.Redemption.ItemName != "运河文创" {
		t.Errorf("ItemName = %q", result.Redemption.ItemName)
	}
	if result.Redemption.PointsSpent != 50 {
		t.Errorf("PointsSpent = %d, want 50", result.Redemption.PointsSpent)
	}
	if result.UserPoints != 50 {
		t.Errorf("Remaining points = %d, want 50", result.UserPoints)
	}
}

func TestShopService_Redeem_GuestForbidden(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{}
	shopRepo := &mockShopRepo{}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	_, err := svc.Redeem(ctx, "u1", "guest", "i1")
	if err == nil {
		t.Fatal("Expected error for guest role")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrForbidden {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrForbidden)
	}
}

func TestShopService_Redeem_InsufficientPoints(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "u1", Points: 10}, nil
		},
	}
	shopRepo := &mockShopRepo{
		getItemFn: func(ctx context.Context, id string) (*model.ShopItem, error) {
			return &model.ShopItem{
				ID: "i1", Name: "贵", PointsCost: 100, Stock: 5, IsActive: true,
			}, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	_, err := svc.Redeem(ctx, "u1", "user", "i1")
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInsufficientPoints {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrInsufficientPoints)
	}
}

func TestShopService_Redeem_OutOfStock(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "u1", Points: 1000}, nil
		},
	}
	shopRepo := &mockShopRepo{
		getItemFn: func(ctx context.Context, id string) (*model.ShopItem, error) {
			return &model.ShopItem{
				ID: "i1", Name: "断货", PointsCost: 10, Stock: 0, IsActive: true,
			}, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	_, err := svc.Redeem(ctx, "u1", "user", "i1")
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrOutOfStock {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrOutOfStock)
	}
}

func TestShopService_Redeem_InactiveItem(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "u1", Points: 100}, nil
		},
	}
	shopRepo := &mockShopRepo{
		getItemFn: func(ctx context.Context, id string) (*model.ShopItem, error) {
			return &model.ShopItem{
				ID: "i1", Name: "下架", PointsCost: 10, Stock: 5, IsActive: false,
			}, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	_, err := svc.Redeem(ctx, "u1", "user", "i1")
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrItemNotFound {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrItemNotFound)
	}
}

func TestShopService_Redeem_TransactionFailure(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: "u1", Points: 100}, nil
		},
	}
	shopRepo := &mockShopRepo{
		getItemFn: func(ctx context.Context, id string) (*model.ShopItem, error) {
			return &model.ShopItem{
				ID: "i1", Name: "商品", PointsCost: 50, Stock: 5, IsActive: true,
			}, nil
		},
		transactionFn: func(ctx context.Context, fn func(tx *gorm.DB) error) error {
			return errors.New("db connection lost")
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	_, err := svc.Redeem(ctx, "u1", "user", "i1")
	if err == nil {
		t.Fatal("Expected error from failed transaction")
	}
}

func TestShopService_GetHistory_Pagination(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{}
	shopRepo := &mockShopRepo{
		listRedemptionsFn: func(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error) {
			return []model.Redemption{
				{ID: "r1", ItemName: "item1"},
			}, 1, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())

	// Test default pagination (page=0 should become 1, pageSize=0 should become 20)
	result, err := svc.GetHistory(ctx, "u1", 0, 0)
	if err != nil {
		t.Fatalf("GetHistory error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

func TestShopService_GetHistory_PageSizeCap(t *testing.T) {
	ctx := context.Background()
	userRepo := &mockUserRepo{}
	shopRepo := &mockShopRepo{
		listRedemptionsFn: func(ctx context.Context, userID string, offset, limit int) ([]model.Redemption, int64, error) {
			// limit should be capped to 20
			if limit != 20 {
				t.Errorf("Expected limit=20 after capping, got limit=%d", limit)
			}
			return nil, 0, nil
		},
	}
	svc := NewShopService(shopRepo, userRepo, zap.NewNop())
	_, _ = svc.GetHistory(ctx, "u1", 1, 200)
}

func TestShopService_ListItems(t *testing.T) {
	ctx := context.Background()
	shopRepo := &mockShopRepo{
		listItemsFn: func(ctx context.Context) ([]model.ShopItem, error) {
			return []model.ShopItem{
				{ID: "i1", Name: "Item 1"},
				{ID: "i2", Name: "Item 2"},
			}, nil
		},
	}
	svc := NewShopService(shopRepo, &mockUserRepo{}, zap.NewNop())

	items, err := svc.ListItems(ctx)
	if err != nil {
		t.Fatalf("ListItems error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len = %d, want 2", len(items))
	}
}
