package service

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// ─── Mock Post Repository ───

type mockPostRepo struct {
	createFn          func(ctx context.Context, post *model.Post) error
	findByIDFn        func(ctx context.Context, id string) (*model.Post, error)
	listFn            func(ctx context.Context, opts repository.PostListOptions) ([]model.Post, int64, error)
	updateFn          func(ctx context.Context, post *model.Post) error
	softDeleteFn      func(ctx context.Context, id string) error
	toggleLikeFn      func(ctx context.Context, postID, userID string) (bool, int, error)
	isLikedByUserFn   func(ctx context.Context, postID, userID string) (bool, error)
	createCommentFn   func(ctx context.Context, comment *model.Comment) error
	listCommentsFn    func(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error)
	softDeleteCommentFn func(ctx context.Context, id string) error
	updateStatusFn    func(ctx context.Context, id, status string) error
	findUsersByIDsFn  func(ctx context.Context, ids []string) (map[string]model.UserJSON, error)
	countFn           func(ctx context.Context) (int64, error)
	countByStatusFn   func(ctx context.Context, status string) (int64, error)
}

func (m *mockPostRepo) Create(ctx context.Context, post *model.Post) error {
	if m.createFn != nil {
		return m.createFn(ctx, post)
	}
	return nil
}
func (m *mockPostRepo) FindByID(ctx context.Context, id string) (*model.Post, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, apperrors.NewDefault(apperrors.ErrPostNotFound)
}
func (m *mockPostRepo) List(ctx context.Context, opts repository.PostListOptions) ([]model.Post, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, opts)
	}
	return nil, 0, nil
}
func (m *mockPostRepo) Update(ctx context.Context, post *model.Post) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, post)
	}
	return nil
}
func (m *mockPostRepo) SoftDelete(ctx context.Context, id string) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, id)
	}
	return nil
}
func (m *mockPostRepo) ToggleLike(ctx context.Context, postID, userID string) (bool, int, error) {
	if m.toggleLikeFn != nil {
		return m.toggleLikeFn(ctx, postID, userID)
	}
	return false, 0, nil
}
func (m *mockPostRepo) IsLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	if m.isLikedByUserFn != nil {
		return m.isLikedByUserFn(ctx, postID, userID)
	}
	return false, nil
}
func (m *mockPostRepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	if m.createCommentFn != nil {
		return m.createCommentFn(ctx, comment)
	}
	return nil
}
func (m *mockPostRepo) ListComments(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error) {
	if m.listCommentsFn != nil {
		return m.listCommentsFn(ctx, postID, offset, limit)
	}
	return nil, 0, nil
}
func (m *mockPostRepo) SoftDeleteComment(ctx context.Context, id string) error {
	if m.softDeleteCommentFn != nil {
		return m.softDeleteCommentFn(ctx, id)
	}
	return nil
}
func (m *mockPostRepo) UpdateStatus(ctx context.Context, id, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *mockPostRepo) FindUsersByIDs(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
	if m.findUsersByIDsFn != nil {
		return m.findUsersByIDsFn(ctx, ids)
	}
	return map[string]model.UserJSON{}, nil
}
func (m *mockPostRepo) Count(ctx context.Context) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 0, nil
}
func (m *mockPostRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, status)
	}
	return 0, nil
}

// ─── Shared test user repo ───

func newTestUserRepo() *mockUserRepo {
	return &mockUserRepo{}
}

func newTestPostRepo() *mockPostRepo {
	return &mockPostRepo{}
}

func newTestPostService(postRepo repository.PostRepository, userRepo *mockUserRepo) PostService {
	return NewPostService(postRepo, userRepo, zap.NewNop())
}

// ─── Tests: Create ───

func TestPostService_Create_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		createFn: func(ctx context.Context, post *model.Post) error {
			if post.ID == "" {
				t.Error("expected post ID to be set")
			}
			if post.Status != "pending" {
				t.Errorf("expected status pending, got %s", post.Status)
			}
			return nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{
				"u1": {ID: "u1", Username: "test", Nickname: "测试用户"},
			}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.Create(ctx, "u1", "标题", "内容", "share", nil, nil)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.Title != "标题" {
		t.Errorf("Title = %q, want %q", result.Title, "标题")
	}
	if result.Status != "pending" {
		t.Errorf("Status = %q, want pending", result.Status)
	}
	if result.Author.Username != "test" {
		t.Errorf("Author.Username = %q", result.Author.Username)
	}
}

