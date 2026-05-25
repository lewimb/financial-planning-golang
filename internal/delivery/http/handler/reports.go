package handler

import (
	"strconv"
	"time"

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

// --- Year-based endpoints ---

func (h *ReportsHandler) GetIncomeExpenseTrend(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := time.Now().Year()
	if y, err := strconv.Atoi(c.DefaultQuery("year", "")); err == nil && y > 2000 {
		year = y
	}
	data, err := h.uc.GetIncomeExpenseTrend(userID, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *ReportsHandler) GetNetworthHistory(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := time.Now().Year()
	if y, err := strconv.Atoi(c.DefaultQuery("year", "")); err == nil && y > 2000 {
		year = y
	}
	data, err := h.uc.GetNetworthHistory(userID, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *ReportsHandler) GetSavingsRateHistory(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := time.Now().Year()
	if y, err := strconv.Atoi(c.DefaultQuery("year", "")); err == nil && y > 2000 {
		year = y
	}
	data, err := h.uc.GetSavingsRateHistory(userID, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *ReportsHandler) GetMonthComparisonByDate(c *gin.Context) {
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
	data, err := h.uc.GetMonthComparisonByDate(userID, month, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

func (h *ReportsHandler) GetCategoryBreakdownDetailed(c *gin.Context) {
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
	data, err := h.uc.GetCategoryBreakdownDetailed(userID, month, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}
