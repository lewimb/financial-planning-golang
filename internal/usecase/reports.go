package usecase

import (
	"log"
	"time"

	"github.com/financial-planning/internal/domain"
)

type ReportsUseCase struct {
	txRepo     domain.TransactionRepository
	budgetRepo domain.BudgetRepository
	goalRepo   domain.GoalRepository
}

func NewReportsUseCase(
	txRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
	goalRepo domain.GoalRepository,
) *ReportsUseCase {
	return &ReportsUseCase{txRepo: txRepo, budgetRepo: budgetRepo, goalRepo: goalRepo}
}

func (uc *ReportsUseCase) GetMonthlySummary(userID int, months int) ([]domain.MonthlySummaryItem, error) {
	items, err := uc.txRepo.GetMonthlySummary(userID, months)
	if err != nil {
		log.Printf("reports: GetMonthlySummary userID=%d months=%d: %v", userID, months, err)
	}
	return items, err
}

func (uc *ReportsUseCase) GetCategoryBreakdown(userID int, year, month string) (map[string]float64, error) {
	txs, _, err := uc.txRepo.GetByUserID(userID, 0, 0, year, month)
	if err != nil {
		log.Printf("reports: GetCategoryBreakdown userID=%d year=%s month=%s: %v", userID, year, month, err)
		return nil, err
	}

	breakdown := make(map[string]float64)
	var totalExpense float64
	for _, t := range txs {
		if t.Type == "EXPENSE" {
			breakdown[t.Category] += t.Amount
			totalExpense += t.Amount
		}
	}

	if totalExpense > 0 {
		for cat := range breakdown {
			breakdown[cat] = breakdown[cat] / totalExpense * 100
		}
	}
	return breakdown, nil
}

type SavingsRatePoint struct {
	Month        int     `json:"month"`
	Year         int     `json:"year"`
	SavingsRate  float64 `json:"savings_rate"`
	NetSavings   float64 `json:"net_savings"`
	Income       float64 `json:"income"`
}

func (uc *ReportsUseCase) GetSavingsRate(userID int, months int) ([]SavingsRatePoint, error) {
	summary, err := uc.txRepo.GetMonthlySummary(userID, months)
	if err != nil {
		log.Printf("reports: GetSavingsRate userID=%d months=%d: %v", userID, months, err)
		return nil, err
	}

	points := make([]SavingsRatePoint, 0, len(summary))
	for _, s := range summary {
		var rate float64
		net := s.Income - s.Expense
		if s.Income > 0 {
			rate = net / s.Income * 100
		}
		points = append(points, SavingsRatePoint{
			Month:       s.Month,
			Year:        s.Year,
			SavingsRate: rate,
			NetSavings:  net,
			Income:      s.Income,
		})
	}
	return points, nil
}

type NetWorthPoint struct {
	Month    int     `json:"month"`
	Year     int     `json:"year"`
	NetWorth float64 `json:"net_worth"`
}

func (uc *ReportsUseCase) GetNetWorth(userID int, months int) ([]NetWorthPoint, error) {
	summary, err := uc.txRepo.GetMonthlySummary(userID, months)
	if err != nil {
		log.Printf("reports: GetNetWorth userID=%d months=%d: %v", userID, months, err)
		return nil, err
	}

	points := make([]NetWorthPoint, 0, len(summary))
	var cumulative float64
	for _, s := range summary {
		cumulative += s.Income - s.Expense
		points = append(points, NetWorthPoint{
			Month:    s.Month,
			Year:     s.Year,
			NetWorth: cumulative,
		})
	}
	return points, nil
}

type MonthComparisonResponse struct {
	CurrentMonth  domain.MonthlySummaryItem `json:"current_month"`
	PreviousMonth domain.MonthlySummaryItem `json:"previous_month"`
	IncomeChange  float64                   `json:"income_change_pct"`
	ExpenseChange float64                   `json:"expense_change_pct"`
}

func (uc *ReportsUseCase) GetMonthComparison(userID int) (*MonthComparisonResponse, error) {
	now := time.Now()
	summary, err := uc.txRepo.GetMonthlySummary(userID, 2)
	if err != nil {
		log.Printf("reports: GetMonthComparison userID=%d: %v", userID, err)
		return nil, err
	}

	resp := &MonthComparisonResponse{}
	for _, s := range summary {
		if s.Month == int(now.Month()) && s.Year == now.Year() {
			resp.CurrentMonth = s
		} else {
			resp.PreviousMonth = s
		}
	}

	if resp.PreviousMonth.Income > 0 {
		resp.IncomeChange = (resp.CurrentMonth.Income - resp.PreviousMonth.Income) / resp.PreviousMonth.Income * 100
	}
	if resp.PreviousMonth.Expense > 0 {
		resp.ExpenseChange = (resp.CurrentMonth.Expense - resp.PreviousMonth.Expense) / resp.PreviousMonth.Expense * 100
	}
	return resp, nil
}
