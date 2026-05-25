package usecase

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/financial-planning/internal/domain"
)

type InsightsUseCase struct {
	txRepo     domain.TransactionRepository
	budgetRepo domain.BudgetRepository
	goalRepo   domain.GoalRepository
}

func NewInsightsUseCase(
	txRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
	goalRepo domain.GoalRepository,
) *InsightsUseCase {
	return &InsightsUseCase{txRepo: txRepo, budgetRepo: budgetRepo, goalRepo: goalRepo}
}

// --- Financial Health ---

type FinancialHealthComponents struct {
	SavingsRate     float64 `json:"savings_rate"`
	BudgetAdherence float64 `json:"budget_adherence"`
	GoalProgress    float64 `json:"goal_progress"`
}

type FinancialHealthResponse struct {
	Score          int                       `json:"score"`
	Rating         string                   `json:"rating"`
	Components     FinancialHealthComponents `json:"components"`
	Trend          string                   `json:"trend"`
	LastCalculated time.Time                `json:"last_calculated"`
}

func (uc *InsightsUseCase) GetFinancialHealth(userID int) (*FinancialHealthResponse, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	summary, err := uc.txRepo.GetMonthlySummary(userID, 1)
	if err != nil {
		log.Printf("insights: GetFinancialHealth tx userID=%d: %v", userID, err)
		return nil, err
	}

	var savingsRate float64
	if len(summary) > 0 {
		s := summary[len(summary)-1]
		if s.Income > 0 {
			savingsRate = (s.Income - s.Expense) / s.Income
		}
	}

	usage, err := uc.budgetRepo.GetUsage(userID, month, year)
	if err != nil {
		log.Printf("insights: GetFinancialHealth budget userID=%d: %v", userID, err)
		return nil, err
	}
	budgetAdherence := 1.0
	if len(usage) > 0 {
		safe := 0
		for _, u := range usage {
			if u.Status != "EXCEEDED" {
				safe++
			}
		}
		budgetAdherence = float64(safe) / float64(len(usage))
	}

	goals, err := uc.goalRepo.GetAll(userID, false)
	if err != nil {
		log.Printf("insights: GetFinancialHealth goals userID=%d: %v", userID, err)
		return nil, err
	}
	goalProgress := 1.0
	if len(goals) > 0 {
		var total float64
		for _, g := range goals {
			if g.TargetAmount > 0 {
				p := float64(g.CurrentAmount) / float64(g.TargetAmount)
				if p > 1 {
					p = 1
				}
				total += p
			}
		}
		goalProgress = total / float64(len(goals))
	}

	savingsScore := math.Min(savingsRate/0.20*40, 40)
	budgetScore := budgetAdherence * 30
	goalScore := goalProgress * 30
	score := int(math.Round(savingsScore + budgetScore + goalScore))

	rating := scoreRating(score)

	return &FinancialHealthResponse{
		Score:  score,
		Rating: rating,
		Components: FinancialHealthComponents{
			SavingsRate:     math.Round(savingsRate*1000) / 1000,
			BudgetAdherence: math.Round(budgetAdherence*1000) / 1000,
			GoalProgress:    math.Round(goalProgress*1000) / 1000,
		},
		Trend:          "stable",
		LastCalculated: time.Now().UTC().Truncate(24 * time.Hour),
	}, nil
}

func scoreRating(score int) string {
	switch {
	case score <= 40:
		return "Poor"
	case score <= 60:
		return "Fair"
	case score <= 80:
		return "Good"
	default:
		return "Excellent"
	}
}

// --- Insights ---

type InsightItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type InsightsPeriod struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

type InsightsResponse struct {
	Insights    []InsightItem  `json:"insights"`
	Period      InsightsPeriod `json:"period"`
	GeneratedAt time.Time      `json:"generated_at"`
}

