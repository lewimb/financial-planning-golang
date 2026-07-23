package usecase

import (
	"errors"
	"log"

	"github.com/financial-planning/internal/domain"
)

type BudgetUseCase struct {
	repo domain.BudgetRepository
}

func NewBudgetUseCase(repo domain.BudgetRepository) *BudgetUseCase {
	return &BudgetUseCase{repo: repo}
}

func (uc *BudgetUseCase) GetBudgets(userID int, category, month, year string) ([]domain.Budget, error) {
	budgets, err := uc.repo.GetAll(userID, category, month, year)
	if err != nil {
		log.Printf("budget: GetBudgets userID=%d: %v", userID, err)
	}
	return budgets, err
}

func (uc *BudgetUseCase) GetByID(id int) (*domain.BudgetResponse, error) {
	b, err := uc.repo.GetByID(id)
	if err != nil {
		log.Printf("budget: GetByID id=%d: %v", id, err)
	}
	return b, err
}

func (uc *BudgetUseCase) GetUsage(userID, month, year int) ([]domain.BudgetUsage, error) {
	usage, err := uc.repo.GetUsage(userID, month, year)
	if err != nil {
		log.Printf("budget: GetUsage userID=%d month=%d year=%d: %v", userID, month, year, err)
	}
	return usage, err
}

func (uc *BudgetUseCase) Create(userID int, req domain.CreateBudgetRequest) error {
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Period != "MONTHLY" && req.Period != "YEARLY" {
		return errors.New("invalid period")
	}
	if req.LimitAmount <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if req.Period == "MONTHLY" {
		// Monthly budgets recur every month — no fixed month/year is stored,
		// so any client-supplied month/year is ignored rather than validated.
		req.Month = nil
		req.Year = nil
	} else {
		// Yearly budgets are scoped to one calendar year.
		req.Month = nil
		if req.Year == nil || *req.Year == 0 {
			return errors.New("year is required for yearly budget")
		}
	}
	if req.AlertThreshold == 0 {
		req.AlertThreshold = 80
	}
	if err := uc.repo.Create(userID, req); err != nil {
		log.Printf("budget: Create userID=%d category=%s: %v", userID, req.Category, err)
		return err
	}
	return nil
}

func (uc *BudgetUseCase) Update(userID, id, limitAmount, alertThreshold int, category string) (*domain.UpdateBudgetResponse, error) {
	resp, err := uc.repo.Update(userID, id, limitAmount, alertThreshold, category)
	if err != nil {
		log.Printf("budget: Update userID=%d id=%d: %v", userID, id, err)
	}
	return resp, err
}

func (uc *BudgetUseCase) Delete(userID, id int) error {
	if err := uc.repo.Delete(userID, id); err != nil {
		log.Printf("budget: Delete userID=%d id=%d: %v", userID, id, err)
		return err
	}
	return nil
}
