# Frontend Integration Notes — API v1.4

> Backend version: **1.4.0**
> Date: 2026-05-25
> Context: These changes resolve every hardcoded-data gap identified in `frontend_backend_gap_analysis.md`.
> All new endpoints are under `/api/auth/v1/` and require JWT auth.

---

## Quick Summary

| Area | Old behaviour | New endpoint |
|------|---------------|-------------|
| Dashboard graph | Hardcoded 6-month data | `GET /reports/income-expense-trend?year=` |
| Dashboard budget overview | Hardcoded constants | Already existed: `GET /budgets/usage` |
| AI Coach – health score | Static `78` | `GET /financial-health` |
| AI Coach – key insights | Static strings | `GET /insights?month=&year=` |
| AI Coach – recommendations | Static strings | `GET /recommendations` |
| Reports – category pie chart | Wrong field names + hardcoded | `GET /transactions/category-breakdown?month=&year=` |
| Reports – MoM comparison | Hardcoded | `GET /reports/month-comparison-v2?month=&year=` |
| Reports – net worth trend | Hardcoded | `GET /reports/networth-history?year=` |
| Reports – savings rate chart | Hardcoded | `GET /reports/savings-rate-history?year=` |
| Reports – income/expense bar | Hardcoded | `GET /reports/income-expense-trend?year=` |
| Settings – profile form | No submit handler / no endpoint | `PATCH /users/profile` |
| Notification settings | UI only | `GET/POST /notification-settings` |
| Token refresh | In-memory lost on reload | **Frontend-only fix** (see §8) |

---

## 1. Dashboard Graph (`DashboardGraph.tsx`)

**Replace** the hardcoded `chartData` array.

```typescript
// Remove this:
const chartData = [
  { month: "January", income: 186, expense: 80 },
  ...
]

// Fetch from:
GET /api/auth/v1/reports/income-expense-trend?year=2026
```

**Response shape:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "income": 5000000, "expense": 3200000, "savings": 1800000 },
    ...
  ]
}
```

Map `data[i].month_name` → x-axis, `income` / `expense` → bar values. All 12 months always present (zero-filled for months with no transactions).

---

## 2. Dashboard Budget Overview (`DashboardBudgetOverview.tsx`)

**Remove** the local `budgetData` constant. The `GetUsageBudgets` action already exists in `app/actions/budgets.ts`. Wire it into the dashboard route loader.

```typescript
// In dashboard loader, add:
const budgetUsage = await GetUsageBudgets(token, currentYear, currentMonth);
// Pass as prop to DashboardBudgetOverview
```

Response is an **array** (not wrapped in `data`), matching `BudgetUsageResponse` schema.

---

## 3. AI Coach – Financial Health (`FinancialHealth.tsx`)

**Replace** the static `financial_health` object.

```typescript
GET /api/auth/v1/financial-health
```

**Response shape:**
```json
{
  "score": 72,
  "rating": "Good",
  "components": {
    "savings_rate": 0.21,
    "budget_adherence": 0.85,
    "goal_progress": 0.45
  },
  "trend": "stable",
  "last_calculated": "2026-05-25T00:00:00Z"
}
```

Map `score` to the gauge, `rating` to the label chip, `components.*` to the three sub-bars.

---

## 4. AI Coach – Key Insights (`FinancialKeyInsights.tsx`)

**Replace** the hardcoded `keyInsights` array.

```typescript
GET /api/auth/v1/insights?month=5&year=2026
```

**Response shape:**
```json
{
  "insights": [
    { "type": "goal_progress", "title": "3 of 4 goals on track", "description": "...", "status": "success" },
    { "type": "budget_exceeded", "title": "FOOD budget exceeded", "description": "...", "status": "warning" }
  ],
  "period": { "month": 5, "year": 2026 },
  "generated_at": "2026-05-25T10:00:00Z"
}
```

`status` maps to your existing `"success" | "warning" | "info"` colour tokens.

> **Also available:** `GET /api/auth/v1/ml/insights` returns ML-derived patterns (`top_category`, `category_breakdown`, `spike_category`). Consider displaying both in the same panel or merging them.

---

## 5. AI Coach – Recommendations (`FinancialRecommendation.tsx`)

**Replace** the hardcoded `recomendations` array (fix the typo too).

```typescript
GET /api/auth/v1/recommendations
```

**Response shape:**
```json
{
  "recommendations": [
    {
      "priority": "high",
      "category": "savings",
      "title": "Increase savings rate",
      "action": "Your current savings rate is 8%. Increasing to 20% would save an additional Rp 600.000/month.",
      "potential_impact": "Rp 600.000/month"
    }
  ],
  "generated_at": "2026-05-25T10:00:00Z"
}
```

`priority` → badge colour (`high` = red, `medium` = yellow, `low` = grey).

---

## 6. Reports – Category Spending (`ReportCategoriesSpending.tsx`)

**Critical:** Current code uses `browser` field from a shadcn template — wrong field names AND hardcoded.

```typescript
// Remove entirely:
const chartData = [
  { browser: "chrome", expenses: 275, ... },
]

