# System Architecture

Last audited against codebase: 2026-05-20. Updated for v1.3 — notifications, activity log, reports, chat history, recurring transactions, financial health score, monthly summary.

---

## 1. High-Level Overview

Three-tier application:

- **Go Backend** (Gin, port 8080) — authentication, business logic, data access, API routing
- **ML Service** (Python/FastAPI, port 8000) — stateless computation: spending analysis, anomaly detection, Prophet-based forecasting
- **PostgreSQL** (port 5432) — single relational database, raw SQL via pgx/v5 (no ORM)

External services:
- **Google Gemini API** — AI chat via `google.golang.org/genai` SDK (`gemini-3-flash-preview`)
- **Frontend** (React/Vite, port 5173) — not part of this repo; CORS origin hardcoded to `http://localhost:5173`

```mermaid
graph TB
    subgraph Client ["Client Layer"]
        FE["Frontend\n(React / Vite :5173)"]
        SEED["db/seeder\n(standalone CLI binary)"]
    end

    subgraph Backend ["Go Backend (:8080)"]
        direction TB
        Router["Gin Router\n/api/v1/* + /api/auth/v1/*"]
        Auth["AuthMiddleware\nJWT HS256 — middleware/auth.go"]

        subgraph Handlers ["Delivery Layer — internal/delivery/http/handler/"]
            UH["UserHandler\n+ GetMe"]
            TH["TransactionHandler\n+ MonthlySummary / Export / Import"]
            BH["BudgetHandler"]
            GH["GoalHandler\n+ completed_this_year"]
            DH["DashboardHandler\n+ financial_health / has_anomalies"]
            CH["ChatHandler\n+ GetHistory / ClearHistory"]
            MLH["MLHandler"]
            FPH["FinancialProfileHandler"]
            NH["NotificationHandler"]
            AH["ActivityLogHandler"]
            RH["ReportsHandler"]
        end

        subgraph UseCases ["Use Case Layer — internal/usecase/"]
            UUC["UserUseCase\nbcrypt + JWT + GetMe"]
            TUC["TransactionUseCase\nvalidation + bulk import + export\n+ monthly summary + activityRepo"]
            BUC["BudgetUseCase\nperiod + threshold defaults"]
            GUC["GoalUseCase\nnet-savings validation + completed_this_year"]
            DUC["DashboardUseCase\naggregations + financial_health score"]
            CUC["ChatUseCase\nprompt builder + history + clear"]
            MUC["MLUseCase\ntx → ml.Transaction + timeouts"]
            FUC["FinancialProfileUseCase\nvalidation + net_available calc"]
            NUC["NotificationUseCase\nbudget alert checker"]
            ALUC["ActivityLogUseCase\nread-only log query"]
            RUC["ReportsUseCase\nmonthly/category/savings/net-worth"]
        end

        subgraph Repositories ["Repository Layer — internal/repository/postgres/"]
            UR["UserRepository\n+ GetByID"]
            TR["TransactionRepository\n+ GetMonthlySummary / BulkCreate\n+ is_recurring fields"]
            BR["BudgetRepository"]
            GR["GoalRepository\n+ CountCompletedThisYear"]
            LR["AiLogRepository\n+ DeleteByUserID"]
            FR["FinancialProfileRepository"]
            NR["NotificationRepository\n+ preferences + dedup"]
            ALR["ActivityLogRepository\npaginated query"]
        end

        subgraph Domain ["Domain Layer — internal/domain/"]
            Interfaces["Repository Interfaces\n+ DTOs + Sentinel Errors"]
        end
    end

    subgraph External ["External Services"]
        ML["ML Service\nPython / FastAPI :8000\nPOST /analysis /anomaly /forecast"]
        GEMINI["Google Gemini API\ngemini-3-flash-preview\n3 retries + exponential backoff"]
    end

    subgraph Data ["Data Layer"]
        PG[("PostgreSQL :5432\n9 tables\n(7 active, 2 scaffolded)")]
    end

    FE -->|"HTTP JSON — Authorization: Bearer token"| Router
    Router --> Auth
    Auth --> Handlers
    Handlers --> UseCases
    UseCases --> Repositories
    Repositories -->|"raw SQL — pgx/v5 driver"| PG
    SEED -->|"raw SQL — pgx/v5 driver\n-fresh · -only flags"| PG
    MUC -->|"POST /analysis · /anomaly · /forecast\nJSON — no auth\ntimeouts: 5s/10s/60s"| ML
    CUC -->|"GenerateContent\ngRPC/HTTP SDK"| GEMINI
```

---

## 2. Clean Architecture Dependency Rule

```mermaid
flowchart LR
    D["Domain\ninternal/domain/\n(interfaces + DTOs + errors)"]
    R["Repository\ninternal/repository/postgres/\n(SQL implementations)"]
    U["Use Case\ninternal/usecase/\n(business logic)"]
    H["Delivery\ninternal/delivery/http/\n(Gin handlers + router + middleware)"]
    M["main.go\n(wiring only)"]

    D --> R
    D --> U
    U --> H
    M --> R
    M --> U
    M --> H
```

**Dependency direction:** All arrows point inward toward Domain. The Domain layer has zero imports from other internal packages.

---

## 3. Component Wiring (main.go)

