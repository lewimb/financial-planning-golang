package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/financial-planning/internal/domain"
	"google.golang.org/genai"
)

var ErrChatUnavailable = errors.New("AI service unavailable")

// --- Gemini Client ---

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiClient{
		client: client,
		model:  "gemini-3-flash-preview",
	}, nil
}

func (g *GeminiClient) Call(ctx context.Context, prompt string) (string, error) {
	const maxRetries = 3
	backoff := time.Second

	for attempt := range maxRetries {
		result, err := g.client.Models.GenerateContent(
			ctx,
			g.model,
			genai.Text(prompt),
			nil,
		)
		if err != nil {
			if isRateLimitErr(err) {
				if attempt == maxRetries-1 {
					return "", fmt.Errorf("%w: rate limit exceeded", ErrChatUnavailable)
				}
				log.Printf("gemini: rate limited, retrying in %v (attempt %d/%d)", backoff, attempt+1, maxRetries)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					continue
				}
			}
			return "", fmt.Errorf("%w: %v", ErrChatUnavailable, err)
		}

		text := result.Text()
		if text == "" {
			return "", ErrChatUnavailable
		}
		return text, nil
	}

	return "", fmt.Errorf("%w: max retries exceeded", ErrChatUnavailable)
}

func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "quota")
}

// --- ChatUseCase ---

type ChatUseCase struct {
	txRepo      domain.TransactionRepository
	budgetRepo  domain.BudgetRepository
	goalRepo    domain.GoalRepository
	logRepo     domain.AiLogRepository
	profileRepo domain.FinancialProfileRepository
	gemini      *GeminiClient
}

func NewChatUseCase(
	txRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
	goalRepo domain.GoalRepository,
	logRepo domain.AiLogRepository,
	profileRepo domain.FinancialProfileRepository,
	gemini *GeminiClient,
) *ChatUseCase {
	return &ChatUseCase{
		txRepo:      txRepo,
		budgetRepo:  budgetRepo,
		goalRepo:    goalRepo,
		logRepo:     logRepo,
		profileRepo: profileRepo,
		gemini:      gemini,
	}
}

func (uc *ChatUseCase) Ask(ctx context.Context, userID int, message string) (string, error) {
	now := time.Now()

	income, _ := uc.txRepo.GetMonthlyIncome(userID)
	expense, _ := uc.txRepo.GetMonthlyExpenses(userID)
	netSavings, _ := uc.txRepo.GetNetSavings(userID)
	budgetUsage, _ := uc.budgetRepo.GetUsage(userID, int(now.Month()), now.Year())

	exceeded := 0
	for _, b := range budgetUsage {
		if b.Status == "EXCEEDED" {
			exceeded++
		}
	}

	activeGoals, _ := uc.goalRepo.GetAll(userID, true)

	profileSection := ""
	if profile, err := uc.profileRepo.GetByUserID(userID); err == nil {
		profile.NetAvailable = profile.MonthlyIncome - profile.FixedExpenses - profile.Debt
		profileSection = "\n" + BuildFinancialProfileContext(profile)
	}

	prompt := fmt.Sprintf(`You are a helpful financial assistant. Answer the user's question based on their financial data below.
Respond in the same language the user writes in (Indonesian or English).
Be concise and actionable.
%s
Transaction Data (current month: %s %d):
- Monthly income: %.0f
- Monthly expense: %.0f
- Net savings (all-time): %.0f
- Budgets: %d total, %d exceeded budget limit
- Active financial goals: %d

User question: %s`,
		profileSection,
		now.Month().String(), now.Year(),
		income, expense, netSavings,
		len(budgetUsage), exceeded,
		len(activeGoals),
		message,
	)

	reply, err := uc.gemini.Call(ctx, prompt)
	if err != nil {
		log.Printf("gemini: failed to get reply for user %d: %v", userID, err)
		return "", err
	}

	if saveErr := uc.logRepo.Save(userID, message, reply); saveErr != nil {
		log.Printf("ai_logs: failed to save chat log for user %d: %v", userID, saveErr)
	}

	return reply, nil
}
