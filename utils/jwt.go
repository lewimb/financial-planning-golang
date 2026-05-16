package utils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwt(userId int, name string, email string) (string, error) {
	fmt.Println("Generating JWT for user:", userId, name, email)
	claims := MyCustomClaims{
		Id:    userId,
		Name:  name,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "lewimb",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
