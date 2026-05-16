package usecase

import (
	"errors"
	"strings"

	"github.com/financial-planning/internal/domain"
)

type TransactionUseCase struct {
	repo domain.TransactionRepository
}

func NewTransactionUseCase(repo domain.TransactionRepository) *TransactionUseCase {
	return &TransactionUseCase{repo: repo}
}

func (uc *TransactionUseCase) GetTransactions(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error) {
	return uc.repo.GetByUserID(userID, limit, offset, year, month)
}

func (uc *TransactionUseCase) Create(userID int, req domain.TransactionRequest) error {
	req.Type = strings.ToUpper(req.Type)
	if req.Type != "INCOME" && req.Type != "EXPENSE" {
		return errors.New("invalid transaction type: must be INCOME or EXPENSE")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Date.IsZero() {
		return errors.New("date is required")
	}
	return uc.repo.Create(userID, req)
}

func (uc *TransactionUseCase) Update(userID, id int, req domain.TransactionRequest) error {
	req.Type = strings.ToUpper(req.Type)
	if req.Type != "INCOME" && req.Type != "EXPENSE" {
		return errors.New("invalid transaction type: must be INCOME or EXPENSE")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Date.IsZero() {
		return errors.New("date is required")
	}
	return uc.repo.Update(userID, id, req)
}

func (uc *TransactionUseCase) Delete(userID, id int) error {
	return uc.repo.Delete(userID, id)
}

func (uc *TransactionUseCase) GetMonthlyExpenses(userID int) (float64, error) {
	return uc.repo.GetMonthlyExpenses(userID)
}

func (uc *TransactionUseCase) GetMonthlyIncome(userID int) (float64, error) {
	return uc.repo.GetMonthlyIncome(userID)
}
