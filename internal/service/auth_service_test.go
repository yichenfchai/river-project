package service

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/pkg/auth"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// ─── Mock Implementations ───

type mockUserRepo struct {
	users           map[string]*model.User
	findByUsername  func(ctx context.Context, username string) (*model.User, error)
	existsByUsername func(ctx context.Context, username string) (bool, error)
	existsByEmail   func(ctx context.Context, email string) (bool, error)
	create          func(ctx context.Context, user *model.User) error
	findByID        func(ctx context.Context, id string) (*model.User, error)
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.findByUsername != nil {
		return m.findByUsername(ctx, username)
	}
	return nil, nil
}
func (m *mockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.existsByUsername != nil {
		return m.existsByUsername(ctx, username)
	}
	return false, nil
}
func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.existsByEmail != nil {
		return m.existsByEmail(ctx, email)
	}
	return false, nil
}
func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.create != nil {
		return m.create(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	if m.findByID != nil {
		return m.findByID(ctx, id)
	}
	return nil, nil
}

type mockTokenManager struct {
	issueTokens func(userID, role, deviceID string) (*auth.TokenPair, error)
}

func (m *mockTokenManager) IssueTokens(userID, role, deviceID string) (*auth.TokenPair, error) {
	if m.issueTokens != nil {
		return m.issueTokens(userID, role, deviceID)
	}
	return &auth.TokenPair{
		AccessToken:  "mock-access",
		RefreshToken: "mock-refresh",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

var testLogger = zap.NewNop()

// ─── Tests ───

func TestAuthService_Register_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockUserRepo{
		existsByUsername: func(ctx context.Context, username string) (bool, error) { return false, nil },
		existsByEmail:    func(ctx context.Context, email string) (bool, error) { return false, nil },
		create:           func(ctx context.Context, user *model.User) error { return nil },
	}
	tm := &mockTokenManager{}
	svc := NewAuthService(repo, nil, testLogger)
	// override with our tm
	svc = &authService{repo: repo, tm: tm, log: testLogger}

	result, err := svc.Register(ctx, RegisterInput{
		Username: "newuser",
		Password: "password123",
		Email:    "new@test.com",
		Nickname: "New User",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if result.User.Username != "newuser" {
		t.Errorf("Username = %q, want %q", result.User.Username, "newuser")
	}
	if result.User.Role != "user" {
		t.Errorf("Role = %q, want %q", result.User.Role, "user")
	}
	if result.AccessToken != "mock-access" {
		t.Errorf("AccessToken = %q", result.AccessToken)
	}
	// Password should be hashed (not plaintext)
	if result.User.Password == "password123" {
		t.Error("Password was stored as plaintext!")
	}
	// Verify hash
	if err := bcrypt.CompareHashAndPassword([]byte(result.User.Password), []byte("password123")); err != nil {
		t.Error("Password hash does not match original password")
	}
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	repo := &mockUserRepo{
		existsByUsername: func(ctx context.Context, username string) (bool, error) { return true, nil },
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	_, err := svc.Register(ctx, RegisterInput{
		Username: "existing",
		Password: "password",
		Email:    "e@test.com",
	})

	if err == nil {
		t.Fatal("Expected error for duplicate username")
	}
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrUsernameExists {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrUsernameExists)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := &mockUserRepo{
		existsByUsername: func(ctx context.Context, username string) (bool, error) { return false, nil },
		existsByEmail:    func(ctx context.Context, email string) (bool, error) { return true, nil },
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	_, err := svc.Register(ctx, RegisterInput{
		Username: "newuser",
		Password: "password",
		Email:    "taken@test.com",
	})

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrEmailExists {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrEmailExists)
	}
}

func TestAuthService_Register_DefaultNickname(t *testing.T) {
	ctx := context.Background()
	var createdUser *model.User
	repo := &mockUserRepo{
		existsByUsername: func(ctx context.Context, u string) (bool, error) { return false, nil },
		existsByEmail:    func(ctx context.Context, e string) (bool, error) { return false, nil },
		create: func(ctx context.Context, user *model.User) error {
			createdUser = user
			return nil
		},
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	result, err := svc.Register(ctx, RegisterInput{
		Username: "nonick",
		Password: "password",
		Email:    "a@b.com",
		Nickname: "",
	})
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if result.User.Nickname != "nonick" {
		t.Errorf("Nickname = %q, want %q (fallback to username)", result.User.Nickname, "nonick")
	}
	if createdUser.Nickname != "nonick" {
		t.Errorf("Stored nickname = %q, want %q", createdUser.Nickname, "nonick")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()
	// Pre-create a user with hashed password
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:       "user-1",
		Username: "testuser",
		Password: string(hash),
		Role:     "user",
	}

	repo := &mockUserRepo{
		findByUsername: func(ctx context.Context, username string) (*model.User, error) {
			return testUser, nil
		},
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	result, err := svc.Login(ctx, LoginInput{
		Username: "testuser",
		Password: "correctpass",
	})
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}
	if result.User.ID != "user-1" {
		t.Errorf("User ID = %q", result.User.ID)
	}
	if result.AccessToken != "mock-access" {
		t.Errorf("AccessToken = %q", result.AccessToken)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	testUser := &model.User{
		ID:       "u1",
		Username: "u",
		Password: string(hash),
		Role:     "user",
	}
	repo := &mockUserRepo{
		findByUsername: func(ctx context.Context, u string) (*model.User, error) { return testUser, nil },
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	_, err := svc.Login(ctx, LoginInput{Username: "u", Password: "wrong"})

	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrPasswordWrong {
		t.Errorf("Code = %d, want %d", appErr.Code, apperrors.ErrPasswordWrong)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockUserRepo{
		findByUsername: func(ctx context.Context, u string) (*model.User, error) {
			return nil, apperrors.NewDefault(apperrors.ErrUserNotFound)
		},
	}
	tm := &mockTokenManager{}
	svc := &authService{repo: repo, tm: tm, log: testLogger}

	_, err := svc.Login(ctx, LoginInput{Username: "ghost", Password: "x"})

	// Security: should return "password wrong" not "user not found"
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("Expected *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrPasswordWrong {
		t.Errorf("Code = %d, want %d (security: hide existence check)", appErr.Code, apperrors.ErrPasswordWrong)
	}
}

func TestAuthService_GetProfile(t *testing.T) {
	ctx := context.Background()
	expectedUser := &model.User{ID: "u42", Username: "alice"}
	repo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			if id == "u42" {
				return expectedUser, nil
			}
			return nil, apperrors.NewDefault(apperrors.ErrUserNotFound)
		},
	}
	svc := &authService{repo: repo, tm: &mockTokenManager{}, log: testLogger}

	user, err := svc.GetProfile(ctx, "u42")
	if err != nil {
		t.Fatalf("GetProfile error: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q", user.Username)
	}
}

func TestAuthService_GetProfile_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockUserRepo{
		findByID: func(ctx context.Context, id string) (*model.User, error) {
			return nil, apperrors.NewDefault(apperrors.ErrUserNotFound)
		},
	}
	svc := &authService{repo: repo, tm: &mockTokenManager{}, log: testLogger}

	_, err := svc.GetProfile(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent user")
	}
}
