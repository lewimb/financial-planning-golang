# Frontend Developer Guide

**Base URL:** `http://localhost:8080`  
**Auth:** All `/api/auth/v1/…` routes require a JWT. Send it as:
- `Authorization: Bearer <token>` header, **or**
- `accessToken` cookie (set automatically on login)

> **Cookie-first:** The server sets `accessToken` as an httpOnly cookie on login. In React Router v7, read the cookie server-side via `tokenParser(request)` and forward it as a `Bearer` header to all backend calls. Do **not** store the token in `localStorage`.

---

## 1. App Overview

AI Financial Planning app. Users:
1. Register / login
2. Complete an onboarding profile (income, expenses, goals)
3. Record transactions (income + expenses)
4. Set budgets per category
5. Track savings goals
6. Get ML-powered insights (spending forecast, anomaly alerts, analysis)
7. Chat with an AI assistant about their finances

---

## 2. Authentication

### Register

```
POST /api/v1/register
```

```json
// Request
{ "email": "user@example.com", "password": "secret123", "name": "John" }

// Response 200
{ "message": "User registered successfully" }

// Error 409 — email already in use
{ "error": "already exists" }
```

---

### Login

```
POST /api/v1/login
```

```json
// Request
{ "email": "user@example.com", "password": "secret123" }

// Response 200
{
  "message": "Login Successfully",
  "status": "200",
  "data": { "token": "eyJhbGci..." }
}
```

The server also sets an `accessToken` httpOnly cookie. Read from the cookie in protected route loaders — never prompt users to paste tokens.

---

## 3. Financial Profile (Onboarding)

Collect this **once** after registration. Used to personalize AI responses.

**Onboarding gate:** In your root protected layout loader, call `GET /financial-profile`. On 404, redirect the user to `/auth/onboarding`. Block all other routes until the profile exists.

```
GET /financial-profile → 404 → redirect to /auth/onboarding
GET /financial-profile → 200 → proceed to dashboard
```

### Create / Update Profile

```
POST /api/auth/v1/financial-profile
```

```json
// Request — all fields except spending_habit and risk_level are required
{
  "monthly_income": 8000000,
  "fixed_expenses": 3000000,
  "current_savings": 5000000,
  "debt": 1000000,
  "employment_status": "employed",
  "financial_goals": ["emergency_fund", "house"],
  "spending_habit": "moderate",
  "risk_level": "medium"
}

// Response 200
{
  "message": "profile saved",
  "data": {
    "monthly_income": 8000000,
    "fixed_expenses": 3000000,
    "current_savings": 5000000,
    "debt": 1000000,
    "employment_status": "employed",
    "financial_goals": ["emergency_fund", "house"],
    "spending_habit": "moderate",
    "risk_level": "medium",
    "net_available": 4000000,
    "created_at": "2026-05-19T10:00:00Z",
    "updated_at": "2026-05-19T10:00:00Z"
  }
}
```

**`net_available`** = `monthly_income − fixed_expenses − debt`. Server computes it — never send it.

**`financial_goals`** — replace the whole list on every save. Accepted values:
`emergency_fund`, `house`, `investment`, `education`, `travel`, `retirement` (or any string).

**Validation errors (400):**
- `"monthly_income must be >= 0"`
- `"financial_goals cannot be empty"`
- `"employment_status is required"`

---

### Get Profile

```
GET /api/auth/v1/financial-profile
```

```json
// Response 200
{ "data": { ...same shape as above... } }

// Response 404 — profile not set yet → show onboarding form
{ "error": "profile not found" }
```

---

## 4. Transactions

### List

```
GET /api/auth/v1/transactions?month=5&year=2026&limit=10&offset=0
```

All query params optional. `limit=0` returns all records.

```json
// Response 200
{
  "data": [
    {
      "id": 1,
      "amount": 500000,
      "category": "Food",
      "type": "EXPENSE",
      "date": "2026-05-01T00:00:00Z",
      "description": "Lunch"
    }
  ],
  "total": 42
}
```

