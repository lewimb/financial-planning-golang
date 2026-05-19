package usecase

import (
	"errors"
	"fmt"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/ml"
)

// ErrMLUnavailable is returned when the ML service cannot be reached or returns an error.
var ErrMLUnavailable = errors.New("ML service unavailable")

// MLUseCase orchestrates: fetch transactions from DB → convert → call ML client → return result.
type MLUseCase struct {
	txRepo domain.TransactionRepository
	client *ml.Client
}

func NewMLUseCase(txRepo domain.TransactionRepository, client *ml.Client) *MLUseCase {
	return &MLUseCase{txRepo: txRepo, client: client}
}

// fetchMLTransactions retrieves all non-deleted transactions for the user (scoped to
// the supplied year/month if provided) and converts them to the ML service format.
// limit=0 means no LIMIT clause — all matching rows are returned.
func (uc *MLUseCase) fetchMLTransactions(userID int, year, month string) ([]ml.Transaction, error) {
	txs, _, err := uc.txRepo.GetByUserID(userID, 0, 0, year, month)
	if err != nil {
		return nil, fmt.Errorf("ml usecase: fetch transactions: %w", err)
	}

	result := make([]ml.Transaction, 0, len(txs))
	for _, t := range txs {
		result = append(result, ml.Transaction{
			Date:     t.Date.Format("2006-01-02"),
			Amount:   t.Amount,
			Type:     t.Type,
			Category: t.Category,
		})
	}
	return result, nil
}

// GetAnalysis fetches the user's transactions and returns spending analysis from the ML service.
// year and month are optional filters (empty string = no filter).
func (uc *MLUseCase) GetAnalysis(userID int, year, month string) (*ml.AnalysisResponse, error) {
	txs, err := uc.fetchMLTransactions(userID, year, month)
	if err != nil {
		return nil, err
	}
	result, err := uc.client.Analysis(txs)
	if err != nil {
		return nil, ErrMLUnavailable
	}
	return result, nil
}

// GetAnomaly fetches the user's transactions and returns anomaly detection results.
// Requires at least 5 unique expense days — returns empty anomaly list otherwise.
func (uc *MLUseCase) GetAnomaly(userID int, year, month string) (*ml.AnomalyResponse, error) {
	txs, err := uc.fetchMLTransactions(userID, year, month)
	if err != nil {
		return nil, err
	}
	result, err := uc.client.Anomaly(txs)
	if err != nil {
		return nil, ErrMLUnavailable
	}
	return result, nil
}

// GetForecast fetches the user's transactions and returns a daily spending forecast.
// periods controls how many days ahead to forecast (1–365, default 30).
func (uc *MLUseCase) GetForecast(userID int, periods int, year, month string) (*ml.ForecastResponse, error) {
	txs, err := uc.fetchMLTransactions(userID, year, month)
	if err != nil {
		return nil, err
	}
	result, err := uc.client.Forecast(txs, periods)
	if err != nil {
		return nil, ErrMLUnavailable
	}
	return result, nil
}
