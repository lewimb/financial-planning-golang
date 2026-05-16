package handler

import (
	"errors"
	"strconv"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	uc *usecase.GoalUseCase
}

func NewGoalHandler(uc *usecase.GoalUseCase) *GoalHandler {
	return &GoalHandler{uc: uc}
}

func (h *GoalHandler) GetAll(c *gin.Context) {
	userID := utils.ClaimId(c)
	active := c.DefaultQuery("active", "false")
	goals, err := h.uc.GetGoals(userID, active == "true")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": goals})
}

func (h *GoalHandler) GetByID(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid goal ID"})
		return
	}
	goal, err := h.uc.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": goal})
}

func (h *GoalHandler) GetOverview(c *gin.Context) {
	userID := utils.ClaimId(c)
	overview, err := h.uc.GetOverview(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": overview})
}

func (h *GoalHandler) GetMilestones(c *gin.Context) {
	userID := utils.ClaimId(c)
	// Milestones are included in GetOverview; expose them directly from the repo via use case
	// Use GetOverview and return Goals field as milestones
	overview, err := h.uc.GetOverview(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": overview.Goals})
}

func (h *GoalHandler) Create(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
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
	c.JSON(201, gin.H{"message": "Goal created successfully"})
}

func (h *GoalHandler) Update(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid goal id"})
		return
	}
	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.Update(id, userID, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Goal updated successfully"})
}

func (h *GoalHandler) Delete(c *gin.Context) {
	userID := utils.ClaimId(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid goal ID"})
		return
	}
	if err := h.uc.Delete(id, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Goal deleted successfully"})
}

func (h *GoalHandler) Contribute(c *gin.Context) {
	userID := utils.ClaimId(c)
	var req domain.GoalContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.GoalId <= 0 {
		c.JSON(400, gin.H{"error": "Invalid goal ID"})
		return
	}
	if err := h.uc.Contribute(req.GoalId, userID, req.Contribution); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Contribution successful"})
}
