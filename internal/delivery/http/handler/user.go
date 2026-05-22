package handler

import (
	"errors"
	"fmt"

	"github.com/financial-planning/internal/usecase"
	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	uc *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if err := h.uc.Register(req.Email, req.Password, req.Name); err != nil {
		if errors.Is(err, usecase.ErrUserExists) {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(500, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "User created successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	fmt.Println("Login request received for email:", req.Email, req.Password)

	token, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		fmt.Printf("Login failed for email %s: %v\n", req.Email, err)
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": "Login failed"})
		return
	}
	c.SetCookie("accessToken", token, 3600, "/", "*", false, false)
	c.JSON(200, gin.H{"message": "Login Successfully", "status": "200", "data": gin.H{"token": token}})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.uc.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": users, "status": "200", "message": "Get all users successfully"})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := utils.ClaimId(c)
	user, err := h.uc.GetMe(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"created_at": user.CreatedAt,
		},
	})
}
