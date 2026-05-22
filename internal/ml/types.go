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

// InsightsResponse mirrors POST /insights on the ML service.
type InsightsResponse struct {
	TopCategory       *string            `json:"top_category"`
	CategoryBreakdown map[string]float64 `json:"category_breakdown"` // % per category, sums to ~100
	SpikeCategory     *string            `json:"spike_category"`     // null if no spike detected
}
