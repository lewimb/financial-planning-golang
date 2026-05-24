package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwt(userId int, name string, email string) (string, error) {
	claims := MyCustomClaims{
		Id:    userId,
		Name:  name,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "lewimb",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
