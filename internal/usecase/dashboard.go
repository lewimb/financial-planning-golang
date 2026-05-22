package usecase

import (
	"log"
	"time"

	"github.com/financial-planning/internal/domain"
)

type DashboardUseCase struct {
	txRepo     domain.TransactionRepository
	budgetRepo domain.BudgetRepository
	goalRepo   domain.GoalRepository
}

func NewDashboardUseCase(
	txRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
	goalRepo domain.GoalRepository,
) *DashboardUseCase {
	return &DashboardUseCase{txRepo: txRepo, budgetRepo: budgetRepo, goalRepo: goalRepo}
}

func (uc *DashboardUseCase) Get(userID int) (*domain.DashboardResponse, error) {
	now := time.Now()

	income, err := uc.txRepo.GetMonthlyIncome(userID)
	if err != nil {
		log.Printf("dashboard: GetMonthlyIncome userID=%d: %v", userID, err)
		return nil, err
	}

	expense, err := uc.txRepo.GetMonthlyExpenses(userID)
	if err != nil {
		log.Printf("dashboard: GetMonthlyExpenses userID=%d: %v", userID, err)
		return nil, err
	}

	netSavings, err := uc.txRepo.GetNetSavings(userID)
	if err != nil {
		log.Printf("dashboard: GetNetSavings userID=%d: %v", userID, err)
		return nil, err
	}

	budgetUsage, err := uc.budgetRepo.GetUsage(userID, int(now.Month()), now.Year())
	if err != nil {
		log.Printf("dashboard: GetBudgetUsage userID=%d: %v", userID, err)
		return nil, err
	}

	budgetSummary := domain.BudgetStatusSummary{Total: len(budgetUsage)}
	for _, b := range budgetUsage {
		switch b.Status {
		case "SAFE":
			budgetSummary.Safe++
		case "WARNING":
			budgetSummary.Warning++
		case "EXCEEDED":
			budgetSummary.Exceeded++
		}
	}

	activeGoals, err := uc.goalRepo.GetAll(userID, true)
	if err != nil {
		log.Printf("dashboard: GetGoals userID=%d: %v", userID, err)
		return nil, err
	}

	total, err := uc.goalRepo.CountActive(userID)
	if err != nil {
		log.Printf("dashboard: CountActiveGoals userID=%d: %v", userID, err)
		return nil, err
	}

	completed := 0
	for _, g := range activeGoals {
		if g.Status == "COMPLETED" {
			completed++
		}
	}

	health := computeFinancialHealth(income, netSavings, budgetSummary, activeGoals)

	hasAnomalies := false
	for _, b := range budgetUsage {
		if b.Status == "EXCEEDED" {
			hasAnomalies = true
			break
		}
	}

	return &domain.DashboardResponse{
		MonthlyIncome:  income,
		MonthlyExpense: expense,
		NetSavings:     netSavings,
		BudgetSummary:  budgetSummary,
		GoalSummary: domain.GoalProgressSummary{
			Total:     total,
			Active:    total - completed,
			Completed: completed,
		},
		ActiveGoals:     activeGoals,
		FinancialHealth: health,
		HasAnomalies:    hasAnomalies,
	}, nil
}

func computeFinancialHealth(income, netSavings float64, budgets domain.BudgetStatusSummary, goals []domain.GoalResponse) domain.FinancialHealth {
	var savingsRate float64
	if income > 0 {
		savingsRate = (netSavings / income) * 100
		if savingsRate > 100 {
			savingsRate = 100
		}
		if savingsRate < 0 {
			savingsRate = 0
		}
	}

	var budgetAdherence float64
	if budgets.Total > 0 {
		budgetAdherence = float64(budgets.Safe) / float64(budgets.Total) * 100
	} else {
		budgetAdherence = 100
	}

	var goalProgress float64
	if len(goals) > 0 {
		var total float64
		for _, g := range goals {
			if g.TargetAmount > 0 {
				pct := float64(g.CurrentAmount) / float64(g.TargetAmount) * 100
				if pct > 100 {
					pct = 100
				}
				total += pct
			}
		}
		goalProgress = total / float64(len(goals))
	}

	score := savingsRate*0.4 + budgetAdherence*0.35 + goalProgress*0.25

	var label string
	switch {
	case score >= 80:
		label = "Excellent"
	case score >= 60:
		label = "Good"
	case score >= 40:
		label = "Fair"
	default:
		label = "Needs Attention"
	}

	return domain.FinancialHealth{
		Score:           score,
		SavingsRate:     savingsRate,
		BudgetAdherence: budgetAdherence,
		GoalProgress:    goalProgress,
		Label:           label,
	}
}
