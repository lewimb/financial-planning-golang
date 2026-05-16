package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/financial-planning/internal/domain"
)

var ErrChatUnavailable = errors.New("AI service unavailable")

type ChatUseCase struct {
	txRepo     domain.TransactionRepository
	budgetRepo domain.BudgetRepository
	goalRepo   domain.GoalRepository
}

func NewChatUseCase(
	txRepo domain.TransactionRepository,
	budgetRepo domain.BudgetRepository,
	goalRepo domain.GoalRepository,
) *ChatUseCase {
	return &ChatUseCase{txRepo: txRepo, budgetRepo: budgetRepo, goalRepo: goalRepo}
}

func (uc *ChatUseCase) Ask(userID int, message string) (string, error) {
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

	context := fmt.Sprintf(`You are a helpful financial assistant. Answer the user's question based on their financial data below.
Respond in the same language the user writes in (Indonesian or English).
Be concise and actionable.

Financial Data (current month: %s %d):
- Monthly income: %.0f
- Monthly expense: %.0f
- Net savings (all-time): %.0f
- Budgets: %d total, %d exceeded budget limit
- Active financial goals: %d

User question: %s`,
		now.Month().String(), now.Year(),
		income, expense, netSavings,
		len(budgetUsage), exceeded,
		len(activeGoals),
		message,
	)

	return callGemini(context)
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func callGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", ErrChatUnavailable
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s",
		apiKey,
	)

	body := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", ErrChatUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API error: status %d", resp.StatusCode)
	}

	var gemResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return "", err
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return "", ErrChatUnavailable
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}
