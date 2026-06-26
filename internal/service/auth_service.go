package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	"github.com/yichenfchai/river-project/pkg/auth"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// AuthService 认证业务逻辑接口
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	GetProfile(ctx context.Context, userID string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*model.User, error)
	Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	Logout(ctx context.Context, userID, jti string, expiresAt time.Time) error
}

type RegisterInput struct {
	Username string
	Password string
	Email    string
	Nickname string
}

type LoginInput struct {
	Username string
	Password string
}

type UpdateProfileInput struct {
	Nickname  *string
	Bio       *string
	AvatarURL *string
}

type AuthResult struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int        `json:"expires_in"`
}

type authService struct {
	repo UserRepository
	tm   TokenManager
	log  *zap.Logger
}

// 接口分离：Service 层依赖最小接口
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type TokenManager interface {
	IssueTokens(userID, role, deviceID string) (*auth.TokenPair, error)
	ParseToken(tokenStr string) (*auth.Claims, error)
	BlacklistJTI(jti string, expiresAt time.Time)
}

var _ UserRepository = (repository.UserRepository)(nil)

func NewAuthService(repo repository.UserRepository, tm *auth.TokenManager, log *zap.Logger) AuthService {
	return &authService{repo: repo, tm: tm, log: log}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	exists, err := s.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.NewDefault(apperrors.ErrUsernameExists)
	}

	exists, err = s.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.NewDefault(apperrors.ErrEmailExists)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("密码哈希失败", zap.Error(err))
		return nil, apperrors.Internal(err)
	}

	nickname := input.Nickname
	if nickname == "" {
		nickname = input.Username
	}

	user := model.User{
		ID:       uuid.New().String(),
		Username: input.Username,
		Password: string(hash),
		Email:    input.Email,
		Nickname: nickname,
		Role:     "user",
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		s.log.Error("创建用户失败", zap.Error(err))
		return nil, err
	}

	tokens, err := s.tm.IssueTokens(user.ID, user.Role, "")
	if err != nil {
		s.log.Error("签发Token失败", zap.Error(err))
		return nil, apperrors.Internal(err)
	}

	s.log.Info("注册成功", zap.String("user_id", user.ID))
	return &AuthResult{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.ErrUserNotFound {
			s.log.Warn("登录失败-用户不存在", zap.String("username", input.Username))
			return nil, apperrors.NewDefault(apperrors.ErrPasswordWrong)
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		s.log.Warn("登录失败-密码错误", zap.String("username", input.Username))
		return nil, apperrors.NewDefault(apperrors.ErrPasswordWrong)
	}

	tokens, err := s.tm.IssueTokens(user.ID, user.Role, "")
	if err != nil {
		s.log.Error("签发Token失败", zap.Error(err))
		return nil, apperrors.Internal(err)
	}

	s.log.Info("登录成功", zap.String("user_id", user.ID), zap.String("role", user.Role))
	return &AuthResult{
		User:         *user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *authService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Nickname != nil {
		if len(*input.Nickname) > 128 {
			return nil, apperrors.BadRequest("昵称不能超过128字")
		}
		user.Nickname = *input.Nickname
	}
	if input.Bio != nil {
		if len(*input.Bio) > 500 {
			return nil, apperrors.BadRequest("简介不能超过500字")
		}
		user.Bio = *input.Bio
	}
	if input.AvatarURL != nil {
		if len(*input.AvatarURL) > 512 {
			return nil, apperrors.BadRequest("头像URL过长")
		}
		user.AvatarURL = *input.AvatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	s.log.Info("用户资料更新", zap.String("user_id", userID))
	return user, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	oldClaims, err := s.tm.ParseToken(refreshToken)
	if err != nil {
		s.log.Warn("Refresh失败-Token无效", zap.Error(err))
		return nil, apperrors.NewDefault(apperrors.ErrTokenInvalid)
	}

	user, err := s.repo.FindByID(ctx, oldClaims.Subject)
	if err != nil {
		s.log.Warn("Refresh失败-用户不存在", zap.String("user_id", oldClaims.Subject))
		return nil, err
	}

	// 以数据库最新角色签发新 Token
	tokens, err := s.tm.IssueTokens(user.ID, user.Role, "")
	if err != nil {
		s.log.Error("签发Token失败", zap.Error(err))
		return nil, apperrors.Internal(err)
	}

	// 令牌旋转：旧 refresh token 立即失效
	if oldClaims.ExpiresAt != nil {
		s.tm.BlacklistJTI(oldClaims.ID, oldClaims.ExpiresAt.Time)
	}

	s.log.Info("Token刷新成功", zap.String("user_id", user.ID))
	return tokens, nil
}

func (s *authService) Logout(ctx context.Context, userID, jti string, expiresAt time.Time) error {
	if jti != "" {
		s.tm.BlacklistJTI(jti, expiresAt)
	}
	s.log.Info("用户登出", zap.String("user_id", userID), zap.String("jti", jti))
	return nil
}
