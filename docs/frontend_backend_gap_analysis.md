# Frontend ↔ Backend Gap Analysis

> **Generated**: 2026-05-25  
> **Project**: Financial Planning App (React Router v7 + TypeScript)  
> **Backend Base URL**: `http://localhost:8080/api`  
> **Auth Pattern**: JWT Bearer token, in-memory store + `accessToken` cookie  
> **Scan Coverage**: 124 TypeScript/React files across routes, components, hooks, actions, utils

---

## Overall Health: 70/100

| Category | Status |
|---|---|
| Core CRUD (Transactions, Budgets, Goals) | ✅ Fully integrated |
| Dashboard | ⚠️ Partial — graph uses hardcoded data |
| Reports / Analytics | ❌ Almost entirely hardcoded |
| AI Coach | ⚠️ Chat works; Health/Insights/Recommendations are static |
| Auth / Token | ❌ In-memory token lost on page refresh |
| Settings | ⚠️ Financial profile works; basic profile form non-functional |
| Push Notifications | ❌ UI only, no backend |
| Realtime | ❌ Not implemented |

---

## 1. Static Data Usage

### 1.1 `app/lib/components/section/dashboard/DashboardGraph.tsx` — Lines 24–31

**Static data detected:**
```typescript
const chartData = [
  { month: "January", income: 186, expense: 80 },
  { month: "February", income: 305, expense: 200 },
  { month: "March", income: 237, expense: 120 },
  { month: "April", income: 73, expense: 190 },
  { month: "May", income: 209, expense: 130 },
  { month: "June", income: 214, expense: 140 },
];
```

**Why backend-driven:** Dashboard graph is the primary visual KPI for the user. Static values show fictional trends and erode trust.  
**Suggested endpoint:** `GET /auth/v1/transactions/monthly-trend?year={year}`  
**Suggested DB:** Aggregate query on `transactions` table — `GROUP BY MONTH(date), type SUM(amount)`.  
**Notes:** Backend already exposes `/auth/v1/transactions/monthly` and `/auth/v1/transactions/monthly-income`. A dedicated trend endpoint combining both would be ideal, or the frontend can compose them.

---

### 1.2 `app/lib/components/section/dashboard/DashboardBudgetOverview.tsx` — Lines 5–30

**Static data detected:**
```typescript
export const budgetData = [
  { category: "Food & Dining", used: 420, limit: 600, percentage: 70 },
  { category: "Transportation", used: 180, limit: 200, percentage: 90 },
  { category: "Entertainment", used: 95, limit: 150, percentage: 63 },
  { category: "Utilities", used: 120, limit: 200, percentage: 60 },
];
```

**Why backend-driven:** Budget overview on dashboard must reflect the user's real spending. Static data is misleading.  
**Suggested endpoint:** `GET /auth/v1/budgets/usage?month={month}&year={year}` (already exists — just needs to be wired into this component)  
**Suggested DB:** `budgets` + `transactions` JOIN with aggregation.  
**Notes:** The endpoint already exists in `app/actions/budgets.ts` (`GetUsageBudgets`). This component simply needs to receive the data from the dashboard loader instead of using a local constant.

---

### 1.3 `app/lib/components/section/ai-coach/FinancialHealth.tsx` — Lines 4–12

**Static data detected:**
```typescript
const financial_health = {
  score: 78,
  rating: "Good",
  components: {
    savings_rate: 0.14,
    budget_adherence: 0.68,
    goal_progress: 0.62,
  },
};
```

**Why backend-driven:** Financial health score is a personalized metric. Static score `78` is displayed to every user regardless of their actual finances.  
**Suggested endpoint:** `GET /auth/v1/financial-health`  
**Suggested DB:** Derived from existing tables — no new tables needed. Calculated from: `transactions` (savings rate), `budgets`/budget usage (adherence), `goals` (progress).

---

### 1.4 `app/lib/components/section/ai-coach/FinancialKeyInsights.tsx` — Lines 4–20

**Static data detected:**
```typescript
const keyInsights = [
  { title: "On track for goals", description: "4 of 5 goals progressing well", status: "success" },
  { title: "Food spending high", description: "15% above your typical average", status: "warning" },
  { title: "Income increased", description: "$850 freelance income this month", status: "info" },
];
```

**Why backend-driven:** Insights reference specific user numbers ("4 of 5 goals", "$850 freelance income"). Hardcoded values are factually wrong for every user.  
**Suggested endpoint:** `GET /auth/v1/insights` (or reuse `GET /auth/v1/ml/insights`)  
**Suggested DB:** Derived — no new table. ML insights endpoint already exists; this component needs to consume it.

---

### 1.5 `app/lib/components/section/ai-coach/FinancialRecommendation.tsx` — Lines 3–16

**Static data detected:**
```typescript
const recomendations = [
  { title: "Increase savings rate", action: "Boost from 14% to 20% of monthly income" },
  { title: "Review subscriptions", action: "Cancel unused services to save ~$45/month" },
  { title: "Emergency fund ready", action: "Consider moving to high-yield savings" },
];
```

**Why backend-driven:** Recommendations cite the user's actual savings rate and specific dollar amounts. Hardcoded values break user trust.  
**Suggested endpoint:** `GET /auth/v1/recommendations` (or derive from `/auth/v1/ml/insights`)  
**Suggested DB:** None needed — rule-based engine or LLM output.

---

### 1.6 `app/lib/components/section/reports/ReportCategoriesSpending.tsx` — Lines 16–22

**Static data detected:**
```typescript
const chartData = [
  { browser: "chrome", expenses: 275, fill: "var(--color-chrome)" },
  { browser: "safari", expenses: 200, fill: "var(--color-safari)" },
  { browser: "firefox", expenses: 287, fill: "var(--color-firefox)" },
  { browser: "edge", expenses: 173, fill: "var(--color-edge)" },
  { browser: "other", expenses: 190, fill: "var(--color-other)" },
];
```

