package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/financial-planning/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) GetByUserID(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error) {
	query := `SELECT id, amount, category, type, "date"::date, description FROM transactions WHERE user_id=$1 AND deleted_at IS NULL`
	args := []interface{}{userID}

	if month != "" && year != "" {
		monthInt, err := strconv.Atoi(month)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid month: %v", err)
		}
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid year: %v", err)
		}
		args = append(args, monthInt, yearInt)
		query += fmt.Sprintf(` AND EXTRACT(MONTH FROM "date"::date) = $%d AND EXTRACT(YEAR FROM "date"::date) = $%d`, len(args)-1, len(args))
	}

	query += ` ORDER BY "date" DESC`

	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
		if offset > 0 {
			args = append(args, offset)
			query += fmt.Sprintf(" OFFSET $%d", len(args))
		}
	}

	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id=$1 AND deleted_at IS NULL`
	countArgs := []interface{}{userID}

	if month != "" && year != "" {
		monthInt, _ := strconv.Atoi(month)
		yearInt, _ := strconv.Atoi(year)
		countArgs = append(countArgs, monthInt, yearInt)
		countQuery += fmt.Sprintf(` AND EXTRACT(MONTH FROM "date"::date) = $%d AND EXTRACT(YEAR FROM "date"::date) = $%d`, len(countArgs)-1, len(countArgs))
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count query error: %v", err)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("main query error: %v", err)
	}
	defer rows.Close()

	var transactions []domain.TransactionResponse
	for rows.Next() {
		var t domain.TransactionResponse
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Type, &t.Date, &t.Description); err != nil {
			return nil, 0, fmt.Errorf("scan error: %v", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, totalCount, rows.Err()
}

func (r *transactionRepository) Create(userID int, req domain.TransactionRequest) error {
	_, err := r.db.Exec(
		"INSERT INTO transactions (amount, category, type, date, description, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
		req.Amount, req.Category, req.Type, req.Date, req.Description, userID,
	)
	return err
}

func (r *transactionRepository) Update(userID, id int, req domain.TransactionRequest) error {
	result, err := r.db.Exec(`
		UPDATE transactions
		SET amount = $1, category = $2, type = $3, date = $4, description = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, req.Amount, req.Category, req.Type, req.Date, req.Description, id, userID)
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

func (r *transactionRepository) Delete(userID, id int) error {
	result, err := r.db.Exec(`
		UPDATE transactions SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
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

func (r *transactionRepository) GetMonthlyExpenses(userID int) (float64, error) {
	now := time.Now()
	var total float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = $1
		AND EXTRACT(MONTH FROM date) = $2
		AND EXTRACT(YEAR FROM date) = $3
		AND type = 'EXPENSE'
		AND deleted_at IS NULL
	`, userID, int(now.Month()), now.Year()).Scan(&total)
	return total, err
}

func (r *transactionRepository) GetMonthlyIncome(userID int) (float64, error) {
	now := time.Now()
	var total float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = $1
		AND EXTRACT(MONTH FROM date) = $2
		AND EXTRACT(YEAR FROM date) = $3
		AND type = 'INCOME'
		AND deleted_at IS NULL
	`, userID, int(now.Month()), now.Year()).Scan(&total)
	return total, err
}

func (r *transactionRepository) GetNetSavings(userID int) (float64, error) {
	var net float64
	err := r.db.QueryRow(`
		SELECT
		COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount END), 0) -
		COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount END), 0) AS net_savings
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID).Scan(&net)
	return net, err
}
