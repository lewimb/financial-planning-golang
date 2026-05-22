package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type aiLogRepository struct {
	db *sql.DB
}

func NewAiLogRepository(db *sql.DB) domain.AiLogRepository {
	return &aiLogRepository{db: db}
}

func (r *aiLogRepository) Save(userID int, question, response string) error {
	_, err := r.db.Exec(
		`INSERT INTO ai_logs (user_id, question, response) VALUES ($1, $2, $3)`,
		userID, question, response,
	)
	return err
}

func (r *aiLogRepository) GetByUserID(userID int) ([]domain.AiLog, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, question, response, created_at
		FROM ai_logs
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AiLog
	for rows.Next() {
		var l domain.AiLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Question, &l.Response, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *aiLogRepository) DeleteByUserID(userID int) error {
	_, err := r.db.Exec(`UPDATE ai_logs SET deleted_at = NOW() WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	return err
}
