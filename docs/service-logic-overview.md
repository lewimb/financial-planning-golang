# Service Logic Overview

Complete reference for all business logic in `internal/usecase/`. Each use case is described with its validation rules, cross-domain dependencies, and error contracts.

---

## UserUseCase

**File:** `internal/usecase/user.go`  
**Dependencies:** `UserRepository`, `bcrypt`, `utils.GenerateJwt`

### Register
1. Hash password with `bcrypt.GenerateFromPassword(cost=10)`
2. Call `repo.Create(email, hashedPassword, name)`
3. On `domain.ErrConflict` (pgErr 23505) → return `ErrUserExists`

### Login
1. Fetch user by email via `repo.FindByEmail`
2. On not found → `ErrInvalidCredentials` (no distinction between "wrong email" vs "wrong password")
3. `bcrypt.CompareHashAndPassword` — on mismatch → `ErrInvalidCredentials`
4. `utils.GenerateJwt(userID, name, email)` — HS256, claims: `{id, name, email, iss:lewimb}`, **no `exp` claim**

---

## TransactionUseCase

**File:** `internal/usecase/transaction.go`  
**Dependencies:** `TransactionRepository`

### Create / Update — Validation Rules
| Field | Rule |
|---|---|
| `type` | `strings.ToUpper` then must be `INCOME` or `EXPENSE` |
| `amount` | must be `> 0` |
| `category` | must be non-empty string |
| `date` | must be non-zero `time.Time` |

### Delete
Soft delete — no business logic, delegates to repo.

### GetMonthlyExpenses / GetMonthlyIncome
Server-side `time.Now()` determines the current period. No date parameters accepted.

### GetTransactions
Optional `year` and `month` string params, passed to repo for `EXTRACT` filtering. Default `limit=10`, `offset=0` if not parseable.

---

## BudgetUseCase

**File:** `internal/usecase/budget.go`  
**Dependencies:** `BudgetRepository`

### Create — Validation Rules
| Field | Rule |
|---|---|
| `category` | non-empty |
| `period` | must be `MONTHLY` or `YEARLY` |
| `year` | must be `> 0` |
| `limit_amount` | must be `> 0` |
| `month` | required if `MONTHLY`; forced to `nil` if `YEARLY` |
| `alert_threshold` | defaults to `80` if zero |

### Duplicate Detection
Repository checks `SELECT EXISTS` before inserting — returns `domain.ErrConflict` if `(user_id, category, period, month, year)` already exists.

### Update (NULLIF Pattern)
Zero values for `limitAmount` / `alertThreshold` and empty string for `category` preserve the existing DB value. True partial update.

### GetUsage — Budget Status Logic
```
percentage = ROUND(used / limit × 100)
if percentage >= 100 → EXCEEDED
if percentage >= alert_threshold → WARNING
else → SAFE
change_percent = ROUND((used - prev_used) / prev_used × 100)
  if prev_used == 0 and used > 0 → change_percent = 100
```

---

## GoalUseCase

**File:** `internal/usecase/goal.go`  
**Dependencies:** `GoalRepository`, **`TransactionRepository`** (cross-domain)

### Create — Validation Rules
| Field | Rule |
|---|---|
| `name` | non-empty |
| `target_amount` | must be `> 0` |
| `deadline` | must be in the future (`time.Now()`) |

Duplicate check: repo uses `WHERE NOT EXISTS (... name=$7 AND deadline >= NOW())` — same name + future deadline = conflict.

### Update
Same name and target_amount validations. No deadline future check on update. Target amount update is constrained by repo: only updates if `$2 > current_amount`.

### Contribute — Cross-Domain Validation
```
if contribution <= 0 → error
net = TransactionRepository.GetNetSavings(userID)  ← cross-domain
if net <= 0 → error "no net savings"
if contribution > int(net) → error "exceeds available savings"
→ GoalRepository.Contribute sets current_amount directly (NOT incremental)
→ Auto-completes goal if current_amount >= target_amount
```

### GetOverview
Aggregates 3 repo calls:
1. `GetSavingsTotal` — `SUM(current_amount)` all goals (not just active)
2. `GetUpcomingMilestones` — next 4 goals by deadline, `target_amount ≠ current_amount`
3. `CountActive` — `COUNT WHERE deadline >= NOW()`

---

## DashboardUseCase

**File:** `internal/usecase/dashboard.go`  
**Dependencies:** `TransactionRepository`, `BudgetRepository`, `GoalRepository`

### Get — Sequential Aggregation (6 DB calls)
```
1. TransactionRepository.GetMonthlyIncome     → income
2. TransactionRepository.GetMonthlyExpenses   → expense
3. TransactionRepository.GetNetSavings        → netSavings (all-time)
4. BudgetRepository.GetUsage                  → []BudgetUsage (with status)
5. GoalRepository.GetAll(active=true)         → []GoalResponse (deadline >= NOW)
6. GoalRepository.CountActive                 → total int
```

