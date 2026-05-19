package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/financial-planning/internal/domain"
)

type FinancialProfileUseCase struct {
	repo domain.FinancialProfileRepository
}

func NewFinancialProfileUseCase(repo domain.FinancialProfileRepository) *FinancialProfileUseCase {
	return &FinancialProfileUseCase{repo: repo}
}

func (uc *FinancialProfileUseCase) Upsert(userID int, req domain.UpsertFinancialProfileRequest) (*domain.FinancialProfileResponse, error) {
	if err := validateProfileRequest(req); err != nil {
		return nil, err
	}
	if err := uc.repo.Upsert(userID, req); err != nil {
		return nil, err
	}
	return uc.Get(userID)
}

func (uc *FinancialProfileUseCase) Get(userID int) (*domain.FinancialProfileResponse, error) {
	p, err := uc.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	p.NetAvailable = p.MonthlyIncome - p.FixedExpenses - p.Debt
	return p, nil
}

// BuildFinancialProfileContext returns a formatted string for Gemini prompt injection.
func BuildFinancialProfileContext(p *domain.FinancialProfileResponse) string {
	var sb strings.Builder
	sb.WriteString("User Financial Profile:\n")
	sb.WriteString(fmt.Sprintf("- Monthly Income: %.0f\n", p.MonthlyIncome))
	sb.WriteString(fmt.Sprintf("- Fixed Expenses: %.0f\n", p.FixedExpenses))
	sb.WriteString(fmt.Sprintf("- Current Savings: %.0f\n", p.CurrentSavings))
	sb.WriteString(fmt.Sprintf("- Debt: %.0f\n", p.Debt))
	sb.WriteString(fmt.Sprintf("- Net Available (income - expenses - debt): %.0f\n", p.NetAvailable))
	sb.WriteString(fmt.Sprintf("- Employment Status: %s\n", p.EmploymentStatus))
	if len(p.FinancialGoals) > 0 {
		sb.WriteString(fmt.Sprintf("- Financial Goals: %s\n", strings.Join(p.FinancialGoals, ", ")))
	}
	if p.SpendingHabit != nil && *p.SpendingHabit != "" {
		sb.WriteString(fmt.Sprintf("- Spending Habit: %s\n", *p.SpendingHabit))
	}
	if p.RiskLevel != nil && *p.RiskLevel != "" {
		sb.WriteString(fmt.Sprintf("- Risk Level: %s\n", *p.RiskLevel))
	}
	return sb.String()
}

func validateProfileRequest(req domain.UpsertFinancialProfileRequest) error {
	if req.MonthlyIncome < 0 {
		return errors.New("monthly_income must be >= 0")
	}
	if req.FixedExpenses < 0 {
		return errors.New("fixed_expenses must be >= 0")
	}
	if req.CurrentSavings < 0 {
		return errors.New("current_savings must be >= 0")
	}
	if req.Debt < 0 {
		return errors.New("debt must be >= 0")
	}
	if strings.TrimSpace(req.EmploymentStatus) == "" {
		return errors.New("employment_status is required")
	}
	if len(req.FinancialGoals) == 0 {
		return errors.New("financial_goals cannot be empty")
	}
	for _, g := range req.FinancialGoals {
		if strings.TrimSpace(g) == "" {
			return errors.New("financial_goals cannot contain blank entries")
		}
	}
	return nil
}
