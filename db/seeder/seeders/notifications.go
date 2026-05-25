package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedNotifications(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed notifications: no users found — run SeedUsers first")
	}

	total := 0
	for i, uid := range userIDs {
		if i >= len(factories.Notifications) {
			break
		}
		for _, n := range factories.Notifications[i] {
			if _, err := s.DB.Exec(`
				INSERT INTO notifications (user_id, type, title, message, entity_type, entity_id, is_read)
				VALUES ($1,$2,$3,$4,$5,NULL,$6)`,
				uid, n.Type, n.Title, n.Message, n.EntityType, n.IsRead,
			); err != nil {
				log.Fatalf("seed notifications: user %d: %v", uid, err)
			}
			total++
		}
	}
	fmt.Printf("  seeded %d notifications\n", total)
}
