# Forecasting & ML Pipeline Flow

Derived from `internal/usecase/ml.go`, `internal/ml/client.go`, `internal/ml/types.go`.

---

## 1. Complete Forecast Pipeline

```mermaid
flowchart TD
    A([GET /api/auth/v1/ml/forecast]) --> B[JWT → userID]
    B --> C[Parse query params\nyear · month optional\nperiods default 30]
    C --> D{periods > 0 and valid int?}
    D -- No/missing --> E[periods = 30]
    D -- Yes --> F[periods = parsed value]
    E --> G[MLUseCase.GetForecast\nuserID · periods · year · month]
    F --> G

    G --> H[fetchMLTransactions\nuserID · year · month]
    H --> I[TransactionRepository.GetByUserID\nlimit=0 offset=0\nno LIMIT clause — ALL rows]
    I --> J[PostgreSQL\nSELECT all non-deleted transactions\nordered by date DESC]
    J --> K[Convert domain.TransactionResponse\n→ ml.Transaction\ndate: YYYY-MM-DD · amount · type · category]

    K --> L[ml.Client.Forecast\ntransactions · periods]
    L --> M[Clamp periods to 1–365]
    M --> N[context.WithTimeout 60 seconds]
    N --> O[POST ML_SERVICE_URL/forecast?periods=N\nContent-Type: application/json\nBody: JSON array of transactions]

    O --> P{HTTP 200?}
    P -- No --> Q[fmt.Errorf unexpected status]
    Q --> R[MLUseCase wraps → ErrMLUnavailable]
    R --> S[503 ML service unavailable]

    P -- Yes --> T[json.Decode → ForecastResponse]
    T --> U[200 Response Body\npredicted_monthly_spending float64\ndaily_forecast array date + predicted_amount]
```

---

## 2. Spending Analysis Pipeline

```mermaid
flowchart TD
    A([GET /api/auth/v1/ml/analysis]) --> B[JWT → userID]
    B --> C[Parse year · month optional]
    C --> D[MLUseCase.GetAnalysis]
    D --> E[fetchMLTransactions same as forecast]
    E --> F[ml.Client.Analysis\ncontext.WithTimeout 5 seconds]
    F --> G[POST ML_SERVICE_URL/analysis\nJSON transaction array]
    G --> H{HTTP 200?}
    H -- No --> I[503 service unavailable]
    H -- Yes --> J[Decode AnalysisResponse]
    J --> K[200 total_expense · avg_daily\ntop_category · spending_distribution map]
```

---

## 3. Anomaly Detection Pipeline

```mermaid
flowchart TD
    A([GET /api/auth/v1/ml/anomaly]) --> B[JWT → userID]
    B --> C[Parse year · month optional]
    C --> D[MLUseCase.GetAnomaly]
    D --> E[fetchMLTransactions same as forecast]
    E --> F[ml.Client.Anomaly\ncontext.WithTimeout 10 seconds]
    F --> G[POST ML_SERVICE_URL/anomaly\nJSON transaction array]
    G --> H{HTTP 200?}
    H -- No --> I[503 service unavailable]
    H -- Yes --> J[Decode AnomalyResponse]
    J --> K[200 anomalies array date + amount\nsummary string]
```

> **ML Service note:** The Python FastAPI service at `:8000` performs the actual Prophet-based forecasting and statistical anomaly detection. The Go backend acts as a proxy — it fetches transaction data from PostgreSQL, transforms it to the ML schema, and forwards it. No ML computation happens in Go.

---

## 4. Data Transformation: Domain → ML Schema

```mermaid
flowchart LR
    subgraph Domain ["domain.TransactionResponse"]
        D1[ID int]
        D2[Amount float64]
        D3[Category string]
        D4[Type string]
        D5[Date time.Time]
        D6[Description string]
    end
    subgraph ML ["ml.Transaction"]
        M1[Date string YYYY-MM-DD]
        M2[Amount float64]
        M3[Type string INCOME/EXPENSE]
        M4[Category string]
    end
    D5 -->|".Format('2006-01-02')"| M1
    D2 --> M2
    D4 --> M3
    D3 --> M4
```

---

## 5. ML Client Timeout Strategy

```mermaid
flowchart LR
    A[/analysis] -->|context timeout| B[5 seconds]
    C[/anomaly] -->|context timeout| D[10 seconds]
    E[/forecast] -->|context timeout| F[60 seconds]
    B --> G[Fast — simple aggregation]
    D --> H[Medium — statistical detection]
    F --> I[Slow — Prophet time-series model]
```

---

## 6. Error Handling — ML Unavailable

```mermaid
flowchart TD
    A[ml.Client method returns error] --> B[MLUseCase wraps error]
    B --> C[return ErrMLUnavailable]
    C --> D{errors.Is ErrMLUnavailable?}
    D -- Yes --> E[503 Service Unavailable\nML service unavailable]
    D -- No --> F[500 Internal Server Error]
```
