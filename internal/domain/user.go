package domain

import "time"

type User struct {
	ID       int
	Email    string
	Password string
	Name     string
}

type UserResponse struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Password  string     `json:"password"`
}

type UserRepository interface {
	GetAll() ([]UserResponse, error)
	FindByEmail(email string) (*User, error)
	GetByID(id int) (*UserResponse, error)
	Create(email, hashedPassword, name string) error
}
