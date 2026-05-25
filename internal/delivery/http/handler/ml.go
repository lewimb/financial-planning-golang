package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type MLHandler struct {
	uc *usecase.MLUseCase
}

func NewMLHandler(uc *usecase.MLUseCase) *MLHandler {
	return &MLHandler{uc: uc}
}

func (h *MLHandler) GetAnalysis(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")

	result, err := h.uc.GetAnalysis(userID, year, month)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MLHandler) GetAnomaly(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")

	result, err := h.uc.GetAnomaly(userID, year, month)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MLHandler) GetForecast(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")

	periods := 30
	if raw := c.Query("periods"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			periods = n
		}
	}

	result, err := h.uc.GetForecast(userID, periods, year, month)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MLHandler) GetInsights(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")

	result, err := h.uc.GetInsights(userID, year, month)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MLHandler) StartForecast(c *gin.Context) {
	userID := utils.ClaimId(c)
	year := c.Query("year")
	month := c.Query("month")

	periods := 30
	if raw := c.Query("periods"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			periods = n
		}
	}

	result, err := h.uc.StartForecast(userID, periods, year, month)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusAccepted, result)
}

func (h *MLHandler) GetForecastStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	result, err := h.uc.GetForecastStatus(jobID)
	if err != nil {
		if errors.Is(err, usecase.ErrMLUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ML service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}
