package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/financial-planning/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"message": "Missing Authorization!", "code": 401})
			c.Abort()
			return
		}
		arr := strings.Split(authHeader, " ")
		if len(arr) != 2 {
			c.JSON(401, gin.H{"error": "authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}
		authToken := arr[1]

		claims := utils.MyCustomClaims{}
		token, err := jwt.ParseWithClaims(authToken, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_KEY")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