**Critical bug:** Field names use `browser` (copied from shadcn template) instead of `category`. The chart is completely fake.  
**Suggested endpoint:** `GET /auth/v1/transactions/category-breakdown?month={month}&year={year}`  
**Suggested DB:** Aggregate on `transactions` GROUP BY `category`.

---

### 1.7 `app/lib/components/section/reports/ReportsMonthComparison.tsx` — Lines 7–36

**Static data detected:**
```typescript
export const M_O_M_DATA = [
  { title: "Income", currentValue: 5420, previousValue: 5300, percentageChange: 2.3 },
  { title: "Expenses", currentValue: 3240, previousValue: 3500, percentageChange: -7.4 },
  { title: "Savings", currentValue: 2180, previousValue: 1800, percentageChange: 21.1 },
  { title: "Net Worth", currentValue: 45200, previousValue: 43000, percentageChange: 5.1 },
];
```

**Why backend-driven:** Month-over-month deltas are critical financial KPIs. Exported as a constant — any component importing this gets fake data.  
**Suggested endpoint:** `GET /auth/v1/reports/month-comparison?month={month}&year={year}`  
**Suggested DB:** Two-month aggregate from `transactions`.

---

### 1.8 `app/lib/components/section/reports/ReportsNetworth.tsx` — Lines 23–30

**Static data detected:**
```typescript
const chartData = [
  { month: "January", netWorth: 15200 },
  { month: "February", netWorth: 16650 },
  { month: "March", netWorth: 18100 },
  { month: "April", netWorth: 17500 },
  { month: "May", netWorth: 19200 },
  { month: "June", netWorth: 20800 },
];
```

**Why backend-driven:** Net worth trend is a core financial health indicator. Currently fabricated.  
**Suggested endpoint:** `GET /auth/v1/reports/networth-history?year={year}`  
**Suggested DB:** Requires either a `networth_snapshots` table or calculated from cumulative transaction history.

---

### 1.9 `app/lib/components/section/reports/ReportsSavingRate.tsx` — Lines 23–30

**Static data detected:**
```typescript
const chartData = [
  { month: "Jan", rate: 15 },
  { month: "Feb", rate: 22 },
  { month: "Mar", rate: 18 },
  { month: "Apr", rate: 25 },
  { month: "May", rate: 20 },
  { month: "Jun", rate: 28 },
];
```

**Suggested endpoint:** `GET /auth/v1/reports/savings-rate-history?year={year}`  
**Suggested DB:** Monthly aggregate of `(income - expense) / income` from `transactions`.

---

### 1.10 `app/lib/components/section/reports/ReportTransactions.tsx` — Lines 15–22

**Static data detected:**
```typescript
const chartData = [
  { month: "January", expense: 15200, income: 18600, savings: 3400 },
  { month: "February", expense: 14500, income: 18600, savings: 4100 },
  { month: "March", expense: 16200, income: 19200, savings: 3000 },
  { month: "April", expense: 13800, income: 18600, savings: 4800 },
  { month: "May", expense: 15600, income: 20100, savings: 4500 },
  { month: "June", expense: 14200, income: 19800, savings: 5600 },
];
```

**Suggested endpoint:** `GET /auth/v1/reports/income-expense-trend?year={year}` (or reuse `/auth/v1/transactions/monthly`)  
**Suggested DB:** Aggregate from `transactions`.

---

### 1.11 `app/lib/dummies/transactionDummies.ts` — Dead mock file

**Static data detected:** 6 hardcoded transaction records (June 2024, static amounts).  
**Status:** Not imported anywhere in active code. Dead code.  
**Action:** Delete or move to `__tests__/` directory.

---

## 2. Missing Backend Features

### 2.1 Financial Health Score Calculation

**Frontend behavior:** Shows hardcoded `score: 78` with `savings_rate: 0.14`, `budget_adherence: 0.68`, `goal_progress: 0.62`.  
**Expected backend behavior:** Calculate score from user's real data each request (or cache daily).  

**Required endpoints:**
```
GET /auth/v1/financial-health
```

**Suggested request/response:**
```json
// Request: Authorization: Bearer {token}

// Response 200:
{
  "score": 72,
  "rating": "Good",
  "components": {
    "savings_rate": 0.21,
    "budget_adherence": 0.85,
    "goal_progress": 0.45
  },
  "trend": "improving",
  "last_calculated": "2026-05-25T00:00:00Z"
}
```

**Scoring formula suggestion:**
- `savings_rate_score = min(savings_rate * 200, 40)` (max 40 pts)
- `budget_adherence_score = budget_adherence * 30` (max 30 pts)
- `goal_progress_score = goal_progress * 30` (max 30 pts)
- `total = sum`

---

### 2.2 Financial Key Insights

**Frontend behavior:** Shows 3 hardcoded insight strings with hardcoded dollar amounts.  
**Expected backend behavior:** Generate personalized insights from user transaction and goal data.

**Required endpoints:**
```
GET /auth/v1/insights
```

**Suggested response:**
```json
{
  "insights": [
    {
      "type": "goal_progress",
      "title": "On track for goals",
      "description": "3 of 4 goals progressing on schedule",
      "status": "success"
    },
    {
      "type": "budget_warning",
      "title": "Food spending elevated",
      "description": "12% above your 3-month average",
      "status": "warning"
    }
  ],
  "generated_at": "2026-05-25T10:00:00Z"
}
```

**Note:** The ML insights endpoint (`/auth/v1/ml/insights`) already exists and returns spending patterns. This endpoint can either re-shape that output or call it internally.

---

### 2.3 Financial Recommendations

**Frontend behavior:** Shows 3 generic hardcoded recommendation strings.  
**Expected backend behavior:** Generate rule-based or LLM-powered recommendations.

