package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/config"
	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
)

func SeedFinancialProfiles(s goseeder.Seeder) {
	userIDs := config.GetUserIDs(s.DB)
	if len(userIDs) == 0 {
		log.Fatal("seed financial_profiles: no users found — run SeedUsers first")
	}

	for i, uid := range userIDs {
		if i >= len(factories.Profiles) {
			break
		}
		p := factories.Profiles[i]

		_, err := s.DB.Exec(`
			INSERT INTO user_financial_profiles
				(user_id, monthly_income, fixed_expenses, current_savings, debt,
				 employment_status, spending_habit, risk_level, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				monthly_income    = EXCLUDED.monthly_income,
				fixed_expenses    = EXCLUDED.fixed_expenses,
				current_savings   = EXCLUDED.current_savings,
				debt              = EXCLUDED.debt,
				employment_status = EXCLUDED.employment_status,
				spending_habit    = EXCLUDED.spending_habit,
				risk_level        = EXCLUDED.risk_level,
				updated_at        = NOW()`,
			uid, p.MonthlyIncome, p.FixedExpenses, p.CurrentSavings,
			p.Debt, p.EmploymentStatus, p.SpendingHabit, p.RiskLevel,
		)
		if err != nil {
			log.Fatalf("seed financial_profiles: upsert user %d: %v", uid, err)
		}

		// Replace goal tags
		if _, err = s.DB.Exec(`DELETE FROM user_financial_goals WHERE user_id = $1`, uid); err != nil {
			log.Fatalf("seed financial_goals: delete user %d: %v", uid, err)
		}
		for _, goal := range p.Goals {
			if _, err = s.DB.Exec(
				`INSERT INTO user_financial_goals (user_id, goal_type)
				 VALUES ($1,$2) ON CONFLICT (user_id, goal_type) DO NOTHING`,
				uid, goal,
			); err != nil {
				log.Fatalf("seed financial_goals: insert user %d goal %s: %v", uid, goal, err)
			}
		}
	}
	fmt.Printf("  seeded %d financial profiles\n", min(len(userIDs), len(factories.Profiles)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
