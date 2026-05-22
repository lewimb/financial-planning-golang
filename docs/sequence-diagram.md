# Sequence Diagrams

All sequence diagrams are derived from the actual Go source code in `internal/`.  
Last audited against codebase: 2026-05-19.

---

## 1. User Registration

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as UserHandler
    participant UC as UserUseCase
    participant R as UserRepository
    participant DB as PostgreSQL

    C->>H: POST /api/v1/register { email, password, name }
    H->>UC: Register(email, password, name)
    UC->>UC: bcrypt.GenerateFromPassword(password, cost=10)
    UC->>R: Create(email, hashedPassword, name)
    R->>DB: INSERT INTO users (email, name, password)<br/>ON CONFLICT email → pgErr.Code 23505
    DB-->>R: result
    alt email already exists
        R-->>UC: domain.ErrConflict
        UC-->>H: ErrUserExists
        H-->>C: 409 { "error": "User already exists" }
    else success
        R-->>UC: nil
        UC-->>H: nil
        H-->>C: 200 { "message": "User created successfully" }
    end
```

---

## 2. User Login and JWT Issuance

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as UserHandler
    participant UC as UserUseCase
    participant R as UserRepository
    participant DB as PostgreSQL

    C->>H: POST /api/v1/login { email, password }
    H->>UC: Login(email, password)
    UC->>R: FindByEmail(email)
    R->>DB: SELECT id,email,name,password FROM users WHERE email=$1
    DB-->>R: user row
    alt user not found
        R-->>UC: sql.ErrNoRows
        UC-->>H: ErrInvalidCredentials
        H-->>C: 400 { "error": "invalid email or password" }
    else user found
        R-->>UC: *domain.User
        UC->>UC: bcrypt.CompareHashAndPassword(hash, password)
        alt password mismatch
            UC-->>H: ErrInvalidCredentials
            H-->>C: 400 { "error": "invalid email or password" }
        else password matches
            UC->>UC: utils.GenerateJwt(userID, name, email)<br/>HS256 claims:{id,name,email,iss:lewimb} signed with SECRET_KEY<br/>No exp claim — cookie TTL enforces session length
            UC-->>H: tokenString
            H->>H: c.SetCookie("accessToken", token, 3600, "/", "*", false, false)
            H-->>C: 200 { token: "eyJ..." }
        end
    end
```

---

