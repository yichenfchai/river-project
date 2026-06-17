package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/grand-canal-guardian/pkg/errors"
	"github.com/grand-canal-guardian/services/quiz-service/internal/model"
	"github.com/grand-canal-guardian/services/quiz-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type QuizService struct {
	repo *repository.QuizRepository
	rdb  *redis.Client
}

func NewQuizService(repo *repository.QuizRepository, rdb *redis.Client) *QuizService {
	return &QuizService{repo: repo, rdb: rdb}
}

// StartSession 开始答题会话 — 从 Redis 获取预载题池
func (s *QuizService) StartSession(ctx context.Context, userID, difficulty string, count int) (*model.QuizSession, []*model.Question, *apperrors.AppError) {
	questions, repoErr := s.repo.GetQuestions(ctx, difficulty, count)
	if repoErr != nil {
		return nil, nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, repoErr)
	}
	if len(questions) == 0 {
		return nil, nil, apperrors.NewDefault(apperrors.ErrQuestionNotFound)
	}

	sessionID := uuid.New().String()
	questionIDs := make([]string, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	session := &model.QuizSession{
		UserID:    userID,
		Questions: questionIDs,
	}

	// 缓存 session 到 Redis (TTL 1h)
	data, _ := json.Marshal(session)
	_ = s.rdb.Set(ctx, "quiz:session:"+sessionID, string(data), 3600*1e9).Err()

	session.ID = sessionID // for handler use

	// 隐藏答案
	for _, q := range questions {
		q.Answer = -1
	}

	return session, questions, nil
}

// SubmitAnswer 提交答案
func (s *QuizService) SubmitAnswer(ctx context.Context, userID, sessionID, questionID string, selectedOption int) (*model.SubmitResponse, *apperrors.AppError) {
	// 1. 加载 session
	sessionData, err := s.rdb.Get(ctx, "quiz:session:"+sessionID).Result()
	if err != nil {
		return nil, apperrors.NewDefault(apperrors.ErrSessionExpired)
	}

	var session model.QuizSession
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, apperrors.NewDefault(apperrors.ErrSessionExpired)
	}

	// 2. 校验题目是否属于 session
	if session.CurrentIndex >= len(session.Questions) ||
		session.Questions[session.CurrentIndex] != questionID {
		return nil, apperrors.BadRequest("题目不匹配")
	}

	// 3. 获取正确答案
	question, repoErr := s.repo.GetQuestionByID(ctx, questionID)
	if repoErr != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, repoErr)
	}
	if question == nil {
		return nil, apperrors.NewDefault(apperrors.ErrQuestionNotFound)
	}

	correct := selectedOption == question.Answer
	score := 0
	if correct {
		session.Streak++
		session.CorrectCount++
		score = model.GetScoreForAnswer(question.Difficulty, session.Streak-1)
	} else {
		session.Streak = 0
	}
	session.Score += score
	session.TotalCount++
	session.CurrentIndex++

	// 4. 保存更新后的 session
	updatedData, _ := json.Marshal(session)
	_ = s.rdb.Set(ctx, "quiz:session:"+sessionID, string(updatedData), 3600*1e9).Err()

	// 5. 记录到 DB
	record := &model.QuizRecord{
		UserID:     userID,
		QuestionID: questionID,
		Correct:    correct,
		Score:      score,
		Streak:     session.Streak,
	}
	_ = s.repo.CreateRecord(ctx, record)

	// 6. 更新 Redis 排行榜 (Sorted Set)
	scoreKey := "leaderboard:total"
	if score > 0 {
		s.rdb.ZIncrBy(ctx, scoreKey, float64(score), userID)
	}

	return &model.SubmitResponse{
		Correct:        correct,
		Score:          score,
		TotalScore:     session.Score,
		Streak:         session.Streak,
		TotalIndex:     session.TotalCount,
		TotalQuestions: len(session.Questions),
		HasMore:        session.CurrentIndex < len(session.Questions),
	}, nil
}

// GetLeaderboard 排行榜
func (s *QuizService) GetLeaderboard(ctx context.Context, period string, limit int) ([]*model.LeaderboardEntry, *apperrors.AppError) {
	key := fmt.Sprintf("leaderboard:%s", period)
	if period == "" {
		key = "leaderboard:total"
	}

	results, err := s.rdb.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		// Redis 不可用时降级到 DB
		dbResults, dbErr := s.repo.GetLeaderboard(ctx, limit)
		if dbErr != nil {
			return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, dbErr)
		}
		entries := make([]*model.LeaderboardEntry, len(dbResults))
		for i, r := range dbResults {
			entries[i] = &model.LeaderboardEntry{
				UserID: r.UserID,
				Score:  int(r.Score),
				Rank:   i + 1,
			}
		}
		return entries, nil
	}

	entries := make([]*model.LeaderboardEntry, len(results))
	for i, z := range results {
		entries[i] = &model.LeaderboardEntry{
			UserID: z.Member.(string),
			Score:  int(z.Score),
			Rank:   i + 1,
		}
	}
	return entries, nil
}

// GetUserStats 个人统计
func (s *QuizService) GetUserStats(ctx context.Context, userID string) (*model.UserStats, *apperrors.AppError) {
	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return stats, nil
}
