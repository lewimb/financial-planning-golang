# System Architecture

## 1. High-Level Overview

The system is a three-tier application composed of:

- **Go Backend** (Gin, port 8080) — authentication, business logic, data access, API routing
- **ML Service** (Python/FastAPI, port 8000) — stateless computation: analysis, anomaly detection, forecasting
- **PostgreSQL** (port 5432) — single relational database, raw SQL (no ORM)

External services:
- **Google Gemini API** — AI chat responses via `google.golang.org/genai` SDK
- **Frontend** (React/Vite, port 5173) — not part of this repo; CORS configured for this origin

```mermaid
graph TB
    subgraph Client ["Client Layer"]
        FE["Frontend\n(React / Vite :5173)"]
    end

    subgraph Backend ["Go Backend (:8080)"]
        direction TB
        Router["Gin Router\n/api/v1/* + /api/auth/v1/*"]
        Auth["Auth Middleware\nJWT HS256 verification"]

        subgraph Handlers ["Delivery Layer\ninternal/delivery/http/handler/"]
            UH["UserHandler"]
            TH["TransactionHandler"]
            BH["BudgetHandler"]
            GH["GoalHandler"]
            DH["DashboardHandler"]
            CH["ChatHandler"]
            MLH["MLHandler"]
            FPH["FinancialProfileHandler"]
        end

        subgraph UseCases ["Use Case Layer\ninternal/usecase/"]
            UUC["UserUseCase\nbcrypt + JWT"]
            TUC["TransactionUseCase\nvalidation + normalisation"]
            BUC["BudgetUseCase"]
            GUC["GoalUseCase\nnet savings check"]
            DUC["DashboardUseCase\naggregation"]
            CUC["ChatUseCase\nprompt builder"]
            MUC["MLUseCase\ntx → ML format"]
            FUC["FinancialProfileUseCase\nvalidation + net_available"]
        end

        subgraph Repositories ["Repository Layer\ninternal/repository/postgres/"]
            UR["UserRepository"]
            TR["TransactionRepository"]
            BR["BudgetRepository"]
            GR["GoalRepository"]
            LR["AiLogRepository"]
            FR["FinancialProfileRepository\n(atomic tx for goals)"]
        end

        subgraph Domain ["Domain Layer\ninternal/domain/"]
            Interfaces["Repository Interfaces\n+ DTOs + Sentinel Errors"]
        end
    end

    subgraph External ["External Services"]
        ML["ML Service\nPython / FastAPI :8000\n/analysis /anomaly /forecast"]
        GEMINI["Google Gemini API\ngemini-3-flash-preview"]
    end

    subgraph Data ["Data Layer"]
        PG[("PostgreSQL :5432\n9 tables")]
    end

    FE -->|"HTTP JSON\nAuthorization: Bearer token"| Router
    Router --> Auth
    Auth --> Handlers
    Handlers --> UseCases
    UseCases --> Repositories
    Repositories -->|"raw SQL\npgx/v5 driver"| PG
    MUC -->|"POST /analysis\nPOST /anomaly\nPOST /forecast\n(JSON, no auth)"| ML
    CUC -->|"GenerateContent\n(gRPC/HTTP SDK)"| GEMINI
```

---

## 2. Clean Architecture — Dependency Rules

The codebase strictly follows Clean Architecture with `internal/` enforcing Go's package visibility rules.

```mermaid
graph LR
    subgraph "Dependency Direction (inward only)"
        DEL["Delivery\ninternal/delivery/http/"]
        UC["Use Case\ninternal/usecase/"]
        REPO["Repository\ninternal/repository/postgres/"]
        DOM["Domain\ninternal/domain/"]
        ML_PKG["ML Client\ninternal/ml/"]
    end

    DEL -->|"depends on"| UC
    DEL -->|"depends on"| DOM
    UC -->|"depends on"| DOM
    UC -->|"depends on"| ML_PKG
    REPO -->|"implements"| DOM
    ML_PKG -->|"HTTP client only\nno domain import"| DOM

    MAIN["main.go\n(composition root)"] -->|"imports all layers\nwires dependencies"| DEL
    MAIN --> UC
    MAIN --> REPO
    MAIN --> ML_PKG
```

**Key rule:** `domain/` imports nothing from this project. It defines interfaces, DTOs, and sentinel errors. All other layers depend inward on `domain/`, never outward.

`main.go` is the **composition root** — the only file that imports all layers and wires concrete types to interfaces via constructor injection.

---

## 3. Request Lifecycle

