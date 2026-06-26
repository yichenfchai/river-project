package service

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/repository"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type QuizService interface {
	GetQuestions(ctx context.Context, count int, difficulty, category string) (*model.SessionResponse, error)
	SubmitAnswer(ctx context.Context, userID string, req model.SubmitAnswerRequest) (*model.SubmitAnswerResponse, error)
	SubmitBatch(ctx context.Context, userID string, sessionID string, answers []AnswerItem) (*model.BatchSubmitResult, error)
	GetUserStats(ctx context.Context, userID string) (*model.UserStats, error)
	GetLeaderboard(ctx context.Context, period, currentUserID string, page, pageSize int) (*model.LeaderboardResponse, error)
	CreateQuestion(ctx context.Context, question, category, difficulty string, options []string, answerStr string, explanation string) error
}

type AnswerItem struct {
	QuestionID   string `json:"question_id"`
	Answer       string `json:"answer"`
	AnswerTimeMs int    `json:"answer_time_ms"`
}

type quizSession struct {
	QuestionIDs []string
	CreatedAt   time.Time
}

type quizService struct {
	repo     repository.QuizRepository
	userRepo repository.UserRepository
	sessions sync.Map
	log      *zap.Logger
}

func NewQuizService(repo repository.QuizRepository, userRepo repository.UserRepository, log *zap.Logger) QuizService {
	s := &quizService{repo: repo, userRepo: userRepo, log: log}
	go s.sessionCleanup(10 * time.Minute)
	return s
}

func (s *quizService) sessionCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.sessions.Range(func(key, val interface{}) bool {
			ses, ok := val.(*quizSession)
			if !ok || now.Sub(ses.CreatedAt) > 2*time.Hour {
				s.sessions.Delete(key)
			}
			return true
		})
	}
}

func (s *quizService) normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func difficultyMultiplier(d string) int {
	switch d {
	case "hard":
		return 30
	case "medium":
		return 20
	default:
		return 10
	}
}

func answerLetterToIndex(letter string) int {
	if len(letter) == 0 {
		return -1
	}
	c := letter[0]
	if c >= 'A' && c <= 'Z' {
		return int(c - 'A')
	}
	if c >= 'a' && c <= 'z' {
		return int(c - 'a')
	}
	return -1
}

func indexToLetter(idx int) string {
	if idx >= 0 && idx < 26 {
		return string(rune('A' + idx))
	}
	return "?"
}

func rankTitle(points int) string {
	switch {
	case points >= 15000:
		return "运河守护者"
	case points >= 5000:
		return "钻石守护者"
	case points >= 1500:
		return "黄金守护者"
	case points >= 500:
		return "白银守护者"
	default:
		return "青铜守护者"
	}
}

func toQuestionResponse(q model.Question) model.QuestionResponse {
	opts := parseOptionsStr(q.Options)
	if opts == nil {
		opts = []string{}
	}
	return model.QuestionResponse{
		ID:         q.ID,
		Type:       "choice",
		Difficulty: q.Difficulty,
		Category:   q.Category,
		Question:   q.Content,
		Options:    opts,
	}
}

func (s *quizService) GetQuestions(ctx context.Context, count int, difficulty, category string) (*model.SessionResponse, error) {
	if count < 1 {
		count = 10
	}
	if count > 50 {
		count = 50
	}

	questions, err := s.repo.GetQuestions(ctx, count, difficulty, category)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New().String()
	ids := make([]string, len(questions))
	responses := make([]model.QuestionResponse, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
		responses[i] = toQuestionResponse(q)
	}
	if responses == nil {
		responses = []model.QuestionResponse{}
	}

	s.sessions.Store(sessionID, &quizSession{
		QuestionIDs: ids,
		CreatedAt:   time.Now(),
	})

	return &model.SessionResponse{
		SessionID: sessionID,
		Questions: responses,
	}, nil
}

