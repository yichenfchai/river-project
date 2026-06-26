package service

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type mockQuizRepo struct {
	getQuestionsFn       func(ctx context.Context, count int, difficulty, category string) ([]model.Question, error)
	getQuestionByIDFn    func(ctx context.Context, id string) (*model.Question, error)
	createRecordFn       func(ctx context.Context, record *model.QuizRecord) error
	getUserStatsFn       func(ctx context.Context, userID string) (*repository.QuizUserStats, error)
	getCategoryBreakdownFn func(ctx context.Context, userID string) (map[string]model.CategoryBreakdownItem, error)
	getLastStreakFn      func(ctx context.Context, userID string) (int, error)
	getMaxStreakFn       func(ctx context.Context, userID string) (int, error)
	getLeaderboardFn     func(ctx context.Context, period, userID string, page, pageSize int) ([]repository.LeaderboardRow, int64, *repository.MyRank, error)
	updateUserPointsFn   func(ctx context.Context, userID string, points int, title string) error
	createQuestionFn     func(ctx context.Context, q *model.Question) error
	countDistinctUsersFn func(ctx context.Context) (int64, error)
}

func (m *mockQuizRepo) GetQuestions(ctx context.Context, count int, difficulty, category string) ([]model.Question, error) {
	if m.getQuestionsFn != nil {
		return m.getQuestionsFn(ctx, count, difficulty, category)
	}
	return nil, nil
}
func (m *mockQuizRepo) GetQuestionByID(ctx context.Context, id string) (*model.Question, error) {
	if m.getQuestionByIDFn != nil {
		return m.getQuestionByIDFn(ctx, id)
	}
	return nil, apperrors.NewDefault(apperrors.ErrQuestionNotFound)
}
func (m *mockQuizRepo) CreateRecord(ctx context.Context, record *model.QuizRecord) error {
	if m.createRecordFn != nil {
		return m.createRecordFn(ctx, record)
	}
	return nil
}
func (m *mockQuizRepo) GetUserStats(ctx context.Context, userID string) (*repository.QuizUserStats, error) {
	if m.getUserStatsFn != nil {
		return m.getUserStatsFn(ctx, userID)
	}
	return &repository.QuizUserStats{}, nil
}
func (m *mockQuizRepo) GetCategoryBreakdown(ctx context.Context, userID string) (map[string]model.CategoryBreakdownItem, error) {
	if m.getCategoryBreakdownFn != nil {
		return m.getCategoryBreakdownFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockQuizRepo) GetLastStreak(ctx context.Context, userID string) (int, error) {
	if m.getLastStreakFn != nil {
		return m.getLastStreakFn(ctx, userID)
	}
	return 0, nil
}
func (m *mockQuizRepo) GetMaxStreak(ctx context.Context, userID string) (int, error) {
	if m.getMaxStreakFn != nil {
		return m.getMaxStreakFn(ctx, userID)
	}
	return 0, nil
}
func (m *mockQuizRepo) GetLeaderboard(ctx context.Context, period, userID string, page, pageSize int) ([]repository.LeaderboardRow, int64, *repository.MyRank, error) {
	if m.getLeaderboardFn != nil {
		return m.getLeaderboardFn(ctx, period, userID, page, pageSize)
	}
	return nil, 0, nil, nil
}
func (m *mockQuizRepo) UpdateUserPoints(ctx context.Context, userID string, points int, title string) error {
	if m.updateUserPointsFn != nil {
		return m.updateUserPointsFn(ctx, userID, points, title)
	}
	return nil
}
func (m *mockQuizRepo) CreateQuestion(ctx context.Context, q *model.Question) error {
	if m.createQuestionFn != nil {
		return m.createQuestionFn(ctx, q)
	}
	return nil
}

func (m *mockQuizRepo) CountDistinctUsers(ctx context.Context) (int64, error) {
	if m.countDistinctUsersFn != nil {
		return m.countDistinctUsersFn(ctx)
	}
	return 0, nil
}

func newTestQuizService(quizRepo repository.QuizRepository) QuizService {
	return &quizService{
		repo:     quizRepo,
		userRepo: newTestUserRepo(),
		log:      zap.NewNop(),
	}
}

// ─── Tests ───

func TestQuizService_GetQuestions_Success(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		getQuestionsFn: func(ctx context.Context, count int, difficulty, category string) ([]model.Question, error) {
			return []model.Question{
				{ID: "q1", Content: "问题1", Options: `["A.1","B.2","C.3","D.4"]`, Answer: 2, Difficulty: "easy"},
			}, nil
		},
	}
	svc := newTestQuizService(quizRepo)

	result, err := svc.GetQuestions(ctx, 5, "mixed", "")
	if err != nil {
		t.Fatalf("GetQuestions error: %v", err)
	}
	if result.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if len(result.Questions) != 1 {
		t.Errorf("len = %d, want 1", len(result.Questions))
	}
}

