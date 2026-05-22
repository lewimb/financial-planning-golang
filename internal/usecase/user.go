package usecase

import (
	"errors"
	"fmt"
	"log"

	"github.com/financial-planning/internal/domain"
	"github.com/financial-planning/utils"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("this email is already in use")
var ErrInvalidCredentials = errors.New("invalid email or password")

type UserUseCase struct {
	repo domain.UserRepository
}

func NewUserUseCase(repo domain.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) GetAll() ([]domain.UserResponse, error) {
	users, err := uc.repo.GetAll()
	if err != nil {
		log.Printf("user: GetAll: %v", err)
	}
	return users, err
}

func (uc *UserUseCase) GetMe(userID int) (*domain.UserResponse, error) {
	user, err := uc.repo.GetByID(userID)
	if err != nil {
		log.Printf("user: GetMe userID=%d: %v", userID, err)
	}
	return user, err
}

func (uc *UserUseCase) Register(email, password, name string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Printf("user: Register bcrypt error email=%s: %v", email, err)
		return err
	}
	if err := uc.repo.Create(email, string(hash), name); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return ErrUserExists
		}
		log.Printf("user: Register repo.Create email=%s: %v", email, err)
		return err
	}
	return nil
}

func (uc *UserUseCase) Login(email, password string) (string, error) {
	user, err := uc.repo.FindByEmail(email)
	if err != nil {
		fmt.Printf("user: Login FindByEmail email=%s: %v", email, err)
		log.Printf("user: Login FindByEmail email=%s: %v", email, err)
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		fmt.Printf("user: Login bcrypt mismatch email=%s: %v\n", email, err)
		return "", ErrInvalidCredentials
	}
	return utils.GenerateJwt(user.ID, user.Name, user.Email)
}
