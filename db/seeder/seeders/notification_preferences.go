package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedNotificationPreferences(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed notification_preferences: no users found — run SeedUsers first")
	}

	for i, uid := range userIDs {
		if i >= len(factories.NotificationPrefs) {
			break
		}
		p := factories.NotificationPrefs[i]
		if _, err := s.DB.Exec(`
			INSERT INTO notification_preferences
				(user_id, budget_alerts, goal_reminders, anomaly_alerts, weekly_summary, push_enabled, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				budget_alerts  = EXCLUDED.budget_alerts,
				goal_reminders = EXCLUDED.goal_reminders,
				anomaly_alerts = EXCLUDED.anomaly_alerts,
				weekly_summary = EXCLUDED.weekly_summary,
				push_enabled   = EXCLUDED.push_enabled,
				updated_at     = NOW()`,
			uid, p.BudgetAlerts, p.GoalReminders, p.AnomalyAlerts, p.WeeklySummary, p.PushEnabled,
		); err != nil {
			log.Fatalf("seed notification_preferences: user %d: %v", uid, err)
		}
	}
	fmt.Printf("  seeded %d notification preference rows\n", min(len(userIDs), len(factories.NotificationPrefs)))
}
