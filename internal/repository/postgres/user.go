package postgres

import (
	"database/sql"
	"errors"

	"github.com/financial-planning/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]domain.UserResponse, error) {
	rows, err := r.db.Query("SELECT id,email,name,created_at,deleted_at from users WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.UserResponse{}
	for rows.Next() {
		var u domain.UserResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.DeletedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(
		"SELECT id, email, name, password FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id int) (*domain.UserResponse, error) {
	var u domain.UserResponse
	err := r.db.QueryRow(`
		SELECT id, email, name, created_at, deleted_at,
		       first_name, last_name, phone
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.DeletedAt,
		&u.FirstName, &u.LastName, &u.Phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(email, hashedPassword, name string) error {
	_, err := r.db.Exec(
		"INSERT INTO users (email, password, name) VALUES ($1, $2, $3)",
		email, hashedPassword, name,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *userRepository) UpdateProfile(userID int, req domain.UpdateProfileRequest) (*domain.UserProfileResponse, error) {
	var p domain.UserProfileResponse
	err := r.db.QueryRow(`
		UPDATE users
		SET first_name = COALESCE($1, first_name),
		    last_name  = COALESCE($2, last_name),
		    phone      = COALESCE($3, phone)
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING id, email, name, first_name, last_name, phone`,
		req.FirstName, req.LastName, req.Phone, userID,
	).Scan(&p.ID, &p.Email, &p.Name, &p.FirstName, &p.LastName, &p.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}
