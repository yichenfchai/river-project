package model

import (
	"time"
)

type Question struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	Content    string    `gorm:"type:text;not null" json:"question"`
	Options    string    `gorm:"type:text;not null" json:"-"`
	Answer     int       `gorm:"not null" json:"-"`
	Difficulty string    `gorm:"size:16;default:easy;index" json:"difficulty"`
	Category   string    `gorm:"size:32;default:运河文化" json:"category"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Question) TableName() string { return "questions" }

type QuizRecord struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	UserID     string    `gorm:"size:36;not null;index" json:"user_id"`
	QuestionID string    `gorm:"size:36;not null" json:"question_id"`
	Correct    bool      `gorm:"not null" json:"correct"`
	Score      int       `gorm:"default:0" json:"score"`
	Streak     int       `gorm:"default:0" json:"streak"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (QuizRecord) TableName() string { return "quiz_records" }

// ─── DTOs ───

type QuestionResponse struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Difficulty string   `json:"difficulty"`
	Category   string   `json:"category"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
}

type SessionResponse struct {
	SessionID string             `json:"session_id"`
	Questions []QuestionResponse `json:"questions"`
}

type SubmitAnswerRequest struct {
	SessionID    string `json:"session_id" binding:"required"`
	QuestionID   string `json:"question_id" binding:"required"`
	Answer       string `json:"answer" binding:"required"`
	AnswerTimeMs int    `json:"answer_time_ms"`
}

type SubmitAnswerResponse struct {
	IsCorrect    bool   `json:"is_correct"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation  string `json:"explanation"`
	PointsEarned int    `json:"points_earned"`
	StreakBonus  int    `json:"streak_bonus"`
	TotalPoints  int    `json:"total_points"`
}

type BatchSubmitResult struct {
	CorrectCount      int                    `json:"correct_count"`
	TotalCount        int                    `json:"total_count"`
	TotalPointsEarned int                    `json:"total_points_earned"`
	NewTotalPoints    int                    `json:"new_total_points"`
	RankTitle         string                 `json:"rank_title"`
	Results           []SubmitAnswerResponse `json:"results"`
}

type UserStats struct {
	TotalPoints       int                             `json:"total_points"`
	RankTitle         string                          `json:"rank_title"`
	TotalAnswers      int                             `json:"total_answers"`
	CorrectAnswers    int                             `json:"correct_answers"`
	Accuracy          float64                         `json:"accuracy"`
	CurrentStreak     int                             `json:"current_streak"`
	MaxStreak         int                             `json:"max_streak"`
	CategoryBreakdown map[string]CategoryBreakdownItem `json:"category_breakdown"`
}

type CategoryBreakdownItem struct {
	Total   int `json:"total"`
	Correct int `json:"correct"`
}

type LeaderboardEntry struct {
	Rank        int      `json:"rank"`
	User        UserJSON `json:"user"`
	TotalPoints int      `json:"total_points"`
	RankTitle   string   `json:"rank_title"`
	AnswerCount int      `json:"answer_count"`
	Accuracy    float64  `json:"accuracy"`
}

type LeaderboardResponse struct {
	Period      string            `json:"period"`
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
	MyRank      *MyRankInfo       `json:"my_rank"`
}

type MyRankInfo struct {
	Rank        int `json:"rank"`
	TotalPoints int `json:"total_points"`
}
