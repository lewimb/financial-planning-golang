package domain

import "time"

type Budget struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Category       string    `json:"category"`
	Period         string    `json:"period"`
	Month          *int      `json:"month"`
	Year           *int      `json:"year"`
	LimitAmount    int       `json:"limit_amount"`
	AlertThreshold int       `json:"alert_threshold"`
	CreatedAt      time.Time `json:"created_at"`
}

type BudgetResponse struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	Category       string    `json:"category"`
	Period         string    `json:"period"`
	Month          *int      `json:"month"`
	Year           *int      `json:"year"`
	LimitAmount    int       `json:"limitAmount"`
	AlertThreshold int       `json:"alertThreshold"`
	CreatedAt      time.Time `json:"createdAt"`
}

type UpdateBudgetRequest struct {
	LimitAmount    int    `json:"limitAmount"`
	AlertThreshold int    `json:"alertThreshold"`
	Category       string `json:"category"`
}

type UpdateBudgetResponse struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	Category       string `json:"category"`
	Period         string `json:"period"`
	Month          *int   `json:"month"`
	Year           *int   `json:"year"`
	LimitAmount    int    `json:"limit_amount"`
	AlertThreshold int    `json:"alert_threshold"`
}

type CreateBudgetRequest struct {
	Category       string `json:"category"`
	Period         string `json:"period"`
	Month          *int   `json:"month"`
	Year           *int   `json:"year"`
	LimitAmount    int    `json:"limit_amount"`
	AlertThreshold int    `json:"alert_threshold"`
}

type BudgetUsage struct {
	BudgetID       int     `json:"budget_id"`
	Category       string  `json:"category"`
	Period         string  `json:"period"`
	Limit          int64   `json:"limit"`
	AlertThreshold int     `json:"alert_threshold"`
	Used           int64   `json:"used"`
	Remaining      int64   `json:"remaining"`
	Percentage     float64 `json:"percentage"`
	Status         string  `json:"status"`
	ChangePercent  float64 `json:"change_percent"`
}

type BudgetRepository interface {
	GetAll(userID int, category, month, year string) ([]Budget, error)
	GetByID(id int) (*BudgetResponse, error)
	GetUsage(userID, month, year int) ([]BudgetUsage, error)
	Create(userID int, req CreateBudgetRequest) error
	Update(userID, id, limitAmount, alertThreshold int, category string) (*UpdateBudgetResponse, error)
	Delete(userID, id int) error
}