// Fetch from:
GET /api/auth/v1/transactions/category-breakdown?month=5&year=2026
```

**Response shape:**
```json
{
  "period": { "month": 5, "year": 2026 },
  "total_expense": 3240000,
  "data": [
    { "category": "FOOD", "label": "Food & Dining", "total": 1200000, "percentage": 37.0, "transaction_count": 15 }
  ]
}
```

Use `label` for display names and `total` / `percentage` for chart values. The old `browser` → `category`, `expenses` → `total`.

---

## 7. Reports – Month-over-Month (`ReportsMonthComparison.tsx`)

**Replace** the exported `M_O_M_DATA` constant.

```typescript
GET /api/auth/v1/reports/month-comparison-v2?month=5&year=2026
```

**Response shape:**
```json
{
  "current":  { "month": 5, "year": 2026, "income": 5420000, "expense": 3240000, "savings": 2180000 },
  "previous": { "month": 4, "year": 2026, "income": 5300000, "expense": 3500000, "savings": 1800000 },
  "changes":  { "income_pct": 2.3, "expense_pct": -7.4, "savings_pct": 21.1 }
}
```

`changes.*_pct` is positive = improvement for income/savings, positive = worse for expense. Net worth is derived: `current.savings - previous.savings` if needed.

---

## 8. Reports – Net Worth Trend (`ReportsNetworth.tsx`)

```typescript
GET /api/auth/v1/reports/networth-history?year=2026
```

**Response shape:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "net_worth": 1800000 },
    { "month": 2, "month_name": "February", "net_worth": 4100000 }
  ]
}
```

> **Note:** `net_worth` is cumulative within the requested year only (starts from 0 at January). It does NOT include prior-year balances. Label this chart "Savings Growth" or "Net Savings" if you want to set accurate expectations.

---

## 9. Reports – Savings Rate (`ReportsSavingRate.tsx`)

```typescript
GET /api/auth/v1/reports/savings-rate-history?year=2026
```

**Response shape:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "income": 5000000, "expense": 3200000, "rate": 36.0 }
  ]
}
```

Map `month_name` → x-axis, `rate` → line value.

---

## 10. Reports – Income/Expense Bar (`ReportTransactions.tsx`)

Same endpoint as Dashboard Graph:

```typescript
GET /api/auth/v1/reports/income-expense-trend?year=2026
```

Fields: `income`, `expense`, `savings` per month.

---

## 11. Settings – Profile Form (`SettingsProfileForm.tsx`)

**Add a submit handler.** New endpoint:

```typescript
PATCH /api/auth/v1/users/profile
Content-Type: application/json