```mermaid
graph TD
    REQ["HTTP Request\nfrom Frontend"]
    CORS["corsMiddleware\norigin check + CORS headers"]
    ROUTE{"Route matching\nGin router"}
    PUBLIC["Public route\n/api/v1/*"]
    PROTECTED_MW["AuthRequired middleware\nParse Bearer token\nValidate HS256 signature\nStore claims in context"]
    PROTECTED["Protected handler\n/api/auth/v1/*"]
    HANDLER["Handler\n1. utils.ClaimId(c) → userID\n2. ShouldBindJSON\n3. call use case"]
    USECASE["Use Case\n1. Validate inputs\n2. Apply business rules\n3. Call repositories"]
    REPO["Repository\nBuild SQL query\nExecute via *sql.DB"]
    DB[("PostgreSQL")]
    RESP["JSON Response\ngin.H{} or typed struct"]

    REQ --> CORS
    CORS --> ROUTE
    ROUTE --> PUBLIC
    ROUTE --> PROTECTED_MW
    PROTECTED_MW -->|"valid JWT"| PROTECTED
    PROTECTED_MW -->|"invalid"| RESP
    PUBLIC --> HANDLER
    PROTECTED --> HANDLER
    HANDLER --> USECASE
    USECASE --> REPO
    REPO --> DB
    DB --> REPO
    REPO --> USECASE
    USECASE --> HANDLER
    HANDLER --> RESP
```

---

## 4. ML Integration Architecture

```mermaid
graph LR
    subgraph "Go Backend"
        MLH["MLHandler\nparse query params"]
        MUC["MLUseCase\nfetchMLTransactions"]
        TRepo["TransactionRepository\nlimit=0 fetches all records"]
        CLIENT["ml.Client\nHTTP POST with timeout"]
    end

    subgraph "ML Service (Python/FastAPI)"
        ANALYSIS["POST /analysis\nPandas aggregation\ntimeout: 5s"]
        ANOMALY["POST /anomaly\nIsolationForest\ntimeout: 10s"]
        FORECAST["POST /forecast?periods=N\nFacebook Prophet\ntimeout: 60s"]
    end

    DB[("PostgreSQL")]

    MLH -->|"userID, year, month, periods"| MUC
    MUC --> TRepo
    TRepo -->|"SELECT all transactions\nwhere user_id = $1"| DB
    DB --> TRepo
    TRepo --> MUC
    MUC -->|"[]ml.Transaction\n{date, amount, type, category}"| CLIENT
    CLIENT -->|"POST JSON array"| ANALYSIS
    CLIENT -->|"POST JSON array"| ANOMALY
    CLIENT -->|"POST JSON array"| FORECAST
    ANALYSIS -->|"*AnalysisResponse"| CLIENT
    ANOMALY -->|"*AnomalyResponse"| CLIENT
    FORECAST -->|"*ForecastResponse"| CLIENT
    CLIENT --> MUC
    MUC --> MLH
```

**Key properties of ML integration:**
- **Stateless:** ML service holds no state between calls. Full transaction array sent every time.
- **Timeout per endpoint:** analysis 5s, anomaly 10s, forecast 60s (context.WithTimeout).
- **INCOME silently ignored:** ML service filters only EXPENSE records internally. Go sends all types.
- **URL configuration:** `ML_SERVICE_URL` env var, falls back to `http://localhost:8000`.
- **Error wrapping:** Any ML error → `usecase.ErrMLUnavailable` → HTTP 503.

---

## 5. Authentication Architecture

```mermaid
graph TD
    LOGIN["POST /api/v1/login"]
    PW_CHECK["bcrypt.CompareHashAndPassword\n(cost=10)"]
    JWT_GEN["utils.GenerateJwt\nHS256, claims: {id, name, email}\nIssuer: lewimb\nNo expiry in claims"]
    COOKIE["c.SetCookie(accessToken, token, 3600)\nhttpOnly not set — JS-readable"]
    RESPONSE["Response body\n{ data: { token } }"]

    MIDDLEWARE["AuthRequired middleware\nreads Authorization header ONLY\n(cookie is set but NOT read by middleware)"]
    PARSE["jwt.ParseWithClaims\nverify HS256 signature\ncheck signing method"]
    CONTEXT["c.Set(claims, MyCustomClaims)\nstores {id, name, email}"]
    CLAIMID["utils.ClaimId(c)\nextract userID for DB queries"]

    LOGIN --> PW_CHECK
    PW_CHECK --> JWT_GEN
    JWT_GEN --> COOKIE
    JWT_GEN --> RESPONSE

    MIDDLEWARE --> PARSE
    PARSE -->|"valid"| CONTEXT
    CONTEXT --> CLAIMID
```

---

## 6. Component Dependency Graph

