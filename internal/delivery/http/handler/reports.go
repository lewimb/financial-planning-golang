package handler

import (
	"strconv"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type ReportsHandler struct {
	uc *usecase.ReportsUseCase
}

func NewReportsHandler(uc *usecase.ReportsUseCase) *ReportsHandler {
	return &ReportsHandler{uc: uc}
}

func (h *ReportsHandler) GetMonthlySummary(c *gin.Context) {
	userID := utils.ClaimId(c)
	months := 6
	if m, err := strconv.Atoi(c.DefaultQuery("months", "6")); err == nil && m > 0 && m <= 24 {
		months = m
	}
	data, err := h.uc.GetMonthlySummary(userID, months)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *ReportsHandler) GetCategoryBreakdown(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")
	data, err := h.uc.GetCategoryBreakdown(userID, year, month)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *ReportsHandler) GetSavingsRate(c *gin.Context) {
	userID := utils.ClaimId(c)
	months := 6
	if m, err := strconv.Atoi(c.DefaultQuery("months", "6")); err == nil && m > 0 && m <= 24 {
		months = m
	}
	data, err := h.uc.GetSavingsRate(userID, months)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *ReportsHandler) GetNetWorth(c *gin.Context) {
	userID := utils.ClaimId(c)
	months := 12
	if m, err := strconv.Atoi(c.DefaultQuery("months", "12")); err == nil && m > 0 && m <= 60 {
		months = m
	}
	data, err := h.uc.GetNetWorth(userID, months)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

func (h *ReportsHandler) GetMonthComparison(c *gin.Context) {
	userID := utils.ClaimId(c)
	data, err := h.uc.GetMonthComparison(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}
