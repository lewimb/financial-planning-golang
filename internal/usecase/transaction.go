package usecase

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/financial-planning/internal/domain"
)

type TransactionUseCase struct {
	repo        domain.TransactionRepository
	activityRepo domain.ActivityLogRepository
}

func NewTransactionUseCase(repo domain.TransactionRepository, activityRepo domain.ActivityLogRepository) *TransactionUseCase {
	return &TransactionUseCase{repo: repo, activityRepo: activityRepo}
}

func (uc *TransactionUseCase) GetTransactions(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error) {
	txs, total, err := uc.repo.GetByUserID(userID, limit, offset, year, month)
	if err != nil {
		log.Printf("transaction: GetTransactions userID=%d: %v", userID, err)
	}
	return txs, total, err
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
	if req.IsRecurring {
		req.RecurrenceInterval = strings.ToUpper(req.RecurrenceInterval)
		valid := map[string]bool{"DAILY": true, "WEEKLY": true, "MONTHLY": true, "YEARLY": true}
		if !valid[req.RecurrenceInterval] {
			return errors.New("recurrence_interval must be DAILY, WEEKLY, MONTHLY, or YEARLY when is_recurring is true")
		}
	}
	if err := uc.repo.Create(userID, req); err != nil {
		log.Printf("transaction: Create userID=%d category=%s: %v", userID, req.Category, err)
		return err
	}
	if uc.activityRepo != nil {
		desc := fmt.Sprintf("Created %s transaction: %s %.0f", strings.ToLower(req.Type), req.Category, float64(req.Amount))
		_ = uc.activityRepo.Log(userID, "CREATE", "transaction", nil, desc)
	}
	return nil
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
	if err := uc.repo.Update(userID, id, req); err != nil {
		log.Printf("transaction: Update userID=%d id=%d: %v", userID, id, err)
		return err
	}
	if uc.activityRepo != nil {
		_ = uc.activityRepo.Log(userID, "UPDATE", "transaction", &id, fmt.Sprintf("Updated transaction #%d", id))
	}
	return nil
}

func (uc *TransactionUseCase) Delete(userID, id int) error {
	if err := uc.repo.Delete(userID, id); err != nil {
		log.Printf("transaction: Delete userID=%d id=%d: %v", userID, id, err)
		return err
	}
	if uc.activityRepo != nil {
		_ = uc.activityRepo.Log(userID, "DELETE", "transaction", &id, fmt.Sprintf("Deleted transaction #%d", id))
	}
	return nil
}

func (uc *TransactionUseCase) GetMonthlyExpenses(userID int) (float64, error) {
	v, err := uc.repo.GetMonthlyExpenses(userID)
	if err != nil {
		log.Printf("transaction: GetMonthlyExpenses userID=%d: %v", userID, err)
	}
	return v, err
}

func (uc *TransactionUseCase) GetMonthlyIncome(userID int) (float64, error) {
	v, err := uc.repo.GetMonthlyIncome(userID)
	if err != nil {
		log.Printf("transaction: GetMonthlyIncome userID=%d: %v", userID, err)
	}
	return v, err
}

func (uc *TransactionUseCase) GetMonthlySummary(userID int, months int) ([]domain.MonthlySummaryItem, error) {
	items, err := uc.repo.GetMonthlySummary(userID, months)
	if err != nil {
		log.Printf("transaction: GetMonthlySummary userID=%d months=%d: %v", userID, months, err)
	}
	return items, err
}

func (uc *TransactionUseCase) BulkImport(userID int, items []domain.ImportTransactionRequest) (*domain.ImportResult, error) {
	result := &domain.ImportResult{}
	var valid []domain.TransactionRequest

	for i, item := range items {
		txType := strings.ToUpper(item.Type)
		if txType != "INCOME" && txType != "EXPENSE" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: invalid type %q", i+1, item.Type))
			continue
		}
		if item.Amount <= 0 {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: amount must be > 0", i+1))
			continue
		}
		if item.Category == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: category is required", i+1))
			continue
		}
		t, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: invalid date %q (expected YYYY-MM-DD)", i+1, item.Date))
			continue
		}
		valid = append(valid, domain.TransactionRequest{
			Amount:      int(item.Amount),
			Category:    item.Category,
			Type:        txType,
			Date:        t,
			Description: item.Description,
		})
	}

	if len(valid) > 0 {
		if err := uc.repo.BulkCreate(userID, valid); err != nil {
			log.Printf("transaction: BulkImport BulkCreate userID=%d count=%d: %v", userID, len(valid), err)
			return nil, err
		}
		result.Imported = len(valid)
		if uc.activityRepo != nil {
			_ = uc.activityRepo.Log(userID, "IMPORT", "transaction", nil,
				fmt.Sprintf("Bulk imported %d transactions", result.Imported))
		}
	}
	return result, nil
}
