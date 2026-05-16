package usecase

import (
	"errors"

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
	return uc.repo.GetAll()
}

func (uc *UserUseCase) Register(email, password, name string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	if err := uc.repo.Create(email, string(hash), name); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func (uc *UserUseCase) Login(email, password string) (string, error) {
	user, err := uc.repo.FindByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return utils.GenerateJwt(user.ID, user.Name, user.Email)
}
