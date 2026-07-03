package service

import (
	"context"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
)

// StoryService 科普故事业务层接口
type StoryService interface {
	List(ctx context.Context, page, pageSize int) ([]model.Story, int64, error)
	GetByID(ctx context.Context, id string) (*model.Story, error)
}

type storyService struct {
	repo repository.StoryRepository
}

func NewStoryService(repo repository.StoryRepository) StoryService {
	return &storyService{repo: repo}
}

func (s *storyService) List(ctx context.Context, page, pageSize int) ([]model.Story, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

func (s *storyService) GetByID(ctx context.Context, id string) (*model.Story, error) {
	return s.repo.FindByID(ctx, id)
}