```mermaid
graph TB
    subgraph "main.go (composition root)"
        MAIN["Wire all dependencies"]
    end

    MAIN -->|"creates"| UR["postgres.UserRepository"]
    MAIN -->|"creates"| TR["postgres.TransactionRepository"]
    MAIN -->|"creates"| BR["postgres.BudgetRepository"]
    MAIN -->|"creates"| GR["postgres.GoalRepository"]
    MAIN -->|"creates"| LR["postgres.AiLogRepository"]
    MAIN -->|"creates"| PR["postgres.FinancialProfileRepository"]
    MAIN -->|"creates"| MLC["ml.NewClient()"]
    MAIN -->|"creates"| GEM["usecase.NewGeminiClient(ctx)"]

    MAIN -->|"NewUserUseCase(UR)"| UUC["UserUseCase"]
    MAIN -->|"NewTransactionUseCase(TR)"| TUC["TransactionUseCase"]
    MAIN -->|"NewBudgetUseCase(BR)"| BUC["BudgetUseCase"]
    MAIN -->|"NewGoalUseCase(GR, TR)"| GUC["GoalUseCase"]
    MAIN -->|"NewDashboardUseCase(TR, BR, GR)"| DUC["DashboardUseCase"]
    MAIN -->|"NewChatUseCase(TR, BR, GR, LR, PR, GEM)"| CUC["ChatUseCase"]
    MAIN -->|"NewMLUseCase(TR, MLC)"| MUC["MLUseCase"]
    MAIN -->|"NewFinancialProfileUseCase(PR)"| FUC["FinancialProfileUseCase"]

    MAIN -->|"delivery.Setup(r, Deps{...})"| ROUTER["Gin Router\n+ all handlers"]
```

---

## 7. Environment Configuration

| Variable | Used by | Default | Required |
|---|---|---|---|
| `DB_USER` | main.go initDB | — | Yes |
| `DB_PASSWORD` | main.go initDB | — | Yes |
| `DB_NAME` | main.go initDB | — | Yes |
| `DB_HOST` | main.go initDB | — | Yes |
| `DB_PORT` | main.go initDB | — | Yes |
| `SECRET_KEY` | JWT sign + verify | — | Yes |
| `GEMINI_API_KEY` | genai.NewClient | — | Yes (503 if missing) |
| `ML_SERVICE_URL` | ml.NewClient | `http://localhost:8000` | No |

---

## 8. Architecture Assessment

### Missing Concerns

| Concern | Current State | Impact |
|---|---|---|
| **No interface for ml.Client** | `MLUseCase` takes `*ml.Client` (concrete), not an interface | Cannot mock ML client in tests; violates DIP |
| **No caching** | Dashboard hits 5 DB queries on every request; ML results recomputed every call | Latency and DB load at scale |
| **No connection pool config** | `sql.Open` with defaults (unlimited connections) | Connection exhaustion under load |
| **No rate limiting** | `/chat` endpoint calls Gemini without per-user throttle | Gemini quota exhaustion possible |
| **`reports` / `settings` scaffolded** | Tables exist, no Go implementation | Dead schema; maintenance confusion |
| **Duplicate `.env` load in main.go** | `godotenv.Load()` called twice | Minor bug; no functional impact |
| **JWT no expiry in claims** | Only cookie expiry enforced | Token valid indefinitely if extracted from cookie |
| **Middleware reads header only** | Cookie set on login but AuthRequired ignores it | Cookie-based auth silently fails |

### Tight Coupling

| Coupling | Location | Severity |
|---|---|---|
| `ChatUseCase` has 6 dependencies | `usecase/chat.go` NewChatUseCase | Medium — hard to test in isolation |
| `GoalUseCase` injects `TransactionRepository` for net savings | Cross-domain dependency | Low — intentional, documented |
| `DashboardUseCase` aggregates from 3 repos | `usecase/dashboard.go` | Low — acceptable for read aggregation |
| Budget↔Transaction join by string category | `repository/postgres/budget.go` GetUsage | High — typo silently breaks tracking |

### Scalability Concerns

| Concern | Detail |
|---|---|
| **ML forecast 60s timeout** | Blocks a goroutine per in-flight request; no queuing |
| **Full transaction list to ML** | `limit=0` fetches all records; grows unbounded with user history |
| **No async processing** | Dashboard, forecast, chat are all synchronous request-response |
| **Single DB instance** | No read replica or connection pooling config |

### Suggested Improvements

1. **Define `MLClientInterface`** in `internal/ml/` so `MLUseCase` depends on an interface → enables testing and swapping implementations.
2. **Add Redis cache** for dashboard (TTL 1–5 min) and ML results (TTL per forecast period).
3. **Set `sql.DB` pool params** (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`) in `main.go`.
4. **Move forecast to async job** — accept request, return job ID, poll for result — to avoid 60s HTTP timeout.
5. **Add token expiry to JWT claims** (`ExpiresAt` in `RegisteredClaims`) and refresh flow.
6. **Implement `settings` and `reports` routes** or remove the tables to reduce schema confusion.
7. **Category normalisation** — enforce consistent casing on transaction insert rather than relying on `LOWER()` in JOIN.
