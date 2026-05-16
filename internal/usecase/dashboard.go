package usecase

import (
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
		return nil, err
	}

	expense, err := uc.txRepo.GetMonthlyExpenses(userID)
	if err != nil {
		return nil, err
	}

	netSavings, err := uc.txRepo.GetNetSavings(userID)
	if err != nil {
		return nil, err
	}

	budgetUsage, err := uc.budgetRepo.GetUsage(userID, int(now.Month()), now.Year())
	if err != nil {
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
		return nil, err
	}

	total, err := uc.goalRepo.CountActive(userID)
	if err != nil {
		return nil, err
	}

	completed := 0
	for _, g := range activeGoals {
		if g.Status == "COMPLETED" {
			completed++
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
		ActiveGoals: activeGoals,
	}, nil
}
