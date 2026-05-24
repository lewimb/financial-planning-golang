# UML Class Diagram — Database Entity Model

```mermaid
---
title: Financial Planning Platform — Entity & Operations Model
---
classDiagram
    namespace Core {
        class users {
            <<table>>
            +id: SERIAL PK
            +email: VARCHAR(255) UNIQUE NOT NULL
            +name: VARCHAR(255) NOT NULL
            +password: VARCHAR(255) NOT NULL
            +created_at: TIMESTAMP
            +deleted_at: TIMESTAMP
            +Register(email, password, name) error
            +Login(email, password) string
            +GetAll() []UserResponse
            +GetMe(userID) UserResponse
        }
    }

    namespace "Financial Transactions" {
        class transactions {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +amount: BIGINT NOT NULL
            +category: VARCHAR(255) NOT NULL
            +type: VARCHAR(10) NOT NULL
            +date: DATE NOT NULL
            +description: TEXT
            +is_recurring: BOOLEAN
            +recurrence_interval: VARCHAR(20)
            +created_at: TIMESTAMP
            +updated_at: TIMESTAMP
            +deleted_at: TIMESTAMP
            +Create(userID, req) error
            +Update(userID, id, req) error
            +Delete(userID, id) error
            +GetAll(userID, limit, offset, year, month) []TransactionResponse
            +GetMonthlyExpenses(userID) float64
            +GetMonthlyIncome(userID) float64
            +GetMonthlySummary(userID, months) []MonthlySummaryItem
            +Export(userID) string
            +BulkImport(userID, items) ImportResult
        }

        class budgets {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +category: VARCHAR(255) NOT NULL
            +period: VARCHAR(10) NOT NULL
            +month: INTEGER
            +year: INTEGER NOT NULL
            +limit_amount: INTEGER NOT NULL
            +alert_threshold: INTEGER
            +created_at: TIMESTAMP
            +updated_at: TIMESTAMP
            +deleted_at: TIMESTAMP
            +Create(userID, req) error
            +Update(userID, id, limit, threshold, category) UpdateBudgetResponse
            +Delete(userID, id) error
            +GetAll(userID, category, month, year) []Budget
            +GetByID(id) BudgetResponse
            +GetUsage(userID, month, year) []BudgetUsage
        }
    }

    namespace "Goals & Planning" {
        class goals {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +name: VARCHAR(255) NOT NULL
            +target_amount: INTEGER NOT NULL
            +current_amount: INTEGER
            +deadline: DATE
            +status: VARCHAR(20)
            +description: TEXT
            +created_at: TIMESTAMP
            +updated_at: TIMESTAMP
            +Create(userID, req) error
            +Update(id, userID, req) error
            +Delete(id, userID) error
            +GetAll(userID, active) []GoalResponse
            +GetByID(id, userID) GoalResponse
            +GetOverview(userID) GoalOverviewResponse
            +GetMilestones(userID) []GoalResponse
            +Contribute(id, userID, amount) error
        }

        class user_financial_profiles {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK UNIQUE
            +monthly_income: NUMERIC(15,2)
            +fixed_expenses: NUMERIC(15,2)
            +current_savings: NUMERIC(15,2)
            +debt: NUMERIC(15,2)
            +employment_status: VARCHAR(100)
            +spending_habit: VARCHAR(100)
            +risk_level: VARCHAR(50)
            +created_at: TIMESTAMPTZ
            +updated_at: TIMESTAMPTZ
            +Upsert(userID, req) FinancialProfileResponse
            +Get(userID) FinancialProfileResponse
        }

        class user_financial_goals {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +goal_type: VARCHAR(100) NOT NULL
            +Create(userID, goal_type) error
            +GetAll(userID) []UserFinancialGoal
        }
    }

    namespace "AI & Reports" {
        class ai_logs {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +question: TEXT
            +response: TEXT
            +created_at: TIMESTAMP
            +deleted_at: TIMESTAMP
            +Save(userID, question, response) error
            +GetHistory(userID) []AiLog
            +ClearHistory(userID) error
        }

        class reports {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +type: VARCHAR(50)
            +month: INTEGER
            +year: INTEGER
            +content: TEXT
            +status: VARCHAR(20)
            +generated_at: TIMESTAMP
            +GetMonthlySummary(userID, months) []MonthlySummaryItem
            +GetCategoryBreakdown(userID, year, month) map
            +GetSavingsRate(userID, months) []SavingsRatePoint
            +GetNetWorth(userID, months) []NetWorthPoint
            +GetMonthComparison(userID) MonthComparisonResponse
        }
    }

    namespace Settings {
        class settings {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK UNIQUE
            +currency: VARCHAR(10)
            +language: VARCHAR(10)
            +notification_enabled: BOOLEAN
            +Upsert(userID, currency, lang, notifEnabled) error
            +Get(userID) Settings
        }
    }

    namespace "Notifications & Activity" {
        class notifications {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +type: VARCHAR(50) NOT NULL
            +title: VARCHAR(255) NOT NULL
            +message: TEXT NOT NULL
            +entity_type: VARCHAR(50)
            +entity_id: INTEGER
            +is_read: BOOLEAN
            +created_at: TIMESTAMPTZ
            +GetAll(userID, unreadOnly) []Notification
            +MarkRead(id, userID) error
            +MarkAllRead(userID) error
            +Delete(id, userID) error
            +GetPreferences(userID) NotificationPreferences
            +UpdatePreferences(userID, prefs) error
            +GetUnreadCount(userID) int
        }

        class notification_preferences {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK UNIQUE
            +budget_alerts: BOOLEAN
            +goal_reminders: BOOLEAN
            +anomaly_alerts: BOOLEAN
            +updated_at: TIMESTAMPTZ
        }

        class activity_logs {
            <<table>>
            +id: SERIAL PK
            +user_id: INTEGER FK
            +action: VARCHAR(50) NOT NULL
            +entity_type: VARCHAR(50) NOT NULL
            +entity_id: INTEGER
            +description: TEXT NOT NULL
            +created_at: TIMESTAMPTZ
            +Log(userID, action, entityType, entityID, description) error
            +GetActivity(userID, limit, offset) []ActivityLog
        }
    }

    %% FK Relationships
    users "1" --> "0..*" transactions : user_id
    users "1" --> "0..*" budgets : user_id
    users "1" --> "0..*" goals : user_id
    users "1" --> "0..*" ai_logs : user_id
    users "1" --> "0..*" reports : user_id
    users "1" --> "0..*" user_financial_goals : user_id
    users "1" --> "0..*" notifications : user_id
    users "1" --> "0..*" activity_logs : user_id
    users "1" --> "0..1" settings : user_id
    users "1" --> "0..1" user_financial_profiles : user_id
    users "1" --> "0..1" notification_preferences : user_id

    %% Cross-Entity Business Operations
    transactions ..> budgets : GetUsage()\nin-memory join by category
    transactions ..> goals : Contribute()\nvalidates ≤ net_savings
    budgets ..> notifications : CheckBudgetAlerts()\ngenerates warnings
    transactions ..> activity_logs : CRUD logs activity
    budgets ..> activity_logs : CRUD logs activity
    goals ..> activity_logs : CRUD logs activity
    user_financial_profiles ..> transactions : GetNetSavings()\nincome - expense
```

