package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

type QuizRepository interface {
	GetQuestions(ctx context.Context, count int, difficulty, category string) ([]model.Question, error)
	GetQuestionByID(ctx context.Context, id string) (*model.Question, error)

	CreateRecord(ctx context.Context, record *model.QuizRecord) error
	GetUserStats(ctx context.Context, userID string) (*QuizUserStats, error)
	GetCategoryBreakdown(ctx context.Context, userID string) (map[string]model.CategoryBreakdownItem, error)
	GetLastStreak(ctx context.Context, userID string) (int, error)
	GetMaxStreak(ctx context.Context, userID string) (int, error)

	GetLeaderboard(ctx context.Context, period string, userID string, page, pageSize int) ([]LeaderboardRow, int64, *MyRank, error)
	UpdateUserPoints(ctx context.Context, userID string, points int, title string) error

	CreateQuestion(ctx context.Context, q *model.Question) error

	CountDistinctUsers(ctx context.Context) (int64, error)
}

type QuizUserStats struct {
	TotalAnswers   int
	CorrectAnswers int
}

type LeaderboardRow struct {
	UserID      string  `gorm:"column:user_id"`
	Username    string  `gorm:"column:username"`
	Nickname    string  `gorm:"column:nickname"`
	AvatarURL   string  `gorm:"column:avatar_url"`
	Role        string  `gorm:"column:role"`
	TotalPoints int     `gorm:"column:total_points"`
	RankTitle   string  `gorm:"column:rank_title"`
	AnswerCount int     `gorm:"column:answer_count"`
	Accuracy    float64 `gorm:"column:accuracy"`
}

type MyRank struct {
	Rank        int
	TotalPoints int
}

type quizRepo struct {
	db *gorm.DB
}

func NewQuizRepo(db *gorm.DB) QuizRepository {
	return &quizRepo{db: db}
}

func (r *quizRepo) GetQuestions(ctx context.Context, count int, difficulty, category string) ([]model.Question, error) {
	q := r.db.WithContext(ctx).Model(&model.Question{})

	if difficulty != "" && difficulty != "mixed" {
		q = q.Where("difficulty = ?", difficulty)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	q = q.Order("RANDOM()").Limit(count)

	var questions []model.Question
	if err := q.Find(&questions).Error; err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return questions, nil
}

func (r *quizRepo) GetQuestionByID(ctx context.Context, id string) (*model.Question, error) {
	var q model.Question
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&q).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrQuestionNotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &q, nil
}

func (r *quizRepo) CreateRecord(ctx context.Context, record *model.QuizRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *quizRepo) GetUserStats(ctx context.Context, userID string) (*QuizUserStats, error) {
	var stats QuizUserStats
	err := r.db.WithContext(ctx).Model(&model.QuizRecord{}).
		Select("COUNT(*) as total_answers, COUNT(CASE WHEN correct THEN 1 END) as correct_answers").
		Where("user_id = ?", userID).
		Scan(&stats).Error
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &stats, nil
}

func (r *quizRepo) GetCategoryBreakdown(ctx context.Context, userID string) (map[string]model.CategoryBreakdownItem, error) {
	type row struct {
		Category string
		Total    int
		Correct  int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("quiz_records").
		Select("q.category, COUNT(*) as total, COUNT(CASE WHEN quiz_records.correct THEN 1 END) as correct").
		Joins("JOIN questions q ON q.id = quiz_records.question_id").
		Where("quiz_records.user_id = ?", userID).
		Group("q.category").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	result := make(map[string]model.CategoryBreakdownItem, len(rows))
	for _, r := range rows {
		result[r.Category] = model.CategoryBreakdownItem{Total: r.Total, Correct: r.Correct}
	}
	return result, nil
}

func (r *quizRepo) GetLastStreak(ctx context.Context, userID string) (int, error) {
	var records []model.QuizRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&records).Error
	if err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	streak := 0
	for _, rec := range records {
		if rec.Correct {
			streak++
		} else {
			break
		}
	}
	return streak, nil
}

func (r *quizRepo) GetMaxStreak(ctx context.Context, userID string) (int, error) {
	var records []model.QuizRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&records).Error
	if err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	maxStreak := 0
	currentStreak := 0
	for _, rec := range records {
		if rec.Correct {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 0
		}
	}
	return maxStreak, nil
}

func (r *quizRepo) GetLeaderboard(ctx context.Context, period string, userID string, page, pageSize int) ([]LeaderboardRow, int64, *MyRank, error) {
	base := r.db.WithContext(ctx).
		Table("quiz_records").
		Select("quiz_records.user_id, u.username, u.nickname, u.avatar_url, u.role, u.points as total_points, u.rank_title, COUNT(*) as answer_count, ROUND(COUNT(CASE WHEN quiz_records.correct THEN 1 END)*1.0/NULLIF(COUNT(*),0), 2) as accuracy").
		Joins("JOIN users u ON u.id = quiz_records.user_id AND u.deleted_at IS NULL")

	switch period {
	case "daily":
		base = base.Where("quiz_records.created_at >= CURRENT_DATE")
	case "weekly":
		base = base.Where("quiz_records.created_at >= date_trunc('week', CURRENT_DATE)")
	case "monthly":
		base = base.Where("quiz_records.created_at >= date_trunc('month', CURRENT_DATE)")
	}

	base = base.Group("quiz_records.user_id, u.username, u.nickname, u.avatar_url, u.role, u.points, u.rank_title")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	offset := (page - 1) * pageSize
	var rows []LeaderboardRow
	if err := base.Order("total_points DESC, accuracy DESC").Offset(offset).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, 0, nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}

	var myRank *MyRank
	if userID != "" {
		type rankRow struct {
			UserID      string
			TotalPoints int
		}
		var all []rankRow
		if err := base.Order("total_points DESC, accuracy DESC").Scan(&all).Error; err == nil {
			for i, r := range all {
				if r.UserID == userID {
					myRank = &MyRank{Rank: i + 1, TotalPoints: r.TotalPoints}
					break
				}
			}
		}
	}

	return rows, total, myRank, nil
}

func (r *quizRepo) UpdateUserPoints(ctx context.Context, userID string, points int, title string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"points":     gorm.Expr("points + ?", points),
			"rank_title": title,
		}).Error
}

func (r *quizRepo) CreateQuestion(ctx context.Context, q *model.Question) error {
	if err := r.db.WithContext(ctx).Create(q).Error; err != nil {
		return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return nil
}

func (r *quizRepo) CountDistinctUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QuizRecord{}).
		Distinct("user_id").Count(&count).Error
	if err != nil {
		return 0, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return count, nil
}

func parseOptions(raw string) []string {
	if raw == "" {
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
