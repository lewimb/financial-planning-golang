package usecase

import (
	"errors"

	"github.com/financial-planning/internal/domain"
)

type BudgetUseCase struct {
	repo domain.BudgetRepository
}

func NewBudgetUseCase(repo domain.BudgetRepository) *BudgetUseCase {
	return &BudgetUseCase{repo: repo}
}

func (uc *BudgetUseCase) GetBudgets(userID int, category, month, year string) ([]domain.Budget, error) {
	return uc.repo.GetAll(userID, category, month, year)
}

func (uc *BudgetUseCase) GetByID(id int) (*domain.BudgetResponse, error) {
	return uc.repo.GetByID(id)
}

func (uc *BudgetUseCase) GetUsage(userID, month, year int) ([]domain.BudgetUsage, error) {
	return uc.repo.GetUsage(userID, month, year)
}

func (uc *BudgetUseCase) Create(userID int, req domain.CreateBudgetRequest) error {
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Period != "MONTHLY" && req.Period != "YEARLY" {
		return errors.New("invalid period")
	}
	if req.Year == 0 {
		return errors.New("year is required")
	}
	if req.LimitAmount <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if req.Period == "MONTHLY" && req.Month == nil {
		return errors.New("month required for monthly budget")
	}
	if req.Period == "YEARLY" {
		req.Month = nil
	}
	if req.AlertThreshold == 0 {
		req.AlertThreshold = 80
	}
	return uc.repo.Create(userID, req)
}

func (uc *BudgetUseCase) Update(userID, id, limitAmount, alertThreshold int, category string) (*domain.UpdateBudgetResponse, error) {
	return uc.repo.Update(userID, id, limitAmount, alertThreshold, category)
}

func (uc *BudgetUseCase) Delete(userID, id int) error {
	return uc.repo.Delete(userID, id)
}