---

## Detailed Entity Specification

### Core

#### `users`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `email` | `VARCHAR(255)` | `NOT NULL, UNIQUE` | User email (login credential) |
| `name` | `VARCHAR(255)` | `NOT NULL` | Display name |
| `password` | `VARCHAR(255)` | `NOT NULL` | bcrypt-hashed password |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | Row creation timestamp |
| `deleted_at` | `TIMESTAMP` | — | Soft-delete flag |

---

### Financial Transactions

#### `transactions`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL` | Owner |
| `amount` | `BIGINT` | `NOT NULL` | Transaction amount (in IDR, no decimals) |
| `category` | `VARCHAR(255)` | `NOT NULL` | Expense/income category |
| `type` | `VARCHAR(10)` | `NOT NULL, CHECK(INCOME, EXPENSE)` | Transaction direction |
| `date` | `DATE` | `NOT NULL` | Transaction date |
| `description` | `TEXT` | — | Optional note |
| `is_recurring` | `BOOLEAN` | `DEFAULT FALSE` | Recurring transaction flag |
| `recurrence_interval` | `VARCHAR(20)` | — | e.g., monthly, weekly |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | Row creation timestamp |
| `updated_at` | `TIMESTAMP` | — | Last update timestamp |
| `deleted_at` | `TIMESTAMP` | — | Soft-delete flag |

