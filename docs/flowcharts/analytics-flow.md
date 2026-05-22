# Analytics Flow

Covers Dashboard aggregation and ML analytics. Derived from `internal/usecase/dashboard.go`, `internal/usecase/ml.go`, and `internal/repository/postgres/`.

---

## 1. Dashboard Aggregation

```mermaid
flowchart TD
    A([GET /api/auth/v1/dashboard]) --> B[JWT → userID]
    B --> C[DashboardUseCase.Get\ntime.Now determines current month+year]

    C --> D1[TransactionRepository.GetMonthlyIncome\nSELECT SUM INCOME current month]
    C --> D2[TransactionRepository.GetMonthlyExpenses\nSELECT SUM EXPENSE current month]
    C --> D3[TransactionRepository.GetNetSavings\nSELECT SUM INCOME - SUM EXPENSE all-time]

    D1 --> E[income float64]
    D2 --> F[expense float64]
    D3 --> G[netSavings float64]

    E --> H[BudgetRepository.GetUsage\ncurrentMonth · currentYear]
    F --> H
    G --> H

    H --> I[budgetUsage BudgetUsage array\nwith computed status per budget]

    I --> J[Aggregate BudgetStatusSummary\ncount each SAFE · WARNING · EXCEEDED]

    J --> K[GoalRepository.GetAll\nuserID active=true\nWHERE deadline >= NOW]

    K --> L[activeGoals GoalResponse array]

    L --> M[GoalRepository.CountActive\nSELECT COUNT WHERE deadline >= NOW]

    M --> N[total int]

    N --> O[Count COMPLETED goals\nfrom activeGoals slice in-memory]

    O --> P[Build DashboardResponse\nmonthly_income · monthly_expense\nnet_savings · budget_summary\ngoal_summary · active_goals]

    P --> Q[200 response]
```

> **Sequential queries:** All 6 DB calls in Dashboard run sequentially — no goroutine parallelism. A future improvement could use `errgroup` to run them concurrently.

---

## 2. Dashboard Response Structure

```mermaid
flowchart LR
    subgraph DashboardResponse
        A[monthly_income float64\ncurrent month SUM INCOME]
        B[monthly_expense float64\ncurrent month SUM EXPENSE]
        C[net_savings float64\nall-time net position]
        subgraph BudgetSummary
            D[total int]
            E[safe int]
            F[warning int]
            G[exceeded int]
        end
        subgraph GoalSummary
            H[total int\nactive goals count]
            I[active = total - completed]
            J[completed int]
        end
        K[active_goals GoalResponse array\ndeadline >= NOW]
    end
```

---

## 3. ML Analysis — Spending Breakdown

```mermaid
flowchart TD
    A([GET /ml/analysis?year=&month=]) --> B[JWT → userID]
    B --> C[MLUseCase.GetAnalysis\noptional year+month filter]
    C --> D[Fetch all matching transactions\nfrom PostgreSQL]
    D --> E[POST ML_SERVICE_URL/analysis\ntimeout 5s]
    E --> F{Success?}
    F -- No --> G[503 ML service unavailable]
    F -- Yes --> H[200 AnalysisResponse\ntotal_expense · avg_daily\ntop_category · spending_distribution map]
```

**`spending_distribution`** is a map of category → total spend, e.g.:
```json
{ "Makanan & Minuman": 1500000, "Transportasi": 450000, ... }
```

---

## 4. ML Anomaly Detection

```mermaid
flowchart TD
    A([GET /ml/anomaly?year=&month=]) --> B[JWT → userID]
    B --> C[MLUseCase.GetAnomaly\noptional year+month filter]
    C --> D[Fetch all matching transactions]
    D --> E[POST ML_SERVICE_URL/anomaly\ntimeout 10s]
    E --> F{Success?}
    F -- No --> G[503 ML service unavailable]
    F -- Yes --> H[200 AnomalyResponse\nanomalies: date+amount pairs\nsummary string]
```

> **ML service note:** The `/anomaly` endpoint requires at least 5 unique expense days to perform statistical detection. Below that threshold the ML service returns an empty anomaly list (see `ML_SERVICE_INTEGRATION.md`).

---

## 5. Data Flow: PostgreSQL → ML Service → Client

```mermaid
flowchart LR
    PG[(PostgreSQL\ntransactions table)] -->|"Raw rows\ndomain.TransactionResponse"| UC[MLUseCase\nfetchMLTransactions]
    UC -->|"ml.Transaction slice\ndate·amount·type·category"| ML[ML Service FastAPI :8000]
    ML -->|"Analysis/Anomaly/Forecast\nJSON response"| UC
    UC --> H[MLHandler]
    H -->|"HTTP 200 JSON"| Client
```
