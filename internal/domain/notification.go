package domain

import "time"

type Notification struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	EntityType *string   `json:"entity_type,omitempty"`
	EntityID   *int      `json:"entity_id,omitempty"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type NotificationPreferences struct {
	BudgetAlerts  bool `json:"budget_alerts"`
	GoalReminders bool `json:"goal_reminders"`
	AnomalyAlerts bool `json:"anomaly_alerts"`
}

type NotificationRepository interface {
	GetByUserID(userID int, unreadOnly bool) ([]Notification, error)
	Create(userID int, notifType, title, message string, entityType *string, entityID *int) error
	MarkRead(id, userID int) error
	MarkAllRead(userID int) error
	Delete(id, userID int) error
	GetUnreadCount(userID int) (int, error)
	GetPreferences(userID int) (*NotificationPreferences, error)
	UpsertPreferences(userID int, prefs NotificationPreferences) error
	ExistsRecent(userID int, notifType string, entityID int) (bool, error)
}