**Indices:** `(user_id, category, date)`, `(user_id, category, type, date)`, `(user_id, is_recurring) WHERE is_recurring = TRUE AND deleted_at IS NULL`

#### `budgets`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL` | Owner |
| `category` | `VARCHAR(255)` | `NOT NULL` | Budget category |
| `period` | `VARCHAR(10)` | `NOT NULL, CHECK(MONTHLY, YEARLY)` | Budget period type |
| `month` | `INTEGER` | — | Month (1–12) for MONTHLY budgets |
| `year` | `INTEGER` | `NOT NULL` | Budget year |
| `limit_amount` | `INTEGER` | `NOT NULL` | Spending limit |
| `alert_threshold` | `INTEGER` | `DEFAULT 80` | Percentage threshold for warnings |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | Row creation timestamp |
| `updated_at` | `TIMESTAMP` | — | Last update timestamp |
| `deleted_at` | `TIMESTAMP` | — | Soft-delete flag |
| — | — | `UNIQUE(user_id, category, period, month, year)` | Prevents duplicate budgets |

**Indices:** `(user_id)`, `(user_id, year, period)`

---

### Goals & Planning

#### `goals`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL` | Owner |
| `name` | `VARCHAR(255)` | `NOT NULL` | Goal name |
| `target_amount` | `INTEGER` | `NOT NULL` | Target savings amount |
| `current_amount` | `INTEGER` | `DEFAULT 0` | Current progress |
| `deadline` | `DATE` | — | Target completion date |
| `status` | `VARCHAR(20)` | `DEFAULT 'ONGOING'` | ONGOING or COMPLETED |
| `description` | `TEXT` | — | Optional description |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | Row creation timestamp |
| `updated_at` | `TIMESTAMP` | — | Last update timestamp |

**Indices:** `(user_id)`, `(user_id, deadline)`

#### `user_financial_profiles`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, UNIQUE, NOT NULL, ON DELETE CASCADE` | Owner (1-to-1) |
| `monthly_income` | `NUMERIC(15,2)` | `NOT NULL DEFAULT 0` | Monthly income |
| `fixed_expenses` | `NUMERIC(15,2)` | `NOT NULL DEFAULT 0` | Fixed monthly expenses |
| `current_savings` | `NUMERIC(15,2)` | `NOT NULL DEFAULT 0` | Current savings amount |
| `debt` | `NUMERIC(15,2)` | `NOT NULL DEFAULT 0` | Total debt |
| `employment_status` | `VARCHAR(100)` | `NOT NULL` | e.g., employed, self-employed |
| `spending_habit` | `VARCHAR(100)` | — | e.g., frugal, moderate, spender |
| `risk_level` | `VARCHAR(50)` | — | e.g., low, medium, high |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Row creation timestamp |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Last update timestamp |

#### `user_financial_goals`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL, ON DELETE CASCADE` | Owner |
| `goal_type` | `VARCHAR(100)` | `NOT NULL` | e.g., emergency_fund, retirement |
| — | — | `UNIQUE(user_id, goal_type)` | One entry per goal type per user |

---

### AI & Reports

#### `ai_logs`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL` | Owner |
| `question` | `TEXT` | — | User's question to AI |
| `response` | `TEXT` | — | AI's response |
| `created_at` | `TIMESTAMP` | `DEFAULT NOW()` | Row creation timestamp |
| `deleted_at` | `TIMESTAMP` | — | Soft-delete flag |

**Index:** `(user_id)`

#### `reports`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL` | Owner |
| `type` | `VARCHAR(50)` | — | MONTHLY or YEARLY |
| `month` | `INTEGER` | — | Report period month |
| `year` | `INTEGER` | — | Report period year |
| `content` | `TEXT` | — | Report body content |
| `status` | `VARCHAR(20)` | `DEFAULT 'GENERATED'` | Generation state |
| `generated_at` | `TIMESTAMP` | `DEFAULT NOW()` | Generation timestamp |

**Index:** `(user_id)`

---

### Settings

