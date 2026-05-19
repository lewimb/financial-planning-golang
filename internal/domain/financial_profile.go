package domain

import "time"

type FinancialProfileResponse struct {
	MonthlyIncome    float64  `json:"monthly_income"`
	FixedExpenses    float64  `json:"fixed_expenses"`
	CurrentSavings   float64  `json:"current_savings"`
	Debt             float64  `json:"debt"`
	EmploymentStatus string   `json:"employment_status"`
	SpendingHabit    *string  `json:"spending_habit"`
	RiskLevel        *string  `json:"risk_level"`
	FinancialGoals   []string `json:"financial_goals"`
	NetAvailable     float64  `json:"net_available"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UpsertFinancialProfileRequest struct {
	MonthlyIncome    float64  `json:"monthly_income"`
	FixedExpenses    float64  `json:"fixed_expenses"`
	CurrentSavings   float64  `json:"current_savings"`
	Debt             float64  `json:"debt"`
	EmploymentStatus string   `json:"employment_status"`
	FinancialGoals   []string `json:"financial_goals"`
	SpendingHabit    *string  `json:"spending_habit"`
	RiskLevel        *string  `json:"risk_level"`
}

type FinancialProfileRepository interface {
	Upsert(userID int, req UpsertFinancialProfileRequest) error
	GetByUserID(userID int) (*FinancialProfileResponse, error)
}
