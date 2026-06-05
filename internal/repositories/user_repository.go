package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IUserRepository defines the persistence contract for users.
type IUserRepository interface {
	GetAll() ([]domain.User, error)
	GetByID(id int) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Create(name, email, passwordHash string, phone *string) (*domain.User, error)
	Update(id int, name string, phone *string) (*domain.User, error)
	SoftDelete(id int) error
}

// UserRepository is the PostgreSQL implementation of IUserRepository.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll() ([]domain.User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, password_hash, phone, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) GetByID(id int) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, phone, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, phone, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) Create(name, email, passwordHash string, phone *string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, phone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, password_hash, phone, created_at, updated_at`,
		name, email, passwordHash, phone).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepository) Update(id int, name string, phone *string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(`
		UPDATE users
		SET name = $1, phone = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id, name, email, password_hash, phone, created_at, updated_at`,
		name, phone, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Phone, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) SoftDelete(id int) error {
	res, err := r.db.Exec(`
		UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
