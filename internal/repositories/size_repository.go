package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// ISizeRepository defines the persistence contract for global sizes.
type ISizeRepository interface {
	GetAll() ([]domain.Size, error)
	GetByID(id int) (*domain.Size, error)
	Create(name string, sortOrder int) (*domain.Size, error)
	Update(id int, name string, sortOrder int) (*domain.Size, error)
	Delete(id int) error
}

type SizeRepository struct{ db *sql.DB }

func NewSizeRepository(db *sql.DB) ISizeRepository { return &SizeRepository{db: db} }

func (r *SizeRepository) GetAll() ([]domain.Size, error) {
	rows, err := r.db.Query(`
		SELECT id, name, sort_order, created_at
		FROM sizes WHERE deleted_at IS NULL ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Size
	for rows.Next() {
		var s domain.Size
		if err := rows.Scan(&s.ID, &s.Name, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SizeRepository) GetByID(id int) (*domain.Size, error) {
	s := &domain.Size{}
	err := r.db.QueryRow(`
		SELECT id, name, sort_order, created_at
		FROM sizes WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&s.ID, &s.Name, &s.SortOrder, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SizeRepository) Create(name string, sortOrder int) (*domain.Size, error) {
	s := &domain.Size{}
	err := r.db.QueryRow(`
		INSERT INTO sizes (name, sort_order) VALUES ($1, $2)
		RETURNING id, name, sort_order, created_at`,
		name, sortOrder).
		Scan(&s.ID, &s.Name, &s.SortOrder, &s.CreatedAt)
	return s, err
}

func (r *SizeRepository) Update(id int, name string, sortOrder int) (*domain.Size, error) {
	s := &domain.Size{}
	err := r.db.QueryRow(`
		UPDATE sizes SET name = $1, sort_order = $2
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id, name, sort_order, created_at`,
		name, sortOrder, id).
		Scan(&s.ID, &s.Name, &s.SortOrder, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("size not found")
	}
	return s, err
}

func (r *SizeRepository) Delete(id int) error {
	res, err := r.db.Exec(`UPDATE sizes SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("size not found")
	}
	return nil
}
