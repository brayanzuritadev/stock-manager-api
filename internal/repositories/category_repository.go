package repositories

import (
	"database/sql"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// ICategoryRepository defines the persistence contract for categories.
type ICategoryRepository interface {
	GetAll() ([]domain.Category, error)
	GetByID(id int) (*domain.Category, error)
	Create(c domain.Category) (*domain.Category, error)
	Update(id int, c domain.Category) (*domain.Category, error)
	SoftDelete(id int) error
}

// CategoryRepository is the PostgreSQL implementation of ICategoryRepository.
type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) ICategoryRepository {
	return &CategoryRepository{db: db}
}

const categoryColumns = `id, name, description, created_at, updated_at`

func scanCategory(row interface{ Scan(...any) error }) (*domain.Category, error) {
	c := &domain.Category{}
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CategoryRepository) GetAll() ([]domain.Category, error) {
	rows, err := r.db.Query(
		`SELECT ` + categoryColumns + ` FROM categories WHERE deleted_at IS NULL ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		c := domain.Category{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *CategoryRepository) GetByID(id int) (*domain.Category, error) {
	row := r.db.QueryRow(
		`SELECT `+categoryColumns+` FROM categories WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanCategory(row)
}

func (r *CategoryRepository) Create(c domain.Category) (*domain.Category, error) {
	row := r.db.QueryRow(
		`INSERT INTO categories (name, description) VALUES ($1, $2)
		 RETURNING `+categoryColumns,
		c.Name, c.Description,
	)
	return scanCategory(row)
}

func (r *CategoryRepository) Update(id int, c domain.Category) (*domain.Category, error) {
	row := r.db.QueryRow(
		`UPDATE categories SET name = $1, description = $2, updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL
		 RETURNING `+categoryColumns,
		c.Name, c.Description, id,
	)
	return scanCategory(row)
}

func (r *CategoryRepository) SoftDelete(id int) error {
	res, err := r.db.Exec(
		`UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