**Required endpoints:**
```
GET /auth/v1/recommendations
```

**Suggested response:**
```json
{
  "recommendations": [
    {
      "priority": "high",
      "title": "Increase savings rate",
      "action": "Boost from 8% to 20% of monthly income",
      "category": "savings",
      "potential_impact": "Rp 500.000 additional savings/month"
    }
  ],
  "generated_at": "2026-05-25T10:00:00Z"
}
```

---

### 2.4 Net Worth History

**Frontend behavior:** Shows 6-month hardcoded net worth trend.  
**Expected backend behavior:** Track cumulative net savings over time.

**Required endpoints:**
```
GET /auth/v1/reports/networth-history?year={year}
```

**Suggested DB:** Either a `networth_snapshots` table (persisted monthly) or calculate on-the-fly:
```sql
SELECT 
  DATE_TRUNC('month', date) AS month,
  SUM(CASE WHEN type = 'INCOME' THEN amount ELSE -amount END) AS net_monthly,
  SUM(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE -amount END)) 
    OVER (ORDER BY DATE_TRUNC('month', date)) AS cumulative_net_worth
FROM transactions
WHERE user_id = ? AND YEAR(date) = ?
GROUP BY DATE_TRUNC('month', date)
ORDER BY month;
```

---

### 2.5 Savings Rate History

**Required endpoints:**
```
GET /auth/v1/reports/savings-rate-history?year={year}
```

**Suggested response:**
```json
{
  "data": [
    { "month": "2026-01", "income": 5000000, "expense": 3200000, "rate": 36.0 },
    { "month": "2026-02", "income": 5200000, "expense": 3800000, "rate": 26.9 }
  ]
}
```

---

### 2.6 User Profile Update

**File:** `app/lib/components/section/settings/SettingsProfileForm.tsx`  
**Frontend behavior:** Form renders first name, last name, email, phone fields but has **no `onSubmit` handler** — submits to nothing.  
**Expected backend behavior:** `PATCH /auth/v1/users/profile` to update user display name and phone.  

**Required endpoints:**
```
PATCH /auth/v1/users/profile
```

**Suggested request:**
```json
{
  "first_name": "Lewi",
  "last_name": "Borosi",
  "phone": "+6281234567890"
}
```

---

### 2.7 Push Notifications

**File:** `app/lib/components/section/settings/SettingsPushNotifications.tsx`  
**Frontend behavior:** Toggle switches for notification preferences exist in UI.  
**Expected backend behavior:** Store preferences, trigger push/email notifications on budget exceeded, goal achieved, etc.

**Required endpoints:**
```
GET  /auth/v1/notification-settings
POST /auth/v1/notification-settings
```

**Suggested DB:** `notification_settings` table with columns: `user_id`, `budget_alerts`, `goal_reminders`, `weekly_summary`, `push_enabled`.

---

### 2.8 Month-over-Month Comparison

**Required endpoints:**
```
GET /auth/v1/reports/month-comparison?month={month}&year={year}
```

**Suggested response:**
```json
{
  "current": { "month": "2026-05", "income": 5420000, "expense": 3240000, "savings": 2180000 },
  "previous": { "month": "2026-04", "income": 5300000, "expense": 3500000, "savings": 1800000 },
  "changes": {
    "income_pct": 2.3,
    "expense_pct": -7.4,
    "savings_pct": 21.1
  }
}
```

---

### 2.9 Category Spending Breakdown

**Required endpoints:**
```
GET /auth/v1/transactions/category-breakdown?month={month}&year={year}
```

**Suggested response:**
```json
{
  "data": [
    { "category": "FOOD", "total": 1200000, "percentage": 37.0, "transaction_count": 15 },
    { "category": "TRANSPORT", "total": 450000, "percentage": 13.9, "transaction_count": 8 }
  ],
  "total_expense": 3240000
}
```

---

## 3. Incomplete Integrations

### 3.1 Dashboard Budget Overview — Data Not Wired

**What is partially connected:** `GetUsageBudgets` action exists in `app/actions/budgets.ts`. The budget usage API is called in `budgets.tsx` loader.  
**What is missing:** `dashboard.tsx` (or the dashboard clientLoader) does not call `GetUsageBudgets`. `DashboardBudgetOverview.tsx` uses a local hardcoded constant instead of receiving props from the loader.  
**Risk:** User sees wrong budget data on the most important page (dashboard).

---

### 3.2 Dashboard Graph — Endpoint Exists, Not Wired

**What is partially connected:** `/auth/v1/transactions/monthly` is fetched in `budgets.tsx` to find monthly expense. The endpoint exists.  
**What is missing:** `DashboardGraph.tsx` renders a `BarChart` with completely hardcoded `chartData`. It never receives real data.  
**Risk:** Users believe they have growing income trends that don't exist.

---

### 3.3 AI Coach — Chat Works, Other Panels Static

**What is connected:** `Chatbot.tsx` calls `POST /auth/v1/chat` ✅  
**What is missing:** Three sibling panels (`FinancialHealth`, `FinancialKeyInsights`, `FinancialRecommendation`) all render hardcoded constants with no API calls.  
**Risk:** AI Coach page appears functional but shows fabricated data for 75% of its content.

---

### 3.4 Reports Page — ML Endpoints Connected, Chart Data Not

**What is connected:** `MLInsightsPanel` fetches `/auth/v1/ml/analysis`, `/auth/v1/ml/anomaly`, `/auth/v1/ml/insights` ✅  
**What is missing:** All chart components (`ReportCategoriesSpending`, `ReportsMonthComparison`, `ReportsNetworth`, `ReportsSavingRate`, `ReportTransactions`) render hardcoded data despite the ML data being available.  
**Risk:** Disconnect between ML insights (accurate) and visual charts (fake).

---

### 3.5 Token Persistence — Cookie Set, In-Memory Store Volatile

