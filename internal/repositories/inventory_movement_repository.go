package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IInventoryMovementRepository defines persistence for the unified inventory_movements table.
type IInventoryMovementRepository interface {
	GetAll() ([]domain.InventoryMovement, error)
	GetByID(id int) (*domain.InventoryMovement, error)
	GetByVariantID(variantID int) ([]domain.InventoryMovement, error)
	HasMovements(variantID int) (bool, error)
	Create(variantID int, mvType string, quantity int, unitCost *float64, refType *string, refID *int) (*domain.InventoryMovement, error)
}

type InventoryMovementRepository struct{ db *sql.DB }

func NewInventoryMovementRepository(db *sql.DB) IInventoryMovementRepository {
	return &InventoryMovementRepository{db: db}
}

const movementJoin = `
	SELECT im.id, im.product_variant_id, pv.product_id, im.type, im.quantity, im.unit_cost,
	       im.reference_type, im.reference_id, im.created_at,
	       p.name, sz.name, c.name, COALESCE(c.hex_code, ''),
	       COALESCE(img.url, '')
	FROM inventory_movements im
	JOIN product_variants pv ON pv.id = im.product_variant_id
	JOIN products p           ON p.id  = pv.product_id
	JOIN sizes sz             ON sz.id = pv.size_id
	JOIN colors c             ON c.id  = pv.color_id
	LEFT JOIN LATERAL (
		SELECT url FROM product_images
		WHERE product_id = pv.product_id
		  AND (color_id = pv.color_id OR color_id IS NULL)
		ORDER BY CASE WHEN color_id = pv.color_id THEN 0 ELSE 1 END, sort_order
		LIMIT 1
	) img ON true`

func scanMovement(row interface{ Scan(...any) error }) (*domain.InventoryMovement, error) {
	m := &domain.InventoryMovement{}
	err := row.Scan(&m.ID, &m.ProductVariantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
		&m.ReferenceType, &m.ReferenceID, &m.CreatedAt,
		&m.ProductName, &m.SizeName, &m.ColorName, &m.ColorHex, &m.ImageURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

func (r *InventoryMovementRepository) GetAll() ([]domain.InventoryMovement, error) {
	rows, err := r.db.Query(movementJoin + ` ORDER BY im.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.InventoryMovement
	for rows.Next() {
		m := domain.InventoryMovement{}
		if err := rows.Scan(&m.ID, &m.ProductVariantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
			&m.ReferenceType, &m.ReferenceID, &m.CreatedAt,
			&m.ProductName, &m.SizeName, &m.ColorName, &m.ColorHex, &m.ImageURL); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *InventoryMovementRepository) GetByID(id int) (*domain.InventoryMovement, error) {
	row := r.db.QueryRow(movementJoin+` WHERE im.id = $1`, id)
	return scanMovement(row)
}

func (r *InventoryMovementRepository) GetByVariantID(variantID int) ([]domain.InventoryMovement, error) {
	rows, err := r.db.Query(movementJoin+` WHERE im.product_variant_id = $1 ORDER BY im.created_at DESC`, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.InventoryMovement
	for rows.Next() {
		m := domain.InventoryMovement{}
		if err := rows.Scan(&m.ID, &m.ProductVariantID, &m.ProductID, &m.Type, &m.Quantity, &m.UnitCost,
			&m.ReferenceType, &m.ReferenceID, &m.CreatedAt,
			&m.ProductName, &m.SizeName, &m.ColorName, &m.ColorHex, &m.ImageURL); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *InventoryMovementRepository) HasMovements(variantID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM inventory_movements WHERE product_variant_id = $1)`, variantID).Scan(&exists)
	return exists, err
}

func (r *InventoryMovementRepository) Create(variantID int, mvType string, quantity int, unitCost *float64, refType *string, refID *int) (*domain.InventoryMovement, error) {
	if mvType != "IN" && mvType != "OUT" {
		return nil, fmt.Errorf("invalid movement type: %s (must be IN or OUT)", mvType)
	}
	row := r.db.QueryRow(`
		WITH ins AS (
			INSERT INTO inventory_movements (product_variant_id, type, quantity, unit_cost, reference_type, reference_id)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING *
		)
		SELECT ins.id, ins.product_variant_id, pv.product_id, ins.type, ins.quantity, ins.unit_cost,
		       ins.reference_type, ins.reference_id, ins.created_at,
		       p.name, sz.name, c.name, COALESCE(c.hex_code, ''), COALESCE(img.url, '')
		FROM ins
		JOIN product_variants pv ON pv.id = ins.product_variant_id
		JOIN products p           ON p.id  = pv.product_id
		JOIN sizes sz             ON sz.id = pv.size_id
		JOIN colors c             ON c.id  = pv.color_id
		LEFT JOIN LATERAL (
			SELECT url FROM product_images
			WHERE product_id = pv.product_id
			  AND (color_id = pv.color_id OR color_id IS NULL)
			ORDER BY CASE WHEN color_id = pv.color_id THEN 0 ELSE 1 END, sort_order
			LIMIT 1
		) img ON true`,
		variantID, mvType, quantity, unitCost, refType, refID)
	return scanMovement(row)
}
