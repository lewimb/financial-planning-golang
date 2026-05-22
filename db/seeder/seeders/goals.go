package seeders

import (
	"fmt"
	"log"
	"time"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedGoals(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed goals: no users found — run SeedUsers first")
	}

	total := 0
	for i, uid := range userIDs {
		if i >= len(factories.Goals) {
			break
		}
		for _, g := range factories.Goals[i] {
			deadline, err := time.Parse("2006-01-02", g.Deadline)
			if err != nil {
				log.Fatalf("seed goals: parse deadline %q: %v", g.Deadline, err)
			}
			if _, err = s.DB.Exec(`
				INSERT INTO goals (user_id, name, target_amount, current_amount, status, deadline, description)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				uid, g.Name, g.TargetAmount, g.CurrentAmount, g.Status, deadline, g.Description,
			); err != nil {
				log.Fatalf("seed goals: user %d goal %q: %v", uid, g.Name, err)
			}
			total++
		}
	}
	fmt.Printf("  seeded %d goals\n", total)
}
