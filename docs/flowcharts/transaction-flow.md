# Transaction Flow

Derived from `internal/usecase/transaction.go`, `internal/repository/postgres/transaction.go`, and `internal/delivery/http/handler/transaction.go`.

---

## 1. Create Transaction

```mermaid
flowchart TD
    A([POST /api/auth/v1/transactions]) --> B[JWT Middleware\nextract userID]
    B --> C[Parse JSON body\namount · category · type · date · description]
    C --> D{ShouldBindJSON ok?}
    D -- No --> E[400 Invalid input]
    D -- Yes --> F[TransactionUseCase.Create]

    F --> G[strings.ToUpper type]
    G --> H{type == INCOME\nor EXPENSE?}
    H -- No --> I[400 invalid type]
    H -- Yes --> J{amount > 0?}
    J -- No --> K[400 amount must be greater than 0]
    J -- Yes --> L{category non-empty?}
    L -- No --> M[400 category is required]
    L -- Yes --> N{date non-zero?}
    N -- No --> O[400 date is required]
    N -- Yes --> P[TransactionRepository.Create]
    P --> Q[INSERT INTO transactions\namount · category · type\ndate · description · user_id]
    Q --> R{DB error?}
    R -- Yes --> S[400 error]
    R -- No --> T[200 Transaction created successfully]
```

---

## 2. List Transactions (Paginated + Filterable)

```mermaid
flowchart TD
    A([GET /api/auth/v1/transactions]) --> B[JWT → userID]
    B --> C[Parse query params\nlimit default 10 · offset default 0\nmonth · year optional]
    C --> D[TransactionUseCase.GetTransactions]
    D --> E[TransactionRepository.GetByUserID]

    E --> F{month AND year provided?}
    F -- Yes --> G[Add EXTRACT MONTH and YEAR filters\nto WHERE clause]
    F -- No --> H[No date filter]
    G --> I[Build query]
    H --> I

    I --> J[Execute count query\nSELECT COUNT * same filters]
    J --> K[Execute main query\nSELECT id·amount·category·type·date·description\nORDER BY date DESC\nLIMIT offset if limit > 0]
    K --> L[200 { data: transactions, total: count }]
```

---

## 3. Update Transaction

```mermaid
flowchart TD
    A([PUT /api/auth/v1/transactions/:id]) --> B[JWT → userID]
    B --> C[Parse :id param]
    C --> D{id valid int?}
    D -- No --> E[400 invalid transaction id]
    D -- Yes --> F[Parse JSON body\nsame fields as create]
    F --> G{Bind ok?}
    G -- No --> H[400 invalid request body]
    G -- Yes --> I[TransactionUseCase.Update\nvalidation identical to Create]
    I --> J{Valid?}
    J -- No --> K[400 validation error]
    J -- Yes --> L[TransactionRepository.Update]
    L --> M[UPDATE transactions\nSET amount·category·type·date·description\nupdated_at = NOW\nWHERE id=$6 AND user_id=$7 AND deleted_at IS NULL]
    M --> N{RowsAffected == 0?}
    N -- Yes --> O[ErrNotFound → 404]
    N -- No --> P[200 transaction updated successfully]
```

---

## 4. Soft-Delete Transaction

```mermaid
flowchart TD
    A([DELETE /api/auth/v1/transactions/:id]) --> B[JWT → userID]
    B --> C[Parse :id param]
    C --> D{id valid int?}
    D -- No --> E[400 invalid id]
    D -- Yes --> F[TransactionUseCase.Delete]
    F --> G[TransactionRepository.Delete]
    G --> H[UPDATE transactions\nSET deleted_at = NOW\nWHERE id=$1 AND user_id=$2 AND deleted_at IS NULL]
    H --> I{RowsAffected == 0?}
    I -- Yes --> J[ErrNotFound → 404]
    I -- No --> K[200 transaction deleted successfully]
```

> **Soft delete:** The row is NOT removed. `deleted_at` is stamped with `NOW()`. All subsequent queries filter `WHERE deleted_at IS NULL`.

---

## 5. Monthly Aggregations

```mermaid
flowchart TD
    A1([GET /transactions/monthly]) --> B1[JWT → userID]
    A2([GET /transactions/monthly-income]) --> B2[JWT → userID]

    B1 --> C1[GetMonthlyExpenses\nSELECT COALESCE SUM amount 0\nWHERE type = EXPENSE\nAND EXTRACT MONTH = now.Month\nAND EXTRACT YEAR = now.Year\nAND deleted_at IS NULL]
    C1 --> D1[200 { total: float64 }]

    B2 --> C2[GetMonthlyIncome\nSELECT COALESCE SUM amount 0\nWHERE type = INCOME\nAND EXTRACT MONTH = now.Month\nAND EXTRACT YEAR = now.Year\nAND deleted_at IS NULL]
    C2 --> D2[200 { total: float64 }]
```

---

## 6. Net Savings (Cross-domain use)

```mermaid
flowchart TD
    A[GoalUseCase.Contribute\nor ChatUseCase.Ask\nor DashboardUseCase.Get] --> B[TransactionRepository.GetNetSavings]
    B --> C[SELECT\nCOALESCE SUM INCOME 0 -\nCOALESCE SUM EXPENSE 0\nFROM transactions\nWHERE user_id = $1\nAND deleted_at IS NULL\nno date filter — all-time]
    C --> D[float64 net savings returned]
```

> **All-time scope:** `GetNetSavings` has no month/year filter. It represents the user's total financial position since account creation.
