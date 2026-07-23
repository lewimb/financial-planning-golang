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
				// Recurring — one per user+category, no month/year stored.
				_, err = s.DB.Exec(`
					INSERT INTO budgets (user_id, category, period, limit_amount, alert_threshold)
					SELECT $1::int,$2::varchar,$3::varchar,$4::int,$5::int
					WHERE NOT EXISTS (
						SELECT 1 FROM budgets
						WHERE user_id=$1 AND category=$2 AND period='MONTHLY' AND deleted_at IS NULL
					)`,
					uid, b.Category, b.Period, b.LimitAmount, b.AlertThreshold,
				)
			} else {
				// YEARLY: scoped to b.Year. PostgreSQL unique constraints treat
				// NULLs as distinct, and we guard with WHERE NOT EXISTS anyway
				// to keep the seeder idempotent on re-run.
				_, err = s.DB.Exec(`
					INSERT INTO budgets (user_id, category, period, year, limit_amount, alert_threshold)
					SELECT $1::int,$2::varchar,$3::varchar,$4::int,$5::int,$6::int
					WHERE NOT EXISTS (
						SELECT 1 FROM budgets
						WHERE user_id=$1 AND category=$2 AND period='YEARLY' AND year=$4 AND deleted_at IS NULL
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