func (uc *InsightsUseCase) GetInsights(userID, month, year int) (*InsightsResponse, error) {
	var insights []InsightItem

	goals, err := uc.goalRepo.GetAll(userID, false)
	if err != nil {
		log.Printf("insights: GetInsights goals userID=%d: %v", userID, err)
		return nil, err
	}
	if len(goals) > 0 {
		onTrack := 0
		for _, g := range goals {
			if g.TargetAmount > 0 && float64(g.CurrentAmount)/float64(g.TargetAmount) >= 0.5 {
				onTrack++
			}
		}
		status := "success"
		if onTrack < len(goals)/2 {
			status = "warning"
		}
		insights = append(insights, InsightItem{
			Type:        "goal_progress",
			Title:       fmt.Sprintf("%d of %d goals on track", onTrack, len(goals)),
			Description: "Goals with at least 50% progress toward target amount.",
			Status:      status,
		})
	}

	usage, err := uc.budgetRepo.GetUsage(userID, month, year)
	if err != nil {
		log.Printf("insights: GetInsights budget userID=%d: %v", userID, err)
		return nil, err
	}
	for _, u := range usage {
		if u.Status == "EXCEEDED" {
			insights = append(insights, InsightItem{
				Type:        "budget_exceeded",
				Title:       fmt.Sprintf("%s budget exceeded", u.Category),
				Description: fmt.Sprintf("You spent %.0f%% over your %s budget this month.", u.Percentage-100, u.Category),
				Status:      "warning",
			})
		} else if u.Status == "WARNING" {
			insights = append(insights, InsightItem{
				Type:        "budget_warning",
				Title:       fmt.Sprintf("%s budget near limit", u.Category),
				Description: fmt.Sprintf("You have used %.0f%% of your %s budget.", u.Percentage, u.Category),
				Status:      "warning",
			})
		}
		if len(insights) >= 5 {
			break
		}
	}

	summary, err := uc.txRepo.GetMonthlySummary(userID, 2)
	if err != nil {
		log.Printf("insights: GetInsights summary userID=%d: %v", userID, err)
		return nil, err
	}
	if len(summary) == 2 {
		cur := summary[1]
		prev := summary[0]
		if prev.Income > 0 {
			incomeChange := (cur.Income - prev.Income) / prev.Income * 100
			if incomeChange > 10 {
				insights = append(insights, InsightItem{
					Type:        "income_increase",
					Title:       "Income increased",
					Description: fmt.Sprintf("Your income is up %.1f%% compared to last month.", incomeChange),
					Status:      "success",
				})
			} else if incomeChange < -10 {
				insights = append(insights, InsightItem{
					Type:        "income_decrease",
					Title:       "Income decreased",
					Description: fmt.Sprintf("Your income dropped %.1f%% compared to last month.", -incomeChange),
					Status:      "warning",
				})
			}
		}
	}

	if len(insights) == 0 {
		insights = append(insights, InsightItem{
			Type:        "no_data",
			Title:       "No insights yet",
			Description: "Add transactions and budgets to get personalized insights.",
			Status:      "info",
		})
	}

	return &InsightsResponse{
		Insights:    insights,
		Period:      InsightsPeriod{Month: month, Year: year},
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// --- Recommendations ---

type RecommendationItem struct {
	Priority        string `json:"priority"`
	Category        string `json:"category"`
	Title           string `json:"title"`
	Action          string `json:"action"`
	PotentialImpact string `json:"potential_impact"`
}

type RecommendationsResponse struct {
	Recommendations []RecommendationItem `json:"recommendations"`
	GeneratedAt     time.Time            `json:"generated_at"`
}

func (uc *InsightsUseCase) GetRecommendations(userID int) (*RecommendationsResponse, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	var recs []RecommendationItem

	summary, err := uc.txRepo.GetMonthlySummary(userID, 1)
	if err != nil {
		log.Printf("insights: GetRecommendations summary userID=%d: %v", userID, err)
		return nil, err
	}
	if len(summary) > 0 {
		s := summary[len(summary)-1]
		if s.Income > 0 {
			rate := (s.Income - s.Expense) / s.Income
			if rate < 0.10 {
				target := s.Income * 0.20
				additional := target - (s.Income - s.Expense)
				recs = append(recs, RecommendationItem{
					Priority: "high",
					Category: "savings",
					Title:    "Increase savings rate",
					Action: fmt.Sprintf("Your current savings rate is %.0f%%. Increasing to 20%% would save an additional Rp %.0f/month.",
						rate*100, additional),
					PotentialImpact: fmt.Sprintf("Rp %.0f/month", additional),
				})
			}
		}
	}

	usage, err := uc.budgetRepo.GetUsage(userID, month, year)
	if err != nil {
		log.Printf("insights: GetRecommendations budget userID=%d: %v", userID, err)
		return nil, err
	}
	for _, u := range usage {
		if u.Percentage >= 90 {
			recs = append(recs, RecommendationItem{
				Priority: "medium",
				Category: "budget",
				Title:    fmt.Sprintf("Review %s spending", u.Category),
				Action:   fmt.Sprintf("Your %s budget is %.0f%% used. Consider reducing non-essential spending in this category.", u.Category, u.Percentage),
				PotentialImpact: fmt.Sprintf("Rp %.0f potential savings", float64(u.Used)*0.1),
			})
			if len(recs) >= 4 {
				break
			}
		}
	}

	goals, err := uc.goalRepo.GetAll(userID, false)
	if err != nil {
		log.Printf("insights: GetRecommendations goals userID=%d: %v", userID, err)
		return nil, err
	}
	for _, g := range goals {
		if g.TargetAmount > 0 {
			progress := float64(g.CurrentAmount) / float64(g.TargetAmount)
			daysLeft := time.Until(g.Deadline).Hours() / 24
			if daysLeft > 0 && daysLeft < 60 && progress < 0.8 {
				remaining := g.TargetAmount - g.CurrentAmount
				recs = append(recs, RecommendationItem{
					Priority: "high",
					Category: "goals",
					Title:    fmt.Sprintf("Boost contributions to \"%s\"", g.Name),
					Action:   fmt.Sprintf("Deadline in %.0f days. You need Rp %d more to reach 80%% progress.", daysLeft, remaining),
					PotentialImpact: fmt.Sprintf("Rp %d to reach target", remaining),
				})
			}
		}
		if len(recs) >= 4 {
			break
		}
	}

	if len(recs) == 0 {
		recs = append(recs, RecommendationItem{
			Priority:        "low",
			Category:        "general",
			Title:           "Keep up the good work",
			Action:          "Your finances look healthy. Consider setting new savings goals.",
			PotentialImpact: "",
		})
	}

	return &RecommendationsResponse{
		Recommendations: recs,
		GeneratedAt:     time.Now().UTC(),
	}, nil
}