`total` = total matching records (not page count). Pages = `Math.ceil(total / limit)`.

---

### Create

```
POST /api/auth/v1/transactions
```

```json
// Request
{
  "amount": 500000,
  "category": "Food",
  "type": "EXPENSE",
  "date": "2026-05-01T00:00:00Z",
  "description": "Lunch"
}

// Response 200
{ "message": "Transaction created successfully" }
```

`type` must be `"INCOME"` or `"EXPENSE"` (case-insensitive, normalized server-side).

---

### Update

```
PUT /api/auth/v1/transactions/:id
```

Same body as create. Returns `{ "message": "..." }`.

---

### Delete (soft)

```
DELETE /api/auth/v1/transactions/:id
```

```json
// Response 200
{ "message": "Transaction deleted successfully" }
```

---

### Monthly Totals

```
GET /api/auth/v1/transactions/monthly          → total expenses this month
GET /api/auth/v1/transactions/monthly-income   → total income this month
```

```json
// Response
{ "total": 3200000, "message": "success" }
```

---

## 5. Budgets

### List

```
GET /api/auth/v1/budgets?category=Food&month=5&year=2026
```

```json
// Response 200 — snake_case fields
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "category": "Food",
      "period": "MONTHLY",
      "month": 5,
      "year": 2026,
      "limit_amount": 2000000,
      "alert_threshold": 80,
      "created_at": "2026-05-01T00:00:00Z"
    }
  ]
}
```

---

### Create

```
POST /api/auth/v1/budgets
```

```json
// Request
{
  "category": "Food",
  "period": "MONTHLY",
  "month": 5,
  "year": 2026,
  "limit_amount": 2000000,
  "alert_threshold": 80
}

// Response 201
{ "message": "Budget created successfully" }
```

`period`: `"MONTHLY"` or `"YEARLY"`. For `MONTHLY`, include `month` (1–12).

---

### Budget Usage

```
GET /api/auth/v1/budgets/usage?year=2026&month=5
```

```json
// Response — direct array (no wrapper object)
[
  {
    "budget_id": 1,
    "category": "Food",
    "period": "MONTHLY",
    "limit": 2000000,
    "alert_threshold": 80,
    "used": 1400000,
    "remaining": 600000,
    "percentage": 70.0,
    "status": "SAFE",
    "change_percent": -5.2
  }
]
```

`status` values: `SAFE` / `WARNING` (>= alert_threshold%) / `EXCEEDED` (>= 100%).

---

### Get by ID

```
GET /api/auth/v1/budgets/:id
```

> ⚠️ **camelCase inconsistency:** This endpoint returns camelCase fields (`userId`, `limitAmount`, `alertThreshold`, `createdAt`). All other budget endpoints use snake_case. Map to snake_case when normalizing state client-side.

---

### Update

```
PUT /api/auth/v1/budgets/:id
```

```json
// Request — send only what changes; zero means "keep existing"
// ⚠️ camelCase body (known backend inconsistency — matches GET /budgets/:id)
{ "limitAmount": 2500000, "alertThreshold": 75 }

// Response 200 — snake_case
{ "data": { ...updated budget in snake_case... } }
```

---

### Delete (soft)

```
DELETE /api/auth/v1/budgets/:id
```

---

## 6. Goals

### List

```
GET /api/auth/v1/goals?active=true
```

`active=true` returns only goals with `deadline >= now`.

```json
// Response
{
  "data": [
    {
      "id": 1,
      "name": "Emergency Fund",
      "target_amount": 10000000,
      "current_amount": 3000000,
      "status": "ONGOING",
      "deadline": "2026-12-31T00:00:00Z",
      "description": "3 months of expenses",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### Create

```
POST /api/auth/v1/goals
```

```json
// Request
{
  "name": "Emergency Fund",
  "target_amount": 10000000,
  "description": "3 months of expenses",
  "deadline": "2026-12-31T00:00:00Z"
}

