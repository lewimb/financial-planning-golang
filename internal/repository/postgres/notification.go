package postgres

import (
	"database/sql"
	"time"

	"github.com/financial-planning/internal/domain"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetByUserID(userID int, unreadOnly bool) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, entity_type, entity_id, is_read, created_at
		FROM notifications
		WHERE user_id = $1`
	if unreadOnly {
		query += ` AND is_read = FALSE`
	}
	query += ` ORDER BY created_at DESC LIMIT 50`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
			&n.EntityType, &n.EntityID, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, rows.Err()
}

func (r *notificationRepository) Create(userID int, notifType, title, message string, entityType *string, entityID *int) error {
	_, err := r.db.Exec(`
		INSERT INTO notifications (user_id, type, title, message, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, notifType, title, message, entityType, entityID,
	)
	return err
}

func (r *notificationRepository) MarkRead(id, userID int) error {
	_, err := r.db.Exec(`
		UPDATE notifications SET is_read = TRUE
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

func (r *notificationRepository) MarkAllRead(userID int) error {
	_, err := r.db.Exec(`
		UPDATE notifications SET is_read = TRUE
		WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	)
	return err
}

func (r *notificationRepository) Delete(id, userID int) error {
	result, err := r.db.Exec(`
		DELETE FROM notifications WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *notificationRepository) GetUnreadCount(userID int) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *notificationRepository) GetPreferences(userID int) (*domain.NotificationPreferences, error) {
	var p domain.NotificationPreferences
	err := r.db.QueryRow(`
		SELECT budget_alerts, goal_reminders, anomaly_alerts, weekly_summary, push_enabled
		FROM notification_preferences
		WHERE user_id = $1`,
		userID,
	).Scan(&p.BudgetAlerts, &p.GoalReminders, &p.AnomalyAlerts, &p.WeeklySummary, &p.PushEnabled)
	if err == sql.ErrNoRows {
		return &domain.NotificationPreferences{BudgetAlerts: true, GoalReminders: true, AnomalyAlerts: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *notificationRepository) UpsertPreferences(userID int, prefs domain.NotificationPreferences) error {
	_, err := r.db.Exec(`
		INSERT INTO notification_preferences (user_id, budget_alerts, goal_reminders, anomaly_alerts, weekly_summary, push_enabled, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			budget_alerts  = EXCLUDED.budget_alerts,
			goal_reminders = EXCLUDED.goal_reminders,
			anomaly_alerts = EXCLUDED.anomaly_alerts,
			weekly_summary = EXCLUDED.weekly_summary,
			push_enabled   = EXCLUDED.push_enabled,
			updated_at     = EXCLUDED.updated_at`,
		userID, prefs.BudgetAlerts, prefs.GoalReminders, prefs.AnomalyAlerts,
		prefs.WeeklySummary, prefs.PushEnabled, time.Now(),
	)
	return err
}

func (r *notificationRepository) ExistsRecent(userID int, notifType string, entityID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM notifications
			WHERE user_id = $1
			  AND type = $2
			  AND entity_id = $3
			  AND created_at > NOW() - INTERVAL '24 hours'
		)`,
		userID, notifType, entityID,
	).Scan(&exists)
	return exists, err
}
