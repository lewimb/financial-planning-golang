# Clean Architecture Migration — Design Spec

**Date:** 2026-05-16  
**Project:** financial-planning-golang  
**Approach:** Full rewrite under `internal/`, by-layer organization, Pragmatic Go use case style

---

## 1. Goals

- Enforce strict dependency direction: outer layers depend on inner layers, never the reverse
- Decouple business logic from frameworks (Gin) and database driver (`database/sql`)
- Make each layer independently testable via repository interfaces
- Fix cross-domain SQL leakage (`GoalContributions` queries transactions table through goal repo's DB)
- Delete dead/duplicate code (`controller/` package duplicates `handler/`)

---

## 2. Folder Structure

```
financial-planning-golang/
├── internal/
│   ├── domain/
│   │   ├── user.go          # User entity, DTOs, UserRepository interface
│   │   ├── transaction.go   # Transaction entity, DTOs, TransactionRepository interface
│   │   ├── budget.go        # Budget entity, DTOs, BudgetRepository interface
│   │   ├── goal.go          # Goal entity, DTOs, GoalRepository interface
│   │   └── errors.go        # Shared sentinel errors
│   ├── usecase/
│   │   ├── user.go          # UserUseCase struct + methods
│   │   ├── transaction.go   # TransactionUseCase struct + methods
│   │   ├── budget.go        # BudgetUseCase struct + methods
│   │   └── goal.go          # GoalUseCase struct + methods
│   ├── repository/
│   │   └── postgres/
│   │       ├── user.go      # Implements domain.UserRepository
│   │       ├── transaction.go
│   │       ├── budget.go
│   │       └── goal.go
│   └── delivery/
│       └── http/
│           ├── handler/
│           │   ├── user.go
│           │   ├── transaction.go
│           │   ├── budget.go
│           │   └── goal.go
│           ├── middleware/
│           │   └── auth.go
│           └── router.go
├── utils/              # unchanged (jwt.go, bcrypt.go, claims.go)
├── db/migrations/      # unchanged
├── main.go
├── go.mod
└── .env
```

**Packages deleted:** `model/`, `service/`, `handler/`, `controller/`, `routes/`, `repository/` (top-level aggregator)

---

## 3. Dependency Rule

```
domain/ ← usecase/ ← delivery/http/
domain/ ← repository/postgres/
```

- `domain/` imports nothing from this project
- `usecase/` imports only `domain/`
- `repository/postgres/` imports `domain/` and `database/sql`
- `delivery/http/` imports `domain/` (errors) and `usecase/` (concrete structs)
- `main.go` imports all layers for wiring

Go's `internal/` enforces that nothing outside the module can import these packages.

---

## 4. Domain Layer

### `domain/errors.go`

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrConflict     = errors.New("already exists")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)
```

### `domain/transaction.go`

```go
type Transaction struct {
    ID          int
    UserID      int
    Amount      int
    Category    string
    Type        string  // "INCOME" | "EXPENSE"
    Date        time.Time
    Description string
}

type TransactionRequest struct {
    Amount      int
    Category    string
    Type        string
    Date        time.Time
    Description string
}

type TransactionResponse struct {
    ID          int
    Amount      float64
    Category    string
    Type        string
    Date        time.Time
    Description string
}

