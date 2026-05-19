package handler

import (
	"errors"
	"net/http"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type FinancialProfileHandler struct {
	uc *usecase.FinancialProfileUseCase
}

func NewFinancialProfileHandler(uc *usecase.FinancialProfileUseCase) *FinancialProfileHandler {
	return &FinancialProfileHandler{uc: uc}
}

func (h *FinancialProfileHandler) Upsert(c *gin.Context) {
	userID := utils.ClaimId(c)

	var req domain.UpsertFinancialProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile, err := h.uc.Upsert(userID, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// validation errors from use case are user-facing
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile saved", "data": profile})
}

func (h *FinancialProfileHandler) Get(c *gin.Context) {
	userID := utils.ClaimId(c)

	profile, err := h.uc.Get(userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": profile})
}
