# Clean Architecture Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the financial-planning-golang backend from a flat layered structure to clean architecture under `internal/` with strict dependency direction enforced by Go's visibility rules.

**Architecture:** By-layer organization under `internal/` — `domain/` (entities + repository interfaces), `usecase/` (business logic, concrete structs), `repository/postgres/` (SQL implementations), `delivery/http/` (Gin handlers + router). Pragmatic Go style: repository interfaces in `domain/`, use case structs are concrete (no use case interfaces). `GoalUseCase` receives both `GoalRepository` and `TransactionRepository` to fix current cross-domain SQL leakage.

**Tech Stack:** Go 1.25, Gin, pgx/v5 (raw SQL via `database/sql`), golang-jwt/v5, bcrypt, godotenv

**Spec:** `docs/superpowers/specs/2026-05-16-clean-architecture-design.md`

**Build verification strategy:** Each task verifies its package compiles with `go build ./internal/...`. Full `go build .` only passes after Task 9 (delete old packages) + Task 10 (rewrite main.go).

---

## File Map

**Create:**
- `utils/claims.go` — rewrite: add `MyCustomClaims` struct here, remove `model` import
- `utils/jwt.go` — rewrite: use `utils.MyCustomClaims` instead of `model.MyCustomClaims`
- `internal/domain/errors.go` — sentinel errors
- `internal/domain/user.go` — User entity, DTOs, UserRepository interface
- `internal/domain/transaction.go` — Transaction DTOs, TransactionRepository interface
- `internal/domain/budget.go` — Budget entities, DTOs, BudgetRepository interface
- `internal/domain/goal.go` — Goal entities, DTOs, GoalRepository interface
- `internal/repository/postgres/user.go` — implements domain.UserRepository
- `internal/repository/postgres/transaction.go` — implements domain.TransactionRepository
- `internal/repository/postgres/budget.go` — implements domain.BudgetRepository
- `internal/repository/postgres/goal.go` — implements domain.GoalRepository
- `internal/usecase/user.go` — Register, Login, GetAll
- `internal/usecase/transaction.go` — CRUD + monthly expenses
- `internal/usecase/budget.go` — CRUD + usage
- `internal/usecase/goal.go` — CRUD + overview + contribute (injects TransactionRepository)
- `internal/delivery/http/middleware/auth.go` — JWT guard
- `internal/delivery/http/handler/user.go` — UserHandler
- `internal/delivery/http/handler/transaction.go` — TransactionHandler
- `internal/delivery/http/handler/budget.go` — BudgetHandler
- `internal/delivery/http/handler/goal.go` — GoalHandler
- `internal/delivery/http/router.go` — Setup function + Deps struct

**Rewrite:**
- `main.go` — thin wiring: initDB → repos → usecases → Setup → Run

**Delete (after main.go rewrite):**
- `model/` (auth.go, user.go, transaction.go, budget.go, goal.go)
- `service/` (auth.go, user.go, transaction.go, budget.go, goal.go)
- `handler/` (auth.go, transactions.go, budget.go, goal.go)
- `controller/` (user.go)
- `routes/` (routes.go, auth_routes.go, transaction_routes.go, budget_routes.go, goal_routes.go)
- `repository/repository.go`
- `middleware/middleware.go`

---

## Task 1: Update utils/ — move MyCustomClaims out of model

`model/auth.go` currently defines `MyCustomClaims`. Both `utils/claims.go` and `utils/jwt.go` import `model` for it. Move the struct to `utils/` so `utils` has no dependency on `model/`.

**Files:**
- Rewrite: `utils/claims.go`
- Rewrite: `utils/jwt.go`

- [ ] **Step 1: Rewrite `utils/claims.go`**

```go
package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func ClaimId(c *gin.Context) int {
	claims, exist := c.Get("claims")
	if !exist {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return 0
	}
	userClaims := claims.(MyCustomClaims)
	return userClaims.Id
}
```

- [ ] **Step 2: Rewrite `utils/jwt.go`**

```go
package utils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwt(userId int, name string, email string) (string, error) {
	fmt.Println("Generating JWT for user:", userId, name, email)
	claims := MyCustomClaims{
		Id:    userId,
		Name:  name,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "lewimb",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
```

- [ ] **Step 3: Verify utils compiles**

```
go build ./utils/...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```
git add utils/claims.go utils/jwt.go
git commit -m "refactor(utils): move MyCustomClaims into utils, remove model dependency"
```

---

## Task 2: Create domain layer

All domain files are pure Go — no framework imports, no `database/sql`. They define entities, DTOs, and repository interfaces.

**Files:**
- Create: `internal/domain/errors.go`
- Create: `internal/domain/user.go`
- Create: `internal/domain/transaction.go`
- Create: `internal/domain/budget.go`
- Create: `internal/domain/goal.go`

- [ ] **Step 1: Create `internal/domain/errors.go`**

```go
package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)
```

- [ ] **Step 2: Create `internal/domain/user.go`**

```go
package domain

import "time"

type User struct {
	ID       int
	Email    string
	Password string
	Name     string
}

type UserResponse struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Password  string     `json:"password"`
}

type UserRepository interface {
	GetAll() ([]UserResponse, error)
	FindByEmail(email string) (*User, error)
	Create(email, hashedPassword, name string) error
}
```

- [ ] **Step 3: Create `internal/domain/transaction.go`**

```go
package domain

