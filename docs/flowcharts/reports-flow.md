# Reports Flow

Added in v1.3 to support full Reports tab with Overview, Categories, Monthly Comparison, Savings Rate, Net Worth tabs.

---

## Reports Data Flow

```mermaid
flowchart TD
    FE["Frontend Reports Tab"] --> Tabs{Active Tab}

    Tabs -- Overview --> MS["GET /reports/monthly-summary?months=6"]
    Tabs -- Categories --> CB["GET /reports/category-breakdown\n?year&month"]
    Tabs -- Monthly Comparison --> MC["GET /reports/month-comparison"]
    Tabs -- Savings Rate --> SR["GET /reports/savings-rate?months=6"]
    Tabs -- Net Worth --> NW["GET /reports/net-worth?months=12"]

    MS & CB & MC & SR & NW --> RUC["ReportsUseCase"]
    RUC --> TR["TransactionRepository"]
    TR --> DB[(transactions table)]
    DB --> TR
    TR --> RUC
    RUC --> API["API Response JSON"]
    API --> FE
```

---

## Category Breakdown Computation

```mermaid
flowchart TD
    Input["All EXPENSE transactions\n(optional month/year filter)"] --> Sum["SUM amount per category"]
    Sum --> Total["total_expense = SUM all categories"]
    Total --> Pct["percentage = category_sum / total * 100"]
    Pct --> Output["{ Food: 40.0, Transport: 33.4, ... }"]
```

---

## Financial Health Score Computation (Dashboard)

```mermaid
flowchart TD
    Income[monthly_income] & NetSavings[net_savings] --> SR["savings_rate = net_savings/income * 100"]
    Budgets[budget_summary] --> BA["budget_adherence = safe_count/total * 100"]
    Goals[active_goals] --> GP["goal_progress = avg(current/target * 100)"]
    SR -->|weight 40%| Score
    BA -->|weight 35%| Score
    GP -->|weight 25%| Score
    Score["score = SR*0.4 + BA*0.35 + GP*0.25"] --> Label{Score range}
    Label -- ">=80" --> Excellent
    Label -- ">=60" --> Good
    Label -- ">=40" --> Fair
    Label -- "<40" --> NeedsAttention["Needs Attention"]
```
