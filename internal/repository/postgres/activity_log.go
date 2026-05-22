package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type activityLogRepository struct {
	db *sql.DB
}

func NewActivityLogRepository(db *sql.DB) domain.ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Log(userID int, action, entityType string, entityID *int, description string) error {
	_, err := r.db.Exec(`
		INSERT INTO activity_logs (user_id, action, entity_type, entity_id, description)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, action, entityType, entityID, description,
	)
	return err
}

func (r *activityLogRepository) GetByUserID(userID int, limit, offset int) ([]domain.ActivityLog, int, error) {
	var total int
	if err := r.db.QueryRow(
		`SELECT COUNT(*) FROM activity_logs WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, user_id, action, entity_type, entity_id, description, created_at
		FROM activity_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []domain.ActivityLog
	for rows.Next() {
		var l domain.ActivityLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &l.Description, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}
