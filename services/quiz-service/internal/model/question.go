package model

import "time"

// Question 题目
type Question struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Options    string    `gorm:"type:text;not null" json:"options"`   // JSON array
	Answer     int       `gorm:"not null" json:"-"`                   // 0-indexed correct option
	Difficulty string    `gorm:"size:16;index" json:"difficulty"`    // easy|medium|hard
	Category   string    `gorm:"size:32" json:"category"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Question) TableName() string { return "questions" }

// QuizRecord 答题记录
type QuizRecord struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID     string    `gorm:"index;not null" json:"user_id"`
	QuestionID string    `gorm:"not null" json:"question_id"`
	Correct    bool      `gorm:"not null" json:"correct"`
	Score      int       `gorm:"default:0" json:"score"`
	Streak     int       `gorm:"default:0" json:"streak"`
	CreatedAt  time.Time `json:"created_at"`
}

func (QuizRecord) TableName() string { return "quiz_records" }

// UserStats 用户统计 (Redis + DB 双写)
type UserStats struct {
	UserID        string  `json:"user_id"`
	TotalScore    int     `json:"total_score"`
	TotalQuestions int    `json:"total_questions"`
	CorrectCount  int     `json:"correct_count"`
	Accuracy      float64 `json:"accuracy"`
	Streak        int     `json:"streak"`
	MaxStreak     int     `json:"max_streak"`
	RankTitle     string  `json:"rank_title"`
}

// SubmitRequest 提交答案请求
type SubmitRequest struct {
	SessionID      string `json:"session_id" binding:"required"`
	QuestionID     string `json:"question_id" binding:"required"`
	SelectedOption int    `json:"selected_option" binding:"min=0"`
}

// Session 答题会话 (Redis)
type QuizSession struct {
	UserID       string   `json:"user_id"`
	Questions    []string `json:"questions"`     // question IDs in order
	CurrentIndex int      `json:"current_index"`
	Score        int      `json:"score"`
	Streak       int      `json:"streak"`
	CorrectCount int      `json:"correct_count"`
	TotalCount   int      `json:"total_count"`
}


// SubmitResponse 提交答案响应
type SubmitResponse struct {
	Correct        bool `json:"correct"`
	Score          int  `json:"score"`
	TotalScore     int  `json:"total_score"`
	Streak         int  `json:"streak"`
	TotalIndex     int  `json:"total_index"`
	TotalQuestions int  `json:"total_questions"`
	HasMore        bool `json:"has_more"`
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Score    int    `json:"score"`
	Rank     int    `json:"rank"`
}

// GetRankTitle 根据积分获取段位
func GetRankTitle(score int) string {
	switch {
	case score >= 15000: return "运河守护者"
	case score >= 5000:  return "钻石守护者"
	case score >= 1500:  return "黄金守护者"
	case score >= 500:   return "白银守护者"
	default:             return "青铜守护者"
	}
}

// GetScoreForAnswer 根据难度和连续正确计算得分
func GetScoreForAnswer(difficulty string, streak int) int {
	base := 100
	switch difficulty {
	case "hard":
		base = 300
	case "medium":
		base = 200
	}
	// 连续正确加成: streak 0-2: 1x, 3-5: 1.5x, 6+: 2x
	multiplier := 1.0
	if streak >= 6 {
		multiplier = 2.0
	} else if streak >= 3 {
		multiplier = 1.5
	}
	return int(float64(base) * multiplier)
}
