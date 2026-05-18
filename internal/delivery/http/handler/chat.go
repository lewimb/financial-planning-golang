package handler

import (
	"errors"
	"net/http"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	uc *usecase.ChatUseCase
}

func NewChatHandler(uc *usecase.ChatUseCase) *ChatHandler {
	return &ChatHandler{uc: uc}
}

func (h *ChatHandler) Ask(c *gin.Context) {
	userID := utils.ClaimId(c)

	var req domain.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	reply, err := h.uc.Ask(c.Request.Context(), userID, req.Message)
	if err != nil {
		if errors.Is(err, usecase.ErrChatUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, domain.ChatResponse{Reply: reply})
}
