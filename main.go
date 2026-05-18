package main

import (
	"context"
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

	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not loaded: %v", err)
	}

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
	userRepo := postgres.NewUserRepository(db)
	txRepo := postgres.NewTransactionRepository(db)
	budgetRepo := postgres.NewBudgetRepository(db)
	goalRepo := postgres.NewGoalRepository(db)
	aiLogRepo := postgres.NewAiLogRepository(db)

	// use cases
	userUC := usecase.NewUserUseCase(userRepo)
	txUC := usecase.NewTransactionUseCase(txRepo)
	budgetUC := usecase.NewBudgetUseCase(budgetRepo)
	goalUC := usecase.NewGoalUseCase(goalRepo, txRepo)
	dashboardUC := usecase.NewDashboardUseCase(txRepo, budgetRepo, goalRepo)
	chatUC := usecase.NewChatUseCase(txRepo, budgetRepo, goalRepo, aiLogRepo, gemini)

	// delivery
	r := gin.Default()
	r.Use(corsMiddleware())
	delivery.Setup(r, delivery.Deps{
		UserUC:        userUC,
		TransactionUC: txUC,
		BudgetUC:      budgetUC,
		GoalUC:        goalUC,
		DashboardUC:   dashboardUC,
		ChatUC:        chatUC,
	})

	fmt.Println("Server is running on port 8080")
	r.Run()
}
