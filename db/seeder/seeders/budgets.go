package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedBudgets(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed budgets: no users found — run SeedUsers first")
	}

	total := 0
	for i, uid := range userIDs {
		if i >= len(factories.Budgets) {
			break
		}
		for _, b := range factories.Budgets[i] {
			var err error
			if b.Period == "MONTHLY" {
				// Unique constraint fires on (user_id, category, period, month, year)
				_, err = s.DB.Exec(`
					INSERT INTO budgets (user_id, category, period, month, year, limit_amount, alert_threshold)
					VALUES ($1,$2,$3,$4,$5,$6,$7)
					ON CONFLICT (user_id, category, period, month, year) DO NOTHING`,
					uid, b.Category, b.Period, b.Month, b.Year, b.LimitAmount, b.AlertThreshold,
				)
			} else {
				// YEARLY: month is NULL — PostgreSQL unique constraint treats NULLs as distinct,
				// so we guard with WHERE NOT EXISTS to prevent duplicates on re-run.
				_, err = s.DB.Exec(`
					INSERT INTO budgets (user_id, category, period, month, year, limit_amount, alert_threshold)
					SELECT $1,$2,$3,NULL,$4,$5,$6
					WHERE NOT EXISTS (
						SELECT 1 FROM budgets
						WHERE user_id=$1 AND category=$2 AND period='YEARLY' AND month IS NULL AND year=$4 AND deleted_at IS NULL
					)`,
					uid, b.Category, b.Period, b.Year, b.LimitAmount, b.AlertThreshold,
				)
			}
			if err != nil {
				log.Fatalf("seed budgets: user %d category %s: %v", uid, b.Category, err)
			}
			total++
		}
	}
	fmt.Printf("  seeded %d budgets\n", total)
}