func TestQuizService_SubmitAnswer_Correct(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		getQuestionByIDFn: func(ctx context.Context, id string) (*model.Question, error) {
			return &model.Question{ID: "q1", Content: "问题1", Options: `["A.1","B.2","C.3","D.4"]`, Answer: 2, Difficulty: "easy"}, nil
		},
	}
	svc := &quizService{
		repo:     quizRepo,
		userRepo: newTestUserRepo(),
		log:      zap.NewNop(),
	}
	// pre-store session
	svc.sessions.Store("s1", &quizSession{QuestionIDs: []string{"q1"}})

	result, err := svc.SubmitAnswer(ctx, "u1", model.SubmitAnswerRequest{
		SessionID: "s1", QuestionID: "q1", Answer: "C",
	})
	if err != nil {
		t.Fatalf("SubmitAnswer error: %v", err)
	}
	if !result.IsCorrect {
		t.Error("expected correct answer")
	}
	if result.CorrectAnswer != "C" {
		t.Errorf("CorrectAnswer = %q, want C", result.CorrectAnswer)
	}
}

func TestQuizService_SubmitAnswer_Wrong(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		getQuestionByIDFn: func(ctx context.Context, id string) (*model.Question, error) {
			return &model.Question{ID: "q1", Options: `["A","B","C","D"]`, Answer: 0, Difficulty: "medium"}, nil
		},
	}
	svc := &quizService{
		repo:     quizRepo,
		userRepo: newTestUserRepo(),
		log:      zap.NewNop(),
	}
	svc.sessions.Store("s1", &quizSession{QuestionIDs: []string{"q1"}})

	result, err := svc.SubmitAnswer(ctx, "u1", model.SubmitAnswerRequest{
		SessionID: "s1", QuestionID: "q1", Answer: "B",
	})
	if err != nil {
		t.Fatalf("SubmitAnswer error: %v", err)
	}
	if result.IsCorrect {
		t.Error("expected wrong answer")
	}
	if result.PointsEarned != 0 {
		t.Errorf("PointsEarned = %d, want 0", result.PointsEarned)
	}
}

func TestQuizService_SubmitBatch_Success(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		getQuestionByIDFn: func(ctx context.Context, id string) (*model.Question, error) {
			switch id {
			case "q1":
				return &model.Question{ID: "q1", Content: "问题1", Options: `["A","B","C","D"]`, Answer: 0, Difficulty: "easy"}, nil
			case "q2":
				return &model.Question{ID: "q2", Content: "问题2", Options: `["A","B","C","D"]`, Answer: 3, Difficulty: "hard"}, nil
			default:
				return nil, apperrors.NewDefault(apperrors.ErrQuestionNotFound)
			}
		},
	}
	svc := &quizService{
		repo:     quizRepo,
		userRepo: newTestUserRepo(),
		log:      zap.NewNop(),
	}
	svc.sessions.Store("s1", &quizSession{QuestionIDs: []string{"q1", "q2"}})

	result, err := svc.SubmitBatch(ctx, "u1", "s1", []AnswerItem{
		{QuestionID: "q1", Answer: "A"},
		{QuestionID: "q2", Answer: "D"},
	})
	if err != nil {
		t.Fatalf("SubmitBatch error: %v", err)
	}
	if result.CorrectCount != 2 {
		t.Errorf("CorrectCount = %d, want 2", result.CorrectCount)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
	// session should be cleaned up
	if _, ok := svc.sessions.Load("s1"); ok {
		t.Error("session should be deleted after batch submit")
	}
}

func TestQuizService_GetUserStats(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		getUserStatsFn: func(ctx context.Context, userID string) (*repository.QuizUserStats, error) {
			return &repository.QuizUserStats{TotalAnswers: 50, CorrectAnswers: 45}, nil
		},
	}
	svc := newTestQuizService(quizRepo)

	stats, err := svc.GetUserStats(ctx, "u1")
	if err != nil {
		t.Fatalf("GetUserStats error: %v", err)
	}
	if stats.TotalAnswers != 50 {
		t.Errorf("TotalAnswers = %d, want 50", stats.TotalAnswers)
	}
	if mathAbs(stats.Accuracy-0.90) > 0.01 {
		t.Errorf("Accuracy = %.2f, want 0.90", stats.Accuracy)
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func TestQuizService_CreateQuestion(t *testing.T) {
	ctx := context.Background()
	quizRepo := &mockQuizRepo{
		createQuestionFn: func(ctx context.Context, q *model.Question) error {
			if q.Answer != 0 {
				t.Errorf("Answer index = %d, want 0", q.Answer)
			}
			return nil
		},
	}
	svc := newTestQuizService(quizRepo)

	err := svc.CreateQuestion(ctx, "测试题?", "history", "easy", []string{"A", "B", "C"}, "A", "")
	if err != nil {
		t.Fatalf("CreateQuestion error: %v", err)
	}
}

func TestQuizService_CreateQuestion_InvalidAnswer(t *testing.T) {
	ctx := context.Background()
	svc := newTestQuizService(&mockQuizRepo{})

	err := svc.CreateQuestion(ctx, "测试题?", "history", "easy", []string{"A", "B"}, "D", "")
	if err == nil {
		t.Fatal("expected error for invalid answer letter")
	}
}

func TestRankTitle(t *testing.T) {
	tests := []struct {
		points int
		want   string
	}{
		{0, "青铜守护者"},
		{499, "青铜守护者"},
		{500, "白银守护者"},
		{1499, "白银守护者"},
		{1500, "黄金守护者"},
		{4999, "黄金守护者"},
		{5000, "钻石守护者"},
		{14999, "钻石守护者"},
		{15000, "运河守护者"},
		{99999, "运河守护者"},
	}
	for _, tt := range tests {
		got := rankTitle(tt.points)
		if got != tt.want {
			t.Errorf("rankTitle(%d) = %q, want %q", tt.points, got, tt.want)
		}
	}
}
