package handler

import (
	"strconv"
	"time"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type InsightsHandler struct {
	uc *usecase.InsightsUseCase
}

func NewInsightsHandler(uc *usecase.InsightsUseCase) *InsightsHandler {
	return &InsightsHandler{uc: uc}
}

func (h *InsightsHandler) GetFinancialHealth(c *gin.Context) {
	userID := utils.ClaimId(c)
	data, err := h.uc.GetFinancialHealth(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *InsightsHandler) GetInsights(c *gin.Context) {
	userID := utils.ClaimId(c)
	now := time.Now()
	month := int(now.Month())
	year := now.Year()
	if m, err := strconv.Atoi(c.Query("month")); err == nil && m >= 1 && m <= 12 {
		month = m
	}
	if y, err := strconv.Atoi(c.Query("year")); err == nil && y > 2000 {
		year = y
	}
	data, err := h.uc.GetInsights(userID, month, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *InsightsHandler) GetRecommendations(c *gin.Context) {
	userID := utils.ClaimId(c)
	data, err := h.uc.GetRecommendations(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}