```mermaid
flowchart TD
    ENV[".env / environment variables"] --> DB[initDB → *sql.DB]
    DB --> UR[NewUserRepository]
    DB --> TR[NewTransactionRepository]
    DB --> BR[NewBudgetRepository]
    DB --> GR[NewGoalRepository]
    DB --> LR[NewAiLogRepository]
    DB --> FR[NewFinancialProfileRepository]

    UR --> UUC[NewUserUseCase]
    TR --> TUC[NewTransactionUseCase]
    BR --> BUC[NewBudgetUseCase]
    GR --> GUC[NewGoalUseCase]
    TR --> GUC
    TR --> DUC[NewDashboardUseCase]
    BR --> DUC
    GR --> DUC
    TR --> CUC[NewChatUseCase]
    BR --> CUC
    GR --> CUC
    LR --> CUC
    FR --> CUC
    GEMINI2[GeminiClient] --> CUC
    TR --> MUC[NewMLUseCase]
    ML2[ml.NewClient] --> MUC
    FR --> FUC[NewFinancialProfileUseCase]

    UUC --> delivery[delivery.Setup → Gin router]
    TUC --> delivery
    BUC --> delivery
    GUC --> delivery
    DUC --> delivery
    CUC --> delivery
    MUC --> delivery
    FUC --> delivery
```

---

## 4. Authentication Architecture

```mermaid
flowchart TD
    subgraph JWT Token ["JWT Token (HS256)"]
        C1[id: int]
        C2[name: string]
        C3[email: string]
        C4[iss: lewimb]
        C5[NO exp claim]
    end
    subgraph Cookie
        K1[name: accessToken]
        K2[MaxAge: 3600 seconds]
        K3[path: /]
        K4[domain: *]
        K5[secure: false]
        K6[httpOnly: false]
    end
    L[Login success] --> JWT Token
    L --> Cookie
    Note1[Session length enforced by cookie TTL only]
```

---

## 5. ML Service Integration

```mermaid
flowchart LR
    subgraph Go Backend
        MUC[MLUseCase\nfetchMLTransactions]
        MC[ml.Client\nHTTP client]
    end
    subgraph ML Service FastAPI :8000
        A[POST /analysis\ntimeout 5s\nspending stats]
        B[POST /anomaly\ntimeout 10s\nstatistical detection\nmin 5 expense days]
        C[POST /forecast?periods=N\ntimeout 60s\nProphet time-series\nperiods clamped 1–365]
    end

    MUC --> MC
    MC -->|"[]ml.Transaction\ndate·amount·type·category"| A
    MC -->|"[]ml.Transaction"| B
    MC -->|"[]ml.Transaction"| C
    A --> MA[AnalysisResponse\ntotal_expense · avg_daily\ntop_category · distribution]
    B --> MB[AnomalyResponse\nanomalies array · summary]
    C --> MC2[ForecastResponse\npredicted_monthly_spending\ndaily_forecast array]
```

---

## 6. Database Schema Summary (9 tables)

| Table | Status | Delete Strategy | Notes |
|---|---|---|---|
| `users` | Active | soft (deleted_at) | Central entity |
| `transactions` | Active | soft (deleted_at) | BIGINT amount, IDR |
| `budgets` | Active | soft (deleted_at) | MONTHLY/YEARLY periods |
| `goals` | Active | hard (DELETE) | deleted_at dropped in migration 012 |
| `ai_logs` | Active | soft (deleted_at) | Not queried in app — only written |
| `user_financial_profiles` | Active | CASCADE on user delete | 1:1 per user |
| `user_financial_goals` | Active | CASCADE + DELETE before re-insert | FK tags for profile |
| `reports` | Scaffolded | — | No app routes |
| `settings` | Scaffolded | — | No app routes |

---

## 7. Architecture Issues and Improvement Opportunities

### Identified Issues

| Issue | Location | Severity | Notes |
|---|---|---|---|
| Sequential DB calls in Dashboard | `DashboardUseCase.Get` | Medium | 6 queries run one-by-one; `errgroup` could parallelize |
| Logical JOIN between budget and transaction categories | `BudgetRepository.GetUsage` | Medium | No FK enforcement; case-insensitive string match; typos silently break budget tracking |
| GoalUseCase cross-domain dependency on TransactionRepository | `GoalUseCase.Contribute` | Low | Intentional but creates tight coupling; documented |
| JWT has no expiry claim | `utils/jwt.go` | Medium | Token is valid until cookie expires; if cookie is stolen, token remains valid indefinitely |
| CORS hardcoded to localhost:5173 | `main.go` | Low | Production deployment requires env-var-driven CORS config |
| Reports and Settings exist in DB but have no API routes | migrations 006/007 | Low | Dead schema increases cognitive load |
| `GetAll` for users returns passwords in response | `UserRepository.GetAll` | High | `UserResponse.Password` is included in JSON — credential leak |
| Dashboard `GoalProgressSummary.Active` miscounts | `DashboardUseCase.Get` | Medium | Counts completed goals from `GetAll(active=true)` which only returns goals with `deadline >= NOW` — COMPLETED goals with past deadlines are excluded from denominator |

### Scalability Notes

- No pagination on `GetAll` for goals or budgets — unbounded result sets
- No background jobs or scheduled tasks — all computation is request-scoped
- ML service has no auth — anyone who can reach port 8000 can call it directly
- Single PostgreSQL instance — no read replicas or connection pooling beyond pgx defaults
