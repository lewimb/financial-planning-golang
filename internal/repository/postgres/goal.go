package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type goalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) domain.GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) GetAll(userID int, active bool) ([]domain.GoalResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, name, target_amount, current_amount, status, deadline, description, created_at
		FROM goals
		WHERE user_id = $1 AND ($2 = false OR deadline >= NOW())
	`, userID, active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []domain.GoalResponse
	for rows.Next() {
		var g domain.GoalResponse
		if err := rows.Scan(&g.Id, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Status, &g.Deadline, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

func (r *goalRepository) GetByID(id, userID int) (*domain.GoalResponse, error) {
	var g domain.GoalResponse
	err := r.db.QueryRow(`
		SELECT id, name, target_amount, current_amount, description, deadline, status, created_at
		FROM goals WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&g.Id, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Description, &g.Deadline, &g.Status, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *goalRepository) GetSavingsTotal(userID int) (float64, error) {
	var total float64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(current_amount), 0) FROM goals WHERE user_id = $1`, userID,
	).Scan(&total)
	return total, err
}

func (r *goalRepository) CountActive(userID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(id) FROM goals WHERE user_id = $1 AND deadline >= NOW()`, userID,
	).Scan(&count)
	return count, err
}

func (r *goalRepository) GetUpcomingMilestones(userID int) ([]domain.GoalResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, target_amount, current_amount, deadline, created_at, status
		FROM goals WHERE user_id = $1 AND deadline > NOW() AND target_amount <> current_amount
		ORDER BY deadline ASC LIMIT 4
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []domain.GoalResponse
	for rows.Next() {
		var m domain.GoalResponse
		if err := rows.Scan(&m.Id, &m.Name, &m.Description, &m.TargetAmount, &m.CurrentAmount, &m.Deadline, &m.CreatedAt, &m.Status); err != nil {
			return nil, err
		}
		milestones = append(milestones, m)
	}
	return milestones, rows.Err()
}

func (r *goalRepository) Create(userID int, req domain.CreateGoalRequest) error {
	result, err := r.db.Exec(`
		INSERT INTO goals (user_id, name, target_amount, description, deadline)
		SELECT $1, $2, $3, $4, $5 WHERE NOT EXISTS (
			SELECT 1 FROM goals WHERE user_id = $6 AND name = $7 AND deadline >= NOW()
		)
	`, userID, req.Name, req.TargetAmount, req.Description, req.Deadline, userID, req.Name)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *goalRepository) Update(id, userID int, req domain.CreateGoalRequest) error {
	_, err := r.db.Exec(`
		UPDATE goals SET
			name = $1,
			target_amount = CASE WHEN $2 > current_amount THEN $2 ELSE target_amount END,
			description = $3,
			deadline = $4,
			status = CASE WHEN $2 > current_amount THEN 'ONGOING' ELSE status END,
			updated_at = NOW()
		WHERE id = $5 AND user_id = $6
	`, req.Name, req.TargetAmount, req.Description, req.Deadline, id, userID)
	return err
}

func (r *goalRepository) Delete(id, userID int) error {
	result, err := r.db.Exec(`DELETE FROM goals WHERE id = $1 AND user_id = $2`, id, userID)
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

func (r *goalRepository) Contribute(id, userID, amount int) error {
	_, err := r.db.Exec(`
		UPDATE goals
		SET current_amount = $1,
		    status = CASE WHEN $1 >= target_amount THEN 'COMPLETED' ELSE status END,
		    updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`, amount, id, userID)
	return err
}
