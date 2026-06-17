package repository

import (
	"context"
	"errors"

	"github.com/grand-canal-guardian/services/main-service/internal/model"
	"gorm.io/gorm"
)

type QuizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

func (r *QuizRepository) GetQuestions(ctx context.Context, difficulty string, count int) ([]*model.Question, error) {
	var questions []*model.Question
	query := r.db.WithContext(ctx).Model(&model.Question{})
	if difficulty != "" && difficulty != "all" {
		query = query.Where("difficulty = ?", difficulty)
	}
	err := query.Order("RANDOM()").Limit(count).Find(&questions).Error
	return questions, err
}

func (r *QuizRepository) GetQuestionByID(ctx context.Context, id string) (*model.Question, error) {
	var q model.Question
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&q).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &q, err
}

func (r *QuizRepository) CreateRecord(ctx context.Context, record *model.QuizRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *QuizRepository) GetUserStats(ctx context.Context, userID string) (*model.UserStats, error) {
	stats := &model.UserStats{UserID: userID}

	type aggResult struct {
		Total  int64
		Correct int64
		Score  int64
	}

	var result aggResult
	err := r.db.WithContext(ctx).Model(&model.QuizRecord{}).
		Select("COUNT(*) as total, SUM(CASE WHEN correct THEN 1 ELSE 0 END) as correct, SUM(score) as score").
		Where("user_id = ?", userID).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	stats.TotalScore = int(result.Score)
	stats.TotalQuestions = int(result.Total)
	stats.CorrectCount = int(result.Correct)
	if result.Total > 0 {
		stats.Accuracy = float64(result.Correct) / float64(result.Total) * 100
	}
	stats.RankTitle = model.GetRankTitle(stats.TotalScore)
	return stats, nil
}

func (r *QuizRepository) GetLeaderboard(ctx context.Context, limit int) ([]struct {
	UserID  string
	Score   int64
}, error) {
	type row struct {
		UserID string
		Score  int64
	}
	var rows []row
	// Group by user and sum scores
	err := r.db.WithContext(ctx).Model(&model.QuizRecord{}).
		Select("user_id, SUM(score) as score").
		Group("user_id").Order("score DESC").Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]struct {
		UserID string
		Score  int64
	}, len(rows))
	for i, r := range rows {
		result[i].UserID = r.UserID
		result[i].Score = r.Score
	}
	return result, nil
}
