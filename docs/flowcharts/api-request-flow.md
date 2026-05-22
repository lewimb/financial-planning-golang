# API Request Flow

Generic lifecycle of every HTTP request through the system. Derived from `main.go`, `internal/delivery/http/router.go`, and `internal/delivery/http/middleware/`.

---

## 1. Generic Request Lifecycle

```mermaid
flowchart TD
    A([Incoming HTTP Request]) --> B[Gin Engine]
    B --> C[CORS Middleware\nsets Access-Control headers\nOrigin: http://localhost:5173\nMethods: GET POST PUT PATCH DELETE OPTIONS]
    C --> D{OPTIONS preflight?}
    D -- Yes --> E[204 No Content\nAbort]
    D -- No --> F[Route matching]

    F --> G{Route prefix?}
    G -->|"/api/v1/*"| H[Public routes\nNo auth required]
    G -->|"/api/auth/v1/*"| I[JWT AuthMiddleware\nmiddleware.AuthRequired]
    G -->|No match| J[404 Not Found]

    H --> K[UserHandler\nRegister · Login · GetAll]

    I --> L{Token valid?}
    L -- No --> M[400/401 error\nAbort]
    L -- Yes --> N[c.Set claims MyCustomClaims\nuserID · name · email]
    N --> O[Route handler dispatch]

    O --> P[Handler extracts userID\nvia utils.ClaimId c]
    P --> Q[Parse + validate request\nbody or query params]
    Q --> R{Validation ok?}
    R -- No --> S[400 Bad Request]
    R -- Yes --> T[UseCase call]
    T --> U[Repository call]
    U --> V[PostgreSQL query\nraw SQL · pgx driver]
    V --> W[Response marshalled\nJSON]
    W --> X[HTTP response to client]
```

---

## 2. Route Map

```mermaid
flowchart TD
    subgraph Public ["/api/v1/ — No Auth"]
        P1[POST /register → UserHandler.Register]
        P2[POST /login → UserHandler.Login]
        P3[GET /users → UserHandler.GetAll]
    end

    subgraph Protected ["/api/auth/v1/ — JWT Required"]
        subgraph Transactions
            T1[GET /transactions]
            T2[POST /transactions]
            T3[PUT /transactions/:id]
            T4[DELETE /transactions/:id]
            T5[GET /transactions/monthly]
            T6[GET /transactions/monthly-income]
        end
        subgraph Budgets
            B1[GET /budgets]
            B2[POST /budgets]
            B3[GET /budgets/usage]
            B4[GET /budgets/:id]
            B5[PUT /budgets/:id]
            B6[DELETE /budgets/:id]
        end
        subgraph Goals
            G1[GET /goals]
            G2[POST /goals]
            G3[GET /goals/overview]
            G4[GET /goals/milestones]
            G5[GET /goals/:id]
            G6[PUT /goals/:id]
            G7[DELETE /goals/:id]
            G8[PATCH /goals/contribute]
        end
        subgraph Dashboard
            D1[GET /dashboard]
        end
        subgraph Chat
            C1[POST /chat]
        end
        subgraph ML
            M1[GET /ml/analysis]
            M2[GET /ml/anomaly]
            M3[GET /ml/forecast]
        end
        subgraph Profile
            FP1[POST /financial-profile]
            FP2[GET /financial-profile]
        end
    end
```

---

## 3. Handler → UseCase → Repository Dependency Map

```mermaid
flowchart TD
    subgraph Handlers
        UH[UserHandler]
        TH[TransactionHandler]
        BH[BudgetHandler]
        GH[GoalHandler]
        DH[DashboardHandler]
        CH[ChatHandler]
        MLH[MLHandler]
        FPH[FinancialProfileHandler]
    end

    subgraph UseCases
        UUC[UserUseCase]
        TUC[TransactionUseCase]
        BUC[BudgetUseCase]
        GUC[GoalUseCase]
        DUC[DashboardUseCase]
        CUC[ChatUseCase]
        MUC[MLUseCase]
        FUC[FinancialProfileUseCase]
    end

    subgraph Repositories
        UR[UserRepository]
        TR[TransactionRepository]
        BR[BudgetRepository]
        GR[GoalRepository]
        LR[AiLogRepository]
        FR[FinancialProfileRepository]
    end

    UH --> UUC --> UR
    TH --> TUC --> TR
    BH --> BUC --> BR
    GH --> GUC --> GR
    GUC -->|"net savings check"| TR
    DH --> DUC --> TR
    DUC --> BR
    DUC --> GR
    CH --> CUC --> TR
    CUC --> BR
    CUC --> GR
    CUC --> FR
    CUC --> LR
    MLH --> MUC --> TR
    FPH --> FUC --> FR
```

---

## 4. Error → HTTP Status Mapping

```mermaid
flowchart LR
    subgraph Sentinel Errors ["domain/errors.go"]
        E1[ErrNotFound]
        E2[ErrConflict]
        E3[ErrUnauthorized]
        E4[ErrInvalidInput]
    end
    subgraph UseCase Errors
        U1[ErrUserExists]
        U2[ErrInvalidCredentials]
        U3[ErrChatUnavailable]
        U4[ErrMLUnavailable]
    end
    subgraph HTTP Status
        S404[404 Not Found]
        S409[409 Conflict]
        S401[401 Unauthorized]
        S400[400 Bad Request]
        S503[503 Service Unavailable]
    end
    E1 --> S404
    E2 --> S409
    E3 --> S401
    E4 --> S400
    U1 --> S409
    U2 --> S400
    U3 --> S503
    U4 --> S503
```
