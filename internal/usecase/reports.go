package usecase

import (
	"log"
	"time"

	"github.com/financial-planning/internal/domain"
)

var monthNames = [13]string{
	"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

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
	Month       int     `json:"month"`
	Year        int     `json:"year"`
	SavingsRate float64 `json:"savings_rate"`
	NetSavings  float64 `json:"net_savings"`
	Income      float64 `json:"income"`
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

// --- Year-based report types (frontend-compatible format) ---

type IncomeExpenseTrendPoint struct {
	Month     int     `json:"month"`
	MonthName string  `json:"month_name"`
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Savings   float64 `json:"savings"`
}

type IncomeExpenseTrendResponse struct {
	Year int                       `json:"year"`
	Data []IncomeExpenseTrendPoint `json:"data"`
}

func (uc *ReportsUseCase) GetIncomeExpenseTrend(userID, year int) (*IncomeExpenseTrendResponse, error) {
	summary, err := uc.txRepo.GetYearlySummary(userID, year)
	if err != nil {
		log.Printf("reports: GetIncomeExpenseTrend userID=%d year=%d: %v", userID, year, err)
		return nil, err
	}

	data := make([]IncomeExpenseTrendPoint, 0, len(summary))
	for _, s := range summary {
		name := ""
		if s.Month >= 1 && s.Month <= 12 {
			name = monthNames[s.Month]
		}
		data = append(data, IncomeExpenseTrendPoint{
			Month:     s.Month,
			MonthName: name,
			Income:    s.Income,
			Expense:   s.Expense,
			Savings:   s.Income - s.Expense,
		})
	}
	return &IncomeExpenseTrendResponse{Year: year, Data: data}, nil
}

type NetWorthHistoryPoint struct {
	Month     int     `json:"month"`
	MonthName string  `json:"month_name"`
	NetWorth  float64 `json:"net_worth"`
}

type NetWorthHistoryResponse struct {
	Year int                    `json:"year"`
	Data []NetWorthHistoryPoint `json:"data"`
}

func (uc *ReportsUseCase) GetNetworthHistory(userID, year int) (*NetWorthHistoryResponse, error) {
	summary, err := uc.txRepo.GetYearlySummary(userID, year)
	if err != nil {
		log.Printf("reports: GetNetworthHistory userID=%d year=%d: %v", userID, year, err)
		return nil, err
	}

	data := make([]NetWorthHistoryPoint, 0, len(summary))
	var cumulative float64
	for _, s := range summary {
		cumulative += s.Income - s.Expense
		name := ""
		if s.Month >= 1 && s.Month <= 12 {
			name = monthNames[s.Month]
		}
		data = append(data, NetWorthHistoryPoint{
			Month:     s.Month,
			MonthName: name,
			NetWorth:  cumulative,
		})
	}
	return &NetWorthHistoryResponse{Year: year, Data: data}, nil
}

type SavingsRateHistoryPoint struct {
	Month     int     `json:"month"`
	MonthName string  `json:"month_name"`
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Rate      float64 `json:"rate"`
}

type SavingsRateHistoryResponse struct {
	Year int                       `json:"year"`
	Data []SavingsRateHistoryPoint `json:"data"`
}

func (uc *ReportsUseCase) GetSavingsRateHistory(userID, year int) (*SavingsRateHistoryResponse, error) {
	summary, err := uc.txRepo.GetYearlySummary(userID, year)
	if err != nil {
		log.Printf("reports: GetSavingsRateHistory userID=%d year=%d: %v", userID, year, err)
		return nil, err
	}

	data := make([]SavingsRateHistoryPoint, 0, len(summary))
	for _, s := range summary {
		var rate float64
		if s.Income > 0 {
			rate = (s.Income - s.Expense) / s.Income * 100
		}
		name := ""
		if s.Month >= 1 && s.Month <= 12 {
			name = monthNames[s.Month]
		}
		data = append(data, SavingsRateHistoryPoint{
			Month:     s.Month,
			MonthName: name,
			Income:    s.Income,
			Expense:   s.Expense,
			Rate:      rate,
		})
	}
	return &SavingsRateHistoryResponse{Year: year, Data: data}, nil
}

type MonthComparisonPeriod struct {
	Month   int     `json:"month"`
	Year    int     `json:"year"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Savings float64 `json:"savings"`
}

type MonthComparisonChanges struct {
	IncomePct  float64 `json:"income_pct"`
	ExpensePct float64 `json:"expense_pct"`
	SavingsPct float64 `json:"savings_pct"`
}

type NewMonthComparisonResponse struct {
	Current  MonthComparisonPeriod  `json:"current"`
	Previous MonthComparisonPeriod  `json:"previous"`
	Changes  MonthComparisonChanges `json:"changes"`
}

func (uc *ReportsUseCase) GetMonthComparisonByDate(userID, month, year int) (*NewMonthComparisonResponse, error) {
	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear = year - 1
	}

	curSummary, err := uc.txRepo.GetYearlySummary(userID, year)
	if err != nil {
		log.Printf("reports: GetMonthComparisonByDate userID=%d: %v", userID, err)
		return nil, err
	}
	prevSummary, err := uc.txRepo.GetYearlySummary(userID, prevYear)
	if err != nil {
		log.Printf("reports: GetMonthComparisonByDate prev userID=%d: %v", userID, err)
		return nil, err
	}

	findMonth := func(items []domain.MonthlySummaryItem, m int) domain.MonthlySummaryItem {
		for _, s := range items {
			if s.Month == m {
				return s
			}
		}
		return domain.MonthlySummaryItem{}
	}

	cur := findMonth(curSummary, month)
	prev := findMonth(prevSummary, prevMonth)

	resp := &NewMonthComparisonResponse{
		Current: MonthComparisonPeriod{
			Month: month, Year: year,
			Income: cur.Income, Expense: cur.Expense, Savings: cur.Income - cur.Expense,
		},
		Previous: MonthComparisonPeriod{
			Month: prevMonth, Year: prevYear,
			Income: prev.Income, Expense: prev.Expense, Savings: prev.Income - prev.Expense,
		},
	}

	pct := func(cur, prev float64) float64 {
		if prev == 0 {
			return 0
		}
		return (cur - prev) / prev * 100
	}
	resp.Changes = MonthComparisonChanges{
		IncomePct:  pct(resp.Current.Income, resp.Previous.Income),
		ExpensePct: pct(resp.Current.Expense, resp.Previous.Expense),
		SavingsPct: pct(resp.Current.Savings, resp.Previous.Savings),
	}
	return resp, nil
}

type CategoryBreakdownResponse struct {
	Period       struct{ Month, Year int }       `json:"period"`
	TotalExpense float64                         `json:"total_expense"`
	Data         []domain.CategoryBreakdownItem  `json:"data"`
}

func (uc *ReportsUseCase) GetCategoryBreakdownDetailed(userID, month, year int) (*CategoryBreakdownResponse, error) {
	items, err := uc.txRepo.GetCategoryBreakdownDetailed(userID, month, year)
	if err != nil {
		log.Printf("reports: GetCategoryBreakdownDetailed userID=%d month=%d year=%d: %v", userID, month, year, err)
		return nil, err
	}
	var total float64
	for _, item := range items {
		total += item.Total
	}
	resp := &CategoryBreakdownResponse{
		TotalExpense: total,
		Data:         items,
	}
	resp.Period.Month = month
	resp.Period.Year = year
	return resp, nil
}
