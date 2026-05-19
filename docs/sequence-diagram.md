# Sequence Diagrams

All sequence diagrams are derived from the actual Go source code in `internal/`.

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

    C->>H: POST /api/v1/register<br/>{ email, password, name }
    H->>UC: Register(email, password, name)
    UC->>UC: bcrypt.GenerateFromPassword(password, cost=10)
    UC->>R: Create(email, hashedPassword, name)
    R->>DB: INSERT INTO users (email, name, password)<br/>ON CONFLICT email → ErrConflict
    DB-->>R: result
    alt email already exists
        R-->>UC: domain.ErrConflict
        UC-->>H: ErrUserExists
        H-->>C: 409 { "error": "already exists" }
    else success
        R-->>UC: nil
        UC-->>H: nil
        H-->>C: 200 { "message": "User registered successfully" }
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

    C->>H: POST /api/v1/login<br/>{ email, password }
    H->>UC: Login(email, password)
    UC->>R: FindByEmail(email)
    R->>DB: SELECT id,email,name,password FROM users<br/>WHERE email=$1 AND deleted_at IS NULL
    DB-->>R: user row
    alt user not found
        R-->>UC: error
        UC-->>H: ErrInvalidCredentials
        H-->>C: 400 { "error": "invalid email or password" }
    else user found
        R-->>UC: *domain.User
        UC->>UC: bcrypt.CompareHashAndPassword(hash, password)
        alt password mismatch
            UC-->>H: ErrInvalidCredentials
            H-->>C: 400 { "error": "invalid email or password" }
        else password matches
            UC->>UC: utils.GenerateJwt(userID, name, email)<br/>HS256, signed with SECRET_KEY
            UC-->>H: tokenString
            H->>H: c.SetCookie("accessToken", token, 3600, ...)
            H-->>C: 200 { "message": "Login Successfully",<br/>"data": { "token": "eyJ..." } }
        end
    end
```

---

## 3. Authentication Middleware (all protected routes)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant MW as AuthMiddleware
    participant H as Handler

    C->>MW: Any /api/auth/v1/* request<br/>Authorization: Bearer <token>
    MW->>MW: Extract Authorization header
    alt header missing
        MW-->>C: 400 { "message": "Missing Authorization!" }
    else header present
        MW->>MW: Split "Bearer <token>"
        alt wrong format
            MW-->>C: 401 { "error": "authorization header format must be Bearer {token}" }
        else correct format
            MW->>MW: jwt.ParseWithClaims(token, &claims, SECRET_KEY)<br/>Verify HS256 signature
            alt token invalid or expired
                MW-->>C: 401 { "error": "Invalid token" }
            else token valid
                MW->>MW: c.Set("claims", claims)<br/>stores {id, name, email}
                MW->>H: c.Next() → handler proceeds
                Note over H: utils.ClaimId(c) extracts<br/>userID from context
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

    C->>MW: POST /api/auth/v1/transactions<br/>Authorization: Bearer <token><br/>{ amount, category, type, date, description }
    MW->>H: JWT valid → userID extracted
    H->>H: c.ShouldBindJSON(&req)
    alt bind fails
        H-->>C: 400 { "error": "..." }
    else bind succeeds
        H->>UC: Create(userID, req)
        UC->>UC: strings.ToUpper(req.Type)
        UC->>UC: validate: amount>0, type in [INCOME,EXPENSE],<br/>category non-empty, date non-zero
        alt validation fails
            UC-->>H: error message
            H-->>C: 400 { "error": "..." }
        else valid
            UC->>R: Create(userID, req)
            R->>DB: INSERT INTO transactions<br/>(amount, category, type, date, description, user_id)
            DB-->>R: nil
            R-->>UC: nil
            UC-->>H: nil
            H-->>C: 200 { "message": "Transaction created successfully" }
        end
    end
```

---

