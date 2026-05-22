package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() *sql.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("seeder: open db: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("seeder: ping db: %v", err)
	}
	fmt.Println("Connected to PostgreSQL")
	return db
}

// Truncate clears all seeded tables in FK-safe order and resets sequences.
func Truncate(db *sql.DB) {
	tables := []string{
		"ai_logs",
		"user_financial_goals",
		"user_financial_profiles",
		"goals",
		"budgets",
		"transactions",
		"users",
	}
	fmt.Println("Truncating tables...")
	for _, t := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", t)); err != nil {
			log.Fatalf("seeder: truncate %s: %v", t, err)
		}
		fmt.Printf("  cleared: %s\n", t)
	}
}

// GetUserIDs returns all non-deleted user IDs ordered by id.
func GetUserIDs(db *sql.DB) []int {
	rows, err := db.Query("SELECT id FROM users WHERE deleted_at IS NULL ORDER BY id")
	if err != nil {
		log.Fatalf("seeder: get user ids: %v", err)
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("seeder: scan user id: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}
