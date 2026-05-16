package handler

import (
	"errors"

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
		c.JSON(400, gin.H{"error": "message is required"})
		return
	}
	reply, err := h.uc.Ask(userID, req.Message)
	if err != nil {
		if errors.Is(err, usecase.ErrChatUnavailable) {
			c.JSON(503, gin.H{"error": "AI service unavailable"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, domain.ChatResponse{Reply: reply})
}