func (s *quizService) SubmitAnswer(ctx context.Context, userID string, req model.SubmitAnswerRequest) (*model.SubmitAnswerResponse, error) {
	if _, ok := s.sessions.Load(req.SessionID); !ok {
		return nil, apperrors.New(apperrors.ErrQuestionNotFound, "会话已过期，请重新获取题目")
	}

	question, err := s.repo.GetQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return nil, err
	}

	userAnswer := answerLetterToIndex(req.Answer)
	correct := userAnswer == question.Answer

	mult := difficultyMultiplier(question.Difficulty)
	pointsEarned := 0
	streakBonus := 0
	currentStreak := 0

	if correct {
		pointsEarned = mult
		streak, _ := s.repo.GetLastStreak(ctx, userID)
		currentStreak = streak + 1
		streakBonus = int(math.Min(float64(currentStreak-1), 5.0)) * 5
	} else {
		currentStreak = 0
	}

	record := &model.QuizRecord{
		ID:         uuid.New().String(),
		UserID:     userID,
		QuestionID: req.QuestionID,
		Correct:    correct,
		Score:      pointsEarned + streakBonus,
		Streak:     currentStreak,
	}

	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}

	totalAdded := pointsEarned + streakBonus

	user, _ := s.userRepo.FindByID(ctx, userID)
	newTotal := 0
	if user != nil {
		newTotal = user.Points + totalAdded
	}

	title := rankTitle(newTotal)
	_ = s.repo.UpdateUserPoints(ctx, userID, totalAdded, title)

	correctAnswerLetter := indexToLetter(question.Answer)

	return &model.SubmitAnswerResponse{
		IsCorrect:     correct,
		CorrectAnswer: correctAnswerLetter,
		Explanation:   "",
		PointsEarned:  pointsEarned,
		StreakBonus:   streakBonus,
		TotalPoints:   newTotal,
	}, nil
}

