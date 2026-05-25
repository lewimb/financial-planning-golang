package seeders

import (
	"fmt"
	"log"

	"github.com/financial-planning/db/seeder/factories"
	"github.com/kristijorgji/goseeder"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers(s goseeder.Seeder) {
	for _, u := range factories.Users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("seed users: bcrypt %s: %v", u.Email, err)
		}
		if _, err = s.DB.Exec(
			`INSERT INTO users (email, name, password, first_name, last_name, phone)
			 VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (email) DO NOTHING`,
			u.Email, u.Name, string(hash), u.FirstName, u.LastName, u.Phone,
		); err != nil {
			log.Fatalf("seed users: insert %s: %v", u.Email, err)
		}
	}
	fmt.Printf("  seeded %d users\n", len(factories.Users))
}
