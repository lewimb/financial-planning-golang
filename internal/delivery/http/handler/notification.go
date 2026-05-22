package handler

import (
	"errors"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	uc *usecase.NotificationUseCase
}

func NewNotificationHandler(uc *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"

	notifs, unreadCount, err := h.uc.GetNotifications(userID, unreadOnly)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"data":         notifs,
		"unread_count": unreadCount,
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid notification id"})
		return
	}
	if err := h.uc.MarkRead(id, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := utils.ClaimId(c)
	if err := h.uc.MarkAllRead(userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid notification id"})
		return
	}
	if err := h.uc.Delete(id, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "notification deleted"})
}

func (h *NotificationHandler) GetPreferences(c *gin.Context) {
	userID := utils.ClaimId(c)
	prefs, err := h.uc.GetPreferences(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": prefs})
}

func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	userID := utils.ClaimId(c)
	var prefs domain.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.uc.UpdatePreferences(userID, prefs); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "preferences updated"})
}
