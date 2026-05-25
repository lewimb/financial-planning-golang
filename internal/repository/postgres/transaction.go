package postgres

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	query := `SELECT id, amount, category, type, "date"::date, description, is_recurring, COALESCE(recurrence_interval, '') FROM transactions WHERE user_id=$1 AND deleted_at IS NULL`
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
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Type, &t.Date, &t.Description, &t.IsRecurring, &t.RecurrenceInterval); err != nil {
			return nil, 0, fmt.Errorf("scan error: %v", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, totalCount, rows.Err()
}

func (r *transactionRepository) Create(userID int, req domain.TransactionRequest) error {
	_, err := r.db.Exec(
		`INSERT INTO transactions (amount, category, type, date, description, user_id, is_recurring, recurrence_interval)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))`,
		req.Amount, req.Category, req.Type, req.Date, req.Description, userID,
		req.IsRecurring, req.RecurrenceInterval,
	)
	return err
}

func (r *transactionRepository) Update(userID, id int, req domain.TransactionRequest) error {
	result, err := r.db.Exec(`
		UPDATE transactions
		SET amount = $1, category = $2, type = $3, date = $4, description = $5,
		    updated_at = NOW()
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

func (r *transactionRepository) GetMonthlySummary(userID int, months int) ([]domain.MonthlySummaryItem, error) {
	if months <= 0 {
		months = 6
	}
	rows, err := r.db.Query(`
		SELECT
			EXTRACT(MONTH FROM date)::int AS month,
			EXTRACT(YEAR FROM date)::int  AS year,
			COALESCE(SUM(CASE WHEN type = 'INCOME'  THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE user_id = $1
		  AND deleted_at IS NULL
		  AND date >= DATE_TRUNC('month', NOW()) - ($2 - 1) * INTERVAL '1 month'
		GROUP BY EXTRACT(YEAR FROM date), EXTRACT(MONTH FROM date)
		ORDER BY year ASC, month ASC
	`, userID, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.MonthlySummaryItem
	for rows.Next() {
		var item domain.MonthlySummaryItem
		if err := rows.Scan(&item.Month, &item.Year, &item.Income, &item.Expense); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *transactionRepository) GetYearlySummary(userID, year int) ([]domain.MonthlySummaryItem, error) {
	rows, err := r.db.Query(`
		WITH months AS (
			SELECT generate_series(1, 12) AS month
		)
		SELECT
			m.month,
			COALESCE(SUM(CASE WHEN t.type = 'INCOME'  THEN t.amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN t.type = 'EXPENSE' THEN t.amount ELSE 0 END), 0) AS expense
		FROM months m
		LEFT JOIN transactions t
			ON EXTRACT(MONTH FROM t.date) = m.month
			AND EXTRACT(YEAR FROM t.date) = $2
			AND t.user_id = $1
			AND t.deleted_at IS NULL
		GROUP BY m.month
		ORDER BY m.month
	`, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.MonthlySummaryItem
	for rows.Next() {
		var item domain.MonthlySummaryItem
		if err := rows.Scan(&item.Month, &item.Income, &item.Expense); err != nil {
			return nil, err
		}
		item.Year = year
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *transactionRepository) GetCategoryBreakdownDetailed(userID, month, year int) ([]domain.CategoryBreakdownItem, error) {
	rows, err := r.db.Query(`
		SELECT
			category,
			SUM(amount)::float8 AS total,
			COUNT(*)::int AS transaction_count
		FROM transactions
		WHERE user_id = $1
		  AND type = 'EXPENSE'
		  AND EXTRACT(MONTH FROM date) = $2
		  AND EXTRACT(YEAR FROM date) = $3
		  AND deleted_at IS NULL
		GROUP BY category
		ORDER BY total DESC
	`, userID, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CategoryBreakdownItem
	var grandTotal float64
	for rows.Next() {
		var item domain.CategoryBreakdownItem
		if err := rows.Scan(&item.Category, &item.Total, &item.TransactionCount); err != nil {
			return nil, err
		}
		grandTotal += item.Total
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if grandTotal > 0 {
		for i := range items {
			items[i].Percentage = math.Round(items[i].Total/grandTotal*1000) / 10
			items[i].Label = categoryLabel(items[i].Category)
		}
	}
	return items, nil
}

func categoryLabel(cat string) string {
	labels := map[string]string{
		"FOOD":          "Food & Dining",
		"TRANSPORT":     "Transportation",
		"ENTERTAINMENT": "Entertainment",
		"UTILITIES":     "Utilities",
		"SHOPPING":      "Shopping",
		"HEALTH":        "Healthcare",
		"EDUCATION":     "Education",
		"HOUSING":       "Housing",
		"SAVINGS":       "Savings",
		"INCOME":        "Income",
		"INVESTMENT":    "Investment",
		"TRAVEL":        "Travel",
	}
	if l, ok := labels[strings.ToUpper(cat)]; ok {
		return l
	}
	words := strings.Fields(strings.ToLower(cat))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func (r *transactionRepository) BulkCreate(userID int, reqs []domain.TransactionRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO transactions (amount, category, type, date, description, user_id, is_recurring, recurrence_interval)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, req := range reqs {
		if _, err := stmt.Exec(
			req.Amount, req.Category, req.Type, req.Date, req.Description, userID,
			req.IsRecurring, req.RecurrenceInterval,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
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
