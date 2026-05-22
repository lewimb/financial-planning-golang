package domain

import "time"

type AiLog struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Question  string    `json:"question"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

type AiLogRepository interface {
	Save(userID int, question, response string) error
	GetByUserID(userID int) ([]AiLog, error)
	DeleteByUserID(userID int) error
}
