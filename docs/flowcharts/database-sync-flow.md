# Database Sync Flow

Covers atomic DB operations, cross-table consistency patterns, and the financial profile upsert transaction. Derived from `internal/repository/postgres/`.

---

## 1. Financial Profile Upsert (Atomic DB Transaction)

```mermaid
flowchart TD
    A[FinancialProfileRepository.Upsert] --> B[db.Begin — start DB transaction]
    B --> C[INSERT INTO user_financial_profiles\nmonthly_income · fixed_expenses · current_savings\ndebt · employment_status · spending_habit · risk_level · updated_at=NOW\nON CONFLICT user_id DO UPDATE all fields updated_at=NOW]
    C --> D{error?}
    D -- Yes --> E[tx.Rollback\nreturn error]
    D -- No --> F[DELETE FROM user_financial_goals WHERE user_id=$1\nremove all existing goal tags]
    F --> G{error?}
    G -- Yes --> E
    G -- No --> H[For each goal in req.FinancialGoals]
    H --> I[INSERT INTO user_financial_goals user_id · goal_type\nON CONFLICT user_id goal_type DO NOTHING]
    I --> J{error?}
    J -- Yes --> E
    J -- No --> K{More goals?}
    K -- Yes --> H
    K -- No --> L[tx.Commit]
    L --> M{commit error?}
    M -- Yes --> E
    M -- No --> N[Return nil\nCaller re-reads profile]
```

> **Atomicity guarantee:** The profile fields and goal tags are always consistent — they are either both updated or both rolled back. No partial state is possible.

---

## 2. Goal Contribution — Cross-Domain Net Savings Check

```mermaid
flowchart TD
    A[GoalUseCase.Contribute\ngoalID · userID · amount] --> B{amount > 0?}
    B -- No --> C[400 contribution must be > 0]
    B -- Yes --> D[TransactionRepository.GetNetSavings\nSELECT SUM INCOME - SUM EXPENSE all-time]
    D --> E{net <= 0?}
    E -- Yes --> F[400 cannot add contributions: no net savings]
    E -- No --> G{amount > net int?}
    G -- Yes --> H[400 contribution exceeds available savings]
    G -- No --> I[GoalRepository.Contribute\ngoadID · userID · amount]
    I --> J[UPDATE goals SET\ncurrent_amount = $1\nstatus = CASE WHEN $1 >= target_amount\nTHEN COMPLETED ELSE status END\nupdated_at = NOW\nWHERE id=$2 AND user_id=$3]
    J --> K[200 Contribution successful]
```

> **Cross-domain dependency:** `GoalUseCase` directly injects `TransactionRepository` to validate contribution against net savings. This is an intentional design trade-off to avoid a cross-domain service call.

---

## 3. Budget Usage — Logical JOIN Pattern

```mermaid
flowchart TD
    A[BudgetRepository.GetUsage] --> B[Compute prevMonth and prevYear\ncross-year handling when month == 1]
    B --> C[SQL: FROM budgets b\nLEFT JOIN transactions t\nON t.user_id = b.user_id\nAND LOWER t.category = LOWER b.category\nAND t.type = EXPENSE\nAND t.deleted_at IS NULL]
    C --> D[WHERE b.user_id = $1\nAND b.year = $3\nAND MONTHLY: b.month = $2\nOR YEARLY: period = YEARLY\nAND b.deleted_at IS NULL]
    D --> E[SUM CASE WHEN current period THEN amount ELSE 0 AS used\nSUM CASE WHEN prev period THEN amount ELSE 0 AS prev_used]
    E --> F[GROUP BY b.id b.category b.period b.limit_amount b.alert_threshold]
```

> **No FK between categories:** The JOIN uses `LOWER(t.category) = LOWER(b.category)` — a purely logical match. A misspelling in either table silently breaks the budget tracking for that category.

---

## 4. Soft Delete vs Hard Delete

```mermaid
flowchart LR
    subgraph SoftDelete ["Soft Delete (deleted_at stamp)"]
        U[users]
        T[transactions]
        B[budgets]
        AL[ai_logs]
    end
    subgraph HardDelete ["Hard Delete (DELETE statement)"]
        G[goals]
    end
    subgraph Unused ["Exist but not queried by app"]
        R[reports.deleted_at]
        S[settings.deleted_at]
    end
```

> **Goals:** `deleted_at` was dropped in migration 012. The `goalRepository.Delete()` executes `DELETE FROM goals WHERE id=$1 AND user_id=$2`. All other entities that support deletion use soft delete.

---

## 5. Duplicate Prevention Patterns

```mermaid
flowchart TD
    subgraph Users
        UA[INSERT INTO users\nON CONFLICT email DO NOTHING\nSeeder only]
        UB[SELECT EXISTS before insert\nbudget repository\ncheck by user_id·category·period·month·year]
    end
    subgraph Budgets
        BA[UNIQUE user_id·category·period·month·year\nMonthly: ON CONFLICT DO NOTHING\nYearly: WHERE NOT EXISTS guard]
    end
    subgraph Goals
        GA[CREATE uses\nSELECT 1 WHERE user_id=$6 AND name=$7 AND deadline >= NOW\nif exists → ErrConflict]
    end
    subgraph FinancialProfile
        FA[ON CONFLICT user_id DO UPDATE\nallows re-submit to update profile]
    end
    subgraph UserFinancialGoals
        UA2[DELETE then re-INSERT\nnot incremental — always replace]
    end
```

---

## 6. Index Coverage Map

```mermaid
flowchart LR
    subgraph transactions_indices
        I1[idx_transactions_user_category_date\nuser_id · category · date]
        I2[idx_transactions_full\nuser_id · category · type · date]
    end
    subgraph budgets_indices
        I3[idx_budgets_user\nuser_id]
        I4[idx_budgets_user_year_period\nuser_id · year · period]
    end
    subgraph goals_indices
        I5[idx_goals_user\nuser_id]
        I6[idx_goals_user_deadline\nuser_id · deadline]
    end
    subgraph ai_logs_indices
        I7[idx_ailogs_user\nuser_id]
    end
    subgraph profile_indices
        I8[idx_user_financial_profiles_user_id]
        I9[idx_user_financial_goals_user_id]
    end
