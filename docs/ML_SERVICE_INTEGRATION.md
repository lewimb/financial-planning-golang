# ML Service — Go Backend Integration Guide

## Architecture Overview

```
┌─────────────┐     transactions JSON      ┌──────────────────┐
│  Go Backend │ ─────────────────────────► │   ML Service     │
│             │ ◄───────────────────────── │  (Python/FastAPI)│
└─────────────┘     analysis results       └──────────────────┘
       │
       │ fetches
       ▼
┌─────────────┐
│  Database   │
└─────────────┘
```

**Key principle:** The Go backend owns data. The ML service owns computation. No shared state between them.

- Go backend: authentication, DB access, business logic, routing
- ML service: pure functions — input in, result out

---

## What Each Endpoint Does

| Endpoint | Input | Output | Use case |
|----------|-------|--------|----------|
| `POST /analysis` | Transaction array | Totals + category breakdown | Dashboard summary, chatbot context |
| `POST /anomaly` | Transaction array | Flagged unusual days | Spending alerts, user notifications |
| `POST /forecast` | Transaction array | Predicted daily spend | Budget planning, AI predictions |

---

## Request Flow

```
1. User requests financial summary
2. Go backend fetches transactions from DB
3. Go backend serializes to JSON array
4. Go backend POSTs to ML service endpoint
5. ML service returns computed result
6. Go backend returns result to frontend / passes to AI
```

---

## Go Code Examples

### Shared types

```go
package ml

type Transaction struct {
    Date     string  `json:"date"`
    Amount   float64 `json:"amount"`
    Type     string  `json:"type"`
    Category string  `json:"category"`
}

type AnalysisResult struct {
    TotalExpense        float64            `json:"total_expense"`
    AvgDaily            float64            `json:"avg_daily"`
    TopCategory         *string            `json:"top_category"`
    SpendingDistribution map[string]float64 `json:"spending_distribution"`
}

type AnomalyRecord struct {
    Date   string  `json:"date"`
    Amount float64 `json:"amount"`
}

type AnomalyResult struct {
    Anomalies []AnomalyRecord `json:"anomalies"`
    Summary   string          `json:"summary"`
}

type ForecastRecord struct {
    Date            string  `json:"date"`
    PredictedAmount float64 `json:"predicted_amount"`
}

type ForecastResult struct {
    PredictedMonthlySpending float64          `json:"predicted_monthly_spending"`
    DailyForecast            []ForecastRecord `json:"daily_forecast"`
}
```

### POST /analysis

```go
func GetAnalysis(transactions []Transaction) (*AnalysisResult, error) {
    body, err := json.Marshal(transactions)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "http://ml-service:8000/analysis",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result AnalysisResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

### POST /anomaly

```go
func GetAnomalies(transactions []Transaction) (*AnomalyResult, error) {
    body, _ := json.Marshal(transactions)

    resp, err := http.Post(
        "http://ml-service:8000/anomaly",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result AnomalyResult
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

### POST /forecast (with periods parameter)

```go
func GetForecast(transactions []Transaction, periods int) (*ForecastResult, error) {
    body, _ := json.Marshal(transactions)

    url := fmt.Sprintf("http://ml-service:8000/forecast?periods=%d", periods)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ForecastResult
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

---

## Recommended Timeout Values

| Endpoint | Recommended timeout |
|----------|-------------------|
| `/analysis` | 5s |
| `/anomaly` | 10s |
| `/forecast` | 60s |

Forecast is slower because Prophet fits a model on each request. With 30 days of data, typical response time is 2–5 seconds.

---

## Transaction Mapping

Map your DB model fields to these JSON fields:

| ML service field | Type | Notes |
|-----------------|------|-------|
| `date` | string `YYYY-MM-DD` | Date of transaction |
| `amount` | float | Raw amount — no currency conversion done by ML service |
| `type` | string | Must be `"EXPENSE"` or `"INCOME"` |
| `category` | string | Any label — passed through as-is |

Only `EXPENSE` records affect results. `INCOME` records are filtered internally on every endpoint.

---

## Error Handling

The ML service always returns HTTP 200 with an empty/zero result if input is insufficient. It does **not** return 4xx for empty arrays or edge cases.

Check for meaningful data before displaying results:

```go
result, _ := GetAnalysis(transactions)
if result.TotalExpense == 0 {
    // No expense data — skip rendering or show placeholder
}

anomalies, _ := GetAnomalies(transactions)
if len(anomalies.Anomalies) > 0 {
    // Send user an alert
}
```

---

## Important Notes

- **Stateless:** The ML service holds no data between requests. Send all transactions every time.
- **No auth:** The ML service is an internal service — add network-level access control if needed.
- **Go is source of truth:** Never write data to the ML service. It reads, computes, returns.
- **Idempotent:** Calling the same endpoint twice with the same payload always returns the same result.