import "time"

type TransactionRequest struct {
	Amount      int       `json:"amount"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type TransactionResponse struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Type        string    `json:"type"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
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

- [ ] **Step 4: Create `internal/domain/budget.go`**

```go
package domain

import "time"

type Budget struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Category       string    `json:"category"`
	Period         string    `json:"period"`
	Month          *int      `json:"month"`
	Year           int       `json:"year"`
	LimitAmount    int       `json:"limit_amount"`
	AlertThreshold int       `json:"alert_threshold"`
	CreatedAt      time.Time `json:"created_at"`
}

type BudgetResponse struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	Category       string    `json:"category"`
	Period         string    `json:"period"`
	Month          *int      `json:"month"`
	Year           int       `json:"year"`
	LimitAmount    int       `json:"limitAmount"`
	AlertThreshold int       `json:"alertThreshold"`
	CreatedAt      time.Time `json:"createdAt"`
}

type UpdateBudgetRequest struct {
	LimitAmount    int    `json:"limitAmount" binding:"required"`
	AlertThreshold int    `json:"alertThreshold"`
	Category       string `json:"category"`
}

type UpdateBudgetResponse struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	Category       string `json:"category"`
	Period         string `json:"period"`
	Month          *int   `json:"month"`
	Year           int    `json:"year"`
	LimitAmount    int    `json:"limit_amount"`
	AlertThreshold int    `json:"alert_threshold"`
}

type CreateBudgetRequest struct {
	Category       string `json:"category" binding:"required"`
	Period         string `json:"period" binding:"required"`
	Month          *int   `json:"month"`
	Year           int    `json:"year" binding:"required"`
	LimitAmount    int    `json:"limit_amount" binding:"required"`
	AlertThreshold int    `json:"alert_threshold"`
}

type BudgetUsage struct {
	BudgetID       int     `json:"budget_id"`
	Category       string  `json:"category"`
	Period         string  `json:"period"`
	Limit          int64   `json:"limit"`
	AlertThreshold int     `json:"alert_threshold"`
	Used           int64   `json:"used"`
	Remaining      int64   `json:"remaining"`
	Percentage     float64 `json:"percentage"`
	Status         string  `json:"status"`
	ChangePercent  float64 `json:"change_percent"`
}

type BudgetRepository interface {
	GetAll(userID int, category, month, year string) ([]Budget, error)
	GetByID(id int) (*BudgetResponse, error)
	GetUsage(userID, month, year int) ([]BudgetUsage, error)
	Create(userID int, req CreateBudgetRequest) error
	Update(userID, id, limitAmount, alertThreshold int, category string) (*UpdateBudgetResponse, error)
	Delete(userID, id int) error
}
```

- [ ] **Step 5: Create `internal/domain/goal.go`**

```go
package domain

import "time"

type GoalResponse struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	TargetAmount  int       `json:"target_amount"`
	CurrentAmount int       `json:"current_amount"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateGoalRequest struct {
	Name         string    `json:"name" binding:"required"`
	TargetAmount int       `json:"target_amount" binding:"required"`
	Description  string    `json:"description"`
	Deadline     time.Time `json:"deadline" binding:"required"`
}

type GoalContributionRequest struct {
	GoalId       int `json:"goal_id" binding:"required"`
	Contribution int `json:"contribution" binding:"required"`
}

type GoalOverviewResponse struct {
	TotalGoals int            `json:"total_goals"`
	Savings    int            `json:"savings"`
	Goals      []GoalResponse `json:"goals"`
}

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

- [ ] **Step 6: Verify domain compiles**

```
go build ./internal/domain/...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```
git add internal/domain/
git commit -m "feat: add domain layer — entities, DTOs, repository interfaces"
```

---

## Task 3: Create repository/postgres layer

Each file implements one `domain.XRepository` interface using raw SQL. All SQL is moved verbatim from `service/`. No business logic here.

**Files:**
- Create: `internal/repository/postgres/user.go`
- Create: `internal/repository/postgres/transaction.go`
- Create: `internal/repository/postgres/budget.go`
- Create: `internal/repository/postgres/goal.go`

- [ ] **Step 1: Create `internal/repository/postgres/user.go`**

```go
package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]domain.UserResponse, error) {
	rows, err := r.db.Query("SELECT id,email,name,created_at,deleted_at,password from users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.UserResponse
	for rows.Next() {
		var u domain.UserResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.DeletedAt, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(
		"SELECT id, email, name, password FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(email, hashedPassword, name string) error {
	_, err := r.db.Exec(
		"INSERT INTO users (email, password, name) VALUES ($1, $2, $3)",
		email, hashedPassword, name,
	)
	return err
}
```

- [ ] **Step 2: Create `internal/repository/postgres/transaction.go`**

```go
package postgres

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/financial-planning/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) GetByUserID(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error) {
	query := `SELECT id, amount, category, type, "date"::date, description FROM transactions WHERE user_id=$1 AND deleted_at IS NULL`
	args := []interface{}{userID}

	if month != "" && year != "" {
		monthInt, err := strconv.Atoi(month)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid month: %v", err)
		}
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid year: %v", err)
		}
		args = append(args, monthInt, yearInt)
		query += fmt.Sprintf(` AND EXTRACT(MONTH FROM "date"::date) = $%d AND EXTRACT(YEAR FROM "date"::date) = $%d`, len(args)-1, len(args))
	}

	query += ` ORDER BY "date" DESC`

	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
		if offset > 0 {
			args = append(args, offset)
			query += fmt.Sprintf(" OFFSET $%d", len(args))
		}
	}

	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id=$1 AND deleted_at IS NULL`
	countArgs := []interface{}{userID}

	if month != "" && year != "" {
		monthInt, _ := strconv.Atoi(month)
		yearInt, _ := strconv.Atoi(year)
		countArgs = append(countArgs, monthInt, yearInt)
		countQuery += fmt.Sprintf(` AND EXTRACT(MONTH FROM "date"::date) = $%d AND EXTRACT(YEAR FROM "date"::date) = $%d`, len(countArgs)-1, len(countArgs))
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count query error: %v", err)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("main query error: %v", err)
	}
	defer rows.Close()

	var transactions []domain.TransactionResponse
	for rows.Next() {
		var t domain.TransactionResponse
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Type, &t.Date, &t.Description); err != nil {
			return nil, 0, fmt.Errorf("scan error: %v", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, totalCount, rows.Err()
}

func (r *transactionRepository) Create(userID int, req domain.TransactionRequest) error {
	_, err := r.db.Exec(
		"INSERT INTO transactions (amount, category, type, date, description, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
		req.Amount, req.Category, req.Type, req.Date, req.Description, userID,
	)
	return err
}

func (r *transactionRepository) Update(userID, id int, req domain.TransactionRequest) error {
	result, err := r.db.Exec(`
		UPDATE transactions
		SET amount = $1, category = $2, type = $3, date = $4, description = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, req.Amount, req.Category, req.Type, req.Date, req.Description, id, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *transactionRepository) Delete(userID, id int) error {
	result, err := r.db.Exec(`
		UPDATE transactions SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *transactionRepository) GetMonthlyExpenses(userID int) (float64, error) {
	now := time.Now()
	var total float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM transactions
		WHERE user_id = $1
		AND EXTRACT(MONTH FROM date) = $2
		AND EXTRACT(YEAR FROM date) = $3
		AND type = 'EXPENSE'
		AND deleted_at IS NULL
	`, userID, int(now.Month()), now.Year()).Scan(&total)
	return total, err
}