// Response 201
{ "message": "Goal created successfully" }
```

---

### Update

```
PUT /api/auth/v1/goals/:id
```

Same body as create. If `target_amount < current_amount`, target is NOT lowered.

---

### Delete

```
DELETE /api/auth/v1/goals/:id
```

Permanent (no soft delete for goals).

---

### Contribute

```
PATCH /api/auth/v1/goals/contribute
```

```json
// Request — sets current_amount directly (not an increment)
{ "goal_id": 1, "contribution": 5000000 }

// Response 200
{ "message": "Contribution successful" }
```

If `contribution >= target_amount`, goal auto-completes (status → `COMPLETED`).  
Contribution must not exceed all-time net savings.

> **UI pattern:** Show the current amount and let the user type a new total, OR show a "add Rp X" field and compute `new_total = current_amount + delta` client-side before sending.

---

### Overview

```
GET /api/auth/v1/goals/overview
```

```json
// Response
{
  "message": "success",
  "data": {
    "total_goals": 3,
    "savings": 8000000,
    "goals": [ ...next 4 milestones by deadline... ]
  }
}
```

> **Optimization:** The `goals` array in the overview response is identical to what `GET /goals/milestones` returns. If you're already calling `/overview`, skip the separate `/milestones` call and use `overviewData.goals` directly.

---

### Milestones

```
GET /api/auth/v1/goals/milestones
```

Returns next 4 goals by deadline that are not yet completed. Wraps in `{ "data": [...] }`.

---

## 7. Dashboard

```
GET /api/auth/v1/dashboard
```

No params — always uses current month on the server.

```json
// Response
{
  "data": {
    "monthly_income": 8000000,
    "monthly_expense": 3200000,
    "net_savings": 48000000,
    "budget_summary": {
      "total": 3,
      "safe": 2,
      "warning": 1,
      "exceeded": 0
    },
    "goal_summary": {
      "total": 3,
      "active": 2,
      "completed": 1
    },
    "active_goals": [ ...GoalResponse array... ]
  }
}
```

---

## 8. ML Insights

All three endpoints are GET — the backend fetches transactions and calls the ML service internally. Frontend never sends transaction data to ML.

> **503 handling:** All ML endpoints return 503 when the Python ML service is unreachable. Show a "ML insights unavailable" state rather than an error page. ML service must be running at port 8000.

### Spending Analysis

```
GET /api/auth/v1/ml/analysis?year=2026&month=05
```

```json
// Response
{
  "total_expense": 3200000,
  "avg_daily": 106666.67,
  "top_category": "food",
  "spending_distribution": {
    "food": 1400000,
    "transport": 900000,
    "utilities": 900000
  }
}
```

---

### Anomaly Detection

```
GET /api/auth/v1/ml/anomaly?year=2026&month=05
```

```json
// Response
{
  "anomalies": [
    { "date": "2026-05-15", "amount": 2500000 }
  ],
  "summary": "You spent unusually high on 1 day(s)"
}
```

Empty `anomalies: []` = no unusual days. Requires at least 5 unique expense days in the data.

---

### Spending Forecast

```
GET /api/auth/v1/ml/forecast?periods=30
```

```json
// Response
{
  "predicted_monthly_spending": 3500000,
  "daily_forecast": [
    { "date": "2026-05-20", "predicted_amount": 120000 },
    { "date": "2026-05-21", "predicted_amount": 95000 }
  ]
}
```

`periods` = days ahead (1–365, default 30). Forecast accuracy improves with 30+ days of history.

> ⚠️ **Timeout:** Forecast can take up to **60 seconds**. Show a persistent loading skeleton (not a toast spinner) and do not retry automatically. Accuracy improves significantly with 30+ days of history — warn the user if they have fewer records.

---

## 9. AI Chat

```
POST /api/auth/v1/chat
```

```json
// Request
{ "message": "Berapa pengeluaran saya bulan ini?" }

// Response 200
{ "reply": "Pengeluaran Anda bulan ini adalah Rp 3.200.000." }

