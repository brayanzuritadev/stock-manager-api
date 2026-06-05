package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IProductVariantRepository defines the persistence contract for product variants.
type IProductVariantRepository interface {
	GetByProductID(productID int) ([]domain.ProductVariant, error)
	GetByProductIDs(productIDs []int) (map[int][]domain.ProductVariant, error)
	GetByID(id int) (*domain.ProductVariant, error)
	Create(productID, sizeID, colorID, stock int, priceOverride *float64) (*domain.ProductVariant, error)
	UpdateStock(id, stock int) error
	UpdatePriceOverride(id int, priceOverride *float64) error
	Delete(id int) error
}

type ProductVariantRepository struct{ db *sql.DB }

func NewProductVariantRepository(db *sql.DB) IProductVariantRepository {
	return &ProductVariantRepository{db: db}
}

const variantJoin = `
	SELECT pv.id, pv.product_id, p.name, pv.size_id, sz.name, pv.color_id, c.name, c.hex_code,
	       pv.stock, pv.price_override, pv.created_at, pv.updated_at
	FROM product_variants pv
	JOIN products p    ON p.id  = pv.product_id
	JOIN sizes sz      ON sz.id = pv.size_id
	JOIN colors c      ON c.id  = pv.color_id`

func scanVariant(row interface{ Scan(...any) error }) (*domain.ProductVariant, error) {
	v := &domain.ProductVariant{}
	err := row.Scan(&v.ID, &v.ProductID, &v.ProductName, &v.SizeID, &v.SizeName, &v.ColorID, &v.ColorName, &v.ColorHex,
		&v.Stock, &v.PriceOverride, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func (r *ProductVariantRepository) GetByProductID(productID int) ([]domain.ProductVariant, error) {
	rows, err := r.db.Query(variantJoin+` WHERE pv.product_id = $1 ORDER BY sz.sort_order, c.name`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.ProductVariant
	for rows.Next() {
		v := domain.ProductVariant{}
		if err := rows.Scan(&v.ID, &v.ProductID, &v.ProductName, &v.SizeID, &v.SizeName, &v.ColorID, &v.ColorName, &v.ColorHex,
			&v.Stock, &v.PriceOverride, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *ProductVariantRepository) GetByProductIDs(productIDs []int) (map[int][]domain.ProductVariant, error) {
	result := make(map[int][]domain.ProductVariant)
	if len(productIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := variantJoin + ` WHERE pv.product_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY sz.sort_order, c.name`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v domain.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.ProductName, &v.SizeID, &v.SizeName, &v.ColorID, &v.ColorName, &v.ColorHex,
			&v.Stock, &v.PriceOverride, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		result[v.ProductID] = append(result[v.ProductID], v)
	}
	return result, nil
}

func (r *ProductVariantRepository) GetByID(id int) (*domain.ProductVariant, error) {
	row := r.db.QueryRow(variantJoin+` WHERE pv.id = $1`, id)
	return scanVariant(row)
}

func (r *ProductVariantRepository) Create(productID, sizeID, colorID, stock int, priceOverride *float64) (*domain.ProductVariant, error) {
	row := r.db.QueryRow(`
		WITH ins AS (
			INSERT INTO product_variants (product_id, size_id, color_id, stock, price_override)
			VALUES ($1, $2, $3, $4, $5) RETURNING *
		)
		SELECT ins.id, ins.product_id, p.name, ins.size_id, sz.name, ins.color_id, c.name, c.hex_code,
		       ins.stock, ins.price_override, ins.created_at, ins.updated_at
		FROM ins
		JOIN products p ON p.id  = ins.product_id
		JOIN sizes sz   ON sz.id = ins.size_id
		JOIN colors c   ON c.id  = ins.color_id`,
		productID, sizeID, colorID, stock, priceOverride)
	return scanVariant(row)
}

func (r *ProductVariantRepository) UpdateStock(id, stock int) error {
	res, err := r.db.Exec(`UPDATE product_variants SET stock = $1, updated_at = NOW() WHERE id = $2`, stock, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("variant not found")
	}
	return nil
}

func (r *ProductVariantRepository) UpdatePriceOverride(id int, priceOverride *float64) error {
	res, err := r.db.Exec(`UPDATE product_variants SET price_override = $1, updated_at = NOW() WHERE id = $2`, priceOverride, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("variant not found")
	}
	return nil
}

func (r *ProductVariantRepository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM product_variants WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("variant not found")
	}
	return nil
}