func TestPostService_Create_EmptyTitle(t *testing.T) {
	ctx := context.Background()
	svc := newTestPostService(newTestPostRepo(), newTestUserRepo())

	_, err := svc.Create(ctx, "u1", "", "内容", "share", nil, nil)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestPostService_Create_EmptyContent(t *testing.T) {
	ctx := context.Background()
	svc := newTestPostService(newTestPostRepo(), newTestUserRepo())

	_, err := svc.Create(ctx, "u1", "标题", "", "share", nil, nil)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestPostService_Create_TitleTooLong(t *testing.T) {
	ctx := context.Background()
	svc := newTestPostService(newTestPostRepo(), newTestUserRepo())

	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}
	_, err := svc.Create(ctx, "u1", longTitle, "内容", "share", nil, nil)
	if err == nil {
		t.Fatal("expected error for title > 200 chars")
	}
}

// ─── Tests: GetByID ───

func TestPostService_GetByID_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{
				ID: "p1", AuthorID: "u1", Title: "标题", Content: "内容",
				Topic: "share", Status: "approved", Images: "[]", Tags: "[]",
			}, nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{
				"u1": {ID: "u1", Username: "test", Nickname: "测试"},
			}, nil
		},
		isLikedByUserFn: func(ctx context.Context, postID, userID string) (bool, error) {
			return true, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.GetByID(ctx, "p1", "current_user")
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if result.IsLiked != true {
		t.Error("expected is_liked = true")
	}
}

func TestPostService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestPostService(newTestPostRepo(), newTestUserRepo())

	_, err := svc.GetByID(ctx, "nonexistent", "")
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// ─── Tests: Update ───

func TestPostService_Update_NotOwner(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{ID: "p1", AuthorID: "other_user"}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	_, err := svc.Update(ctx, "p1", "u1", "新标题", "新内容", nil)
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != apperrors.ErrPostNotOwner {
		t.Fatalf("expected ErrPostNotOwner, got %v", err)
	}
}

func TestPostService_Update_Success(t *testing.T) {
	ctx := context.Background()
	post := &model.Post{ID: "p1", AuthorID: "u1", Title: "旧", Content: "旧", Images: "[]", Tags: "[]", Status: "approved"}
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return post, nil
		},
		updateFn: func(ctx context.Context, p *model.Post) error {
			post.Title = p.Title
			post.Content = p.Content
			return nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{
				"u1": {ID: "u1", Username: "test"},
			}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.Update(ctx, "p1", "u1", "新标题", "新内容", nil)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Title != "新标题" {
		t.Errorf("Title = %q, want %q", result.Title, "新标题")
	}
}

// ─── Tests: Delete ───

func TestPostService_Delete_NotOwner(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{ID: "p1", AuthorID: "other"}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	err := svc.Delete(ctx, "p1", "u1")
	appErr, ok := err.(*apperrors.AppError)
	if !ok || appErr.Code != apperrors.ErrPostNotOwner {
		t.Fatalf("expected ErrPostNotOwner, got %v", err)
	}
}

// ─── Tests: ToggleLike ───

func TestPostService_ToggleLike_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{ID: "p1", AuthorID: "u1"}, nil
		},
		toggleLikeFn: func(ctx context.Context, postID, userID string) (bool, int, error) {
			return true, 1, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.ToggleLike(ctx, "p1", "u1")
	if err != nil {
		t.Fatalf("ToggleLike error: %v", err)
	}
	if !result.IsLiked {
		t.Error("expected is_liked = true")
	}
	if result.LikeCount != 1 {
		t.Errorf("LikeCount = %d, want 1", result.LikeCount)
	}
}

// ─── Tests: CreateComment ───

func TestPostService_CreateComment_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{ID: "p1", AuthorID: "u1"}, nil
		},
		createCommentFn: func(ctx context.Context, comment *model.Comment) error {
			if comment.ID == "" {
				t.Error("expected comment ID to be set")
			}
			return nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{
				"u1": {ID: "u1", Username: "test"},
			}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.CreateComment(ctx, "p1", "u1", "评论内容", nil)
	if err != nil {
		t.Fatalf("CreateComment error: %v", err)
	}
	if result.Content != "评论内容" {
		t.Errorf("Content = %q", result.Content)
	}
}

