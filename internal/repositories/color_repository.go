package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IColorRepository defines the persistence contract for colors.
type IColorRepository interface {
	GetAll() ([]domain.Color, error)
	GetByID(id int) (*domain.Color, error)
	Create(name string, hexCode *string) (*domain.Color, error)
	Update(id int, name string, hexCode *string) (*domain.Color, error)
	Delete(id int) error
}

type ColorRepository struct{ db *sql.DB }

func NewColorRepository(db *sql.DB) IColorRepository { return &ColorRepository{db: db} }

func (r *ColorRepository) GetAll() ([]domain.Color, error) {
	rows, err := r.db.Query(`
		SELECT id, name, hex_code, created_at
		FROM colors WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Color
	for rows.Next() {
		var c domain.Color
		if err := rows.Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *ColorRepository) GetByID(id int) (*domain.Color, error) {
	c := &domain.Color{}
	err := r.db.QueryRow(`
		SELECT id, name, hex_code, created_at
		FROM colors WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *ColorRepository) Create(name string, hexCode *string) (*domain.Color, error) {
	c := &domain.Color{}
	err := r.db.QueryRow(`
		INSERT INTO colors (name, hex_code) VALUES ($1, $2)
		RETURNING id, name, hex_code, created_at`,
		name, hexCode).
		Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt)
	return c, err
}

func (r *ColorRepository) Update(id int, name string, hexCode *string) (*domain.Color, error) {
	c := &domain.Color{}
	err := r.db.QueryRow(`
		UPDATE colors SET name = $1, hex_code = $2
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id, name, hex_code, created_at`,
		name, hexCode, id).
		Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("color not found")
	}
	return c, err
}

func (r *ColorRepository) Delete(id int) error {
	res, err := r.db.Exec(`UPDATE colors SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("color not found")
	}
	return nil
}