func (r *transactionRepository) GetNetSavings(userID int) (float64, error) {
	var net float64
	err := r.db.QueryRow(`
		SELECT
		COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount END), 0) -
		COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount END), 0) AS net_savings
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID).Scan(&net)
	return net, err
}
```

- [ ] **Step 3: Create `internal/repository/postgres/budget.go`**

```go
package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/financial-planning/internal/domain"
)

type budgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) GetAll(userID int, category, month, year string) ([]domain.Budget, error) {
	query := `SELECT id, user_id, category, period, month, year, limit_amount, alert_threshold, created_at FROM budgets WHERE user_id = $1 AND deleted_at IS NULL`
	args := []interface{}{userID}
	idx := 2

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", idx)
		args = append(args, category)
		idx++
	}
	if year != "" {
		query += fmt.Sprintf(" AND year = $%d", idx)
		args = append(args, year)
		idx++
	}
	if month != "" {
		query += fmt.Sprintf(" AND month = $%d", idx)
		args = append(args, month)
		idx++
	}
	_ = idx

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold, &b.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *budgetRepository) GetByID(id int) (*domain.BudgetResponse, error) {
	var b domain.BudgetResponse
	err := r.db.QueryRow(`
		SELECT id,user_id,category,period,month,year,limit_amount,alert_threshold,created_at
		FROM budgets WHERE id = $1
	`, id).Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *budgetRepository) GetUsage(userID, month, year int) ([]domain.BudgetUsage, error) {
	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear = year - 1
	}

	rows, err := r.db.Query(`
		SELECT
			b.id, b.category, b.period, b.limit_amount, b.alert_threshold,
			COALESCE(SUM(CASE
				WHEN (
					(b.period = 'MONTHLY' AND EXTRACT(MONTH FROM t.date) = $2 AND EXTRACT(YEAR FROM t.date) = $3)
					OR (b.period = 'YEARLY' AND EXTRACT(YEAR FROM t.date) = $3)
				) THEN t.amount ELSE 0
			END), 0)::BIGINT AS used,
			COALESCE(SUM(CASE
				WHEN (
					(b.period = 'MONTHLY' AND EXTRACT(MONTH FROM t.date) = $4 AND EXTRACT(YEAR FROM t.date) = $5)
					OR (b.period = 'YEARLY' AND EXTRACT(YEAR FROM t.date) = $5)
				) THEN t.amount ELSE 0
			END), 0)::BIGINT AS prev_used
		FROM budgets b
		LEFT JOIN transactions t
			ON t.user_id = b.user_id
			AND LOWER(t.category) = LOWER(b.category)
			AND t.type = 'EXPENSE'
			AND t.deleted_at IS NULL
		WHERE b.user_id = $1
		AND b.year = $3
		AND ((b.period = 'MONTHLY' AND b.month = $2) OR (b.period = 'YEARLY'))
		AND b.deleted_at IS NULL
		GROUP BY b.id, b.category, b.period, b.limit_amount, b.alert_threshold
		ORDER BY b.created_at DESC
	`, userID, month, year, prevMonth, prevYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BudgetUsage
	for rows.Next() {
		var b domain.BudgetUsage
		var prevUsed int64
		if err := rows.Scan(&b.BudgetID, &b.Category, &b.Period, &b.Limit, &b.AlertThreshold, &b.Used, &prevUsed); err != nil {
			return nil, err
		}

		b.Remaining = b.Limit - b.Used
		if b.Remaining < 0 {
			b.Remaining = 0
		}
		if b.Limit > 0 {
			b.Percentage = math.Round((float64(b.Used) / float64(b.Limit)) * 100)
		}
		if b.Percentage >= 100 {
			b.Status = "EXCEEDED"
		} else if b.Percentage >= float64(b.AlertThreshold) {
			b.Status = "WARNING"
		} else {
			b.Status = "SAFE"
		}
		if prevUsed > 0 {
			b.ChangePercent = math.Round(((float64(b.Used) - float64(prevUsed)) / float64(prevUsed)) * 100)
		} else if b.Used > 0 {
			b.ChangePercent = 100
		}

		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *budgetRepository) Create(userID int, req domain.CreateBudgetRequest) error {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM budgets
			WHERE user_id = $1 AND category = $2 AND period = $3 AND year = $4
			AND (month = $5 OR ($5 IS NULL AND month IS NULL))
		)
	`, userID, req.Category, req.Period, req.Year, req.Month).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}
	_, err = r.db.Exec(`
		INSERT INTO budgets (user_id, category, period, month, year, limit_amount, alert_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, req.Category, req.Period, req.Month, req.Year, req.LimitAmount, req.AlertThreshold)
	return err
}

func (r *budgetRepository) Update(userID, id, limitAmount, alertThreshold int, category string) (*domain.UpdateBudgetResponse, error) {
	var b domain.UpdateBudgetResponse
	err := r.db.QueryRow(`
		UPDATE budgets
		SET limit_amount = COALESCE($1, limit_amount),
		    alert_threshold = COALESCE($2, alert_threshold),
		    category = COALESCE($3, category)
		WHERE user_id = $4 AND id = $5
		RETURNING id, user_id, category, period, month, year, limit_amount, alert_threshold
	`, limitAmount, alertThreshold, category, userID, id).
		Scan(&b.ID, &b.UserID, &b.Category, &b.Period, &b.Month, &b.Year, &b.LimitAmount, &b.AlertThreshold)
	if err != nil {
		return nil, errors.New("budget not found or update failed")
	}
	return &b, nil
}

func (r *budgetRepository) Delete(userID, id int) error {
	result, err := r.db.Exec(`
		UPDATE budgets SET deleted_at = NOW()
		WHERE user_id = $1 AND id = $2 AND deleted_at IS NULL
	`, userID, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
```

- [ ] **Step 4: Create `internal/repository/postgres/goal.go`**

```go
package postgres

import (
	"database/sql"

	"github.com/financial-planning/internal/domain"
)

type goalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) domain.GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) GetAll(userID int, active bool) ([]domain.GoalResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, name, target_amount, current_amount, status, deadline, description, created_at
		FROM goals
		WHERE user_id = $1 AND ($2 = false OR deadline >= NOW())
	`, userID, active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []domain.GoalResponse
	for rows.Next() {
		var g domain.GoalResponse
		if err := rows.Scan(&g.Id, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Status, &g.Deadline, &g.Description, &g.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

func (r *goalRepository) GetByID(id, userID int) (*domain.GoalResponse, error) {
	var g domain.GoalResponse
	err := r.db.QueryRow(`
		SELECT id, name, target_amount, current_amount, description, deadline
		FROM goals WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&g.Id, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Description, &g.Deadline)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *goalRepository) GetSavingsTotal(userID int) (float64, error) {
	var total float64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(current_amount), 0) FROM goals WHERE user_id = $1`, userID,
	).Scan(&total)
	return total, err
}

func (r *goalRepository) CountActive(userID int) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(id) FROM goals WHERE user_id = $1 AND deadline >= NOW()`, userID,
	).Scan(&count)
	return count, err
}

func (r *goalRepository) GetUpcomingMilestones(userID int) ([]domain.GoalResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, target_amount, current_amount, deadline, created_at, status
		FROM goals WHERE user_id = $1 AND deadline > NOW() AND target_amount <> current_amount
		ORDER BY deadline ASC LIMIT 4
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []domain.GoalResponse
	for rows.Next() {
		var m domain.GoalResponse
		if err := rows.Scan(&m.Id, &m.Name, &m.Description, &m.TargetAmount, &m.CurrentAmount, &m.Deadline, &m.CreatedAt, &m.Status); err != nil {
			return nil, err
		}
		milestones = append(milestones, m)
	}
	return milestones, rows.Err()
}

func (r *goalRepository) Create(userID int, req domain.CreateGoalRequest) error {
	result, err := r.db.Exec(`
		INSERT INTO goals (user_id, name, target_amount, description, deadline)
		SELECT $1, $2, $3, $4, $5 WHERE NOT EXISTS (
			SELECT 1 FROM goals WHERE user_id = $6 AND name = $7 AND deadline >= NOW()
		)
	`, userID, req.Name, req.TargetAmount, req.Description, req.Deadline, userID, req.Name)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *goalRepository) Update(id, userID int, req domain.CreateGoalRequest) error {
	_, err := r.db.Exec(`
		UPDATE goals SET
			name = $1,
			target_amount = CASE WHEN target_amount > current_amount THEN $2 ELSE target_amount END,
			description = $3,
			deadline = $4,
			status = CASE WHEN target_amount > current_amount THEN 'ONGOING' ELSE status END
		WHERE id = $5 AND user_id = $6
	`, req.Name, req.TargetAmount, req.Description, req.Deadline, id, userID)
	return err
}

func (r *goalRepository) Delete(id, userID int) error {
	_, err := r.db.Exec(`DELETE FROM goals WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *goalRepository) Contribute(id, userID, amount int) error {
	_, err := r.db.Exec(`
		UPDATE goals
		SET current_amount = $1,
		    status = CASE WHEN $1 >= target_amount THEN 'COMPLETED' ELSE status END
		WHERE id = $2 AND user_id = $3
	`, amount, id, userID)
	return err
}
```

- [ ] **Step 5: Verify repository layer compiles**

```
go build ./internal/repository/...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```
git add internal/repository/
git commit -m "feat: add postgres repository implementations"
```

---

## Task 4: Create use case layer

Use case structs receive repository interfaces via constructor injection. Validation logic that was scattered across `service/` and `handler/` is consolidated here.

**Files:**
- Create: `internal/usecase/user.go`
- Create: `internal/usecase/transaction.go`
- Create: `internal/usecase/budget.go`
- Create: `internal/usecase/goal.go`

- [ ] **Step 1: Create `internal/usecase/user.go`**

```go
package usecase

import (
	"errors"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("this email is already in use")
var ErrInvalidCredentials = errors.New("invalid email or password")

type UserUseCase struct {
	repo domain.UserRepository
}

func NewUserUseCase(repo domain.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) GetAll() ([]domain.UserResponse, error) {
	return uc.repo.GetAll()
}

func (uc *UserUseCase) Register(email, password, name string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	if err := uc.repo.Create(email, string(hash), name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (uc *UserUseCase) Login(email, password string) (string, error) {
	user, err := uc.repo.FindByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return utils.GenerateJwt(user.ID, user.Name, user.Email)
}
```

- [ ] **Step 2: Create `internal/usecase/transaction.go`**

```go
package usecase

import (
	"errors"
	"strings"

	"github.com/financial-planning/internal/domain"
)

type TransactionUseCase struct {
	repo domain.TransactionRepository
}

func NewTransactionUseCase(repo domain.TransactionRepository) *TransactionUseCase {
	return &TransactionUseCase{repo: repo}
}

func (uc *TransactionUseCase) GetTransactions(userID, limit, offset int, year, month string) ([]domain.TransactionResponse, int, error) {
	return uc.repo.GetByUserID(userID, limit, offset, year, month)
}

func (uc *TransactionUseCase) Create(userID int, req domain.TransactionRequest) error {
	req.Type = strings.ToUpper(req.Type)
	if req.Type != "INCOME" && req.Type != "EXPENSE" {
		return errors.New("invalid transaction type: must be INCOME or EXPENSE")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Date.IsZero() {
		return errors.New("date is required")
	}
	return uc.repo.Create(userID, req)
}

func (uc *TransactionUseCase) Update(userID, id int, req domain.TransactionRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Type != "INCOME" && req.Type != "EXPENSE" {
		return errors.New("invalid type")
	}
	if req.Date.IsZero() {
		return errors.New("date is required")
	}
	return uc.repo.Update(userID, id, req)
}

func (uc *TransactionUseCase) Delete(userID, id int) error {
	return uc.repo.Delete(userID, id)
}

func (uc *TransactionUseCase) GetMonthlyExpenses(userID int) (float64, error) {
	return uc.repo.GetMonthlyExpenses(userID)
}
```

- [ ] **Step 3: Create `internal/usecase/budget.go`**

```go
package usecase

import (
	"errors"

	"github.com/financial-planning/internal/domain"
)

type BudgetUseCase struct {
	repo domain.BudgetRepository
}

func NewBudgetUseCase(repo domain.BudgetRepository) *BudgetUseCase {
	return &BudgetUseCase{repo: repo}
}

func (uc *BudgetUseCase) GetBudgets(userID int, category, month, year string) ([]domain.Budget, error) {
	return uc.repo.GetAll(userID, category, month, year)
}

func (uc *BudgetUseCase) GetByID(id int) (*domain.BudgetResponse, error) {
	return uc.repo.GetByID(id)
}

func (uc *BudgetUseCase) GetUsage(userID, month, year int) ([]domain.BudgetUsage, error) {
	return uc.repo.GetUsage(userID, month, year)
}

func (uc *BudgetUseCase) Create(userID int, req domain.CreateBudgetRequest) error {
	if req.Category == "" {
		return errors.New("category is required")
	}
	if req.Period != "MONTHLY" && req.Period != "YEARLY" {
		return errors.New("invalid period")
	}
	if req.Year == 0 {
		return errors.New("year is required")
	}
	if req.LimitAmount <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if req.Period == "MONTHLY" && req.Month == nil {
		return errors.New("month required for monthly budget")
	}
	if req.Period == "YEARLY" {
		req.Month = nil
	}
	if req.AlertThreshold == 0 {
		req.AlertThreshold = 80
	}
	return uc.repo.Create(userID, req)
}

func (uc *BudgetUseCase) Update(userID, id, limitAmount, alertThreshold int, category string) (*domain.UpdateBudgetResponse, error) {
	return uc.repo.Update(userID, id, limitAmount, alertThreshold, category)
}

func (uc *BudgetUseCase) Delete(userID, id int) error {
	return uc.repo.Delete(userID, id)
}
```

- [ ] **Step 4: Create `internal/usecase/goal.go`**

`GoalUseCase` takes both `GoalRepository` and `TransactionRepository`. This fixes the current cross-domain SQL leakage where `service/goal.go:GoalContributions` queried the transactions table directly via the goal repo's DB.

```go
package usecase

import (
	"errors"
	"time"

	"github.com/financial-planning/internal/domain"
)

type GoalUseCase struct {
	repo   domain.GoalRepository
	txRepo domain.TransactionRepository
}

func NewGoalUseCase(repo domain.GoalRepository, txRepo domain.TransactionRepository) *GoalUseCase {
	return &GoalUseCase{repo: repo, txRepo: txRepo}
}

func (uc *GoalUseCase) GetGoals(userID int, active bool) ([]domain.GoalResponse, error) {
	return uc.repo.GetAll(userID, active)
}

func (uc *GoalUseCase) GetByID(id, userID int) (*domain.GoalResponse, error) {
	return uc.repo.GetByID(id, userID)
}

func (uc *GoalUseCase) GetOverview(userID int) (*domain.GoalOverviewResponse, error) {
	savings, err := uc.repo.GetSavingsTotal(userID)
	if err != nil {
		return nil, err
	}
	milestones, err := uc.repo.GetUpcomingMilestones(userID)
	if err != nil {
		return nil, err
	}
	total, err := uc.repo.CountActive(userID)
	if err != nil {
		return nil, err
	}
	return &domain.GoalOverviewResponse{
		TotalGoals: total,
		Goals:      milestones,
		Savings:    int(savings),
	}, nil
}

func (uc *GoalUseCase) Create(userID int, req domain.CreateGoalRequest) error {
	if req.TargetAmount <= 0 {
		return errors.New("target amount must be greater than 0")
	}
	if req.Deadline.Before(time.Now()) {
		return errors.New("deadline must be in the future")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	return uc.repo.Create(userID, req)
}

func (uc *GoalUseCase) Update(id, userID int, req domain.CreateGoalRequest) error {
	return uc.repo.Update(id, userID, req)
}

func (uc *GoalUseCase) Delete(id, userID int) error {
	return uc.repo.Delete(id, userID)
}

func (uc *GoalUseCase) Contribute(id, userID, amount int) error {
	if amount <= 0 {
		return errors.New("contribution must be greater than 0")
	}
	net, err := uc.txRepo.GetNetSavings(userID)
	if err != nil {
		return err
	}
	if net <= 0 {
		return errors.New("cannot add contributions: no net savings")
	}
	if amount > int(net) {
		return errors.New("contribution exceeds available savings")
	}
	return uc.repo.Contribute(id, userID, amount)
}
```

- [ ] **Step 5: Verify use case layer compiles**

```
go build ./internal/usecase/...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```
git add internal/usecase/
git commit -m "feat: add use case layer with consolidated validation"
```

---

## Task 5: Create delivery/http layer

Handlers are struct-based (not closure-based like the old code). All handler methods map domain errors to HTTP status codes consistently.

**Files:**
- Create: `internal/delivery/http/middleware/auth.go`
- Create: `internal/delivery/http/handler/user.go`
- Create: `internal/delivery/http/handler/transaction.go`
- Create: `internal/delivery/http/handler/budget.go`
- Create: `internal/delivery/http/handler/goal.go`
- Create: `internal/delivery/http/router.go`

- [ ] **Step 1: Create `internal/delivery/http/middleware/auth.go`**

Logic is identical to `middleware/middleware.go` but uses `utils.MyCustomClaims` instead of `model.MyCustomClaims`.

```go
package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(400, gin.H{"message": "Missing Authorization!", "code": 400})
			c.Abort()
			return
		}

		arr := strings.Split(authHeader, " ")
		if len(arr) != 2 {
			c.JSON(401, gin.H{"error": "authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		claims := utils.MyCustomClaims{}
		token, err := jwt.ParseWithClaims(arr[1], &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_KEY")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
```

- [ ] **Step 2: Create `internal/delivery/http/handler/user.go`**

```go
package handler

import (
	"errors"

	"github.com/financial-planning/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	uc *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if err := h.uc.Register(req.Email, req.Password, req.Name); err != nil {
		if errors.Is(err, usecase.ErrUserExists) {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(500, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "User created successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	token, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(401, gin.H{"error": "Login failed: " + err.Error()})
		return
	}
	c.SetCookie("accessToken", token, 3600, "/", "*", false, false)
	c.JSON(200, gin.H{"message": "Login Successfully", "status": "200", "data": gin.H{"token": token}})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.uc.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": users, "status": "200", "message": "Get all users successfully"})
}
```

- [ ] **Step 3: Create `internal/delivery/http/handler/transaction.go`**

```go
package handler

import (
	"errors"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	uc *usecase.TransactionUseCase
}

func NewTransactionHandler(uc *usecase.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}
	transactions, total, err := h.uc.GetTransactions(userID, limit, offset, c.Query("year"), c.Query("month"))
	if err != nil {
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": transactions, "total": total})
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if err := h.uc.Create(userID, req); err != nil {
		c.JSON(400, gin.H{"error": "Failed to create transaction: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Transaction created successfully"})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction id"})
		return
	}
	var req domain.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.uc.Update(userID, id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction updated successfully"})
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction id"})
		return
	}
	if err := h.uc.Delete(userID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction deleted successfully"})
}

func (h *TransactionHandler) GetMonthlyExpenses(c *gin.Context) {
	userID := utils.ClaimId(c)
	total, err := h.uc.GetMonthlyExpenses(userID)
	if err != nil {
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"total": total, "message": "success"})
}
```

- [ ] **Step 4: Create `internal/delivery/http/handler/budget.go`**

```go
package handler

import (
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type BudgetHandler struct {
	uc *usecase.BudgetUseCase
}

func NewBudgetHandler(uc *usecase.BudgetUseCase) *BudgetHandler {
	return &BudgetHandler{uc: uc}
}

func (h *BudgetHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	budgets, err := h.uc.GetBudgets(userID, c.Query("category"), c.Query("month"), c.Query("year"))
	if err != nil {
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": budgets})
}

func (h *BudgetHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid budget id"})
		return
	}
	budget, err := h.uc.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"message": "Something went wrong", "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": budget})
}

func (h *BudgetHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.uc.Create(userID, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "budget created successfully"})
}

func (h *BudgetHandler) GetUsage(c *gin.Context) {
	userID := utils.ClaimId(c)
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid year"})
		return
	}
	var month int
	if monthStr := c.Query("month"); monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid month"})
			return
		}
	}
	result, err := h.uc.GetUsage(userID, month, year)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *BudgetHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid budget id"})
		return
	}
	var req domain.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	response, err := h.uc.Update(userID, id, req.LimitAmount, req.AlertThreshold, req.Category)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": response})
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid budget id"})
		return
	}
	if err := h.uc.Delete(userID, id); err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "budget deleted successfully"})
}
```

- [ ] **Step 5: Create `internal/delivery/http/handler/goal.go`**

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	uc *usecase.GoalUseCase
}

func NewGoalHandler(uc *usecase.GoalUseCase) *GoalHandler {
	return &GoalHandler{uc: uc}
}

func (h *GoalHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	goals, err := h.uc.GetGoals(userID, c.DefaultQuery("active", "false") == "true")
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to retrieve goals: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": goals})
}

func (h *GoalHandler) GetByID(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}
	goal, err := h.uc.GetByID(id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": goal})
}

func (h *GoalHandler) GetOverview(c *gin.Context) {
	userID := utils.ClaimId(c)
	overview, err := h.uc.GetOverview(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to retrieve goal overview: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": overview})
}

func (h *GoalHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if err := h.uc.Create(userID, req); err != nil {
		c.JSON(400, gin.H{"error": "Failed to create goal: " + err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "Goal created successfully"})
}

func (h *GoalHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal id"})
		return
	}
	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.Update(id, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Goal updated successfully"})
}

func (h *GoalHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}
	if err := h.uc.Delete(id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Goal deleted successfully"})
}

func (h *GoalHandler) Contribute(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.GoalContributionRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Contribution <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contribution must be greater than 0"})
		return
	}
	if req.GoalId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}
	if err := h.uc.Contribute(req.GoalId, userID, req.Contribution); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contribution successful"})
}
```

