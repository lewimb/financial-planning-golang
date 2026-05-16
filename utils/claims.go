package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func ClaimId(c *gin.Context) int {
	claims, exist := c.Get("claims")
	if !exist {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return 0
	}

	userClaims := claims.(MyCustomClaims)
	return userClaims.Id
}
