package domain

import "time"

type TransactionRequest struct {
	Amount             int       `json:"amount"`
	Category           string    `json:"category"`
	Type               string    `json:"type"`
	Date               time.Time `json:"date"`
	Description        string    `json:"description"`
	IsRecurring        bool      `json:"is_recurring"`
	RecurrenceInterval string    `json:"recurrence_interval"`
}

type TransactionResponse struct {
	ID                 int       `json:"id"`
	Amount             float64   `json:"amount"`
	Category           string    `json:"category"`
	Type               string    `json:"type"`
	Date               time.Time `json:"date"`
	Description        string    `json:"description"`
	IsRecurring        bool      `json:"is_recurring"`
	RecurrenceInterval string    `json:"recurrence_interval,omitempty"`
}

type MonthlySummaryItem struct {
	Month   int     `json:"month"`
	Year    int     `json:"year"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type CategoryBreakdownItem struct {
	Category         string  `json:"category"`
	Label            string  `json:"label"`
	Total            float64 `json:"total"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int     `json:"transaction_count"`
}

type ImportTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Type        string  `json:"type"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
}

type ImportResult struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}

type TransactionRepository interface {
	GetByUserID(userID, limit, offset int, year, month string) ([]TransactionResponse, int, error)
	Create(userID int, req TransactionRequest) error
	Update(userID, id int, req TransactionRequest) error
	Delete(userID, id int) error
	GetMonthlyExpenses(userID int) (float64, error)
	GetMonthlyIncome(userID int) (float64, error)
	GetNetSavings(userID int) (float64, error)
	GetMonthlySummary(userID int, months int) ([]MonthlySummaryItem, error)
	GetYearlySummary(userID, year int) ([]MonthlySummaryItem, error)
	GetCategoryBreakdownDetailed(userID, month, year int) ([]CategoryBreakdownItem, error)
	BulkCreate(userID int, reqs []TransactionRequest) error
}
