package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedAiLogs(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed ai_logs: no users found — run SeedUsers first")
	}

	total := 0
	for _, uid := range userIDs {
		for _, entry := range factories.AiLogs {
			if _, err := s.DB.Exec(
				`INSERT INTO ai_logs (user_id, question, response) VALUES ($1,$2,$3)`,
				uid, entry.Question, entry.Response,
			); err != nil {
				log.Fatalf("seed ai_logs: user %d: %v", uid, err)
			}
			total++
		}
	}
	fmt.Printf("  seeded %d AI log entries\n", total)
}
