# Frontend Migration Guide — API v1.1.0

This document covers all breaking changes, new endpoints, and behavioral corrections between the original spec and the actual backend implementation. Frontend developers must address every **BREAKING** item before deploying.

---

## Breaking Changes

### 1. Transaction List — Wrong Path

| | Before (incorrect spec) | After (actual) |
|---|---|---|
| Method | `GET` | `GET` |
| Path | `/api/auth/v1/transactions/users` | `/api/auth/v1/transactions` |

**Action:** Update all transaction list fetch calls to use `/api/auth/v1/transactions`.

---

### 2. Transaction Count Endpoints — Removed & Split

`GET /api/auth/v1/transactions/count` **does not exist**. Two separate endpoints replace it:

| Purpose | Endpoint |
|---|---|
| Monthly expense total | `GET /api/auth/v1/transactions/monthly` |
| Monthly income total | `GET /api/auth/v1/transactions/monthly-income` |

Both return:
```json
{ "total": 5000000 }
```

**Action:** Remove any calls to `/transactions/count`. Use the two separate endpoints above.

---

### 3. Budget Update — Method Changed

| | Before (incorrect spec) | After (actual) |
|---|---|---|
| Method | `PATCH` | `PUT` |
| Path | `/api/auth/v1/budgets/{id}` | `/api/auth/v1/budgets/{id}` |

Request body (all fields optional — omit or send `0`/`""` to keep existing value):
```json
{
  "limit_amount": 3000000,
  "alert_threshold": 80,
  "category": "Food"
}
```

Response (snake_case, no `created_at`):
```json
{
  "id": 1,
  "user_id": 42,
  "category": "Food",
  "period": "MONTHLY",
  "month": 5,
  "year": 2026,
  "limit_amount": 3000000,
  "alert_threshold": 80
}
```

**Action:** Change `PATCH` to `PUT` for budget update calls.

---

### 4. Goal Update — Method Changed

| | Before (incorrect spec) | After (actual) |
|---|---|---|
| Method | `PATCH` | `PUT` |
| Path | `/api/auth/v1/goals/{id}` | `/api/auth/v1/goals/{id}` |

**Action:** Change `PATCH` to `PUT` for goal update calls.

---

### 5. Budget Usage — Response Is a Direct Array

`GET /api/auth/v1/budgets/usage` returns a **direct JSON array**, not a wrapped object.

```json
[
  {
    "id": 1,
    "user_id": 42,
    "category": "Food",
    "period": "MONTHLY",
    "month": 5,
    "year": 2026,
    "limit_amount": 3000000,
    "spent": 2100000,
    "percentage": 70.0,
    "status": "ON_TRACK"
  }
]
```

**Action:** Do not unwrap a `.data` property — iterate the response directly.

---

### 6. Transaction List — `total` Is Record Count, Not Page Count

`GET /api/auth/v1/transactions` response:
```json
{
  "data": [...],
  "total": 47,
  "page": 1,
  "limit": 10
}
```

`total` = total number of matching records across all pages, not the number of pages.

**Action:** If computing page count, calculate `Math.ceil(total / limit)` client-side.

---

### 7. Budget Response Inconsistency (Three Shapes)

The same budget resource returns three different shapes depending on the operation:

| Operation | Endpoint | Key difference |
|---|---|---|
| List | `GET /api/auth/v1/budgets` | snake_case, includes `created_at` |
| Get by ID | `GET /api/auth/v1/budgets/{id}` | camelCase (`limitAmount`, `alertThreshold`, `createdAt`) |
| Update | `PUT /api/auth/v1/budgets/{id}` | snake_case, **no** `created_at` |

**Action:** Normalise budget objects in your data layer after fetching. Do not assume consistent casing.

---

### 8. Goal Contribution — Sets Amount, Does Not Increment

`POST /api/auth/v1/goals/{id}/contribute` **sets** `current_amount` to the value you send. It does not add to the existing amount.

Request:
```json
{ "amount": 500000 }
```

If the goal currently has `current_amount = 200000` and you send `500000`, the result is `current_amount = 500000`, not `700000`.

**Action:** Read the current `current_amount` first, add the desired contribution, then send the total.

When `current_amount >= target_amount`, the goal status automatically becomes `COMPLETED`.

---

### 9. Goals Use Hard Deletes

`DELETE /api/auth/v1/goals/{id}` **permanently deletes** the goal. It cannot be recovered.

**Action:** Show a confirmation dialog before calling this endpoint.

---

## New Endpoints

### Dashboard

```
GET /api/auth/v1/dashboard
Authorization: Bearer <token>  (or cookie)
```

Response:
```json
{
  "monthly_income": 10000000,
  "monthly_expense": 7500000,
  "net_savings": 25000000,
  "budget_status": {
    "total": 5,
    "on_track": 3,
    "warning": 1,
    "exceeded": 1
  },
  "goal_progress": {
    "total": 4,
    "completed": 1,
    "ongoing": 3
  },
  "recent_transactions": [...]
}
```