func TestPostService_CreateComment_EmptyContent(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return &model.Post{ID: "p1"}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	_, err := svc.CreateComment(ctx, "p1", "u1", "", nil)
	if err == nil {
		t.Fatal("expected error for empty comment")
	}
}

// ─── Tests: List (pagination) ───

func TestPostService_List_PaginationDefaults(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		listFn: func(ctx context.Context, opts repository.PostListOptions) ([]model.Post, int64, error) {
			if opts.Page != 1 {
				t.Errorf("expected page=1, got %d", opts.Page)
			}
			if opts.PageSize != 20 {
				t.Errorf("expected pageSize=20, got %d", opts.PageSize)
			}
			if opts.Status != "approved" {
				t.Errorf("expected status=approved, got %s", opts.Status)
			}
			return nil, 0, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.List(ctx, PostListQuery{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}
}

// ─── Tests: Admin Review ───

func TestPostService_Review_Approve(t *testing.T) {
	ctx := context.Background()
	post := &model.Post{ID: "p1", AuthorID: "u1", Title: "待审核", Content: "内容", Status: "pending"}
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return post, nil
		},
		updateStatusFn: func(ctx context.Context, id, status string) error {
			post.Status = status
			return nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{"u1": {ID: "u1", Username: "test"}}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.Review(ctx, "p1", "approve", "")
	if err != nil {
		t.Fatalf("Review error: %v", err)
	}
	if result.Status != "approved" {
		t.Errorf("Status = %q, want approved", result.Status)
	}
}

func TestPostService_Review_Reject(t *testing.T) {
	ctx := context.Background()
	post := &model.Post{ID: "p1", AuthorID: "u1", Status: "pending"}
	postRepo := &mockPostRepo{
		findByIDFn: func(ctx context.Context, id string) (*model.Post, error) {
			return post, nil
		},
		updateStatusFn: func(ctx context.Context, id, status string) error {
			post.Status = status
			return nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{"u1": {ID: "u1", Username: "test"}}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.Review(ctx, "p1", "reject", "违规内容")
	if err != nil {
		t.Fatalf("Review error: %v", err)
	}
	if result.Status != "rejected" {
		t.Errorf("Status = %q, want rejected", result.Status)
	}
}

func TestPostService_Review_InvalidAction(t *testing.T) {
	ctx := context.Background()
	svc := newTestPostService(newTestPostRepo(), newTestUserRepo())

	_, err := svc.Review(ctx, "p1", "invalid", "")
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

// ─── Tests: ListPending ───

func TestPostService_ListPending_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		listFn: func(ctx context.Context, opts repository.PostListOptions) ([]model.Post, int64, error) {
			if opts.Status != "pending" {
				t.Errorf("expected status=pending, got %s", opts.Status)
			}
			return []model.Post{
				{ID: "p1", AuthorID: "u1", Title: "待审核", Status: "pending"},
			}, 1, nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{"u1": {ID: "u1", Username: "test"}}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.ListPending(ctx, 1, 20)
	if err != nil {
		t.Fatalf("ListPending error: %v", err)
	}
	if len(result.Posts) != 1 {
		t.Errorf("len = %d, want 1", len(result.Posts))
	}
}

// ─── Tests: ListComments ───

func TestPostService_ListComments_Success(t *testing.T) {
	ctx := context.Background()
	postRepo := &mockPostRepo{
		listCommentsFn: func(ctx context.Context, postID string, offset, limit int) ([]model.Comment, int64, error) {
			return []model.Comment{
				{ID: "c1", PostID: "p1", AuthorID: "u1", Content: "评论1"},
			}, 1, nil
		},
		findUsersByIDsFn: func(ctx context.Context, ids []string) (map[string]model.UserJSON, error) {
			return map[string]model.UserJSON{"u1": {ID: "u1", Username: "test"}}, nil
		},
	}
	svc := newTestPostService(postRepo, newTestUserRepo())

	result, err := svc.ListComments(ctx, "p1", 1, 20)
	if err != nil {
		t.Fatalf("ListComments error: %v", err)
	}
	if len(result.Comments) != 1 {
		t.Errorf("len = %d, want 1", len(result.Comments))
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}
