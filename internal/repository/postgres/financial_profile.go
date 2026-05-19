package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type financialProfileRepository struct {
	db *sql.DB
}

func NewFinancialProfileRepository(db *sql.DB) domain.FinancialProfileRepository {
	return &financialProfileRepository{db: db}
}

// Upsert inserts or updates the user's financial profile and replaces their goal list
// atomically inside a single DB transaction.
func (r *financialProfileRepository) Upsert(userID int, req domain.UpsertFinancialProfileRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`
		INSERT INTO user_financial_profiles
			(user_id, monthly_income, fixed_expenses, current_savings, debt,
			 employment_status, spending_habit, risk_level, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			monthly_income    = EXCLUDED.monthly_income,
			fixed_expenses    = EXCLUDED.fixed_expenses,
			current_savings   = EXCLUDED.current_savings,
			debt              = EXCLUDED.debt,
			employment_status = EXCLUDED.employment_status,
			spending_habit    = EXCLUDED.spending_habit,
			risk_level        = EXCLUDED.risk_level,
			updated_at        = NOW()
	`, userID, req.MonthlyIncome, req.FixedExpenses, req.CurrentSavings, req.Debt,
		req.EmploymentStatus, req.SpendingHabit, req.RiskLevel)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(`DELETE FROM user_financial_goals WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, goal := range req.FinancialGoals {
		if _, err = tx.Exec(`
			INSERT INTO user_financial_goals (user_id, goal_type)
			VALUES ($1, $2)
			ON CONFLICT (user_id, goal_type) DO NOTHING
		`, userID, goal); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *financialProfileRepository) GetByUserID(userID int) (*domain.FinancialProfileResponse, error) {
	var p domain.FinancialProfileResponse
	err := r.db.QueryRow(`
		SELECT monthly_income, fixed_expenses, current_savings, debt,
		       employment_status, spending_habit, risk_level, created_at, updated_at
		FROM user_financial_profiles
		WHERE user_id = $1
	`, userID).Scan(
		&p.MonthlyIncome, &p.FixedExpenses, &p.CurrentSavings, &p.Debt,
		&p.EmploymentStatus, &p.SpendingHabit, &p.RiskLevel,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT goal_type FROM user_financial_goals
		WHERE user_id = $1 ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	p.FinancialGoals = make([]string, 0)
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		p.FinancialGoals = append(p.FinancialGoals, g)
	}
	return &p, rows.Err()
}