## 5. Dashboard Data Retrieval

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
    Note over UC: All queries run sequentially<br/>No parallel execution

    UC->>TR: GetMonthlyIncome(userID)
    TR->>DB: SELECT SUM(amount) WHERE type='INCOME'<br/>AND month=now AND year=now
    DB-->>TR: income float64
    TR-->>UC: income

    UC->>TR: GetMonthlyExpenses(userID)
    TR->>DB: SELECT SUM(amount) WHERE type='EXPENSE'<br/>AND month=now AND year=now
    DB-->>TR: expense float64
    TR-->>UC: expense

    UC->>TR: GetNetSavings(userID)
    TR->>DB: SELECT SUM(INCOME) - SUM(EXPENSE)<br/>all-time, no date filter
    DB-->>TR: netSavings float64
    TR-->>UC: netSavings

    UC->>BR: GetUsage(userID, currentMonth, currentYear)
    BR->>DB: SELECT b.*, SUM(t.amount) as used, SUM(prev.amount) as prev_used<br/>FROM budgets b LEFT JOIN transactions t<br/>WHERE t.type='EXPENSE' AND category matches
    DB-->>BR: []BudgetUsage (with computed status)
    BR-->>UC: budgetUsage

    UC->>GR: GetAll(userID, active=true)
    GR->>DB: SELECT * FROM goals WHERE user_id=$1<br/>AND deadline >= NOW()
    DB-->>GR: []GoalResponse
    GR-->>UC: activeGoals

    UC->>GR: CountActive(userID)
    GR->>DB: SELECT COUNT(*) FROM goals<br/>WHERE deadline >= NOW()
    DB-->>GR: count
    GR-->>UC: total

    UC->>UC: Aggregate BudgetStatusSummary<br/>(count SAFE/WARNING/EXCEEDED)
    UC->>UC: Aggregate GoalProgressSummary<br/>(total, active, completed)
    UC-->>H: *DashboardResponse
    H-->>C: 200 { "data": { monthly_income, monthly_expense,<br/>net_savings, budget_summary, goal_summary, active_goals } }
```

---

## 6. ML Spending Forecast

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
    H->>H: parse periods query param<br/>(default 30 if missing/invalid)
    H->>UC: GetForecast(userID, periods=30, year="", month="")
    UC->>UC: fetchMLTransactions(userID, "", "")
    UC->>TR: GetByUserID(userID, limit=0, offset=0, year="", month="")
    TR->>DB: SELECT id, amount, category, type, date<br/>FROM transactions WHERE user_id=$1<br/>AND deleted_at IS NULL<br/>ORDER BY date DESC<br/>(no LIMIT — all records)
    DB-->>TR: []TransactionResponse
    TR-->>UC: transactions
    UC->>UC: Convert to []ml.Transaction<br/>{ date: "2006-01-02", amount, type, category }
    UC->>UC: client.Forecast(transactions, periods=30)
    Note over UC,ML: context.WithTimeout(60s)
    UC->>ML: POST http://ML_SERVICE_URL/forecast?periods=30<br/>Content-Type: application/json<br/>Body: [ { date, amount, type, category }, ... ]
    alt ML service unreachable or non-200
        ML-->>UC: error
        UC-->>H: ErrMLUnavailable
        H-->>C: 503 { "error": "ML service unavailable" }
    else ML success
        ML-->>UC: 200 { "predicted_monthly_spending": N,<br/>"daily_forecast": [ { date, predicted_amount }, ... ] }
        UC-->>H: *ForecastResponse
        H-->>C: 200 { "predicted_monthly_spending": 3500000,<br/>"daily_forecast": [ ... ] }
    end
```

---

## 7. AI Chat

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

    C->>H: POST /api/auth/v1/chat<br/>{ "message": "How much did I spend?" }
    H->>H: c.ShouldBindJSON(&req)<br/>validate message non-empty
    H->>UC: Ask(ctx, userID, message)

    par Fetch financial context
        UC->>TR: GetMonthlyIncome(userID)
        TR->>DB: SELECT SUM(INCOME) current month
        DB-->>TR: income
        UC->>TR: GetMonthlyExpenses(userID)
        TR->>DB: SELECT SUM(EXPENSE) current month
        DB-->>TR: expense
        UC->>TR: GetNetSavings(userID)
        TR->>DB: SELECT net all-time
        DB-->>TR: netSavings
    end

    UC->>BR: GetUsage(userID, month, year)
    BR->>DB: budget usage query
    DB-->>BR: []BudgetUsage
    BR-->>UC: budgetUsage

    UC->>GR: GetAll(userID, active=true)
    GR->>DB: active goals
    DB-->>GR: []GoalResponse
    GR-->>UC: activeGoals

    UC->>PR: GetByUserID(userID)
    PR->>DB: SELECT profile + goals
    DB-->>PR: *FinancialProfileResponse (or ErrNotFound)
    alt profile exists
        PR-->>UC: profile
        UC->>UC: BuildFinancialProfileContext(profile)<br/>formats profile as text block
    else no profile
        PR-->>UC: ErrNotFound
        UC->>UC: profileSection = "" (silently skipped)
    end

    UC->>UC: Build Gemini prompt:<br/>"[Profile section]\nTransaction Data:\n- income...\n- expense...\nUser question: [message]"

    UC->>GEM: models.GenerateContent(ctx, model, Text(prompt))
    Note over UC,GEM: Up to 3 retries with exponential backoff<br/>on rate-limit errors (429)
    alt Gemini error (non-rate-limit)
        GEM-->>UC: error
        UC-->>H: ErrChatUnavailable
        H-->>C: 503 { "error": "AI service unavailable" }
    else Gemini success
        GEM-->>UC: GenerateContentResponse
        UC->>UC: result.Text()
        UC->>LR: Save(userID, message, reply)
        LR->>DB: INSERT INTO ai_logs (user_id, question, response)
        DB-->>LR: nil (error logged but not propagated)
        LR-->>UC: (async, non-blocking on error)
        UC-->>H: reply string
        H-->>C: 200 { "reply": "Pengeluaran Anda..." }
    end
