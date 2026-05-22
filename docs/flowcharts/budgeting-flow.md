# Budgeting Flow

Derived from `internal/usecase/budget.go`, `internal/repository/postgres/budget.go`, `internal/delivery/http/handler/budget.go`.

---

## 1. Create Budget

```mermaid
flowchart TD
    A([POST /api/auth/v1/budgets]) --> B[JWT → userID]
    B --> C[Parse JSON body\ncategory · period · month\nyear · limit_amount · alert_threshold]
    C --> D{Bind ok?}
    D -- No --> E[400 invalid request]
    D -- Yes --> F[BudgetUseCase.Create]

    F --> G{category non-empty?}
    G -- No --> H[400 category is required]
    G -- Yes --> I{period == MONTHLY or YEARLY?}
    I -- No --> J[400 invalid period]
    I -- Yes --> K{year > 0?}
    K -- No --> L[400 year is required]
    K -- Yes --> M{limit_amount > 0?}
    M -- No --> N[400 limit must be greater than 0]
    M -- Yes --> O{period == MONTHLY AND month == nil?}
    O -- Yes --> P[400 month required for monthly budget]
    O -- No --> Q{period == YEARLY?}
    Q -- Yes --> R[Force month = nil]
    Q -- No --> S[Keep month as-is]
    R --> T{alert_threshold == 0?}
    S --> T
    T -- Yes --> U[Default alert_threshold = 80]
    T -- No --> V[Keep alert_threshold]
    U --> W[BudgetRepository.Create]
    V --> W

    W --> X[Check duplicate:\nSELECT EXISTS WHERE user_id·category·period·year\nAND month = $5 OR both NULL]
    X --> Y{exists?}
    Y -- Yes --> Z[domain.ErrConflict → 409]
    Y -- No --> AA[INSERT INTO budgets\nuser_id·category·period·month\nyear·limit_amount·alert_threshold]
    AA --> AB[201 budget created successfully]
```

---

## 2. Budget Usage Calculation

```mermaid
flowchart TD
    A([GET /api/auth/v1/budgets/usage?year=2026&month=5]) --> B[JWT → userID]
    B --> C[Parse year required · month optional]
    C --> D{year parseable?}
    D -- No --> E[400 invalid year]
    D -- Yes --> F{month parseable if provided?}
    F -- No --> G[400 invalid month]
    F -- Yes --> H[BudgetUseCase.GetUsage]

    H --> I[BudgetRepository.GetUsage\nuserID · month · year]

    I --> J[Compute prevMonth = month - 1\nif month == 1: prevMonth=12 prevYear=year-1]

    J --> K[Complex SQL JOIN:\nFROM budgets b\nLEFT JOIN transactions t\nON t.user_id = b.user_id\nAND LOWER t.category = LOWER b.category\nAND t.type = EXPENSE\nAND t.deleted_at IS NULL\nWHERE b.user_id = $1\nAND b.year = $3\nAND MONTHLY budget matches month $2\nOR YEARLY budget for year $3\nAND b.deleted_at IS NULL\nGROUP BY b.id]

    K --> L[For each budget row:\ncompute used = SUM current period\ncompute prev_used = SUM previous period]

    L --> M[Calculate derived fields]
    M --> N[remaining = limit - used\nif negative → 0]
    N --> O[percentage = ROUND used/limit * 100]
    O --> P{percentage >= 100?}
    P -- Yes --> Q[status = EXCEEDED]
    P -- No --> R{percentage >= alert_threshold?}
    R -- Yes --> S[status = WARNING]
    R -- No --> T[status = SAFE]
    Q --> U[Calculate change_percent]
    S --> U
    T --> U
    U --> V{prev_used > 0?}
    V -- Yes --> W[change_percent = ROUND used-prev/prev * 100]
    V -- No --> X{used > 0?}
    X -- Yes --> Y[change_percent = 100]
    X -- No --> Z[change_percent = 0]
    W --> AA[200 BudgetUsage array]
    Y --> AA
    Z --> AA
```

---

## 3. Update Budget

```mermaid
flowchart TD
    A([PUT /api/auth/v1/budgets/:id]) --> B[JWT → userID]
    B --> C[Parse :id]
    C --> D{id valid?}
    D -- No --> E[400]
    D -- Yes --> F[Parse body\nlimitAmount · alertThreshold · category]
    F --> G[BudgetUseCase.Update]
    G --> H[BudgetRepository.Update]
    H --> I[UPDATE budgets SET\nlimit_amount = COALESCE NULLIF limitAmount 0 existing\nalert_threshold = COALESCE NULLIF alertThreshold 0 existing\ncategory = COALESCE NULLIF category empty existing\nupdated_at = NOW\nWHERE user_id=$4 AND id=$5\nRETURNING all fields]
    I --> J{row returned?}
    J -- No --> K[ErrNotFound → 404]
    J -- Yes --> L[200 UpdateBudgetResponse]
```

> **NULLIF pattern:** Zero values in `limitAmount`/`alertThreshold` or empty string in `category` are treated as "no change" — the existing value is preserved. This allows partial updates.

---

## 4. Soft-Delete Budget

```mermaid
flowchart TD
    A([DELETE /api/auth/v1/budgets/:id]) --> B[JWT → userID]
    B --> C[Parse :id]
    C --> D[BudgetUseCase.Delete]
    D --> E[UPDATE budgets\nSET deleted_at = NOW\nWHERE user_id=$1 AND id=$2 AND deleted_at IS NULL]
    E --> F{RowsAffected == 0?}
    F -- Yes --> G[ErrNotFound → 404]
    F -- No --> H[200 budget deleted successfully]
```

---

## 5. Budget Status State Machine

```mermaid
flowchart LR
    A([Budget Created\nstatus implicit: SAFE]) --> B{Usage %}
    B -->|"< alert_threshold\ndefault 80%"| C[SAFE]
    B -->|">= alert_threshold\nand < 100%"| D[WARNING]
    B -->|">= 100%"| E[EXCEEDED]
    C --> B
    D --> B
    E --> B
```
