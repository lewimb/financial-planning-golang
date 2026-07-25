package domain

import "time"

type User struct {
	ID        int
	Email     string
	Password  string
	Name      string
	FirstName *string
	LastName  *string
	Phone     *string
}

type UserResponse struct {
	ID        int        `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UserProfileResponse struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
}

type UserRepository interface {
	GetAll() ([]UserResponse, error)
	FindByEmail(email string) (*User, error)
	GetByID(id int) (*UserResponse, error)
	Create(email, hashedPassword, name string) error
	UpdateProfile(userID int, req UpdateProfileRequest) (*UserProfileResponse, error)
}
