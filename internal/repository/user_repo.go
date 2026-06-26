package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// UserRepository 用户数据访问接口 — 上层只依赖此接口，不感知 GORM
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	List(ctx context.Context, page, pageSize int, role, keyword string) ([]model.User, int64, error)
	Count(ctx context.Context) (int64, error)
	CountToday(ctx context.Context) (int64, error)
	UpdateRole(ctx context.Context, id, role string) error
	UpdateStatus(ctx context.Context, id, status string) error
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

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *userRepo) List(ctx context.Context, page, pageSize int, role, keyword string) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{})

	if role != "" {
		q = q.Where("role = ?", role)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	offset := (page - 1) * pageSize
	var users []model.User
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return users, total, nil
}

func (r *userRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func (r *userRepo) CountToday(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("created_at >= CURRENT_DATE").Count(&count).Error; err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func (r *userRepo) UpdateRole(ctx context.Context, id, role string) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("role", role)
	if res.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrUserNotFound)
	}
	return nil
}

func (r *userRepo) UpdateStatus(ctx context.Context, id, status string) error {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.NewDefault(apperrors.ErrUserNotFound)
	}
	return nil
}
