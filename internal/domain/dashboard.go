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

type DashboardResponse struct {
	MonthlyIncome  float64             `json:"monthly_income"`
	MonthlyExpense float64             `json:"monthly_expense"`
	NetSavings     float64             `json:"net_savings"`
	BudgetSummary  BudgetStatusSummary `json:"budget_summary"`
	GoalSummary    GoalProgressSummary `json:"goal_summary"`
	ActiveGoals    []GoalResponse      `json:"active_goals"`
}
