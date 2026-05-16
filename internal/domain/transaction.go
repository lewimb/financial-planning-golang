package domain

import "time"

type TransactionRequest struct {
	Amount      int       `json:"amount"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type TransactionResponse struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type TransactionRepository interface {
	GetByUserID(userID, limit, offset int, year, month string) ([]TransactionResponse, int, error)
	Create(userID int, req TransactionRequest) error
	Update(userID, id int, req TransactionRequest) error
	Delete(userID, id int) error
	GetMonthlyExpenses(userID int) (float64, error)
	GetMonthlyIncome(userID int) (float64, error)
	GetNetSavings(userID int) (float64, error)
}
