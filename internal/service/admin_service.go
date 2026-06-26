package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type AdminService interface {
	Dashboard(ctx context.Context) (*DashboardStats, error)
	ListUsers(ctx context.Context, page, pageSize int, role, keyword string) (*UserListResult, error)
	UpdateUserRole(ctx context.Context, userID, role string) (*model.User, error)
	BanUser(ctx context.Context, userID string, banned bool) error
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

func (s *adminService) normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func (s *adminService) Dashboard(ctx context.Context) (*DashboardStats, error) {
	totalUsers, _ := s.userRepo.Count(ctx)
	activeToday, _ := s.userRepo.CountToday(ctx)

	var totalPosts, pendingReviews, quizPlayers int64
	if s.postRepo != nil {
		totalPosts, _ = s.postRepo.Count(ctx)
		pendingReviews, _ = s.postRepo.CountByStatus(ctx, "pending")
	}
	if s.quizRepo != nil {
		quizPlayers, _ = s.quizRepo.CountDistinctUsers(ctx)
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
	page, pageSize = s.normalizePagination(page, pageSize)

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
