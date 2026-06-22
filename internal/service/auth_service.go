package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/your-org/grand-canal-guardian/internal/model"
	"github.com/your-org/grand-canal-guardian/internal/repository"
	"github.com/your-org/grand-canal-guardian/pkg/auth"
	apperrors "github.com/your-org/grand-canal-guardian/pkg/errors"
)

// AuthService 认证业务逻辑接口
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	GetProfile(ctx context.Context, userID string) (*model.User, error)
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
}

type TokenManager interface {
	IssueTokens(userID, role, deviceID string) (*auth.TokenPair, error)
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
