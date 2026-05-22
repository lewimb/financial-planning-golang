package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	uc      *usecase.TransactionUseCase
	notifUC *usecase.NotificationUseCase
}

func NewTransactionHandler(uc *usecase.TransactionUseCase, notifUC *usecase.NotificationUseCase) *TransactionHandler {
	return &TransactionHandler{uc: uc, notifUC: notifUC}
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
	if h.notifUC != nil {
		go h.notifUC.CheckBudgetAlerts(userID)
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

func (h *TransactionHandler) GetMonthlySummary(c *gin.Context) {
	userID := utils.ClaimId(c)
	months := 6
	if m, err := strconv.Atoi(c.DefaultQuery("months", "6")); err == nil && m > 0 && m <= 24 {
		months = m
	}
	summary, err := h.uc.GetMonthlySummary(userID, months)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": summary, "months": months})
}

func (h *TransactionHandler) Export(c *gin.Context) {
	userID := utils.ClaimId(c)
	month := c.Query("month")
	year := c.Query("year")

	transactions, _, err := h.uc.GetTransactions(userID, 0, 0, year, month)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=\"transactions.csv\"")

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"ID", "Date", "Type", "Category", "Amount", "Description", "IsRecurring", "RecurrenceInterval"})

	for _, t := range transactions {
		_ = w.Write([]string{
			strconv.Itoa(t.ID),
			t.Date.Format("2006-01-02"),
			t.Type,
			t.Category,
			fmt.Sprintf("%.2f", t.Amount),
			t.Description,
			strconv.FormatBool(t.IsRecurring),
			t.RecurrenceInterval,
		})
	}
	w.Flush()
}

func (h *TransactionHandler) Import(c *gin.Context) {
	userID := utils.ClaimId(c)
	var items []domain.ImportTransactionRequest
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	if len(items) == 0 {
		c.JSON(400, gin.H{"error": "No transactions provided"})
		return
	}
	if len(items) > 500 {
		c.JSON(400, gin.H{"error": "Maximum 500 transactions per import"})
		return
	}
	result, err := h.uc.BulkImport(userID, items)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if h.notifUC != nil && result.Imported > 0 {
		go h.notifUC.CheckBudgetAlerts(userID)
	}
	c.JSON(200, gin.H{"data": result, "message": "Import completed"})
}
