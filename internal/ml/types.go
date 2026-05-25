package ml

// Transaction is the payload format required by the ML service.
// JSON tags must match the ML service OpenAPI schema exactly.
type Transaction struct {
	Date     string  `json:"date"`     // YYYY-MM-DD
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`     // "INCOME" | "EXPENSE"
	Category string  `json:"category"`
}

// AnalysisResponse mirrors POST /analysis on the ML service.
type AnalysisResponse struct {
	TotalExpense         float64            `json:"total_expense"`
	AvgDaily             float64            `json:"avg_daily"`
	TopCategory          *string            `json:"top_category"`
	SpendingDistribution map[string]float64 `json:"spending_distribution"`
}

// AnomalyRecord is a single flagged spending day.
type AnomalyRecord struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Severity string  `json:"severity"` // "low" | "medium" | "high" (z-score bands)
}

// AnomalyResponse mirrors POST /anomaly on the ML service.
type AnomalyResponse struct {
	Anomalies []AnomalyRecord `json:"anomalies"`
}

// ForecastRecord is a single day's predicted spend.
type ForecastRecord struct {
	Date            string  `json:"date"`
	PredictedAmount float64 `json:"predicted_amount"`
}

// ForecastResponse mirrors POST /forecast on the ML service.
type ForecastResponse struct {
	PredictedMonthlySpending float64          `json:"predicted_monthly_spending"`
	Confidence               float64          `json:"confidence"` // 0–1 Prophet uncertainty score
	Trend                    string           `json:"trend"`      // "increasing" | "decreasing" | "stable"
	DailyForecast            []ForecastRecord `json:"daily_forecast"`
}

// InsightItem is a single insight from the ML service.
type InsightItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` // "success" | "warning" | "info"
}

// InsightsResponse mirrors POST /insights on the ML service.
type InsightsResponse struct {
	Insights []InsightItem `json:"insights"`
}

// ForecastJobResponse mirrors POST /forecast/start on the ML service.
type ForecastJobResponse struct {
	JobID string `json:"job_id"`
}

// ForecastStatusResponse mirrors GET /forecast/status/{job_id} on the ML service.
type ForecastStatusResponse struct {
	JobID  string           `json:"job_id"`
	Status string           `json:"status"` // "pending" | "running" | "complete" | "failed"
	Result *ForecastResponse `json:"result"`
	Error  *string          `json:"error"`
}