Budget summary built in-memory from `budgetUsage` slice (count SAFE/WARNING/EXCEEDED).  
Goal summary: `completed` counted from goals where `status == "COMPLETED"` in `activeGoals` slice.

> **Note:** `GoalProgressSummary.Active = total - completed`. The `total` comes from `CountActive` (deadline >= NOW), while `completed` is counted from `GetAll(active=true)` which also filters by deadline. COMPLETED goals with past deadlines won't appear in either — this is a minor inconsistency.

---

## ChatUseCase

**File:** `internal/usecase/chat.go`  
**Dependencies:** `TransactionRepository`, `BudgetRepository`, `GoalRepository`, `FinancialProfileRepository`, `AiLogRepository`, `GeminiClient`

### Ask — Prompt Building Pipeline
```
1. GetMonthlyIncome, GetMonthlyExpenses, GetNetSavings (3 queries)
2. BudgetRepository.GetUsage → count exceeded budgets
3. GoalRepository.GetAll(active=true) → count active goals
4. FinancialProfileRepository.GetByUserID → optional; silently skipped if not found
5. BuildFinancialProfileContext → formats profile as text block
6. Build prompt:
   "[Profile section]\nTransaction Data (current month: M Y):\n- income\n- expense\n- netSavings\n- budgets: N total, X exceeded\n- active goals: N\n\nUser question: [message]"
7. GeminiClient.Call(ctx, prompt) with up to 3 retries on rate-limit (HTTP 429)
   → exponential backoff: 1s → 2s
8. Save to ai_logs (non-blocking — error is logged, not propagated)
```

### GeminiClient Retry Logic
```
for attempt in 0..2:
  result = GenerateContent(ctx, model, text)
  if isRateLimitErr(err):  // checks for "429", "rate limit", "quota"
    if attempt == 2: return ErrChatUnavailable "rate limit exceeded"
    sleep(backoff); backoff *= 2; continue
  if err: return ErrChatUnavailable
  if result.Text() == "": return ErrChatUnavailable
  return text
```

---

## MLUseCase

**File:** `internal/usecase/ml.go`  
**Dependencies:** `TransactionRepository`, `ml.Client`

### fetchMLTransactions — Shared Fetch Logic
- `GetByUserID(userID, limit=0, offset=0, year, month)` — `limit=0` bypasses LIMIT clause, returns ALL rows
- Converts `domain.TransactionResponse` → `ml.Transaction{Date, Amount, Type, Category}`
- `Date` formatted as `"2006-01-02"` string

### GetAnalysis / GetAnomaly / GetForecast
All three call `fetchMLTransactions` then delegate to `ml.Client`. Any error is wrapped as `ErrMLUnavailable`.

---

## FinancialProfileUseCase

**File:** `internal/usecase/financial_profile.go`  
**Dependencies:** `FinancialProfileRepository`

### Upsert — Validation Rules
| Field | Rule |
|---|---|
| `monthly_income` | `>= 0` |
| `fixed_expenses` | `>= 0` |
| `current_savings` | `>= 0` |
| `debt` | `>= 0` |
| `employment_status` | non-empty (after `strings.TrimSpace`) |
| `financial_goals` | non-empty slice, no blank entries |

After upsert: re-reads profile from DB and computes `NetAvailable = monthly_income - fixed_expenses - debt`.

### BuildFinancialProfileContext
Formats the profile as a text block injected into the Gemini prompt. Omits `spending_habit` and `risk_level` if nil/empty.

---

## Cross-Domain Dependency Map

```
GoalUseCase ──── uses ────→ TransactionRepository  (net savings validation)
ChatUseCase ──── uses ────→ TransactionRepository  (income/expense/net context)
            ──── uses ────→ BudgetRepository        (budget usage for prompt)
            ──── uses ────→ GoalRepository           (active goals for prompt)
            ──── uses ────→ FinancialProfileRepository (optional profile context)
            ──── uses ────→ AiLogRepository          (persist chat log)
DashboardUseCase ─ uses ─→ TransactionRepository
                 ─ uses ─→ BudgetRepository
                 ─ uses ─→ GoalRepository
MLUseCase ─────── uses ─→ TransactionRepository
```

All cross-domain access goes **through the repository interface** — not through other use cases. This avoids circular dependencies while allowing shared data access.

---

## Sentinel Error Contract

| Error | Package | Meaning | Handler response |
|---|---|---|---|
| `domain.ErrNotFound` | domain | Row not found in DB | 404 |
| `domain.ErrConflict` | domain | Unique constraint violated | 409 |
| `domain.ErrUnauthorized` | domain | Not owner or token issue | 401 |
| `domain.ErrInvalidInput` | domain | Business rule violation | 400 |
| `usecase.ErrUserExists` | usecase | Email already registered | 409 |
| `usecase.ErrInvalidCredentials` | usecase | Wrong email or password | 400 |
| `usecase.ErrChatUnavailable` | usecase | Gemini API error | 503 |
| `usecase.ErrMLUnavailable` | usecase | ML service error | 503 |