- [ ] **Step 6: Create `internal/delivery/http/router.go`**

Route paths are preserved exactly from the old `routes/` package. Note: Gin matches static routes before wildcard routes, so `/overview` and `/contribute` register correctly before `/:id`.

```go
package delivery

import (
	"github.com/financial-planning/internal/delivery/http/handler"
	"github.com/financial-planning/internal/delivery/http/middleware"
	"github.com/financial-planning/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserUC        *usecase.UserUseCase
	TransactionUC *usecase.TransactionUseCase
	BudgetUC      *usecase.BudgetUseCase
	GoalUC        *usecase.GoalUseCase
}

func Setup(r *gin.Engine, deps Deps) {
	userH := handler.NewUserHandler(deps.UserUC)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", userH.Register)
		v1.POST("/login", userH.Login)
		v1.GET("/users", userH.GetAll)
	}

	auth := r.Group("/api/auth/v1")
	auth.Use(middleware.AuthRequired())
	{
		txH := handler.NewTransactionHandler(deps.TransactionUC)
		auth.GET("/transactions/users", txH.GetAll)
		auth.POST("/transactions/", txH.Create)
		auth.PUT("/transactions/:id", txH.Update)
		auth.DELETE("/transactions/:id", txH.Delete)
		auth.GET("/transactions/count", txH.GetMonthlyExpenses)

		bH := handler.NewBudgetHandler(deps.BudgetUC)
		auth.GET("/budgets/", bH.GetAll)
		auth.GET("/budgets/usage", bH.GetUsage)
		auth.GET("/budgets/:id", bH.GetByID)
		auth.POST("/budgets/", bH.Create)
		auth.PATCH("/budgets/:id", bH.Update)
		auth.DELETE("/budgets/:id", bH.Delete)

		gH := handler.NewGoalHandler(deps.GoalUC)
		auth.GET("/goals/", gH.GetAll)
		auth.GET("/goals/overview", gH.GetOverview)
		auth.PATCH("/goals/contribute", gH.Contribute)
		auth.GET("/goals/:id", gH.GetByID)
		auth.POST("/goals/", gH.Create)
		auth.PATCH("/goals/:id", gH.Update)
		auth.DELETE("/goals/:id", gH.Delete)
	}
}
```

