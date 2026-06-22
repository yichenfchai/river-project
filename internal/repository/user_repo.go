package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/your-org/grand-canal-guardian/internal/model"
	apperrors "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// UserRepository 用户数据访问接口 — 上层只依赖此接口，不感知 GORM
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *model.User) error
}

// userRepo GORM 实现
type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrUserNotFound)
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseError, "查询用户失败", err)
	}
	return &u, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrUserNotFound)
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseError, "查询用户失败", err)
	}
	return &u, nil
}

func (r *userRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, apperrors.Wrap(apperrors.ErrDatabaseError, "查询用户名失败", err)
	}
	return count > 0, nil
}

func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, apperrors.Wrap(apperrors.ErrDatabaseError, "查询邮箱失败", err)
	}
	return count > 0, nil
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	if user.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseError, "创建用户失败", err)
	}
	return nil
}
