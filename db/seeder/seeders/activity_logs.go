package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedActivityLogs(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed activity_logs: no users found — run SeedUsers first")
	}

	total := 0
	for i, uid := range userIDs {
		if i >= len(factories.ActivityLogs) {
			break
		}
		for _, l := range factories.ActivityLogs[i] {
			if _, err := s.DB.Exec(`
				INSERT INTO activity_logs (user_id, action, entity_type, entity_id, description)
				VALUES ($1,$2,$3,NULL,$4)`,
				uid, l.Action, l.EntityType, l.Description,
			); err != nil {
				log.Fatalf("seed activity_logs: user %d: %v", uid, err)
			}
			total++
		}
	}
	fmt.Printf("  seeded %d activity log entries\n", total)
}
