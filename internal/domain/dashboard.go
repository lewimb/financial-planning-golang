package domain

type BudgetStatusSummary struct {
	Total    int `json:"total"`
	Safe     int `json:"safe"`
	Warning  int `json:"warning"`
	Exceeded int `json:"exceeded"`
}

type GoalProgressSummary struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
}

type FinancialHealth struct {
	Score           float64 `json:"score"`
	SavingsRate     float64 `json:"savings_rate"`
	BudgetAdherence float64 `json:"budget_adherence"`
	GoalProgress    float64 `json:"goal_progress"`
	Label           string  `json:"label"`
}

type DashboardResponse struct {
	MonthlyIncome   float64             `json:"monthly_income"`
	MonthlyExpense  float64             `json:"monthly_expense"`
	NetSavings      float64             `json:"net_savings"`
	BudgetSummary   BudgetStatusSummary `json:"budget_summary"`
	GoalSummary     GoalProgressSummary `json:"goal_summary"`
	ActiveGoals     []GoalResponse      `json:"active_goals"`
	FinancialHealth FinancialHealth     `json:"financial_health"`
	HasAnomalies    bool                `json:"has_anomalies"`
}