**What is connected:** `accessToken` cookie is set on login. Server-side loaders read from cookie correctly.  
**What is missing:** Client-side code reads from `tokenStore.ts` (in-memory). On page refresh, memory is cleared. The cookie exists but the in-memory store is not re-hydrated from the cookie on app init.  
**Risk:** Client-side features (TanStack Query hooks) lose the token on every refresh, causing 401 errors or blank data.

---

### 3.6 Settings Financial Profile — Partial

**What is connected:** `SettingsFinancialProfile.tsx` fetches and submits `/auth/v1/financial-profile` ✅  
**What is missing:** Basic profile form (`SettingsProfileForm.tsx`) has no submit handler and no backend endpoint. Name and phone number cannot be updated.  
**Risk:** User cannot edit their display name or phone number.

---

## 4. Suggested New APIs

### 4.1 `GET /auth/v1/financial-health`
- **Purpose:** Return calculated financial health score and component breakdown
- **Request:** `Authorization: Bearer {token}`
- **Response:** See §2.1

### 4.2 `GET /auth/v1/insights`
- **Purpose:** Return personalized financial insights
- **Request:** `Authorization: Bearer {token}`, optional `?month=5&year=2026`
- **Response:** See §2.2

### 4.3 `GET /auth/v1/recommendations`
- **Purpose:** Return rule-based or AI recommendations
- **Request:** `Authorization: Bearer {token}`
- **Response:** See §2.3

### 4.4 `GET /auth/v1/reports/networth-history`
- **Purpose:** Return cumulative net worth by month
- **Request:** `Authorization: Bearer {token}`, `?year=2026`
- **Response:** See §2.4

### 4.5 `GET /auth/v1/reports/savings-rate-history`
- **Purpose:** Monthly savings rate trend
- **Request:** `Authorization: Bearer {token}`, `?year=2026`
- **Response:** See §2.5

### 4.6 `GET /auth/v1/reports/month-comparison`
- **Purpose:** Month-over-month income/expense delta
- **Request:** `Authorization: Bearer {token}`, `?month=5&year=2026`
- **Response:** See §2.8

### 4.7 `GET /auth/v1/transactions/category-breakdown`
- **Purpose:** Pie chart data for spending by category
- **Request:** `Authorization: Bearer {token}`, `?month=5&year=2026`
- **Response:** See §2.9

### 4.8 `GET /auth/v1/reports/income-expense-trend`
- **Purpose:** Monthly income/expense/savings bar chart data
- **Request:** `Authorization: Bearer {token}`, `?year=2026`
- **Response:**
```json
{
  "data": [
    { "month": "2026-01", "income": 5000000, "expense": 3200000, "savings": 1800000 }
  ]
}
```

### 4.9 `PATCH /auth/v1/users/profile`
- **Purpose:** Update user display name and phone
- **Request body:**
```json
{ "first_name": "string", "last_name": "string", "phone": "string" }
```
- **Response:** `{ "user": { ... updated user object } }`

### 4.10 `GET /auth/v1/notification-settings` + `POST /auth/v1/notification-settings`
- **Purpose:** Read and persist user notification preferences
- **Request/Response:** See §2.7

---

## 5. Suggested Database Changes

### 5.1 New Table: `notification_settings`
```sql
CREATE TABLE notification_settings (
  id            BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  budget_alerts BOOLEAN NOT NULL DEFAULT TRUE,
  goal_reminders BOOLEAN NOT NULL DEFAULT TRUE,
  weekly_summary BOOLEAN NOT NULL DEFAULT FALSE,
  push_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_user (user_id)
);
```

### 5.2 New Table: `networth_snapshots` (optional, for performance)
```sql
CREATE TABLE networth_snapshots (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  snapshot_date DATE NOT NULL,
  net_worth    DECIMAL(20,2) NOT NULL,
  created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_user_date (user_id, snapshot_date),
  INDEX idx_user_date (user_id, snapshot_date)
);
```

### 5.3 New Columns: `users` table
```sql
ALTER TABLE users 
  ADD COLUMN first_name VARCHAR(100) AFTER email,
  ADD COLUMN last_name  VARCHAR(100) AFTER first_name,
  ADD COLUMN phone      VARCHAR(20)  AFTER last_name;
```
> If `name` already exists as a single field, split it. Check current schema.

### 5.4 New Table: `financial_health_cache` (optional, daily cache)
```sql
CREATE TABLE financial_health_cache (
  id               BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  score            TINYINT NOT NULL,
  rating           ENUM('Poor','Fair','Good','Excellent') NOT NULL,
  savings_rate     DECIMAL(5,4),
  budget_adherence DECIMAL(5,4),
  goal_progress    DECIMAL(5,4),
  calculated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uq_user_date (user_id, DATE(calculated_at))
);
```

### 5.5 Indexes for report queries
```sql
-- Speeds up monthly aggregations on transactions
ALTER TABLE transactions 
  ADD INDEX idx_user_date (user_id, date),
  ADD INDEX idx_user_type_date (user_id, type, date),
  ADD INDEX idx_user_category_date (user_id, category, date);
```

---

## 6. Frontend Logic That Should Move to Backend

### 6.1 Savings Rate Calculation (ReportsOverview.tsx:14–19)
```typescript
// Current frontend code:
const savings_rate = ((monthly_income - monthly_expense) / monthly_income) * 100
```
**Why move:** Acceptable for display only. BUT for historical trend charts and health score calculation, this needs to happen server-side and be stored/cached. Running it purely on frontend means the Reports page only shows the current month's rate.

### 6.2 Budget Percentage / Status Logic (BudgetCategories.tsx, BudgetOverview.tsx)
```typescript
const nearLimit = (item.used / item.limit) * 100 >= 80
const expensePercentage = (totalSpending / totalBudget) * 100
```
**Why move:** Status thresholds (`SAFE` / `WARNING` / `EXCEEDED`) should be authoritative on the backend. If the backend already returns `status` fields, the frontend should use them rather than recalculating. Inconsistency creates bugs if thresholds change.