func (s *quizService) SubmitBatch(ctx context.Context, userID string, sessionID string, answers []AnswerItem) (*model.BatchSubmitResult, error) {
	if _, ok := s.sessions.Load(sessionID); !ok {
		return nil, apperrors.New(apperrors.ErrQuestionNotFound, "会话已过期，请重新获取题目")
	}

	results := make([]model.SubmitAnswerResponse, 0, len(answers))
	correctCount := 0
	totalPointsEarned := 0

	for _, a := range answers {
		question, err := s.repo.GetQuestionByID(ctx, a.QuestionID)
		if err != nil {
			continue
		}

		userAnswer := answerLetterToIndex(a.Answer)
		correct := userAnswer == question.Answer
		mult := difficultyMultiplier(question.Difficulty)
		pointsEarned := 0
		streakBonus := 0
		currentStreak := 0

		if correct {
			correctCount++
			pointsEarned = mult
			streak, _ := s.repo.GetLastStreak(ctx, userID)
			currentStreak = streak + 1
			streakBonus = int(math.Min(float64(currentStreak-1), 5.0)) * 5
		}

		record := &model.QuizRecord{
			ID:         uuid.New().String(),
			UserID:     userID,
			QuestionID: a.QuestionID,
			Correct:    correct,
			Score:      pointsEarned + streakBonus,
			Streak:     currentStreak,
		}
		_ = s.repo.CreateRecord(ctx, record)

		totalAdded := pointsEarned + streakBonus
		totalPointsEarned += totalAdded

		correctAnswerLetter := indexToLetter(question.Answer)
		results = append(results, model.SubmitAnswerResponse{
			IsCorrect:     correct,
			CorrectAnswer: correctAnswerLetter,
			Explanation:   "",
			PointsEarned:  pointsEarned,
			StreakBonus:   streakBonus,
			TotalPoints:   0, // filled per-item but batch returns aggregate
		})
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	newTotal := 0
	if user != nil {
		newTotal = user.Points + totalPointsEarned
	}

	title := rankTitle(newTotal)
	_ = s.repo.UpdateUserPoints(ctx, userID, totalPointsEarned, title)

	s.sessions.Delete(sessionID)

	return &model.BatchSubmitResult{
		CorrectCount:      correctCount,
		TotalCount:        len(answers),
		TotalPointsEarned: totalPointsEarned,
		NewTotalPoints:    newTotal,
		RankTitle:         title,
		Results:           results,
	}, nil
}

func (s *quizService) GetUserStats(ctx context.Context, userID string) (*model.UserStats, error) {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	lastStreak, _ := s.repo.GetLastStreak(ctx, userID)
	maxStreak, _ := s.repo.GetMaxStreak(ctx, userID)
	breakdown, _ := s.repo.GetCategoryBreakdown(ctx, userID)

	user, _ := s.userRepo.FindByID(ctx, userID)
	totalPoints := 0
	rankT := ""
	if user != nil {
		totalPoints = user.Points
		rankT = user.RankTitle
	}

	accuracy := 0.0
	if stats.TotalAnswers > 0 {
		accuracy = float64(stats.CorrectAnswers) / float64(stats.TotalAnswers)
	}

	return &model.UserStats{
		TotalPoints:       totalPoints,
		RankTitle:         rankT,
		TotalAnswers:      stats.TotalAnswers,
		CorrectAnswers:    stats.CorrectAnswers,
		Accuracy:          math.Round(accuracy*100) / 100,
		CurrentStreak:     lastStreak,
		MaxStreak:         maxStreak,
		CategoryBreakdown: breakdown,
	}, nil
}

func (s *quizService) GetLeaderboard(ctx context.Context, period, currentUserID string, page, pageSize int) (*model.LeaderboardResponse, error) {
	page, pageSize = s.normalizePagination(page, pageSize)

	rows, total, myRank, err := s.repo.GetLeaderboard(ctx, period, currentUserID, page, pageSize)
	if err != nil {
		return nil, err
	}

	entries := make([]model.LeaderboardEntry, 0, len(rows))
	for i, row := range rows {
		entries = append(entries, model.LeaderboardEntry{
			Rank: (page-1)*pageSize + i + 1,
			User: model.UserJSON{
				ID: row.UserID, Username: row.Username,
				Nickname: row.Nickname, AvatarURL: row.AvatarURL,
				Role: row.Role,
			},
			TotalPoints: row.TotalPoints,
			RankTitle:   row.RankTitle,
			AnswerCount: row.AnswerCount,
			Accuracy:    row.Accuracy,
		})
	}
	_ = total

	var mr *model.MyRankInfo
	if myRank != nil {
		mr = &model.MyRankInfo{Rank: myRank.Rank, TotalPoints: myRank.TotalPoints}
	}

	return &model.LeaderboardResponse{
		Period:      period,
		Leaderboard: entries,
		MyRank:      mr,
	}, nil
}

func (s *quizService) CreateQuestion(ctx context.Context, question, category, difficulty string, options []string, answerStr string, explanation string) error {
	if question == "" {
		return apperrors.BadRequest("题目内容不能为空")
	}
	if len(options) < 2 {
		return apperrors.BadRequest("至少需要2个选项")
	}

	answerIdx := answerLetterToIndex(answerStr)
	if answerIdx < 0 || answerIdx >= len(options) {
		return apperrors.BadRequest("正确答案不在选项范围内")
	}

	optsJSON, _ := json.Marshal(options)

	q := &model.Question{
		ID:         uuid.New().String(),
		Content:    question,
		Options:    string(optsJSON),
		Answer:     answerIdx,
		Difficulty: difficulty,
		Category:   category,
	}
	_ = explanation

	if err := s.repo.CreateQuestion(ctx, q); err != nil {
		return err
	}

	s.log.Info("题目创建成功", zap.String("id", q.ID), zap.String("category", category))
	return nil
}

func parseOptionsStr(raw string) []string {
	if raw == "" || raw == "null" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return []string{}
	}
	if arr == nil {
		return []string{}
	}
	return arr
}
