# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Product

**AI Financial Planning Platform** — a web backend that helps users manage personal finances with AI-powered insights.

Core MVP features:
1. **Transaction Management** — record income and expenses, categorize, filter by month/year, paginate
2. **Budget Management** — create budgets per category (MONTHLY/YEARLY), track usage with SAFE/WARNING/EXCEEDED status
3. **Goals** — create financial targets with deadline and target amount, track progress via manual contributions
4. **Dashboard** — single endpoint aggregating monthly income, expense, net savings, budget summary, active goals
5. **AI Chatbot** — Gemini-powered Q&A about the user's financial data (net savings, expense, budget status)

Out of scope: multi-currency, bank integrations, auto-investment, advanced analytics.

## Commands

```bash
# Run with hot reload (development)
air

# Run directly
go run main.go

# Build
go build -o ./tmp/main.exe .

# Run tests
go test ./...
```

## Environment Setup

Copy `.env` and populate — all vars are required at startup:

```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=financial_planning
DB_HOST=localhost
DB_PORT=5432
SECRET_KEY=<jwt-signing-secret>
GEMINI_API_KEY=<gemini-api-key>
CORS_ORIGIN=http://localhost:5173
```

`GEMINI_API_KEY` is used by `ChatUseCase` to call the Gemini REST API for the AI chatbot endpoint.
`CORS_ORIGIN` sets the `Access-Control-Allow-Origin` header (defaults to `http://localhost:5173`).

Database migrations in `db/migrations/` use golang-migrate format (`NNNNN_desc.up.sql` / `.down.sql`). Run them manually or wire up the migrate CLI.

## Architecture

Clean architecture under `internal/`, strict dependency direction enforced by Go's `internal/` visibility:

```
HTTP Request
  → internal/delivery/http/   (Gin handlers + router + middleware)
  → internal/usecase/          (business logic, validation — no framework imports)
  → internal/repository/postgres/ (raw SQL — no business logic)
  → internal/domain/           (entities, DTOs, repository interfaces, sentinel errors)
```

**Dependency rule:**
```
domain/ ← usecase/ ← delivery/http/
domain/ ← repository/postgres/
```

`main.go` is the only place that imports all layers for wiring.

**Struct-based handlers:** Each domain has a handler struct that depends on the concrete use case struct:
```go
type TransactionHandler struct {
    uc *usecase.TransactionUseCase
}
```

**Constructor injection in main.go:**
```go
txRepo  := postgres.NewTransactionRepository(db)
txUC    := usecase.NewTransactionUseCase(txRepo)
// GoalUseCase receives txRepo for cross-domain net savings check
goalUC  := usecase.NewGoalUseCase(goalRepo, txRepo)
// DashboardUseCase and ChatUseCase aggregate data from all three repos
dashboardUC := usecase.NewDashboardUseCase(txRepo, budgetRepo, goalRepo)
chatUC      := usecase.NewChatUseCase(txRepo, budgetRepo, goalRepo)
```