- [ ] **Step 7: Verify delivery layer compiles**

```
go build ./internal/...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```
git add internal/delivery/
git commit -m "feat: add delivery layer — handlers, middleware, router"
```

---

## Task 6: Rewrite main.go

`main.go` becomes thin: init DB, wire repos → usecases → Setup → Run. The `delivery` package alias resolves the import path `internal/delivery/http` to the package name `delivery` (declared in `router.go`).

**Files:**
- Rewrite: `main.go`

- [ ] **Step 1: Rewrite `main.go`**

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	delivery "github.com/financial-planning/internal/delivery/http"
	"github.com/financial-planning/internal/repository/postgres"
	"github.com/financial-planning/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func initDB(user, password, name, host, port string) *sql.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")
	return db
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	db := initDB(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)
	defer db.Close()

	userRepo   := postgres.NewUserRepository(db)
	txRepo     := postgres.NewTransactionRepository(db)
	budgetRepo := postgres.NewBudgetRepository(db)
	goalRepo   := postgres.NewGoalRepository(db)

	userUC   := usecase.NewUserUseCase(userRepo)
	txUC     := usecase.NewTransactionUseCase(txRepo)
	budgetUC := usecase.NewBudgetUseCase(budgetRepo)
	goalUC   := usecase.NewGoalUseCase(goalRepo, txRepo)

	r := gin.Default()
	r.Use(corsMiddleware())
	delivery.Setup(r, delivery.Deps{
		UserUC:        userUC,
		TransactionUC: txUC,
		BudgetUC:      budgetUC,
		GoalUC:        goalUC,
	})

	fmt.Println("Server is running on port 8080")
	r.Run()
}
```

