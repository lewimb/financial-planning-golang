package postgres

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/financial-planning/internal/domain"
)

type budgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) GetAll(userID int, category, month, year string) ([]domain.Budget, error) {
	query := `SELECT id, user_id, category, period, month, year, limit_amount, alert_threshold, created_at FROM budgets WHERE user_id = $1 AND deleted_at IS NULL`
	args := []interface{}{userID}
	idx := 2

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, category)
		idx++
	}
	if year != "" {
		query += fmt.Sprintf(" AND year = $%d", idx)
		args = append(args, year)
		idx++
	}
	if month != "" {
		query += fmt.Sprintf(" AND month = $%d", idx)
		args = append(args, month)
		idx++
	}
	_ = idx

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold, &b.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *budgetRepository) GetByID(id int) (*domain.BudgetResponse, error) {
	var b domain.BudgetResponse
	err := r.db.QueryRow(`
		SELECT id,user_id,category,period,month,year,limit_amount,alert_threshold,created_at
		FROM budgets WHERE id = $1
	`, id).Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *budgetRepository) GetUsage(userID, month, year int) ([]domain.BudgetUsage, error) {
	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear = year - 1
	}

	rows, err := r.db.Query(`
		SELECT
			b.id, b.category, b.period, b.limit_amount, b.alert_threshold,
			COALESCE(SUM(CASE
				WHEN (
					(b.period = 'MONTHLY' AND EXTRACT(MONTH FROM t.date) = $2 AND EXTRACT(YEAR FROM t.date) = $3)
					OR (b.period = 'YEARLY' AND EXTRACT(YEAR FROM t.date) = $3)
				) THEN t.amount ELSE 0
			END), 0)::BIGINT AS used,
			COALESCE(SUM(CASE
				WHEN (
					(b.period = 'MONTHLY' AND EXTRACT(MONTH FROM t.date) = $4 AND EXTRACT(YEAR FROM t.date) = $5)
					OR (b.period = 'YEARLY' AND EXTRACT(YEAR FROM t.date) = $5)
				) THEN t.amount ELSE 0
			END), 0)::BIGINT AS prev_used
		FROM budgets b
		LEFT JOIN transactions t
			ON t.user_id = b.user_id
			AND LOWER(t.category) = LOWER(b.category)
			AND t.type = 'EXPENSE'
			AND t.deleted_at IS NULL
		WHERE b.user_id = $1
		AND b.year = $3
		AND ((b.period = 'MONTHLY' AND b.month = $2) OR (b.period = 'YEARLY'))
		AND b.deleted_at IS NULL
		GROUP BY b.id, b.category, b.period, b.limit_amount, b.alert_threshold
		ORDER BY b.created_at DESC
	`, userID, month, year, prevMonth, prevYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BudgetUsage
	for rows.Next() {
		var b domain.BudgetUsage
		var prevUsed int64
		if err := rows.Scan(&b.BudgetID, &b.Category, &b.Period, &b.Limit, &b.AlertThreshold, &b.Used, &prevUsed); err != nil {
			return nil, err
		}

		b.Remaining = b.Limit - b.Used
		if b.Remaining < 0 {
			b.Remaining = 0
		}
		if b.Limit > 0 {
			b.Percentage = math.Round((float64(b.Used) / float64(b.Limit)) * 100)
		}
		if b.Percentage >= 100 {
			b.Status = "EXCEEDED"
		} else if b.Percentage >= float64(b.AlertThreshold) {
			b.Status = "WARNING"
		} else {
			b.Status = "SAFE"
		}
		if prevUsed > 0 {
			b.ChangePercent = math.Round(((float64(b.Used) - float64(prevUsed)) / float64(prevUsed)) * 100)
		} else if b.Used > 0 {
			b.ChangePercent = 100
		}

		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *budgetRepository) Create(userID int, req domain.CreateBudgetRequest) error {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM budgets
			WHERE user_id = $1 AND category = $2 AND period = $3 AND year = $4
			AND (month = $5 OR ($5 IS NULL AND month IS NULL))
		)
	`, userID, req.Category, req.Period, req.Year, req.Month).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}
	_, err = r.db.Exec(`
		INSERT INTO budgets (user_id, category, period, month, year, limit_amount, alert_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, req.Category, req.Period, req.Month, req.Year, req.LimitAmount, req.AlertThreshold)
	return err
}

func (r *budgetRepository) Update(userID, id, limitAmount, alertThreshold int, category string) (*domain.UpdateBudgetResponse, error) {
	var b domain.UpdateBudgetResponse
	err := r.db.QueryRow(`
		UPDATE budgets
		SET limit_amount = COALESCE(NULLIF($1, 0), limit_amount),
		    alert_threshold = COALESCE(NULLIF($2, 0), alert_threshold),
		    category = COALESCE(NULLIF($3, ''), category),
		    updated_at = NOW()
		WHERE user_id = $4 AND id = $5
		RETURNING id, user_id, category, period, month, year, limit_amount, alert_threshold
	`, limitAmount, alertThreshold, category, userID, id).
		Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return &b, nil
}

func (r *budgetRepository) Delete(userID, id int) error {
	result, err := r.db.Exec(`
		UPDATE budgets SET deleted_at = NOW()
		WHERE user_id = $1 AND id = $2 AND deleted_at IS NULL
	`, userID, id)
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
