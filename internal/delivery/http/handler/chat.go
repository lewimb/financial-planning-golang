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

func (h *ChatHandler) GetHistory(c *gin.Context) {
	userID := utils.ClaimId(c)
	logs, err := h.uc.GetHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *ChatHandler) ClearHistory(c *gin.Context) {
	userID := utils.ClaimId(c)
	if err := h.uc.ClearHistory(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "chat history cleared"})
}

func (h *ChatHandler) AskStream(c *gin.Context) {
	userID := utils.ClaimId(c)

	var req domain.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	streamErr := h.uc.AskStream(c.Request.Context(), userID, req.Message, func(chunk string) {
		c.SSEvent("message", gin.H{"content": chunk})
		c.Writer.Flush()
	})
	if streamErr != nil {
		if errors.Is(streamErr, usecase.ErrChatUnavailable) {
			c.SSEvent("error", gin.H{"error": "AI service unavailable"})
		} else {
			c.SSEvent("error", gin.H{"error": "stream failed"})
		}
		c.Writer.Flush()
		return
	}

	c.SSEvent("done", gin.H{})
	c.Writer.Flush()
}