- [ ] **Step 2: Commit**

```
git add main.go
git commit -m "refactor(main): wire clean architecture layers"
```

---

## Task 7: Delete old packages and final verification

With `main.go` no longer importing old packages, they can be safely deleted.

**Files deleted:**
- `model/` (entire directory)
- `service/` (entire directory)
- `handler/` (entire directory)
- `controller/` (entire directory)
- `routes/` (entire directory)
- `repository/repository.go`
- `middleware/middleware.go`

- [ ] **Step 1: Delete old packages**

```
git rm -r model/ service/ handler/ controller/ routes/ middleware/
git rm repository/repository.go
```

- [ ] **Step 2: Final build**

```
go build .
```

Expected: no errors. Binary produced at `main.exe` (Windows) or `main` (Linux/Mac).

- [ ] **Step 3: Run the server briefly to confirm startup**

```
go run main.go
```

Expected output contains:
```
Successfully connected to PostgreSQL!
Server is running on port 8080
```

Ctrl+C to stop.

- [ ] **Step 4: Commit**

```
git add -A
git commit -m "chore: delete old model/service/handler/routes/controller/middleware packages"
```

---

## Self-Review Checklist (completed inline)

**Spec coverage:**
- ✅ By-layer folder structure under `internal/` (Tasks 2–5)
- ✅ `domain/` has no framework imports (Task 2)
- ✅ `usecase/` imports only `domain/` (Task 4)
- ✅ `repository/postgres/` implements domain interfaces (Task 3)
- ✅ `delivery/http/` imports `usecase/` and `domain/` only (Task 5)
- ✅ `GoalUseCase` takes `TransactionRepository` to fix cross-domain SQL leakage (Task 4, Step 4)
- ✅ `GetAllBudgetUsage` JOIN kept in repository layer as read-model optimization (Task 3, Step 3)
- ✅ `MyCustomClaims` moved from `model/` to `utils/` (Task 1)
- ✅ Sentinel errors in `domain/errors.go` (Task 2, Step 1)
- ✅ Handlers map domain errors to HTTP status codes (Tasks 5)
- ✅ Route paths preserved exactly (Task 5, Step 6)
- ✅ Old packages deleted (Task 7)
- ✅ `controller/` (duplicate dead code) deleted (Task 7)

**Type consistency check:**
- `domain.GoalRepository.Contribute(id, userID, amount int)` → called as `uc.repo.Contribute(id, userID, amount)` in `usecase/goal.go` ✅
- `postgres.NewGoalRepository` returns `domain.GoalRepository` → assigned to `goalRepo domain.GoalRepository` in main ✅
- `delivery.Setup` takes `delivery.Deps` with `*usecase.GoalUseCase` → `NewGoalUseCase` returns `*GoalUseCase` ✅
- `handler.GoalHandler.Contribute` calls `h.uc.Contribute(req.GoalId, userID, req.Contribution)` → matches `GoalUseCase.Contribute(id, userID, amount int)` ✅