#### `settings`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, UNIQUE, NOT NULL` | Owner (1-to-1) |
| `currency` | `VARCHAR(10)` | `DEFAULT 'IDR'` | Preferred currency |
| `language` | `VARCHAR(10)` | `DEFAULT 'EN'` | Preferred language |
| `notification_enabled` | `BOOLEAN` | `DEFAULT TRUE` | Global notification toggle |

---

### Notifications & Activity

#### `notifications`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL, ON DELETE CASCADE` | Recipient |
| `type` | `VARCHAR(50)` | `NOT NULL` | e.g., BUDGET_WARNING, BUDGET_EXCEEDED |
| `title` | `VARCHAR(255)` | `NOT NULL` | Notification title |
| `message` | `TEXT` | `NOT NULL` | Notification body |
| `entity_type` | `VARCHAR(50)` | — | Related entity type |
| `entity_id` | `INTEGER` | — | Related entity ID |
| `is_read` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` | Read status |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Creation timestamp |

**Indices:** `(user_id)`, `(user_id, is_read) WHERE is_read = FALSE`

#### `notification_preferences`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, UNIQUE, NOT NULL, ON DELETE CASCADE` | Owner (1-to-1) |
| `budget_alerts` | `BOOLEAN` | `NOT NULL DEFAULT TRUE` | Budget alert toggle |
| `goal_reminders` | `BOOLEAN` | `NOT NULL DEFAULT TRUE` | Goal reminder toggle |
| `anomaly_alerts` | `BOOLEAN` | `NOT NULL DEFAULT TRUE` | Anomaly alert toggle |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Last update timestamp |

#### `activity_logs`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `SERIAL` | `PK` | Primary key |
| `user_id` | `INTEGER` | `FK → users.id, NOT NULL, ON DELETE CASCADE` | Actor |
| `action` | `VARCHAR(50)` | `NOT NULL` | e.g., CREATE, UPDATE, DELETE |
| `entity_type` | `VARCHAR(50)` | `NOT NULL` | e.g., transaction, budget |
| `entity_id` | `INTEGER` | — | Affected entity ID |
| `description` | `TEXT` | `NOT NULL` | Human-readable log |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | Creation timestamp |

**Indices:** `(user_id)`, `(user_id, created_at DESC)`

---

## Relationship Summary

| # | Parent | Child | Type | FK Column | Child Constraints |
|---|--------|-------|------|-----------|-------------------|
| R1 | `users` | `transactions` | 1 ──< N | `user_id` | — |
| R2 | `users` | `budgets` | 1 ──< N | `user_id` | UNIQUE(category, period, month, year) |
| R3 | `users` | `goals` | 1 ──< N | `user_id` | — |
| R4 | `users` | `ai_logs` | 1 ──< N | `user_id` | — |
| R5 | `users` | `reports` | 1 ──< N | `user_id` | — |
| R6 | `users` | `user_financial_goals` | 1 ──< N | `user_id` | UNIQUE(goal_type), ON DELETE CASCADE |
| R7 | `users` | `notifications` | 1 ──< N | `user_id` | ON DELETE CASCADE |
| R8 | `users` | `activity_logs` | 1 ──< N | `user_id` | ON DELETE CASCADE |
| R9 | `users` | `settings` | 1 ── 1 | `user_id` | UNIQUE |
| R10 | `users` | `user_financial_profiles` | 1 ── 1 | `user_id` | UNIQUE, ON DELETE CASCADE |
| R11 | `users` | `notification_preferences` | 1 ── 1 | `user_id` | UNIQUE, ON DELETE CASCADE |

### Logical Relationships (no formal FK)

| Context | Entities | Join Condition | Description |
|---------|----------|----------------|-------------|
| Budget Usage | `transactions` ↔ `budgets` | `category`, `user_id`, `month`, `year` | In-memory join in `GetBudgetUsage()` |
| Dashboard | `transactions` + `budgets` + `goals` | `user_id` | Aggregated via `DashboardUseCase` |
| Goal Contribution | `goals` ↔ `transactions` | `user_id` | Validates `contribution ≤ net_savings` via `GetNetSavings()` |

---

## Multiplicity Legend

```
1 ── 1   One-to-one (enforced by UNIQUE FK)
1 ──< N  One-to-many
0..1     Optional (nullable columns / no row required)
0..*     Zero or more child rows
```