### 6.3 Goal Progress Percentage (Multiple files: DashboardOverview.tsx, GoalsList.tsx, GoalsMilestone.tsx)
```typescript
Math.round((goal.current_amount / goal.target_amount) * 100)
```
**Why move:** The backend should return a `progress_percentage` field. Three separate components calculate this independently — source of truth divergence risk.

### 6.4 Financial Health Score (FinancialHealth.tsx — currently hardcoded)
**Why must move:** Personalized metric requiring access to all user data. Cannot be meaningfully calculated on the frontend with partial data.

### 6.5 Month-over-Month Delta Calculation
**Why move:** Requires access to two months of transaction history. The frontend currently fakes this. Even if the frontend fetched two months, the calculation should be authoritative server-side.

### 6.6 Net Worth Accumulation
**Why move:** Cumulative sum over months is a running total. Requires all historical transaction data. Frontend pagination means the client never has full history.

### 6.7 Category Spending Aggregation
**Why move:** Pie chart requires `SUM(amount) GROUP BY category` — a DB aggregation. Doing this on the frontend after fetching all transactions would require downloading unbounded data.

---

## 7. Realtime / Async Opportunities

### 7.1 Budget Exceeded Alerts — Push Notification / SSE
When a transaction is recorded and it causes a budget's `used` to exceed the `limit`, the user should be notified immediately.  
**Suggested approach:** Backend triggers a push notification via FCM/APNS or emits an SSE event on `POST /auth/v1/transactions`.  
**Priority:** High — core value proposition of budget tracking.

### 7.2 Goal Milestone Notifications
When `current_amount` reaches 25%, 50%, 75%, 100% of `target_amount`, trigger a notification.  
**Suggested approach:** Check milestones in the `POST /auth/v1/goals/contribute` endpoint. Enqueue a notification job.

### 7.3 Weekly Financial Summary — Cron Job
The notification settings UI has a `weekly_summary` toggle.  
**Suggested approach:** Cron job every Monday 09:00 local time — aggregate prior week's income/expense, email the summary.

### 7.4 ML Forecast — Background Job
The frontend shows a "this may take up to 60 seconds" warning for `/auth/v1/ml/forecast`.  
**Problem:** The frontend has no timeout/abort — the request can hang indefinitely.  
**Suggested approach:**
1. `POST /auth/v1/ml/forecast/start` → returns `job_id`
2. Poll `GET /auth/v1/ml/forecast/{job_id}` every 5s
3. Or use SSE: `GET /auth/v1/ml/forecast/stream`  
**Priority:** Medium — improves UX for slow ML operations.

### 7.5 Chat Response Streaming
Current chat sends full response after completion.  
**Suggested approach:** Use SSE or chunked transfer to stream LLM tokens as they're generated — standard UX pattern for chat interfaces.

---

## 8. Security Concerns

### 8.1 CRITICAL: Token Lost on Page Refresh

**File:** `app/lib/utils/tokenStore.ts`  
```typescript
let _token: string | null = null;
export const getToken = () => _token;
export const setToken = (t: string) => { _token = t; };
```
**Problem:** In-memory only. After refresh, `getToken()` returns `null`. Client-side TanStack Query hooks that call `getToken()` directly will send unauthenticated requests.  
**Fix:** On app init, read the `accessToken` cookie and call `setToken()`. The cookie already exists — just needs hydration. Alternatively, have all client hooks read the cookie directly via `getCookie("accessToken")` (util already exists at `app/lib/utils/cookiesParser.ts`).

### 8.2 MEDIUM: Client-Side JWT Expiry Check Only

**File:** `app/lib/utils/tokenParser.ts`  
The expiry check `exp * 1000 < Date.now()` runs only in server loaders. Client-side hooks send requests without expiry validation. If the token expires mid-session, API calls return 401 but there is no global 401 handler to trigger re-login.  
**Fix:** Add a response interceptor in the TanStack Query hooks (or a shared fetch wrapper) to redirect to `/login` on 401.

### 8.3 LOW: Frontend-Trusted Financial Calculations

Budget status (`SAFE`/`WARNING`/`EXCEEDED`) and goal progress are calculated on the frontend. If the backend doesn't enforce these rules, a malicious API response could show false "SAFE" status for an exceeded budget.  
**Fix:** Backend should return authoritative `status` fields and frontend should use them (some routes already do this — verify consistency).

### 8.4 LOW: No CSRF Protection on Form Actions

Route actions use `request.json()` for mutations. Browser-based CSRF attacks typically target `application/x-www-form-urlencoded` forms. JSON body + Bearer token (not cookie auth for mutations) provides inherent CSRF resistance. Verify the `budget-detail.tsx` FormData action is not CSRF-vulnerable since it uses cookies.

### 8.5 LOW: Sensitive Financial Data in Redux DevTools

The Redux store holds `authUser` and `token`. In development, Redux DevTools exposes these. Ensure the Redux store is not persisted to `localStorage` without encryption in production.

---

## 9. AI/ML Integration Gaps

### 9.1 Implemented and Working ✅
| Feature | Endpoint | Status |
|---|---|---|
| Chat / Q&A | `POST /auth/v1/chat` | ✅ Real Gemini integration |
| Spending analysis | `GET /auth/v1/ml/analysis` | ✅ Returns real metrics |
| Anomaly detection | `GET /auth/v1/ml/anomaly` | ✅ Returns severity-ranked anomalies |
| Spending insights | `GET /auth/v1/ml/insights` | ✅ Pattern detection |
| Forecast | `GET /auth/v1/ml/forecast` | ✅ Returns predictions with confidence |

### 9.2 Simulated / Missing ❌