```

---

## 8. Financial Profile Upsert

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as FinancialProfileHandler
    participant UC as FinancialProfileUseCase
    participant R as FinancialProfileRepository
    participant DB as PostgreSQL

    C->>H: POST /api/auth/v1/financial-profile<br/>{ monthly_income, fixed_expenses, current_savings,<br/>debt, employment_status, financial_goals, ... }
    H->>H: c.ShouldBindJSON(&req)
    H->>UC: Upsert(userID, req)
    UC->>UC: validateProfileRequest(req)<br/>check: all numbers >= 0<br/>employment_status non-empty<br/>financial_goals non-empty, no blank entries
    alt validation fails
        UC-->>H: error (e.g. "monthly_income must be >= 0")
        H-->>C: 400 { "error": "..." }
    else validation passes
        UC->>R: Upsert(userID, req)
        R->>DB: BEGIN TRANSACTION
        R->>DB: INSERT INTO user_financial_profiles (...)<br/>ON CONFLICT (user_id) DO UPDATE SET ..., updated_at=NOW()
        DB-->>R: profile upserted
        R->>DB: DELETE FROM user_financial_goals WHERE user_id=$1
        DB-->>R: old goals cleared
        loop for each goal in financial_goals
            R->>DB: INSERT INTO user_financial_goals (user_id, goal_type)<br/>ON CONFLICT DO NOTHING
            DB-->>R: goal inserted
        end
        R->>DB: COMMIT
        DB-->>R: committed
        R-->>UC: nil
        UC->>R: GetByUserID(userID)
        R->>DB: SELECT profile fields
        DB-->>R: profile row
        R->>DB: SELECT goal_type FROM user_financial_goals<br/>WHERE user_id=$1 ORDER BY id
        DB-->>R: []string goals
        R-->>UC: *FinancialProfileResponse
        UC->>UC: p.NetAvailable = income - fixed_expenses - debt
        UC-->>H: *FinancialProfileResponse
        H-->>C: 200 { "message": "profile saved", "data": { ...profile, net_available } }
    end
```

---

## 9. Goal Contribution

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as GoalHandler
    participant UC as GoalUseCase
    participant TR as TransactionRepository
    participant GR as GoalRepository
    participant DB as PostgreSQL

    C->>H: PATCH /api/auth/v1/goals/contribute<br/>{ "goal_id": 1, "contribution": 5000000 }
    H->>H: ShouldBindJSON(&req)<br/>validate goal_id > 0
    H->>UC: Contribute(goalID, userID, contribution)
    UC->>UC: validate: contribution > 0
    UC->>TR: GetNetSavings(userID)
    TR->>DB: SELECT SUM(INCOME) - SUM(EXPENSE)<br/>all-time
    DB-->>TR: netSavings
    alt netSavings <= 0
        UC-->>H: "cannot add contributions: no net savings"
        H-->>C: 400 { "error": "..." }
    else contribution > netSavings
        UC-->>H: "contribution exceeds available savings"
        H-->>C: 400 { "error": "..." }
    else valid
        UC->>GR: Contribute(goalID, userID, contribution)
        GR->>DB: UPDATE goals<br/>SET current_amount = $1,<br/>status = CASE WHEN $1 >= target_amount THEN 'COMPLETED' ELSE status END,<br/>updated_at = NOW()<br/>WHERE id=$2 AND user_id=$3
        DB-->>GR: nil
        GR-->>UC: nil
        UC-->>H: nil
        H-->>C: 200 { "message": "Contribution successful" }
    end
```
