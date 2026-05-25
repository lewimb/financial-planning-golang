package usecase

import (
	"errors"
	"fmt"
	"log"

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
		log.Printf("ml: fetchMLTransactions userID=%d: %v", userID, err)
		return nil, fmt.Errorf("ml usecase: fetch transactions: %w", err)
	}

	result := make([]ml.Transaction, 0, len(txs))
	for _, t := range txs {
		result = append(result, ml.Transaction{
			Date:     t.Date.UTC().Format("2006-01-02"),
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
		log.Printf("ml: GetAnalysis userID=%d: %v", userID, err)
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
		log.Printf("ml: GetAnomaly userID=%d: %v", userID, err)
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
		log.Printf("ml: GetForecast userID=%d periods=%d: %v", userID, periods, err)
		return nil, ErrMLUnavailable
	}
	return result, nil
}

// GetInsights fetches the user's transactions and returns spending pattern insights.
func (uc *MLUseCase) GetInsights(userID int, year, month string) (*ml.InsightsResponse, error) {
	txs, err := uc.fetchMLTransactions(userID, year, month)
	if err != nil {
		return nil, err
	}
	result, err := uc.client.Insights(txs)
	if err != nil {
		log.Printf("ml: GetInsights userID=%d: %v", userID, err)
		return nil, ErrMLUnavailable
	}
	return result, nil
}

// StartForecast submits an async forecast job to the ML service and returns a job ID.
// Poll GetForecastStatus to retrieve the result.
func (uc *MLUseCase) StartForecast(userID int, periods int, year, month string) (*ml.ForecastJobResponse, error) {
	txs, err := uc.fetchMLTransactions(userID, year, month)
	if err != nil {
		return nil, err
	}
	result, err := uc.client.ForecastStart(txs, periods)
	if err != nil {
		log.Printf("ml: StartForecast userID=%d: %v", userID, err)
		return nil, ErrMLUnavailable
	}
	return result, nil
}

// GetForecastStatus polls the ML service for the status of a previously started forecast job.
func (uc *MLUseCase) GetForecastStatus(jobID string) (*ml.ForecastStatusResponse, error) {
	result, err := uc.client.ForecastStatus(jobID)
	if err != nil {
		log.Printf("ml: GetForecastStatus jobID=%s: %v", jobID, err)
		return nil, ErrMLUnavailable
	}
	return result, nil
}