**FinancialHealth.tsx** — Score `78` is not from any ML model. Static.  
**Fix:** Either calculate rule-based on backend, or call Gemini to evaluate and return a structured score.

**FinancialKeyInsights.tsx** — Three hardcoded insight strings. The ML `/insights` endpoint already exists but this component doesn't call it.  
**Fix:** Wire `GET /auth/v1/ml/insights` into the `ai-coach.tsx` route loader and pass results to this component.

**FinancialRecommendation.tsx** — Three generic hardcoded strings.  
**Fix:** Either create a dedicated recommendations endpoint, or derive from ML insights response by extracting the `recommendations` field if the ML service returns it.

### 9.3 ML Forecast UX Gap

The forecast warning reads "this may take up to 60 seconds" but:
1. No `AbortController` / timeout is implemented.
2. No loading skeleton specific to this slow operation.
3. No retry mechanism.

**Fix:** Implement async job pattern (§7.4) or at minimum add a 60s `AbortController` timeout with a user-friendly error state.

### 9.4 AI Coach Context Gap

The Chat interface does not have access to real-time user financial data context beyond what the Gemini backend injects. The other three panels (Health, Insights, Recommendations) show static data. The user may ask the chatbot a question that contradicts what the static panels show — confusing UX.  
**Fix:** Load all four AI Coach components from a single loader that fetches real data. Chatbot system prompt should be injected with the same data the panels display.

---

## 10. Priority Recommendations

### Critical

| # | Issue | Impact | Effort |
|---|---|---|---|
| C1 | Token in-memory store not hydrated from cookie on refresh | Auth broken after every page refresh for client hooks | 2h |
| C2 | `FinancialHealth`, `FinancialKeyInsights`, `FinancialRecommendation` hardcoded | Entire AI Coach section shows fake data | 3d |
| C3 | All Reports charts hardcoded | Reports page shows completely fabricated analytics | 3d |
| C4 | `DashboardGraph` hardcoded | Primary dashboard chart shows fake trends | 4h |
| C5 | `DashboardBudgetOverview` hardcoded | Budget widget on dashboard shows wrong data | 2h |

### High

| # | Issue | Impact | Effort |
|---|---|---|---|
| H1 | `SettingsProfileForm` has no submit handler | Users cannot update their name/phone | 4h |
| H2 | Missing 401 interceptor in client hooks | Silent auth failures, stale data shown after token expiry | 4h |
| H3 | Month-over-month comparison fully hardcoded | Financial comparison data is fake | 1d |
| H4 | Net worth history hardcoded | Core financial metric fabricated | 1d |
| H5 | Category breakdown chart uses wrong field names + hardcoded | Pie chart broken by design | 4h |

### Medium

| # | Issue | Impact | Effort |
|---|---|---|---|
| M1 | Budget exceeded / goal milestone notifications | Missing engagement feature | 2d |
| M2 | ML Forecast has no timeout/abort | UX hangs on slow ML response | 4h |
| M3 | Push notification settings are UI-only | Feature non-functional | 2d |
| M4 | Goal progress % calculated in 3 separate places | DRY violation, drift risk | 2h |
| M5 | Budget status recalculated on frontend instead of using backend field | Logic duplication | 2h |
| M6 | Delete `app/lib/dummies/transactionDummies.ts` | Dead code / confusion risk | 15m |
| M7 | Chat response not streamed | Slower-feeling chat UX | 1d |

### Low

| # | Issue | Impact | Effort |
|---|---|---|---|
| L1 | Redux store underutilized | Over-engineered state management | 1d |
| L2 | Savings rate calculated frontend-only (display acceptable, history not) | Report accuracy | 4h |
| L3 | No indexes on `transactions` for report queries | Performance at scale | 30m |
| L4 | `networth_snapshots` table missing for historical queries | Performance at scale | 1d |
| L5 | No CSRF analysis for FormData route action | Potential security gap | 2h |

---

# Backend Prompt Suggestions

Use the prompts below verbatim with a backend AI assistant to implement each missing feature.

---

## BP-01: Financial Health Score Endpoint

**Goal:** Implement a `GET /auth/v1/financial-health` endpoint that calculates a personalized financial health score for the authenticated user.

**Required endpoint:**
```
GET /auth/v1/financial-health
Authorization: Bearer {token}
```

**Database changes:**
- No new tables required for MVP.
- Optional: Add `financial_health_cache` table to avoid recalculating on every request (cache by `user_id + DATE(NOW())`).

**Service logic:**
1. Decode JWT, extract `userId`.
2. Fetch current month's transactions for user. Calculate `savings_rate = (total_income - total_expense) / total_income`. If `total_income == 0`, set `savings_rate = 0`.
3. Fetch all budgets for user with usage for current month. Calculate `budget_adherence = count(budgets where used <= limit) / count(all budgets)`. If no budgets, set `budget_adherence = 1.0`.
4. Fetch all active goals. Calculate `goal_progress = avg(current_amount / target_amount)` for goals where `current_amount < target_amount`. If no goals, set `goal_progress = 1.0`.
5. Score formula:
   - `savings_score = min(savings_rate / 0.2 * 40, 40)` (full score at 20% savings rate)
   - `budget_score = budget_adherence * 30`
   - `goal_score = goal_progress * 30`
   - `total = round(savings_score + budget_score + goal_score)`
6. Rating: `[0-40] = "Poor"`, `[41-60] = "Fair"`, `[61-80] = "Good"`, `[81-100] = "Excellent"`.
7. Determine trend by comparing to previous month's cached score (if available). Otherwise `"stable"`.

**Validation rules:**
- Must be authenticated (valid JWT, not expired).
- Return 200 even if user has no transactions/budgets/goals (defaults to neutral values).