// Error 503 — Gemini API key not configured
{ "error": "AI service unavailable" }
```

The AI has access to: current month income/expense, budget status, active goals, and the user's financial profile.  
Responds in the same language the user writes in (Indonesian or English).

---

## 10. Pages to Build

| Page | Purpose | Key APIs |
|---|---|---|
| **Register / Login** | Entry point | `POST /register`, `POST /login` |
| **Onboarding** | Financial profile form (first-time, multi-step) | `POST /financial-profile` |
| **Dashboard** | Overview of finances + ML insights | `GET /dashboard`, `GET /ml/analysis`, `GET /ml/anomaly` |
| **Transactions** | List + add income/expense | `GET /transactions`, `POST /transactions`, `PUT /transactions/:id`, `DELETE /transactions/:id` |
| **Budgets** | Create budgets + view usage | `GET /budgets/usage`, `POST /budgets`, `PUT /budgets/:id`, `DELETE /budgets/:id` |
| **Goals** | Track savings goals + contribute | `GET /goals`, `POST /goals`, `PUT /goals/:id`, `DELETE /goals/:id`, `PATCH /goals/contribute` |
| **Forecast** | Future spending prediction | `GET /ml/forecast` |
| **Profile Settings** | Edit financial profile + account settings | `GET /financial-profile`, `POST /financial-profile` |
| **AI Chat** | Ask questions about finances | `POST /chat` |

---

## 11. Onboarding Flow

```
User registers
    ↓
App calls GET /financial-profile
    ↓
404 → Redirect to /auth/onboarding (block all other routes)
    ↓
User fills multi-step form → POST /financial-profile
    ↓
Redirect to Dashboard
```

**Implementation in React Router v7:** Add the profile check to the `/auth/layout.tsx` loader. On 404, `throw redirect("/auth/onboarding")`. The onboarding route must be excluded from the auth layout redirect so users can access it without a profile.

---

## 12. Data Flow

```
Frontend
  ↓ accessToken cookie (httpOnly) or Authorization: Bearer <token>
Go Backend (port 8080)
  ├── Reads/writes PostgreSQL (transactions, budgets, goals, profile)
  ├── Calls ML Service (port 8000) for analysis, anomaly, forecast
  │     ML Service is stateless — backend sends full transaction list each time
  └── Calls Gemini API for AI chat responses
  ↓
Frontend receives structured JSON
```

---

## 13. Error Handling Reference

| Status | Meaning |
|---|---|
| `200` | Success |
| `201` | Created |
| `400` | Bad request / validation error — show `error` field to user |
| `401` | Not authenticated — redirect to login |
| `404` | Resource not found (or profile not set up) |
| `409` | Conflict — duplicate resource |
| `500` | Server error — show generic message |
| `503` | External service (Gemini / ML) unavailable — show feature-specific fallback |

All errors return: `{ "error": "description" }`

---

## 14. Known API Quirks

- **Budget camelCase inconsistency:** `PUT /budgets/:id` request body uses camelCase (`limitAmount`, `alertThreshold`). `GET /budgets/:id` response is also camelCase. All other budget endpoints (list, create, usage) use snake_case. This is a backend inconsistency — handle it field-by-field, don't assume uniform casing.
- **Budget usage endpoint** returns a raw **array** — no `{ data: [...] }` wrapper.
- **Goal contribution** sets `current_amount` directly — not an increment. To add Rp 200k to an existing Rp 3M goal, send `{ goal_id: 1, contribution: 3200000 }`.
- **`/goals/overview` goals array** is identical to `/goals/milestones`. Avoid both calls — use the overview data for milestones.
- **Pagination:** `GET /transactions?limit=10&offset=0`. Total pages = `Math.ceil(total / limit)`.
- **ML timeout:** forecast can take up to 60 seconds. Show persistent loading state.
- **Profile goals vs savings goals:** `financial_goals` in the profile (strings like `"house"`) are different from Goals (savings targets with amounts and deadlines).
- **`/api/v1/users`** — unauthenticated endpoint. Internal debug use only; do not expose in production UI.
