# Financial Planning Backend — System Overview

> **Codebase**: Go 1.25 · Gin · PostgreSQL · JWT HS256 · Google Gemini 2.0 Flash  
> **Architecture**: Clean Architecture (Domain → Use Case → Delivery)  
> **Base URL**: `http://localhost:8080`  
> **CORS origin**: `http://localhost:5173` (Vite/React frontend, hardcoded)

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [System Logic](#2-system-logic)
3. [MVP Scope](#3-mvp-scope)
4. [Architecture Overview](#4-architecture-overview)
5. [Feature Breakdown](#5-feature-breakdown)
6. [API Reference Summary](#6-api-reference-summary)

---

## 1. System Overview

### Purpose

A personal finance management backend that lets a user track income/expenses, set category budgets, create savings goals, view a dashboard summary, and ask an AI assistant questions about their finances.

### Main Modules

| Module | Description |
|---|---|
| **Auth** | Register, login, JWT issuance |
| **Transactions** | Record and query INCOME/EXPENSE entries |
| **Budgets** | Set spending limits per category per period |
| **Goals** | Define savings targets with deadlines and contribute to them |
| **Dashboard** | Aggregate snapshot: income, expense, savings, budget health, goal progress |
| **AI Chat** | Context-aware financial assistant powered by Google Gemini 2.0 Flash |

### Core Business Flow

```
Register / Login
      │
      ▼
[JWT token]
      │
      ├─► Log transactions (INCOME / EXPENSE)
      │
      ├─► Create budgets per category
      │         └─► Check budget usage vs. actual spend
      │
      ├─► Create savings goals
      │         └─► Contribute to goals (validated against net savings)
      │
      ├─► View dashboard (aggregates all of the above)
      │
      └─► Ask AI assistant (receives financial context automatically)
```

### User Roles

Single-role system. Every authenticated user is an **end user** — no admin role exists in the current implementation.

| Actor | Capabilities |
|---|---|
| **Anonymous** | `POST /register`, `POST /login` |
| **Authenticated user** | All other endpoints (scoped to their own `user_id`) |

---

## 2. System Logic

### 2.1 Authentication Flow

```
Client                    Backend                        DB
  │                          │                            │
  │  POST /api/v1/login       │                            │
  │  { email, password }      │                            │
  │─────────────────────────►│                            │
  │                          │  SELECT ... FROM users      │
  │                          │  WHERE email = $1          │
  │                          │───────────────────────────►│
  │                          │◄───────────────────────────│
  │                          │  bcrypt.CompareHashAndPW   │
  │                          │  GenerateJWT(id,name,email)│
  │◄─────────────────────────│                            │
  │  200 { token }            │                            │
  │  Set-Cookie: accessToken  │                            │
```

**Details:**
- Password hashed at registration with `bcrypt` cost factor `10`.
- JWT signed with `HS256` using `SECRET_KEY` env variable.
- JWT payload: `{ id, name, email, iss: "lewimb" }`. No expiry is set — tokens are valid indefinitely until the secret rotates.
- Login sets an `accessToken` cookie **and** returns the token in the response body. However, the auth middleware **only reads the `Authorization` header** (`Bearer <token>`). The cookie is cosmetic in the current implementation.
- `FindByEmail` does not filter `deleted_at`, so soft-deleted users can still log in. *(Assumption: user deletion is not implemented yet.)*

### 2.2 Auth Middleware

```
Request → middleware.AuthRequired()
    ├── missing Authorization header → 400
    ├── wrong format (not "Bearer X") → 401
    ├── invalid/expired token → 401
    └── valid → set "claims" in gin.Context → next handler
```

All `/api/auth/v1/*` routes are protected. The `ClaimId()` utility extracts `user_id` from the claims context key — every protected handler scopes data by this ID automatically.

### 2.3 Transaction Flow

```
POST /api/auth/v1/transactions
    │
    ├── Validate: type ∈ {INCOME, EXPENSE}, amount > 0, category not empty, date not zero
    ├── Normalize: type = strings.ToUpper(type)
    └── INSERT INTO transactions (user_id, amount, category, type, date, description)
```

Retrieval supports optional `month`/`year` filters and `limit`/`offset` pagination. The total count (not total pages) is returned alongside data — page math is left to the client.

Soft delete: `DELETE` sets `deleted_at = NOW()` and checks `rows_affected == 0` to return 404 if the record doesn't exist or belong to the user.

### 2.4 Budget Flow & Business Rules

**Create:**
1. Validate: category, period ∈ {MONTHLY, YEARLY}, year, limit > 0.
2. If `MONTHLY`, `month` must be present; if `YEARLY`, month is set to `null`.
3. `alert_threshold` defaults to `80` if not supplied.
4. Duplicate check: `UNIQUE(user_id, category, period, month, year)` — returns 409 on conflict.

**Usage calculation (SQL):**
```
for each budget:
    used     = SUM(expense transactions matching category + period)
    prev_used = SUM(same for previous month/year)
    percentage = round((used / limit) * 100)
    status:
        percentage >= 100             → EXCEEDED
        percentage >= alert_threshold → WARNING
        else                          → SAFE
    change_percent = round(((used - prev_used) / prev_used) * 100)
```

Category matching is case-insensitive (`LOWER(t.category) = LOWER(b.category)`).

The `GetByID` endpoint does **not** verify `user_id` ownership — any authenticated user can read any budget by ID if they know it. *(Security gap.)*

### 2.5 Goal Flow & Business Rules

**Create:**
- Name required, target > 0, deadline must be in the future.
- Duplicate guard: an active goal (deadline ≥ NOW) with the same name for the same user is rejected with 409.

**Update:**
- `target_amount` is only updated if `new_value > current_amount`; otherwise the existing value is preserved. This prevents lowering a target below what's already saved.
- If target is updated upward, status is reset to `ONGOING`.

**Contribution (critical business rule):**
```
net_savings = SUM(all-time INCOME) - SUM(all-time EXPENSE) for user
if net_savings <= 0  → error: "no net savings"
if amount > net_savings → error: "exceeds available savings"
repo.Contribute(id, userID, amount)  ← sets current_amount = amount (absolute, not additive)
```

> **Important:** `Contribute` sets `current_amount` to the submitted `amount` value directly, not `current_amount + amount`. The UI must send the new cumulative total, not the increment.

Status auto-transitions on contribution:
```
current_amount >= target_amount → status = 'COMPLETED'
```

**Active filter:** `GET /goals?active=true` filters by `deadline >= NOW()`, not by `status`. A goal past its deadline is excluded even if not COMPLETED.

**Milestones:** Returns up to 4 goals ordered by nearest deadline where `target ≠ current` (not yet completed).

### 2.6 Dashboard Flow

Single endpoint aggregates across 3 repositories synchronously:

```
GET /api/auth/v1/dashboard
    │
    ├── txRepo.GetMonthlyIncome(userID)      ← current calendar month
    ├── txRepo.GetMonthlyExpenses(userID)    ← current calendar month
    ├── txRepo.GetNetSavings(userID)         ← all-time
    ├── budgetRepo.GetUsage(userID, month, year)
    │       └── classify each: SAFE / WARNING / EXCEEDED → count
    ├── goalRepo.GetAll(userID, active=true)
    │       └── count COMPLETED in the result set
    └── goalRepo.CountActive(userID)         ← WHERE deadline >= NOW()
```

No caching — all queries run on every dashboard load.

### 2.7 AI Chat Flow

```
POST /api/auth/v1/chat { message: "..." }
    │
    ├── Fetch user's financial context:
    │       monthly_income, monthly_expense, net_savings
    │       budget_usage (current month) → count exceeded
    │       active goals → count
    │
    ├── Build prompt:
    │       "You are a helpful financial assistant...
    │        [Financial Data summary]
    │        User question: <message>"
    │
    ├── POST to Gemini 2.0 Flash API
    │       (https://generativelanguage.googleapis.com/v1beta/...)
    │
    ├── Save Q&A to ai_logs (best-effort — failure does not fail the response)
    │
    └── Return { reply: "..." }
```

Language: Gemini is prompted to respond in the same language as the user (Indonesian or English).

If `GEMINI_API_KEY` is not set or the API returns an error, the endpoint returns `503 AI service unavailable`.

### 2.8 Request/Response Flow

```
HTTP Request
    └─► gin router
            └─► corsMiddleware (all routes)
                    └─► [AuthRequired middleware] (auth routes only)
                            └─► Handler
                                    ├─► Validate input (gin ShouldBindJSON)
                                    ├─► UseCase (business logic + validation)
                                    │       └─► Repository (SQL)
                                    │               └─► PostgreSQL
                                    └─► gin.JSON(status, payload)
```

### 2.9 State Transitions

**Goal status:**
```
[created] → ONGOING
    └─► Contribute (current >= target) → COMPLETED
    └─► Update (new target > current)  → ONGOING (reset)
```

**Budget status** (computed on read, not stored):
```
SAFE ←→ WARNING ←→ EXCEEDED   (derived from percentage vs. alert_threshold)
```

---

## 3. MVP Scope

### Included in MVP

| Feature | Status |
|---|---|
| User registration & login with JWT | ✅ Complete |
| Transaction CRUD (create, read, update, soft-delete) | ✅ Complete |
| Transaction filtering by month/year | ✅ Complete |
| Transaction pagination (limit/offset) | ✅ Complete |
| Budget CRUD | ✅ Complete |
| Budget usage with status (SAFE/WARNING/EXCEEDED) | ✅ Complete |
| Budget change percentage vs. previous period | ✅ Complete |
| Goal CRUD | ✅ Complete |
| Goal contribution with net savings validation | ✅ Complete |
| Goal milestones (upcoming, 4 nearest) | ✅ Complete |
| Dashboard aggregation | ✅ Complete |
| AI financial assistant (Gemini 2.0 Flash) | ✅ Complete |
| AI conversation logging to `ai_logs` | ✅ Complete |
| Soft deletes (transactions, budgets) | ✅ Complete |
| Performance indices on transactions and budgets | ✅ Complete |

### Partially Implemented

| Feature | What Exists | What's Missing |
|---|---|---|
| **AI chat history** | Q&A saved to `ai_logs` table + `GetByUserID` repo method | No HTTP endpoint to retrieve chat history |
| **Cookie-based auth** | `Set-Cookie: accessToken` on login | Middleware ignores cookies; only reads `Authorization` header |
| **Goal "active" filter** | Filters by `deadline >= NOW()` | Does not account for status=COMPLETED; a completed goal is "active" until deadline passes |

### Planned but Not Implemented

| Feature | Schema Exists | Handler | Notes |
|---|---|---|---|
| **User settings** | ✅ `settings` table | ❌ None | currency, language, notifications |
| **Reports** | ✅ `reports` table | ❌ None | MONTHLY/YEARLY report generation |
| **User profile update** | — | ❌ None | No PATCH /users/:id endpoint |
| **Password change / reset** | — | ❌ None | |
| **Notification system** | `settings.notification_enabled` | ❌ None | |

### Technical Limitations / Shortcuts

| Area | Limitation |
|---|---|
| **JWT expiry** | No `exp` claim set — tokens never expire |
| **No refresh tokens** | Single token for session lifetime |
| **CORS hardcoded** | `Access-Control-Allow-Origin: http://localhost:5173` — not configurable |
| **Budget GetByID no ownership check** | Any user can read any budget by guessing the ID |
| **GetAll users returns password hash** | `UserResponse.Password` is exposed in JSON (tag is set) |
| **Goal contribution is absolute** | UI must send cumulative total, not the increment |
| **Dashboard: no caching** | 6+ DB queries on every load |
| **No rate limiting on AI chat** | Each request calls Gemini API — no throttle |
| **Soft delete inconsistency** | Goals use hard `DELETE`; transactions/budgets use soft delete |
| **No input sanitization on chat** | User message passed directly into Gemini prompt |
| **`GET /api/v1/users` is unauthenticated** | Returns all users including hashed passwords |
| **No DB connection pool config** | Using `sql.Open` defaults — no `SetMaxOpenConns` etc. |

---

## 4. Architecture Overview

### 4.1 Overall Structure

```
financial-planning-golang/
├── main.go                          ← wires everything together
├── utils/                           ← JWT generation, claims extraction
│   ├── jwt.go
│   └── claims.go
└── internal/
    ├── domain/                      ← entities, repository interfaces, error sentinels
    │   ├── user.go
    │   ├── transaction.go
    │   ├── budget.go
    │   ├── goal.go
    │   ├── dashboard.go
    │   ├── chat.go
    │   ├── ai_log.go
    │   └── errors.go
    ├── usecase/                     ← business logic, orchestration
    │   ├── user.go
    │   ├── transaction.go
    │   ├── budget.go
    │   ├── goal.go
    │   ├── dashboard.go
    │   └── chat.go
    ├── repository/postgres/         ← SQL implementations of domain interfaces
    │   ├── user.go
    │   ├── transaction.go
    │   ├── budget.go
    │   ├── goal.go
    │   └── ai_log.go
    └── delivery/http/               ← HTTP layer
        ├── router.go                ← route registration
        ├── middleware/
        │   └── auth.go              ← JWT validation middleware
        └── handler/                 ← one handler struct per domain
            ├── user.go
            ├── transaction.go
            ├── budget.go
            ├── goal.go
            ├── dashboard.go
            └── chat.go
```

### 4.2 Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  Delivery (HTTP)                                         │
│  handler → middleware → router                           │
├─────────────────────────────────────────────────────────┤
│  Use Case                                                │
│  Business logic, validation, orchestration               │
├─────────────────────────────────────────────────────────┤
│  Domain                                                  │
│  Entities · Repository interfaces · Error sentinels     │
├─────────────────────────────────────────────────────────┤
│  Repository (Postgres)                                   │
│  Raw SQL via database/sql + pgx driver                   │
└─────────────────────────────────────────────────────────┘
```

Dependencies flow inward: Delivery → UseCase → Domain ← Repository.  
The domain layer has no external dependencies.

### 4.3 Backend Architecture

- **Framework**: Gin (HTTP router + middleware)
- **DB driver**: `pgx/v5` via `database/sql` stdlib interface
- **Auth**: `golang-jwt/jwt/v5`, HS256, signed with `SECRET_KEY` env var
- **Password hashing**: `golang.org/x/crypto/bcrypt`, cost 10
- **External API**: Google Gemini 2.0 Flash (`generativelanguage.googleapis.com`)
- **Config**: `.env` via `joho/godotenv`

### 4.4 Dependency Injection

All wiring happens in `main.go`:

```go
// Repos implement domain interfaces
userRepo   := postgres.NewUserRepository(db)
txRepo     := postgres.NewTransactionRepository(db)
...

// Use cases receive repo interfaces
goalUC := usecase.NewGoalUseCase(goalRepo, txRepo)  // cross-repo dependency

// Handlers receive use cases
delivery.Setup(r, delivery.Deps{...})
```

Use cases that need data from multiple domains receive multiple repo interfaces (e.g., `GoalUseCase` needs `TransactionRepository` to validate net savings; `DashboardUseCase` needs all three).

### 4.5 Database Interaction

Raw SQL via `database/sql`. No ORM. Parameterized queries throughout — no string concatenation of user values into SQL (safe from SQL injection). Dynamic filters in `GetAll` and `GetByUserID` build queries with positional `$N` placeholders.

Migrations managed via numbered `.up.sql` / `.down.sql` files in `db/migrations/`. Not auto-applied on startup — must be run manually with `migrate` or equivalent.

### 4.6 Important Design Decisions

| Decision | Rationale |
|---|---|
| Clean architecture | Separates HTTP concerns from business logic; repositories are testable via interface mocking |
| Interface-based repositories | `domain.UserRepository` is an interface; postgres struct implements it — swap DB without touching use cases |
| No ORM | Raw SQL for full control over queries (especially the budget usage LEFT JOIN with previous-period comparison) |
| AI context built per request | Gives Gemini accurate, real-time financial data without storing a conversation state |
| Best-effort AI log saving | `ai_logs` save failure is logged but doesn't fail the user's chat response |
| Soft delete on transactions/budgets | Allows future audit/undo; goals use hard delete (different product decision — goals are intentionally removed) |

---

## 5. Feature Breakdown

### 5.1 Authentication

**Purpose:** Identify users and protect all financial data endpoints.

**Components:**
- `internal/usecase/user.go` — `Register`, `Login`
- `internal/repository/postgres/user.go` — `Create`, `FindByEmail`, `GetAll`
- `internal/delivery/http/handler/user.go`
- `internal/delivery/http/middleware/auth.go`
- `utils/jwt.go`, `utils/claims.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/register` | ❌ | Create account |
| `POST` | `/api/v1/login` | ❌ | Get JWT |
| `GET` | `/api/v1/users` | ❌ | List all users ⚠️ |

**DB tables:** `users`

**Business rules:**
- Email must be unique (PostgreSQL constraint + PgError code `23505` mapped to `ErrConflict`)
- Password bcrypt hashed (cost 10) before storage
- JWT contains `id`, `name`, `email`; no expiry

**Edge cases / issues:**
- `GET /api/v1/users` is public and returns `password` field (hashed, but still exposed)
- Login does not check `deleted_at` — a soft-deleted user can still authenticate

---

### 5.2 Transactions

**Purpose:** Core financial ledger. Every income or expense a user records becomes a transaction.

**Components:**
- `internal/usecase/transaction.go`
- `internal/repository/postgres/transaction.go`
- `internal/delivery/http/handler/transaction.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/auth/v1/transactions` | ✅ | List user's transactions (paginated, filterable) |
| `POST` | `/api/auth/v1/transactions` | ✅ | Create transaction |
| `PUT` | `/api/auth/v1/transactions/:id` | ✅ | Update transaction |
| `DELETE` | `/api/auth/v1/transactions/:id` | ✅ | Soft-delete transaction |
| `GET` | `/api/auth/v1/transactions/monthly` | ✅ | Sum of current month expenses |
| `GET` | `/api/auth/v1/transactions/monthly-income` | ✅ | Sum of current month income |

**Query params for GET /transactions:** `month`, `year`, `limit` (default 10), `offset` (default 0)

**DB tables:** `transactions`

**Business rules:**
- `type` must be `INCOME` or `EXPENSE` (case-insensitive input, normalized to uppercase)
- `amount` must be > 0
- `category` required
- `date` required
- Ownership enforced: `WHERE user_id = $userID` on update/delete
- Soft delete: sets `deleted_at`, returns 404 if no rows affected (not found or wrong user)

**Indices:**
```sql
idx_transactions_user_category_date (user_id, category, date)
idx_transactions_full (user_id, category, type, date)
```

---

### 5.3 Budgets

**Purpose:** Set spending limits per category. Track how much of a budget has been consumed.

**Components:**
- `internal/usecase/budget.go`
- `internal/repository/postgres/budget.go`
- `internal/delivery/http/handler/budget.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/auth/v1/budgets` | ✅ | List user's budgets (filterable) |
| `POST` | `/api/auth/v1/budgets` | ✅ | Create budget |
| `GET` | `/api/auth/v1/budgets/usage` | ✅ | Usage summary with status |
| `GET` | `/api/auth/v1/budgets/:id` | ✅ | Get single budget |
| `PUT` | `/api/auth/v1/budgets/:id` | ✅ | Update budget |
| `DELETE` | `/api/auth/v1/budgets/:id` | ✅ | Soft-delete budget |

**Query params:**
- `GET /budgets`: `category`, `month`, `year`
- `GET /budgets/usage`: `year` (required), `month` (optional)

**DB tables:** `budgets`, joined with `transactions` for usage

**Business rules:**
- `period` must be `MONTHLY` or `YEARLY`
- `MONTHLY` requires `month` field
- `YEARLY` ignores `month` (set to null)
- `alert_threshold` defaults to 80 (%)
- Duplicate budget rejected: same `(user_id, category, period, month, year)`
- Update uses `COALESCE(NULLIF($1, 0), existing)` — send `0` to keep existing value
- Status thresholds: `>= 100%` = EXCEEDED, `>= alert_threshold%` = WARNING, else SAFE
- Change percent compares current period vs. previous period (month-1 or year-1)

**Edge cases:**
- `GET /budgets/:id` does not check `user_id` — ownership not enforced on single-item reads

---

### 5.4 Goals

**Purpose:** Define and track personal savings targets with deadlines.

**Components:**
- `internal/usecase/goal.go`
- `internal/repository/postgres/goal.go`
- `internal/delivery/http/handler/goal.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/auth/v1/goals` | ✅ | List goals (`?active=true` for active only) |
| `POST` | `/api/auth/v1/goals` | ✅ | Create goal |
| `GET` | `/api/auth/v1/goals/overview` | ✅ | Summary: totals + milestones + savings |
| `GET` | `/api/auth/v1/goals/milestones` | ✅ | Upcoming milestones (subset of overview) |
| `GET` | `/api/auth/v1/goals/:id` | ✅ | Get single goal |
| `PUT` | `/api/auth/v1/goals/:id` | ✅ | Update goal |
| `DELETE` | `/api/auth/v1/goals/:id` | ✅ | Hard-delete goal |
| `PATCH` | `/api/auth/v1/goals/contribute` | ✅ | Add contribution |

**DB tables:** `goals`, `transactions` (for net savings validation)

**Business rules:**
- Name required, target > 0, deadline must be future
- Duplicate guard: `(user_id, name, deadline >= NOW())` — same active goal name rejected
- Contribution validates `net_savings > 0` and `amount <= net_savings`
- `Contribute` sets `current_amount` to the submitted value (absolute, not additive)
- Auto-complete: `current_amount >= target_amount → status = COMPLETED`
- Update: `target_amount` only updated if `new > current_amount`
- Milestones: up to 4, ordered by nearest deadline, where goal is not yet completed (`target ≠ current`)
- Hard delete (no soft delete) — intentional; removed goals leave no trace

**State transitions:**
```
ONGOING → COMPLETED  (when contribution brings current_amount >= target_amount)
COMPLETED → ONGOING  (when target_amount is raised above current_amount via Update)
```

---

### 5.5 Dashboard

**Purpose:** Single-request summary of the user's financial health.

**Components:**
- `internal/usecase/dashboard.go`
- `internal/delivery/http/handler/dashboard.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/auth/v1/dashboard` | ✅ | Full financial snapshot |

**Response shape:**
```json
{
  "data": {
    "monthly_income":  number,
    "monthly_expense": number,
    "net_savings":     number,
    "budget_summary": { "total": n, "safe": n, "warning": n, "exceeded": n },
    "goal_summary":   { "total": n, "active": n, "completed": n },
    "active_goals":   [ GoalResponse... ]
  }
}
```

**DB tables:** `transactions`, `budgets`, `goals`

**Business rules:**
- `monthly_income` / `monthly_expense`: current calendar month only
- `net_savings`: all-time (not just this month)
- `budget_summary`: counts budgets by status for current month
- `goal_summary.total` = goals with `deadline >= NOW()`
- `goal_summary.completed` = goals in the active list where `status = COMPLETED` (note: `CountActive` counts by deadline, not status — minor inconsistency)
- `active_goals` = all goals with `deadline >= NOW()`, any status

**Performance note:** 6+ sequential DB queries per request, no caching.

---

### 5.6 AI Chat

**Purpose:** Allow users to ask financial questions with automatic context injection.

**Components:**
- `internal/usecase/chat.go` — context builder + Gemini API caller
- `internal/repository/postgres/ai_log.go` — log persistence
- `internal/delivery/http/handler/chat.go`

**Endpoints:**

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/auth/v1/chat` | ✅ | Ask a financial question |

**Request:** `{ "message": "..." }`  
**Response:** `{ "reply": "..." }`

**DB tables:** `transactions`, `budgets`, `goals`, `ai_logs` (write only)

**Context injected into prompt:**
- Current month/year label
- Monthly income, monthly expense, all-time net savings
- Total budgets count, count of EXCEEDED budgets
- Count of active goals

**External dependency:** Google Gemini 2.0 Flash  
**API key:** `GEMINI_API_KEY` environment variable  
**Failure behavior:** Returns `503` if key is missing or Gemini returns non-200

**Log behavior:** Every successful AI response is saved to `ai_logs`. Save failures are logged server-side (`log.Printf`) but do not fail the response to the client.

**Missing:** No endpoint to retrieve past chat history from `ai_logs`.

---

## 6. API Reference Summary

### Public Routes (no auth)

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/register` | Register |
| `POST` | `/api/v1/login` | Login → JWT |
| `GET` | `/api/v1/users` | List all users ⚠️ unauthenticated |

### Protected Routes (`Authorization: Bearer <token>` required)

| Method | Path | Description |
|---|---|---|
| **Transactions** | | |
| `GET` | `/api/auth/v1/transactions` | List (paginated, filterable by month/year) |
| `POST` | `/api/auth/v1/transactions` | Create |
| `PUT` | `/api/auth/v1/transactions/:id` | Update |
| `DELETE` | `/api/auth/v1/transactions/:id` | Soft delete |
| `GET` | `/api/auth/v1/transactions/monthly` | Current month expense total |
| `GET` | `/api/auth/v1/transactions/monthly-income` | Current month income total |
| **Budgets** | | |
| `GET` | `/api/auth/v1/budgets` | List (filterable by category/month/year) |
| `POST` | `/api/auth/v1/budgets` | Create |
| `GET` | `/api/auth/v1/budgets/usage` | Usage + status per budget |
| `GET` | `/api/auth/v1/budgets/:id` | Get by ID |
| `PUT` | `/api/auth/v1/budgets/:id` | Update |
| `DELETE` | `/api/auth/v1/budgets/:id` | Soft delete |
| **Goals** | | |
| `GET` | `/api/auth/v1/goals` | List (`?active=true`) |
| `POST` | `/api/auth/v1/goals` | Create |
| `GET` | `/api/auth/v1/goals/overview` | Summary + milestones + savings total |
| `GET` | `/api/auth/v1/goals/milestones` | Upcoming milestones only |
| `GET` | `/api/auth/v1/goals/:id` | Get by ID |
| `PUT` | `/api/auth/v1/goals/:id` | Update |
| `DELETE` | `/api/auth/v1/goals/:id` | Hard delete |
| `PATCH` | `/api/auth/v1/goals/contribute` | Contribute to goal |
| **Dashboard** | | |
| `GET` | `/api/auth/v1/dashboard` | Aggregated financial snapshot |
| **Chat** | | |
| `POST` | `/api/auth/v1/chat` | Ask AI assistant |

---

*Document generated from codebase analysis — `financial-planning-golang` · 2026-05-17*
