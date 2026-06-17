package service

import (
	"context"

	"github.com/grand-canal-guardian/pkg/auth"
	"github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"github.com/grand-canal-guardian/services/main-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务逻辑层 — 只返回 *errors.AppError
type UserService struct {
	repo     *repository.UserRepository
	tokenMgr *auth.TokenManager
}

func NewUserService(repo *repository.UserRepository, tokenMgr *auth.TokenManager) *UserService {
	return &UserService{repo: repo, tokenMgr: tokenMgr}
}

// Register 注册
func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, *errors.AppError) {
	// 检查用户名
	if existing, _ := s.repo.FindByUsername(ctx, req.Username); existing != nil {
		return nil, errors.NewDefault(errors.ErrUsernameExists)
	}

	// 检查邮箱
	if existing, _ := s.repo.FindByEmail(ctx, req.Email); existing != nil {
		return nil, errors.NewDefault(errors.ErrEmailExists)
	}

	// 哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Internal(err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		Nickname:     req.Nickname,
		Role:         "user",
		Status:       "active",
	}
	if user.Nickname == "" {
		user.Nickname = req.Username
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
	}

	return user, nil
}

// Login 登录 → 返回 TokenPair
func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, *errors.AppError) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	if user == nil {
		return nil, errors.NewDefault(errors.ErrPasswordWrong)
	}

	if user.Status == "banned" {
		return nil, errors.NewDefault(errors.ErrUserBanned)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewDefault(errors.ErrPasswordWrong)
	}

	// 签发 Token
	deviceID := req.DeviceID
	if deviceID == "" {
		deviceID = "default"
	}
	tokenPair, appErr := s.tokenMgr.IssueTokens(ctx, user.ID, user.Role, deviceID)
	if appErr != nil {
		return nil, appErr
	}

	// 更新最后登录时间 (非关键路径，忽略错误)
	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	// 隐藏敏感字段
	user.PasswordHash = ""

	return &model.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// RefreshToken 刷新 Token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, *errors.AppError) {
	tokenPair, appErr := s.tokenMgr.RefreshAccess(ctx, refreshToken)
	if appErr != nil {
		return nil, appErr
	}

	// 从 claims 获取用户信息 (需要解析 access token)
	// 简化: 这里直接返回 token，用户信息由前端缓存
	return &model.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout 登出
func (s *UserService) Logout(ctx context.Context, claims *auth.Claims) *errors.AppError {
	if err := s.tokenMgr.Logout(ctx, claims); err != nil {
		return errors.WrapDefault(errors.ErrRedisError, err)
	}
	return nil
}

// GetProfile 获取个人资料
func (s *UserService) GetProfile(ctx context.Context, userID string) (*model.User, *errors.AppError) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	if user == nil {
		return nil, errors.NewDefault(errors.ErrUserNotFound)
	}
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfile 更新资料
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.User, *errors.AppError) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	if user == nil {
		return nil, errors.NewDefault(errors.ErrUserNotFound)
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	user.PasswordHash = ""
	return user, nil
}

// ListUsers 管理员: 用户列表
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, *errors.AppError) {
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	return users, total, nil
}

// BanUser 管理员: 封禁/解封用户
func (s *UserService) BanUser(ctx context.Context, userID string, ban bool) *errors.AppError {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	if user == nil {
		return errors.NewDefault(errors.ErrUserNotFound)
	}

	if ban {
		user.Status = "banned"
	} else {
		user.Status = "active"
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	return nil
}

// ChangeRole 管理员: 修改用户角色
func (s *UserService) ChangeRole(ctx context.Context, userID, newRole string) *errors.AppError {
	if newRole != "user" && newRole != "monitor" && newRole != "admin" {
		return errors.NewDefault(errors.ErrInvalidRole)
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	if user == nil {
		return errors.NewDefault(errors.ErrUserNotFound)
	}

	user.Role = newRole
	if err := s.repo.Update(ctx, user); err != nil {
		return errors.WrapDefault(errors.ErrDatabaseError, err)
	}
	return nil
}
