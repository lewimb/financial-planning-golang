package domain

import "time"

type ActivityLog struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entity_type"`
	EntityID    *int      `json:"entity_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ActivityLogRepository interface {
	Log(userID int, action, entityType string, entityID *int, description string) error
	GetByUserID(userID int, limit, offset int) ([]ActivityLog, int, error)
}
