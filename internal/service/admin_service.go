package service

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type AdminService interface {
	Dashboard(ctx context.Context) (*DashboardStats, error)
	ListUsers(ctx context.Context, page, pageSize int, role, keyword string) (*UserListResult, error)
	UpdateUserRole(ctx context.Context, userID, role string) (*model.User, error)
	BanUser(ctx context.Context, userID string, banned bool) error
	ResetUserPassword(ctx context.Context, userID, newPassword string) error
}

type DashboardStats struct {
	TotalUsers     int64 `json:"total_users"`
	ActiveToday    int64 `json:"active_today"`
	TotalPosts     int64 `json:"total_posts"`
	PendingReviews int64 `json:"pending_reviews"`
	QuizPlayers    int64 `json:"quiz_players"`
	GarbageReports int64 `json:"garbage_reports"`
}

type UserListResult struct {
	Users    []model.User `json:"users"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type adminService struct {
	userRepo repository.UserRepository
	postRepo repository.PostRepository
	quizRepo repository.QuizRepository
	log      *zap.Logger
}

func NewAdminService(userRepo repository.UserRepository, postRepo repository.PostRepository, quizRepo repository.QuizRepository, log *zap.Logger) AdminService {
	return &adminService{userRepo: userRepo, postRepo: postRepo, quizRepo: quizRepo, log: log}
}

func (s *adminService) Dashboard(ctx context.Context) (*DashboardStats, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		s.log.Error("Dashboard: 查询用户总数失败", zap.Error(err))
	}
	activeToday, err := s.userRepo.CountToday(ctx)
	if err != nil {
		s.log.Error("Dashboard: 查询今日活跃失败", zap.Error(err))
	}

	var totalPosts, pendingReviews, quizPlayers int64
	if s.postRepo != nil {
		totalPosts, err = s.postRepo.Count(ctx)
		if err != nil {
			s.log.Error("Dashboard: 查询帖子总数失败", zap.Error(err))
		}
		pendingReviews, err = s.postRepo.CountByStatus(ctx, "pending")
		if err != nil {
			s.log.Error("Dashboard: 查询待审核数失败", zap.Error(err))
		}
	}
	if s.quizRepo != nil {
		quizPlayers, err = s.quizRepo.CountDistinctUsers(ctx)
		if err != nil {
			s.log.Error("Dashboard: 查询答题用户数失败", zap.Error(err))
		}
	}

	return &DashboardStats{
		TotalUsers:     totalUsers,
		ActiveToday:    activeToday,
		TotalPosts:     totalPosts,
		PendingReviews: pendingReviews,
		QuizPlayers:    quizPlayers,
		GarbageReports: 0,
	}, nil
}

func (s *adminService) ListUsers(ctx context.Context, page, pageSize int, role, keyword string) (*UserListResult, error) {
	page, pageSize = normalizePagination(page, pageSize)

	users, total, err := s.userRepo.List(ctx, page, pageSize, role, keyword)
	if err != nil {
		return nil, err
	}

	return &UserListResult{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *adminService) UpdateUserRole(ctx context.Context, userID, role string) (*model.User, error) {
	if role != "user" && role != "monitor" && role != "admin" {
		return nil, apperrors.BadRequest("无效的用户角色")
	}

	if err := s.userRepo.UpdateRole(ctx, userID, role); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.log.Info("用户角色变更", zap.String("user_id", userID), zap.String("role", role))
	return user, nil
}

func (s *adminService) BanUser(ctx context.Context, userID string, banned bool) error {
	status := "active"
	if banned {
		status = "banned"
	}

	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return err
	}

	s.log.Info("用户封禁状态变更", zap.String("user_id", userID), zap.Bool("banned", banned))
	return nil
}

func (s *adminService) ResetUserPassword(ctx context.Context, userID, newPassword string) error {
	if len(newPassword) < 8 {
		return apperrors.BadRequest("新密码至少 8 位")
	}
	if len(newPassword) > 128 {
		return apperrors.BadRequest("新密码过长")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("密码哈希失败", zap.Error(err))
		return apperrors.Internal(err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	s.log.Info("管理员重置用户密码", zap.String("user_id", userID))
	return nil
}