{ "first_name": "Lewi", "last_name": "Borosi", "phone": "+6281234567890" }
```

**Response:** `{ "user": { "id": 1, "email": "...", "name": "...", "first_name": "Lewi", "last_name": "Borosi", "phone": "..." } }`

All fields optional — send only changed ones. Email cannot be changed via this endpoint.

`GET /api/auth/v1/users/me` now also returns `first_name`, `last_name`, `phone` — use it to pre-fill the form.

---

## 12. Notification Settings (`SettingsPushNotifications.tsx`)

New alias endpoints (same functionality as `/notifications/preferences`):

```typescript
GET  /api/auth/v1/notification-settings
POST /api/auth/v1/notification-settings
```

**Response / request body:**
```json
{
  "budget_alerts": true,
  "goal_reminders": true,
  "anomaly_alerts": true,
  "weekly_summary": false,
  "push_enabled": false
}
```

Two new fields: `weekly_summary` and `push_enabled`. Wire the toggle switches to these.

---

## 13. Token Re-hydration on Page Refresh (FRONTEND ONLY — no backend change needed)

**Problem:** `tokenStore.ts` is in-memory. After reload `getToken()` returns `null` → 401 errors.

**Fix:** Hydrate on app init from the existing `accessToken` cookie.

```typescript
// In app/root.tsx clientLoader, or auth layout clientLoader:
import { getCookie } from "~/lib/utils/cookiesParser";
import { setToken } from "~/lib/utils/tokenStore";

export async function clientLoader() {
  const cookieToken = getCookie("accessToken");
  if (cookieToken) setToken(cookieToken);
  // ...rest of loader
}
```

Or update all client-side hooks to call `getCookie("accessToken")` directly instead of `getToken()`.

> **Security note:** The `accessToken` cookie must NOT be `HttpOnly` for this approach. Verify the cookie flag set in the Go backend login handler.

---

## 14. `app/lib/dummies/transactionDummies.ts`

This file contains 6 hardcoded transaction records and is **not imported anywhere**. Delete it.

```bash
rm app/lib/dummies/transactionDummies.ts
```

---

## Endpoint Reference (v1.4 additions only)

| Method | Path | Purpose | Response schema |
|--------|------|---------|----------------|
| `PATCH` | `/api/auth/v1/users/profile` | Update first_name/last_name/phone | `UserProfileResponse` |
| `GET` | `/api/auth/v1/financial-health` | Health score + components | `FinancialHealthResponse` |
| `GET` | `/api/auth/v1/insights` | Rule-based insights | `InsightsResponse` |
| `GET` | `/api/auth/v1/recommendations` | Rule-based recommendations | `RecommendationsResponse` |
| `GET` | `/api/auth/v1/reports/income-expense-trend` | 12-mo income/expense/savings | `IncomeExpenseTrendResponse` |
| `GET` | `/api/auth/v1/reports/networth-history` | 12-mo cumulative net worth | `NetWorthHistoryResponse` |
| `GET` | `/api/auth/v1/reports/savings-rate-history` | 12-mo savings rate % | `SavingsRateHistoryResponse` |
| `GET` | `/api/auth/v1/reports/month-comparison-v2` | MoM delta with savings | `MonthComparisonV2Response` |
| `GET` | `/api/auth/v1/transactions/category-breakdown` | Category pie chart data | `CategoryBreakdownResponse` |
| `GET` | `/api/auth/v1/notification-settings` | Read notification prefs | `NotificationPreferencesBody` |
| `POST` | `/api/auth/v1/notification-settings` | Save notification prefs | `MessageResponse` |

Full schemas in `openapi.yaml` (version 1.4.0).

---

## Breaking Changes in v1.4

None. All existing endpoints are unchanged. New endpoints are additive only.

The following DB columns were added (nullable, backwards-compatible):
- `users.first_name`, `users.last_name`, `users.phone` — run migration `000019`
- `notification_preferences.weekly_summary`, `notification_preferences.push_enabled` — run migration `000020`

Run migrations before deploying:
```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```