## 3. JWT Middleware (all protected routes)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as AuthMiddleware
    participant H as Handler

    C->>MW: Any /api/auth/v1/* — Authorization: Bearer <token>
    MW->>MW: Extract Authorization header
    alt header missing
        MW-->>C: 400 { "message": "Missing Authorization!" }
    else header present
        MW->>MW: Split on space → ["Bearer", token]
        alt wrong format
            MW-->>C: 401 format must be Bearer token
        else correct format
            MW->>MW: jwt.ParseWithClaims(token, &claims, SECRET_KEY)<br/>Verify HS256 signature
            alt token invalid
                MW-->>C: 401 { "error": "Invalid token" }
            else token valid
                MW->>MW: c.Set("claims", claims) → {id, name, email}
                MW->>H: c.Next() — handler proceeds
                Note over H: utils.ClaimId(c) extracts userID from context
            end
        end
    end
```

---

## 4. Create Transaction

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as AuthMiddleware
    participant H as TransactionHandler
    participant UC as TransactionUseCase
    participant R as TransactionRepository
    participant DB as PostgreSQL

    C->>MW: POST /api/auth/v1/transactions { amount, category, type, date, description }
    MW->>H: JWT valid → userID
    H->>H: c.ShouldBindJSON(&req)
    alt bind fails
        H-->>C: 400
    else bind succeeds
        H->>UC: Create(userID, req)
        UC->>UC: strings.ToUpper(type)
        UC->>UC: validate: type∈{INCOME,EXPENSE}, amount>0,<br/>category non-empty, date non-zero
        alt validation fails
            UC-->>H: error
            H-->>C: 400 { "error": "..." }
        else valid
            UC->>R: Create(userID, req)
            R->>DB: INSERT INTO transactions (amount,category,type,date,description,user_id)
            DB-->>R: nil
            R-->>UC: nil
            UC-->>H: nil
            H-->>C: 200 { "message": "Transaction created successfully" }
        end
    end
```

---

## 5. Goal Contribution

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as GoalHandler
    participant UC as GoalUseCase
    participant TR as TransactionRepository
    participant GR as GoalRepository
    participant DB as PostgreSQL

    C->>H: PATCH /api/auth/v1/goals/contribute { goal_id, contribution }
    H->>H: validate goal_id > 0
    H->>UC: Contribute(goalID, userID, contribution)
    UC->>UC: validate contribution > 0
    UC->>TR: GetNetSavings(userID)
    TR->>DB: SELECT SUM(INCOME)-SUM(EXPENSE) all-time, no date filter
    DB-->>TR: netSavings
    alt netSavings <= 0
        UC-->>H: "cannot add contributions: no net savings"
        H-->>C: 400
    else contribution > netSavings
        UC-->>H: "contribution exceeds available savings"
        H-->>C: 400
    else valid
        UC->>GR: Contribute(goalID, userID, contribution)
        GR->>DB: UPDATE goals SET current_amount=$1,<br/>status=CASE WHEN $1>=target THEN COMPLETED ELSE status END,<br/>updated_at=NOW WHERE id=$2 AND user_id=$3
        DB-->>GR: nil
        GR-->>UC: nil
        UC-->>H: nil
        H-->>C: 200 { "message": "Contribution successful" }
    end
```

---

## 6. Dashboard Data Retrieval

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as DashboardHandler
    participant UC as DashboardUseCase
    participant TR as TransactionRepository
    participant BR as BudgetRepository
    participant GR as GoalRepository
    participant DB as PostgreSQL

    C->>H: GET /api/auth/v1/dashboard
    H->>UC: Get(userID)
    Note over UC: 6 sequential DB calls — time.Now() determines period

    UC->>TR: GetMonthlyIncome(userID)
    TR->>DB: SUM(INCOME) current month/year
    DB-->>TR: income

    UC->>TR: GetMonthlyExpenses(userID)
    TR->>DB: SUM(EXPENSE) current month/year
    DB-->>TR: expense

    UC->>TR: GetNetSavings(userID)
    TR->>DB: SUM(INCOME)-SUM(EXPENSE) all-time
    DB-->>TR: netSavings

    UC->>BR: GetUsage(userID, currentMonth, currentYear)
    BR->>DB: JOIN budgets ← transactions, compute SAFE/WARNING/EXCEEDED
    DB-->>BR: []BudgetUsage

    UC->>GR: GetAll(userID, active=true)
    GR->>DB: SELECT * FROM goals WHERE deadline >= NOW()
    DB-->>GR: []GoalResponse

    UC->>GR: CountActive(userID)
    GR->>DB: SELECT COUNT(*) WHERE deadline >= NOW()
    DB-->>GR: total

    UC->>UC: Aggregate BudgetStatusSummary (SAFE/WARNING/EXCEEDED counts)
    UC->>UC: Aggregate GoalProgressSummary (total, active, completed)
    UC-->>H: *DashboardResponse
    H-->>C: 200 { monthly_income, monthly_expense, net_savings,<br/>budget_summary, goal_summary, active_goals }
```

---

## 7. AI Chat (Gemini Integration)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as ChatHandler
    participant UC as ChatUseCase
    participant TR as TransactionRepository
    participant BR as BudgetRepository
    participant GR as GoalRepository
    participant PR as FinancialProfileRepository
    participant LR as AiLogRepository
    participant DB as PostgreSQL
    participant GEM as Gemini API (External)

    C->>H: POST /api/auth/v1/chat { "message": "..." }
    H->>H: ShouldBindJSON — message required
    H->>UC: Ask(ctx, userID, message)

    UC->>TR: GetMonthlyIncome(userID)
    UC->>TR: GetMonthlyExpenses(userID)
    UC->>TR: GetNetSavings(userID)
    TR->>DB: 3 separate queries
    DB-->>TR: income, expense, netSavings

    UC->>BR: GetUsage(userID, month, year)
    BR->>DB: budget usage JOIN
    DB-->>BR: []BudgetUsage

    UC->>GR: GetAll(userID, active=true)
    GR->>DB: active goals
    DB-->>GR: []GoalResponse

    UC->>PR: GetByUserID(userID)
    PR->>DB: SELECT profile + financial goals
    DB-->>PR: *FinancialProfileResponse or ErrNotFound
    alt profile exists
        UC->>UC: BuildFinancialProfileContext(profile) → text block
    else no profile
        UC->>UC: profileSection = "" (silently skipped)
    end

    UC->>UC: Build Gemini prompt with profile context + transaction data + user message
    UC->>GEM: models.GenerateContent(ctx, "gemini-3-flash-preview", prompt)
    Note over UC,GEM: Up to 3 retries with exponential backoff on HTTP 429

    alt Gemini error (non-retryable)
        GEM-->>UC: error
        UC-->>H: ErrChatUnavailable
        H-->>C: 503 AI service unavailable
    else Gemini success
        GEM-->>UC: GenerateContentResponse
        UC->>UC: result.Text()
        UC->>LR: Save(userID, message, reply)
        LR->>DB: INSERT INTO ai_logs — error logged, NOT propagated to caller
        UC-->>H: reply string
        H-->>C: 200 { "reply": "..." }
    end
```

---

## 8. Financial Profile Upsert (Atomic DB Transaction)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as FinancialProfileHandler
    participant UC as FinancialProfileUseCase
    participant R as FinancialProfileRepository
    participant DB as PostgreSQL

    C->>H: POST /api/auth/v1/financial-profile { monthly_income, ..., financial_goals[] }
    H->>H: ShouldBindJSON(&req)
    H->>UC: Upsert(userID, req)
    UC->>UC: validateProfileRequest(req)<br/>all numbers >= 0, employment_status non-empty,<br/>financial_goals non-empty/no blanks
    alt validation fails
        UC-->>H: error
        H-->>C: 400
    else validation passes
        UC->>R: Upsert(userID, req)
        R->>DB: BEGIN
        R->>DB: INSERT INTO user_financial_profiles ON CONFLICT DO UPDATE
        R->>DB: DELETE FROM user_financial_goals WHERE user_id=$1
        loop for each goal type
            R->>DB: INSERT INTO user_financial_goals ON CONFLICT DO NOTHING
        end
        R->>DB: COMMIT
        DB-->>R: committed
        R-->>UC: nil
        UC->>R: GetByUserID(userID)
        R->>DB: SELECT profile fields
        R->>DB: SELECT goal_type ORDER BY id
        DB-->>R: profile + goals
        UC->>UC: p.NetAvailable = income - fixed_expenses - debt
        UC-->>H: *FinancialProfileResponse
        H-->>C: 200 { "data": { ...profile, net_available } }
    end
```

---

## 9. ML Forecast (Full Pipeline)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as MLHandler
    participant UC as MLUseCase
    participant TR as TransactionRepository
    participant DB as PostgreSQL
    participant ML as ML Service (FastAPI :8000)

    C->>H: GET /api/auth/v1/ml/forecast?periods=30
    H->>H: parse periods (default 30)
    H->>UC: GetForecast(userID, periods, year, month)
    UC->>UC: fetchMLTransactions(userID, year, month)
    UC->>TR: GetByUserID(userID, limit=0, offset=0, year, month)
    TR->>DB: SELECT all non-deleted transactions (no LIMIT)
    DB-->>TR: []TransactionResponse
    TR-->>UC: transactions
    UC->>UC: convert to []ml.Transaction {date, amount, type, category}
    UC->>ML: POST /forecast?periods=30 — timeout 60s
    alt ML unreachable or non-200
        ML-->>UC: error
        UC-->>H: ErrMLUnavailable
        H-->>C: 503
    else ML success
        ML-->>UC: { predicted_monthly_spending, daily_forecast[] }
        UC-->>H: *ForecastResponse
        H-->>C: 200 { predicted_monthly_spending, daily_forecast }
    end
```

---

## 10. Budget Usage Retrieval

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as BudgetHandler
    participant UC as BudgetUseCase
    participant R as BudgetRepository
    participant DB as PostgreSQL

    C->>H: GET /api/auth/v1/budgets/usage?year=2026&month=5
    H->>H: parse year (required), month (optional)
    H->>UC: GetUsage(userID, month=5, year=2026)
    UC->>R: GetUsage(userID, 5, 2026)
    R->>R: compute prevMonth=4, prevYear=2026
    R->>DB: complex JOIN budgets ← transactions<br/>GROUP BY budget, compute used + prev_used
    DB-->>R: raw rows
    R->>R: for each row:<br/>remaining = limit - used (floor 0)<br/>percentage = ROUND(used/limit*100)<br/>status = EXCEEDED/WARNING/SAFE<br/>change_percent vs prev_used
    R-->>UC: []BudgetUsage
    UC-->>H: []BudgetUsage
    H-->>C: 200 []BudgetUsage
```

---

## 10. Transaction Create with Budget Alert Check (v1.3)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant TH as TransactionHandler
    participant TUC as TransactionUseCase
    participant TR as TransactionRepository
    participant ALR as ActivityLogRepository
    participant NUC as NotificationUseCase
    participant BR as BudgetRepository
    participant NR as NotificationRepository
    participant DB as PostgreSQL

    C->>TH: POST /api/auth/v1/transactions { amount, category, type, date }
    TH->>TUC: Create(userID, req)
    TUC->>TUC: validate type/amount/category/date
    TUC->>TR: Create(userID, req)
    TR->>DB: INSERT INTO transactions
    DB-->>TR: ok
    TUC->>ALR: Log(userID, CREATE, transaction, nil, desc)
    ALR->>DB: INSERT INTO activity_logs
    TUC-->>TH: nil
    Note over TH,NUC: goroutine — non-blocking
    TH-)NUC: CheckBudgetAlerts(userID) [async]
    NUC->>NR: GetPreferences(userID)
    NR->>DB: SELECT notification_preferences
    alt budget_alerts enabled
        NUC->>BR: GetUsage(userID, month, year)
        BR->>DB: complex JOIN
        DB-->>BR: []BudgetUsage
        loop for each EXCEEDED/WARNING budget
            NUC->>NR: ExistsRecent(userID, type, budgetID)
            alt no recent notification
                NUC->>NR: Create(userID, type, title, message, entity)
                NR->>DB: INSERT INTO notifications
            end
        end
    end
    TH-->>C: 200 { message: Transaction created successfully }
```

---

## 11. Get Monthly Summary (DashboardGraph data) (v1.3)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant TH as TransactionHandler
    participant TUC as TransactionUseCase
    participant TR as TransactionRepository
    participant DB as PostgreSQL

    C->>TH: GET /api/auth/v1/transactions/monthly-summary?months=6
    TH->>TUC: GetMonthlySummary(userID, 6)
    TUC->>TR: GetMonthlySummary(userID, 6)
    TR->>DB: SELECT EXTRACT(MONTH), EXTRACT(YEAR),<br/>SUM(INCOME), SUM(EXPENSE)<br/>FROM transactions<br/>WHERE user_id=$1 AND date >= NOW() - 5 months<br/>GROUP BY year, month ORDER BY year, month
    DB-->>TR: []MonthlySummaryItem
    TR-->>TUC: items
    TUC-->>TH: items
    TH-->>C: 200 { data: [{month, year, income, expense}...], months: 6 }
```

---

## 12. Dashboard with Financial Health Score (v1.3)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant DH as DashboardHandler
    participant DUC as DashboardUseCase
    participant TR as TransactionRepository
    participant BR as BudgetRepository
    participant GR as GoalRepository
    participant DB as PostgreSQL

    C->>DH: GET /api/auth/v1/dashboard
    DH->>DUC: Get(userID)
    DUC->>TR: GetMonthlyIncome(userID)
    TR->>DB: SELECT SUM(amount) INCOME current month
    DUC->>TR: GetMonthlyExpenses(userID)
    DUC->>TR: GetNetSavings(userID)
    DUC->>BR: GetUsage(userID, month, year)
    DUC->>GR: GetAll(userID, active=true)
    DUC->>GR: CountActive(userID)
    DUC->>DUC: computeFinancialHealth(income, netSavings, budgets, goals)
    Note over DUC: score = savings_rate(40%) + budget_adherence(35%) + goal_progress(25%)
    Note over DUC: label = Excellent(>=80) / Good(>=60) / Fair(>=40) / Needs Attention
    DUC-->>DH: DashboardResponse + FinancialHealth + has_anomalies
    DH-->>C: 200 { data: { ...existing fields, financial_health, has_anomalies } }
```

---

## 13. Transaction Bulk Import (v1.3)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant TH as TransactionHandler
    participant TUC as TransactionUseCase
    participant TR as TransactionRepository
    participant NUC as NotificationUseCase
    participant DB as PostgreSQL

    C->>TH: POST /api/auth/v1/transactions/import [{amount, category, type, date}...]
    TH->>TH: validate body (max 500 items)
    TH->>TUC: BulkImport(userID, items)
    loop for each item
        TUC->>TUC: validate type/amount/category/date format
        alt invalid row
            TUC->>TUC: append error, increment failed
        else valid
            TUC->>TUC: append to valid[]
        end
    end
    TUC->>TR: BulkCreate(userID, valid[])
    TR->>DB: BEGIN
    loop for each valid transaction
        TR->>DB: INSERT INTO transactions
    end
    TR->>DB: COMMIT
    TUC->>ALR: Log(IMPORT, transaction, nil, "Bulk imported N")
    TH-)NUC: CheckBudgetAlerts(userID) [async]
    TH-->>C: 200 { data: { imported: N, failed: M, errors: [...] } }
```