---

### AI Chat

```
POST /api/auth/v1/chat
Authorization: Bearer <token>  (or cookie)
Content-Type: application/json
```

Request:
```json
{ "message": "Am I overspending this month?" }
```

Response `200`:
```json
{ "reply": "Based on your data, you've spent 75% of your monthly income..." }
```

Response `503` — Gemini API key not configured on server:
```json
{ "error": "AI service unavailable" }
```

The AI assistant uses the user's actual financial data (monthly income, expenses, net savings, budget status, active goals) as context. It responds in the same language the user writes in (Indonesian or English).

**Action:** Handle the `503` gracefully — show "AI assistant is currently unavailable" rather than a generic error.

---

### Goal Milestones

```
GET /api/auth/v1/goals/milestones
Authorization: Bearer <token>  (or cookie)
```

Returns milestones across all active goals:
```json
[
  {
    "goal_id": 1,
    "goal_name": "Emergency Fund",
    "milestone_percent": 25,
    "amount_at_milestone": 2500000,
    "reached": true
  }
]
```

---

## Authentication

All `/api/auth/v1/*` routes require authentication. Token accepted via:

1. Cookie: `accessToken=<jwt>`
2. Header: `Authorization: Bearer <jwt>`

On failure: `401 Unauthorized`
```json
{ "error": "unauthorized" }
```

No changes to the auth flow itself. `/api/v1/register` and `/api/v1/login` remain unauthenticated.

---

## Full Endpoint Reference

| Method | Path | Auth | Notes |
|---|---|---|---|
| `POST` | `/api/v1/register` | No | — |
| `POST` | `/api/v1/login` | No | Sets `accessToken` cookie |
| `GET` | `/api/auth/v1/dashboard` | Yes | **NEW** |
| `GET` | `/api/auth/v1/transactions` | Yes | **PATH FIXED** (was `/transactions/users`) |
| `POST` | `/api/auth/v1/transactions` | Yes | — |
| `GET` | `/api/auth/v1/transactions/{id}` | Yes | — |
| `PUT` | `/api/auth/v1/transactions/{id}` | Yes | — |
| `DELETE` | `/api/auth/v1/transactions/{id}` | Yes | Soft delete |
| `GET` | `/api/auth/v1/transactions/monthly` | Yes | **REPLACES** `/transactions/count` |
| `GET` | `/api/auth/v1/transactions/monthly-income` | Yes | **REPLACES** `/transactions/count` |
| `GET` | `/api/auth/v1/budgets` | Yes | Returns snake_case |
| `POST` | `/api/auth/v1/budgets` | Yes | — |
| `GET` | `/api/auth/v1/budgets/{id}` | Yes | Returns camelCase |
| `PUT` | `/api/auth/v1/budgets/{id}` | Yes | **METHOD FIXED** (was PATCH) |
| `DELETE` | `/api/auth/v1/budgets/{id}` | Yes | — |
| `GET` | `/api/auth/v1/budgets/usage` | Yes | Direct array response |
| `GET` | `/api/auth/v1/goals` | Yes | — |
| `POST` | `/api/auth/v1/goals` | Yes | — |
| `GET` | `/api/auth/v1/goals/{id}` | Yes | — |
| `PUT` | `/api/auth/v1/goals/{id}` | Yes | **METHOD FIXED** (was PATCH) |
| `DELETE` | `/api/auth/v1/goals/{id}` | Yes | **Hard delete** |
| `POST` | `/api/auth/v1/goals/{id}/contribute` | Yes | Sets amount, does not increment |
| `GET` | `/api/auth/v1/goals/milestones` | Yes | **NEW** |
| `POST` | `/api/auth/v1/chat` | Yes | **NEW** — 503 if AI unavailable |

---

## Required Frontend Actions Summary

- [ ] Fix transaction list URL: `/transactions/users` → `/transactions`
- [ ] Remove `/transactions/count` calls; replace with `/transactions/monthly` and `/transactions/monthly-income`
- [ ] Change budget update from `PATCH` to `PUT`
- [ ] Change goal update from `PATCH` to `PUT`
- [ ] Stop unwrapping `.data` from budget usage response — iterate directly
- [ ] Fix pagination: compute page count as `Math.ceil(total / limit)`
- [ ] Normalise budget response shape (different casing per operation)
- [ ] Fix goal contribution: send total target value, not delta
- [ ] Add confirmation dialog before deleting goals (hard delete)
- [ ] Add `GET /api/auth/v1/dashboard` call for dashboard screen
- [ ] Integrate `POST /api/auth/v1/chat` for AI assistant feature
- [ ] Handle `503` from chat endpoint gracefully
- [ ] Optionally surface goal milestones via `GET /api/auth/v1/goals/milestones`