## Routing

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/register` | No | Register user |
| POST | `/api/v1/login` | No | Login, returns JWT |
| GET | `/api/v1/users` | No | List all users |
| GET | `/api/auth/v1/transactions` | JWT | List transactions (paginated, filterable) |
| POST | `/api/auth/v1/transactions` | JWT | Create transaction |
| PUT | `/api/auth/v1/transactions/:id` | JWT | Update transaction |
| DELETE | `/api/auth/v1/transactions/:id` | JWT | Soft-delete transaction |
| GET | `/api/auth/v1/transactions/monthly` | JWT | Current month total expenses |
| GET | `/api/auth/v1/transactions/monthly-income` | JWT | Current month total income |
| GET | `/api/auth/v1/budgets` | JWT | List budgets (filterable) |
| POST | `/api/auth/v1/budgets` | JWT | Create budget |
| GET | `/api/auth/v1/budgets/usage` | JWT | Budget usage with SAFE/WARNING/EXCEEDED status |
| GET | `/api/auth/v1/budgets/:id` | JWT | Get budget by ID |
| PUT | `/api/auth/v1/budgets/:id` | JWT | Update budget |
| DELETE | `/api/auth/v1/budgets/:id` | JWT | Soft-delete budget |
| GET | `/api/auth/v1/goals` | JWT | List goals (`?active=true` for active only) |
| POST | `/api/auth/v1/goals` | JWT | Create goal |
| GET | `/api/auth/v1/goals/overview` | JWT | Goals overview (total, savings, milestones) |
| GET | `/api/auth/v1/goals/milestones` | JWT | Upcoming milestones (next 4 by deadline) |
| GET | `/api/auth/v1/goals/:id` | JWT | Get goal by ID |
| PUT | `/api/auth/v1/goals/:id` | JWT | Update goal |
| DELETE | `/api/auth/v1/goals/:id` | JWT | Hard-delete goal |
| PATCH | `/api/auth/v1/goals/contribute` | JWT | Set goal current_amount, auto-completes if >= target |
| GET | `/api/auth/v1/dashboard` | JWT | Aggregated dashboard: income, expense, savings, budget summary, goals |
| POST | `/api/auth/v1/chat` | JWT | AI chatbot — ask questions about your finances |

CORS origin is configured via `CORS_ORIGIN` env var (defaults to `http://localhost:5173` in `main.go`).

## Authentication

- Login returns a JWT (HS256, signed with `SECRET_KEY`) in both the response body and a 1-hour cookie.
- Middleware parses the token and stores claims in Gin context.
- Use `utils.ClaimId(c)` inside any protected handler to get the authenticated user's ID.
- JWT has no expiry in the claims — only the cookie expiry enforces session length.

## Database Access

- Raw SQL only (no ORM). PostgreSQL placeholders are `$1`, `$2`, etc.
- Soft deletes via `deleted_at` timestamp on `users`, `transactions`, `budgets`. Goals use hard deletes.
- All queries filter `WHERE deleted_at IS NULL` for soft-deleted entities.
- No SQL transactions — multi-step operations execute queries independently.

## Domain Notes

- **Transactions** — `type` must be `INCOME` or `EXPENSE` (normalized to uppercase in use case). `amount` is stored as int, returned as float64. Supports pagination (`limit`/`offset`) and optional month/year filtering via `EXTRACT()`.

- **Budgets** — `period` must be `MONTHLY` or `YEARLY`. `alert_threshold` defaults to 80 (%). Usage endpoint (`GET /budgets/usage`) compares current month spending to the previous month and returns `SAFE` / `WARNING` / `EXCEEDED` based on `alert_threshold`. Update uses `NULLIF` so zero values mean "no change".

- **Goals** — `PATCH /goals/contribute` body: `{"goal_id": N, "contribution": M}`. `contribution` sets `current_amount` directly (not an increment). Use case validates `contribution <= net_savings` by querying the transaction repository. Auto-sets status to `COMPLETED` when `current_amount >= target_amount`. `GoalUseCase` injects `TransactionRepository` to avoid cross-domain SQL.

- **Dashboard** — `GET /dashboard` aggregates current month income + expense, all-time net savings, budget status counts for current month, and all active goals. No query params needed — uses server-side `time.Now()`.

- **AI Chatbot** — `POST /chat` body: `{"message": "your question"}`. The use case fetches the user's current financial data, builds a context prompt, and calls the Gemini REST API (`gemini-2.0-flash`). Returns 503 if `GEMINI_API_KEY` is not set. Responds in the language the user writes in.

## Sentinel Errors

Defined in `internal/domain/errors.go`:
- `ErrNotFound` → 404
- `ErrConflict` → 409
- `ErrUnauthorized` → 401
- `ErrInvalidInput` → 400

Use case-specific errors: `usecase.ErrUserExists`, `usecase.ErrInvalidCredentials`, `usecase.ErrChatUnavailable`.
