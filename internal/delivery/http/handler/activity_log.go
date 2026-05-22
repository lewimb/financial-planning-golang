package handler

import (
	"strconv"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type ActivityLogHandler struct {
	uc *usecase.ActivityLogUseCase
}

func NewActivityLogHandler(uc *usecase.ActivityLogUseCase) *ActivityLogHandler {
	return &ActivityLogHandler{uc: uc}
}

func (h *ActivityLogHandler) GetActivity(c *gin.Context) {
	userID := utils.ClaimId(c)

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	logs, total, err := h.uc.GetActivity(userID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": logs, "total": total})
}
