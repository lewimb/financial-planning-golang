package handler

import (
	"errors"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	uc *usecase.TransactionUseCase
}

func NewTransactionHandler(uc *usecase.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

func (h *TransactionHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	month := c.Query("month")
	year := c.Query("year")
	limitParse, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limitParse = 10
	}
	offsetParse, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offsetParse = 0
	}

	transactions, total, err := h.uc.GetTransactions(userID, limitParse, offsetParse, year, month)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": transactions, "total": total})
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if err := h.uc.Create(userID, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Transaction created successfully"})
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction id"})
		return
	}
	var req domain.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.uc.Update(userID, id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transaction updated successfully"})
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction id"})
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
	c.JSON(200, gin.H{"message": "transaction deleted successfully"})
}

func (h *TransactionHandler) GetMonthlyExpenses(c *gin.Context) {
	userID := utils.ClaimId(c)
	total, err := h.uc.GetMonthlyExpenses(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"total": total, "message": "success"})
}

func (h *TransactionHandler) GetMonthlyIncome(c *gin.Context) {
	userID := utils.ClaimId(c)
	total, err := h.uc.GetMonthlyIncome(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"total": total, "message": "success"})
}
