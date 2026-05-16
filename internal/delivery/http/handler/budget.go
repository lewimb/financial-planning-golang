package handler

import (
	"errors"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type BudgetHandler struct {
	uc *usecase.BudgetUseCase
}

func NewBudgetHandler(uc *usecase.BudgetUseCase) *BudgetHandler {
	return &BudgetHandler{uc: uc}
}

func (h *BudgetHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	budgets, err := h.uc.GetBudgets(userID, c.Query("category"), c.Query("month"), c.Query("year"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": budgets})
}

func (h *BudgetHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid budget id"})
		return
	}
	budget, err := h.uc.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": budget})
}

func (h *BudgetHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.uc.Create(userID, req); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(409, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "budget created successfully"})
}

func (h *BudgetHandler) GetUsage(c *gin.Context) {
	userID := utils.ClaimId(c)
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid year"})
		return
	}
	monthStr := c.Query("month")
	var month int
	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid month"})
			return
		}
	}
	result, err := h.uc.GetUsage(userID, month, year)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *BudgetHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid budget id"})
		return
	}
	var req domain.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	response, err := h.uc.Update(userID, id, req.LimitAmount, req.AlertThreshold, req.Category)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": response})
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid budget id"})
		return
	}
	if err := h.uc.Delete(userID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "budget deleted successfully"})
}
