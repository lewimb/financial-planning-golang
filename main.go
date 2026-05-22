package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	delivery "github.com/financial-planning/internal/delivery/http"
	"github.com/financial-planning/internal/ml"
	"github.com/financial-planning/internal/repository/postgres"
	"github.com/financial-planning/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func initDB(dbUser, dbPassword, dbName, dbHost, dbPort string) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")
	return db, nil
}

func skipOptionsLogger() gin.HandlerFunc {
	logger := gin.Logger()
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		logger(c)
	}
}

func corsMiddleware() gin.HandlerFunc {
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
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
		log.Printf("warning: .env file not loaded: %v", err)
	}

	ctx := context.Background()

	gemini, err := usecase.NewGeminiClient(ctx)
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	db, err := initDB(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)
	if err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer db.Close()

	// repositories
	userRepo        := postgres.NewUserRepository(db)
	txRepo          := postgres.NewTransactionRepository(db)
	budgetRepo      := postgres.NewBudgetRepository(db)
	goalRepo        := postgres.NewGoalRepository(db)
	aiLogRepo       := postgres.NewAiLogRepository(db)
	profileRepo     := postgres.NewFinancialProfileRepository(db)
	notifRepo       := postgres.NewNotificationRepository(db)
	activityRepo    := postgres.NewActivityLogRepository(db)

	// use cases
	userUC         := usecase.NewUserUseCase(userRepo)
	txUC           := usecase.NewTransactionUseCase(txRepo, activityRepo)
	budgetUC       := usecase.NewBudgetUseCase(budgetRepo)
	goalUC         := usecase.NewGoalUseCase(goalRepo, txRepo)
	dashboardUC    := usecase.NewDashboardUseCase(txRepo, budgetRepo, goalRepo)
	chatUC         := usecase.NewChatUseCase(txRepo, budgetRepo, goalRepo, aiLogRepo, profileRepo, gemini)
	mlClient       := ml.NewClient()
	mlUC           := usecase.NewMLUseCase(txRepo, mlClient)
	profileUC      := usecase.NewFinancialProfileUseCase(profileRepo)
	notifUC        := usecase.NewNotificationUseCase(notifRepo, budgetRepo)
	activityLogUC  := usecase.NewActivityLogUseCase(activityRepo)
	reportsUC      := usecase.NewReportsUseCase(txRepo, budgetRepo, goalRepo)

	// delivery
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(skipOptionsLogger())
	delivery.Setup(r, delivery.Deps{
		UserUC:             userUC,
		TransactionUC:      txUC,
		BudgetUC:           budgetUC,
		GoalUC:             goalUC,
		DashboardUC:        dashboardUC,
		ChatUC:             chatUC,
		MLUC:               mlUC,
		FinancialProfileUC: profileUC,
		NotificationUC:     notifUC,
		ActivityLogUC:      activityLogUC,
		ReportsUC:          reportsUC,
	})

	fmt.Println("Server is running on port 8080")
	r.Run()
}