type TransactionRepository interface {
    GetByUserID(userID, limit, offset int, year, month string) ([]TransactionResponse, int, error)
    Create(userID int, req TransactionRequest) error
    Update(userID, id int, req TransactionRequest) error
    Delete(userID, id int) error
    GetMonthlyExpenses(userID int) (float64, error)
    GetNetSavings(userID int) (float64, error)
}
```

### `domain/goal.go`

```go
type GoalRepository interface {
    GetAll(userID int, active bool) ([]GoalResponse, error)
    GetByID(id, userID int) (*GoalResponse, error)
    GetSavingsTotal(userID int) (float64, error)
    CountActive(userID int) (int, error)
    GetUpcomingMilestones(userID int) ([]GoalResponse, error)
    Create(userID int, req CreateGoalRequest) error
    Update(id, userID int, req CreateGoalRequest) error
    Delete(id, userID int) error
    Contribute(id, userID, amount int) error
}
```

### `domain/budget.go`

```go
type BudgetRepository interface {
    GetAll(userID int, category, month, year string) ([]Budget, error)
    GetByID(id int) (*BudgetResponse, error)
    GetUsage(userID, month, year int) ([]BudgetUsage, error)  // encapsulates JOIN query
    Create(userID int, req CreateBudgetRequest) error
    Update(userID, id, limitAmount, alertThreshold int, category string) (*UpdateBudgetResponse, error)
    Delete(userID, id int) error
}
```

### `domain/user.go`

```go
type UserRepository interface {
    GetAll() ([]UserResponse, error)
    FindByEmail(email string) (*User, error)
    Create(email, hashedPassword, name string) error
}
```

---

## 5. Use Case Layer

Use case structs receive repository interfaces via constructor injection. No framework imports.

### `usecase/transaction.go`

```go
type TransactionUseCase struct {
    repo domain.TransactionRepository
}

func NewTransactionUseCase(repo domain.TransactionRepository) *TransactionUseCase

func (uc *TransactionUseCase) GetTransactions(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error)
func (uc *TransactionUseCase) Create(userID int, req domain.TransactionRequest) error
func (uc *TransactionUseCase) Update(userID, id int, req domain.TransactionRequest) error
func (uc *TransactionUseCase) Delete(userID, id int) error
func (uc *TransactionUseCase) GetMonthlyExpenses(userID int) (float64, error)
```

Validation (type must be INCOME/EXPENSE, amount > 0, etc.) moves from `service/` into use case methods.

### `usecase/goal.go` — cross-domain fix

```go
type GoalUseCase struct {
    repo   domain.GoalRepository
    txRepo domain.TransactionRepository  // injected for net savings check
}

func NewGoalUseCase(repo domain.GoalRepository, txRepo domain.TransactionRepository) *GoalUseCase

func (uc *GoalUseCase) Contribute(id, userID, amount int) error {
    net, err := uc.txRepo.GetNetSavings(userID)  // no raw SQL, no cross-table access
    if net <= 0 || amount > int(net) {
        return domain.ErrInvalidInput
    }
    return uc.repo.Contribute(id, userID, amount)
}
```

### `usecase/budget.go`

```go
type BudgetUseCase struct {
    repo domain.BudgetRepository
}
// methods: GetBudgets, Create, GetUsage, GetByID, Update, Delete
```

### `usecase/user.go`

```go
type UserUseCase struct {
    repo domain.UserRepository
}
// methods: Register, Login, GetAll
// Register/Login business logic (hash check, JWT generation) stays here
```

---

## 6. Repository Layer

Each file in `repository/postgres/` implements one domain repository interface using raw SQL and `*sql.DB`. No business logic.

**`GetAllBudgetUsage` JOIN query** stays in `repository/postgres/budget.go` as a single optimized SQL query. This is a read-model concern, not a domain violation — the interface (`GetUsage`) hides the implementation detail.

---

## 7. Delivery Layer

### Handlers

Struct-based handlers, one per domain. Depend on concrete use case structs.

```go
type TransactionHandler struct {
    uc *usecase.TransactionUseCase
}

func NewTransactionHandler(uc *usecase.TransactionUseCase) *TransactionHandler
```

Error mapping in handlers:

```go
switch {
case errors.Is(err, domain.ErrNotFound):     c.JSON(404, gin.H{"error": err.Error()})
case errors.Is(err, domain.ErrConflict):     c.JSON(409, gin.H{"error": err.Error()})
case errors.Is(err, domain.ErrInvalidInput): c.JSON(400, gin.H{"error": err.Error()})
default:                                      c.JSON(500, gin.H{"error": "internal error"})
}
```

### Router (`delivery/http/router.go`)

```go
type Deps struct {
    UserUC        *usecase.UserUseCase
    TransactionUC *usecase.TransactionUseCase
    BudgetUC      *usecase.BudgetUseCase
    GoalUC        *usecase.GoalUseCase
}