**Expected response format:**
```json
{
  "score": 72,
  "rating": "Good",
  "components": {
    "savings_rate": 0.21,
    "budget_adherence": 0.85,
    "goal_progress": 0.45
  },
  "trend": "improving",
  "last_calculated": "2026-05-25T00:00:00Z"
}
```

**Security:** Enforce row-level security — only calculate for the authenticated user's data. Never allow `userId` in query params.

**OpenAPI:** Add `GET /auth/v1/financial-health` to OpenAPI spec with `FinancialHealthResponse` schema.

---

## BP-02: Financial Insights Endpoint

**Goal:** Implement `GET /auth/v1/insights` that returns 3–5 personalized text insights derived from the user's recent financial data.

**Required endpoint:**
```
GET /auth/v1/insights?month={month}&year={year}
Authorization: Bearer {token}
```

**Database changes:** None required. Query existing `transactions`, `budgets`, `goals`.

**Service logic:**
Generate insights by running these checks and including the ones that match:
1. **Goal tracking:** Count goals where `current_amount / target_amount >= 0.5` vs total goals → "X of Y goals on track".
2. **Budget warning:** Find categories where `used / limit > 1.0` → "Your [category] budget was exceeded by X%".
3. **Top spending category:** `SELECT category, SUM(amount) FROM transactions WHERE type='EXPENSE' GROUP BY category ORDER BY SUM(amount) DESC LIMIT 1`.
4. **Income change:** Compare this month's income to last month's. If delta > 10%, → "Income up/down by X%".
5. **Savings improvement:** Compare savings rate this month vs last month.

**Validation:** Same as BP-01.

**Expected response:**
```json
{
  "insights": [
    {
      "type": "goal_progress",
      "title": "3 of 4 goals on track",
      "description": "Your savings goals are progressing well this month.",
      "status": "success"
    },
    {
      "type": "budget_exceeded",
      "title": "Food budget exceeded",
      "description": "You spent 23% over your monthly food budget.",
      "status": "warning"
    }
  ],
  "period": { "month": 5, "year": 2026 },
  "generated_at": "2026-05-25T10:00:00Z"
}
```

**Security:** Same as BP-01.

---

## BP-03: Financial Recommendations Endpoint

**Goal:** Implement `GET /auth/v1/recommendations` that returns 2–4 personalized, actionable financial recommendations.

**Required endpoint:**
```
GET /auth/v1/recommendations
Authorization: Bearer {token}
```

**Database changes:** None. Pure derivation from existing data.

**Service logic (rule-based engine):**
1. If `savings_rate < 0.10` → recommend increasing savings with calculated target.
2. If any budget has `used / limit > 0.9` → recommend reviewing that category.
3. If any goal has `deadline` within 60 days and `progress < 0.8` → recommend increasing contributions.
4. If `savings_rate > 0.25` and no investment goal exists → recommend creating an investment goal.
5. Rank by impact. Return top 3.

**Expected response:**
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

**Security:** Same as BP-01.

---

## BP-04: Reports — Income/Expense Trend

**Goal:** Implement `GET /auth/v1/reports/income-expense-trend` returning monthly income, expense, and savings for all months of a given year.

**Required endpoint:**
```
GET /auth/v1/reports/income-expense-trend?year={year}
Authorization: Bearer {token}
```

**Database changes:** Add index `(user_id, type, date)` to `transactions` table.

**Service logic:**
```sql
SELECT 
  MONTH(date) AS month,
  SUM(CASE WHEN type='INCOME' THEN amount ELSE 0 END) AS income,
  SUM(CASE WHEN type='EXPENSE' THEN amount ELSE 0 END) AS expense,
  SUM(CASE WHEN type='INCOME' THEN amount ELSE -amount END) AS savings
FROM transactions
WHERE user_id = ? AND YEAR(date) = ?
GROUP BY MONTH(date)
ORDER BY MONTH(date);
```
Fill missing months with `0` values. Return 12 rows always.

**Expected response:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "income": 5000000, "expense": 3200000, "savings": 1800000 },
    { "month": 2, "month_name": "February", "income": 0, "expense": 0, "savings": 0 }
  ]
}
```

**Validation:** `year` must be a valid 4-digit year. Default to current year if omitted.

---

## BP-05: Reports — Category Breakdown

**Goal:** Implement `GET /auth/v1/transactions/category-breakdown` for pie chart data.

**Required endpoint:**
```
GET /auth/v1/transactions/category-breakdown?month={month}&year={year}
Authorization: Bearer {token}
```

**Service logic:**
```sql
SELECT 
  category,
  SUM(amount) AS total,
  COUNT(*) AS transaction_count,
  ROUND(SUM(amount) / total_expense * 100, 1) AS percentage
FROM transactions
WHERE user_id = ? AND type = 'EXPENSE' 
  AND MONTH(date) = ? AND YEAR(date) = ?
