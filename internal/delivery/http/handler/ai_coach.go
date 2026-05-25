package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type AICoachHandler struct {
	insightsUC *usecase.InsightsUseCase
	profileUC  *usecase.FinancialProfileUseCase
}

func NewAICoachHandler(insightsUC *usecase.InsightsUseCase, profileUC *usecase.FinancialProfileUseCase) *AICoachHandler {
	return &AICoachHandler{insightsUC: insightsUC, profileUC: profileUC}
}

// GetContext returns all four AI Coach data sources in a single response, guaranteeing
// the chat system prompt and displayed panels share identical underlying data.
func (h *AICoachHandler) GetContext(c *gin.Context) {
	userID := utils.ClaimId(c)
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	health, err := h.insightsUC.GetFinancialHealth(userID)
	if err != nil {
		log.Printf("ai_coach: GetFinancialHealth userID=%d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	insights, err := h.insightsUC.GetInsights(userID, month, year)
	if err != nil {
		log.Printf("ai_coach: GetInsights userID=%d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	recommendations, err := h.insightsUC.GetRecommendations(userID)
	if err != nil {
		log.Printf("ai_coach: GetRecommendations userID=%d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	chatContext := ""
	if profile, err := h.profileUC.Get(userID); err == nil {
		chatContext = usecase.BuildFinancialProfileContext(profile)
	}

	c.JSON(http.StatusOK, gin.H{
		"health":          health,
		"insights":        insights,
		"recommendations": recommendations,
		"chat_context":    chatContext,
	})
}
