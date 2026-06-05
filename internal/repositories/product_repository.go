package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IProductRepository defines the persistence contract for products.
type IProductRepository interface {
	GetAll(onlyActive bool) ([]domain.Product, error)
	GetActiveWithStock() ([]domain.Product, error)
	GetByID(id int) (*domain.Product, error)
	GetByCategoryID(categoryID int) ([]domain.Product, error)
	Create(p domain.Product) (*domain.Product, error)
	Update(id int, p domain.Product) (*domain.Product, error)
	SoftDelete(id int) error
}

// ProductRepository is the PostgreSQL implementation of IProductRepository.
type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) IProductRepository {
	return &ProductRepository{db: db}
}

const productColumns = `id, name, description, sale_price, category_id, active, created_at, updated_at`

func scanProduct(row interface{ Scan(...any) error }) (*domain.Product, error) {
	p := &domain.Product{}
	err := row.Scan(
		&p.ID, &p.Name, &p.Description, &p.SalePrice,
		&p.CategoryID, &p.Active, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (r *ProductRepository) GetAll(onlyActive bool) ([]domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE deleted_at IS NULL`
	if onlyActive {
		query += ` AND active = true`
	}
	query += ` ORDER BY id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		p := domain.Product{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SalePrice,
			&p.CategoryID, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetByID(id int) (*domain.Product, error) {
	row := r.db.QueryRow(
		`SELECT `+productColumns+` FROM products WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanProduct(row)
}

func (r *ProductRepository) GetByCategoryID(categoryID int) ([]domain.Product, error) {
	rows, err := r.db.Query(
		`SELECT `+productColumns+` FROM products WHERE category_id = $1 AND deleted_at IS NULL ORDER BY id DESC`,
		categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		p := domain.Product{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SalePrice,
			&p.CategoryID, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) Create(p domain.Product) (*domain.Product, error) {
	row := r.db.QueryRow(`
		INSERT INTO products (name, description, sale_price, category_id, active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING `+productColumns,
		p.Name, p.Description, p.SalePrice, p.CategoryID)
	return scanProduct(row)
}

func (r *ProductRepository) Update(id int, p domain.Product) (*domain.Product, error) {
	row := r.db.QueryRow(`
		UPDATE products
		SET name = $1, description = $2, sale_price = $3,
		    category_id = $4, active = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING `+productColumns,
		p.Name, p.Description, p.SalePrice,
		p.CategoryID, p.Active, id)
	return scanProduct(row)
}

func (r *ProductRepository) SoftDelete(id int) error {
	res, err := r.db.Exec(
		`UPDATE products SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func (r *ProductRepository) GetActiveWithStock() ([]domain.Product, error) {
	query := `
        SELECT DISTINCT p.id, p.name, p.description, p.sale_price, p.category_id, p.active, p.created_at, p.updated_at
        FROM products p
        JOIN product_variants v ON v.product_id = p.id
        WHERE p.active = true AND v.stock > 0 AND p.deleted_at IS NULL
        ORDER BY p.id DESC
    `
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		p := domain.Product{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SalePrice,
			&p.CategoryID, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