GROUP BY category
ORDER BY total DESC;
```

**Expected response:**
```json
{
  "period": { "month": 5, "year": 2026 },
  "total_expense": 3240000,
  "data": [
    { "category": "FOOD", "label": "Food & Dining", "total": 1200000, "percentage": 37.0, "transaction_count": 15 }
  ]
}
```

---

## BP-06: Reports — Net Worth History

**Goal:** Implement `GET /auth/v1/reports/networth-history` returning cumulative net worth per month.

**Required endpoint:**
```
GET /auth/v1/reports/networth-history?year={year}
Authorization: Bearer {token}
```

**Database changes:** 
- Option A (real-time): Calculate from `transactions` using window function.
- Option B (performant): Add `networth_snapshots` table, updated by a monthly cron job or on each transaction.

**Service logic (Option A):**
```sql
SELECT 
  MONTH(date) AS month,
  SUM(SUM(CASE WHEN type='INCOME' THEN amount ELSE -amount END)) 
    OVER (ORDER BY MONTH(date) ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS cumulative_net_worth
FROM transactions
WHERE user_id = ? AND YEAR(date) = ?
GROUP BY MONTH(date)
ORDER BY MONTH(date);
```

**Expected response:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "net_worth": 1800000 },
    { "month": 2, "month_name": "February", "net_worth": 4100000 }
  ]
}
```

---

## BP-07: Reports — Month-over-Month Comparison

**Goal:** Implement `GET /auth/v1/reports/month-comparison` returning delta between current and previous month.

**Required endpoint:**
```
GET /auth/v1/reports/month-comparison?month={month}&year={year}
Authorization: Bearer {token}
```

**Service logic:** Fetch current month aggregate and previous month aggregate. Calculate `((current - previous) / previous) * 100` for each metric.

**Expected response:**
```json
{
  "current": { "month": 5, "year": 2026, "income": 5420000, "expense": 3240000, "savings": 2180000 },
  "previous": { "month": 4, "year": 2026, "income": 5300000, "expense": 3500000, "savings": 1800000 },
  "changes": {
    "income_pct": 2.3,
    "expense_pct": -7.4,
    "savings_pct": 21.1,
    "net_worth_pct": 5.1
  }
}
```

---

## BP-08: Reports — Savings Rate History

**Goal:** Implement `GET /auth/v1/reports/savings-rate-history`.

**Required endpoint:**
```
GET /auth/v1/reports/savings-rate-history?year={year}
Authorization: Bearer {token}
```

**Service logic:** For each month: `rate = (income - expense) / income * 100`. Return 0 if no income.

**Expected response:**
```json
{
  "year": 2026,
  "data": [
    { "month": 1, "month_name": "January", "income": 5000000, "expense": 3200000, "rate": 36.0 },
    { "month": 2, "month_name": "February", "income": 0, "expense": 0, "rate": 0 }
  ]
}
```

---

## BP-09: User Profile Update

**Goal:** Implement `PATCH /auth/v1/users/profile` to allow users to update their basic profile.

**Required endpoint:**
```
PATCH /auth/v1/users/profile
Authorization: Bearer {token}
Content-Type: application/json
```

**Database changes:**
```sql
ALTER TABLE users
  ADD COLUMN first_name VARCHAR(100),
  ADD COLUMN last_name  VARCHAR(100),
  ADD COLUMN phone      VARCHAR(20);
```

**Validation rules:**
- `first_name`: optional, max 100 chars, alphanumeric + spaces
- `last_name`: optional, max 100 chars
- `phone`: optional, E.164 format or local Indonesian format validation
- `email`: NOT updatable via this endpoint (separate flow with verification required)

**Service logic:**
1. Decode JWT → `userId`.
2. Validate request body.
3. `UPDATE users SET first_name=?, last_name=?, phone=? WHERE id=?`.
4. Return updated user object.

**Expected response:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "Lewi",
    "last_name": "Borosi",
    "phone": "+6281234567890"
  }
}
```

**Security:** Only allow updating the authenticated user's own profile. Never accept `userId` from request body.

---

## BP-10: Notification Settings CRUD

**Goal:** Allow users to save and retrieve notification preferences.

**Required endpoints:**
```
GET  /auth/v1/notification-settings
POST /auth/v1/notification-settings
Authorization: Bearer {token}
```

**Database changes:** See §5.1 (`notification_settings` table).

**Service logic (GET):** `SELECT * FROM notification_settings WHERE user_id = ?`. If no row, return defaults.  
**Service logic (POST):** `INSERT ... ON DUPLICATE KEY UPDATE ...` (upsert).

**Validation:** All fields boolean. No required fields — partial update supported.

**Expected response (GET):**
```json
{
  "budget_alerts": true,
  "goal_reminders": true,
  "weekly_summary": false,
  "push_enabled": false
}
```

---

## BP-11: Token Re-hydration on App Init (Frontend Fix)

**Goal:** Fix the in-memory token store so client-side hooks remain authenticated after page refresh.

**No backend change required.** Frontend fix in `app/root.tsx` or the auth layout `app/routes/auth/layout.tsx`:

```typescript
// In clientLoader or useEffect in root:
import { getCookie } from "~/lib/utils/cookiesParser";
import { setToken } from "~/lib/utils/tokenStore";

// On client init:
const cookieToken = getCookie("accessToken");
if (cookieToken) {
  setToken(cookieToken);
}
```

Or alternatively, update all client-side hooks (`use-transaction.ts`, `use-budget.ts`) to call `getCookie("accessToken")` directly rather than `getToken()` from tokenStore.

**Security note:** Ensure the `accessToken` cookie is `HttpOnly: false` (must be readable by JS for this approach) OR use the session cookie approach already implemented in server loaders exclusively for server code, and use a non-HttpOnly cookie for client code.

---

## BP-12: ML Forecast — Async Job Pattern

**Goal:** Prevent the forecast request from hanging indefinitely.

**Required endpoints:**
```
POST /auth/v1/ml/forecast/start           → returns { job_id: "abc123" }
GET  /auth/v1/ml/forecast/status/{job_id} → returns { status: "pending"|"complete"|"failed", result?: {...} }
```

**Service logic:**
1. `POST /start`: Submit forecast job to a queue (Redis/BullMQ/etc). Return `job_id` immediately (202 Accepted).
2. Background worker runs ML forecast, stores result in `forecast_jobs` table.
3. Frontend polls `GET /status/{job_id}` every 5 seconds with a 90-second total timeout.
4. Frontend shows progress indicator during polling.

**Database changes:**
```sql
CREATE TABLE forecast_jobs (
  id         VARCHAR(36) PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id),
  status     ENUM('pending','running','complete','failed') DEFAULT 'pending',
  result     JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

---

*End of report. Total gaps identified: 28 items across 11 categories. Critical path: C1 (token hydration) → C5 (dashboard data) → C4 (graph) → C2–C3 (AI Coach + Reports).*