func Setup(r *gin.Engine, deps Deps) {
    // public routes
    userH := handler.NewUserHandler(deps.UserUC)
    r.POST("/api/v1/register", userH.Register)
    r.POST("/api/v1/login",    userH.Login)
    r.GET("/api/v1/users",     userH.GetAll)

    // protected routes
    api := r.Group("/api/auth/v1", middleware.AuthRequired())

    txH := handler.NewTransactionHandler(deps.TransactionUC)
    api.GET("/transactions",              txH.GetAll)
    api.POST("/transactions",             txH.Create)
    api.PUT("/transactions/:id",          txH.Update)
    api.DELETE("/transactions/:id",       txH.Delete)
    api.GET("/transactions/monthly",      txH.GetMonthlyExpenses)

    bH := handler.NewBudgetHandler(deps.BudgetUC)
    api.GET("/budgets",          bH.GetAll)
    api.POST("/budgets",         bH.Create)
    api.GET("/budgets/usage",    bH.GetUsage)
    api.GET("/budgets/:id",      bH.GetByID)
    api.PUT("/budgets/:id",      bH.Update)
    api.DELETE("/budgets/:id",   bH.Delete)

    gH := handler.NewGoalHandler(deps.GoalUC)
    api.GET("/goals",              gH.GetAll)
    api.POST("/goals",             gH.Create)
    api.GET("/goals/overview",     gH.GetOverview)
    api.GET("/goals/milestones",   gH.GetMilestones)
    api.GET("/goals/:id",          gH.GetByID)
    api.PUT("/goals/:id",          gH.Update)
    api.DELETE("/goals/:id",       gH.Delete)
    api.PATCH("/goals/contribute", gH.Contribute)
}
```

### Middleware (`delivery/http/middleware/auth.go`)

Moved verbatim from `middleware/middleware.go`. No logic changes.

---

## 8. Main.go Wiring

```go
func main() {
    godotenv.Load()
    db := initDB(...)

    // repositories
    userRepo := postgres.NewUserRepository(db)
    txRepo   := postgres.NewTransactionRepository(db)
    budgetRepo := postgres.NewBudgetRepository(db)
    goalRepo := postgres.NewGoalRepository(db)

    // use cases
    userUC   := usecase.NewUserUseCase(userRepo)
    txUC     := usecase.NewTransactionUseCase(txRepo)
    budgetUC := usecase.NewBudgetUseCase(budgetRepo)
    goalUC   := usecase.NewGoalUseCase(goalRepo, txRepo)

    // delivery
    r := gin.Default()
    r.Use(corsMiddleware())
    http.Setup(r, http.Deps{
        UserUC:        userUC,
        TransactionUC: txUC,
        BudgetUC:      budgetUC,
        GoalUC:        goalUC,
    })

    r.Run()
}
```

---

## 9. What Does Not Change

- `utils/` (jwt.go, bcrypt.go, claims.go) — no structural change; `MyCustomClaims` struct moves from `model/auth.go` into `utils/claims.go` where `ClaimId` already lives
- `db/migrations/` — untouched
- `.env`, `.air.toml`, `go.mod` — untouched
- CORS config — moved from inline `main.go` closure to named `corsMiddleware()` function in `main.go`
- All SQL queries — moved verbatim into `repository/postgres/`, no rewrites
- Route paths — preserved exactly as-is

---

## 10. Out of Scope

- Adding new features or endpoints
- Changing SQL queries (beyond moving them)
- Adding tests (separate task)
- ORM adoption
- SQL transactions for multi-step operations
