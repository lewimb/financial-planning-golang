package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/generators"
	"github.com/kristijorgji/goseeder"
)

func SeedTransactions(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed transactions: no users found — run SeedUsers first")
	}

	total := 0
	for i, uid := range userIDs {
		txs := generators.Generate(i)
		for _, tx := range txs {
			var recInterval interface{}
			if tx.RecurrenceInterval != "" {
				recInterval = tx.RecurrenceInterval
			}
			if _, err := s.DB.Exec(
				`INSERT INTO transactions (user_id, amount, category, type, date, description, is_recurring, recurrence_interval)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				uid, tx.Amount, tx.Category, tx.Type, tx.Date, tx.Description, tx.IsRecurring, recInterval,
			); err != nil {
				log.Fatalf("seed transactions: user %d: %v", uid, err)
			}
		}
		fmt.Printf("  user %d (%s): %d transactions\n", uid, generators.UserNames[i], len(txs))
		total += len(txs)
	}
	fmt.Printf("  seeded %d transactions total\n", total)
}
