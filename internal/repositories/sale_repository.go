package repositories

import (
	"database/sql"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

type ISaleRepository interface {
	GetAll() ([]domain.Sale, error)
	GetByID(id int) (*domain.Sale, error)
	Create(userID *int, total, totalDiscount float64, channel *string, status *string, notes *string) (*domain.Sale, error)
	UpdateStatus(id int, status string) error
	AddItem(saleID, productVariantID, quantity int, unitPrice, discount float64) (*domain.SaleItem, error)
	GetItems(saleID int) ([]domain.SaleItem, error)
	Delete(id int) error
}

type SaleRepository struct{ db *sql.DB }

func NewSaleRepository(db *sql.DB) ISaleRepository { return &SaleRepository{db: db} }

func (r *SaleRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM sales WHERE id = $1`, id)
	return err
}

func (r *SaleRepository) GetAll() ([]domain.Sale, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.user_id, s.total, s.total_discount, s.channel, s.status, s.notes, s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM sale_items si WHERE si.sale_id = s.id) AS items_count,
			COALESCE((SELECT SUM(si2.unit_price * si2.quantity) FROM sale_items si2 WHERE si2.sale_id = s.id), 0) AS gross_total
		FROM sales s ORDER BY s.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sales []domain.Sale
	for rows.Next() {
		var s domain.Sale
		if err := rows.Scan(&s.ID, &s.UserID, &s.Total, &s.TotalDiscount, &s.Channel,
			&s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt, &s.ItemsCount, &s.GrossTotal); err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}
	return sales, nil
}

func (r *SaleRepository) GetByID(id int) (*domain.Sale, error) {
	s := &domain.Sale{}
	err := r.db.QueryRow(`
		SELECT id, user_id, total, total_discount, channel, status, notes, created_at, updated_at
		FROM sales WHERE id = $1`, id).
		Scan(&s.ID, &s.UserID, &s.Total, &s.TotalDiscount, &s.Channel, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *SaleRepository) Create(userID *int, total, totalDiscount float64, channel *string, status *string, notes *string) (*domain.Sale, error) {
	s := &domain.Sale{}
	err := r.db.QueryRow(`
		INSERT INTO sales (user_id, total, total_discount, channel, status, notes)
		VALUES ($1, $2, $3, $4, COALESCE($5, 'pending'), $6)
		RETURNING id, user_id, total, total_discount, channel, status, notes, created_at, updated_at`,
		userID, total, totalDiscount, channel, status, notes).
		Scan(&s.ID, &s.UserID, &s.Total, &s.TotalDiscount, &s.Channel, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *SaleRepository) UpdateStatus(id int, status string) error {
	res, err := r.db.Exec(`UPDATE sales SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("sale not found")
	}
	return nil
}

func (r *SaleRepository) AddItem(saleID, productVariantID, quantity int, unitPrice, discount float64) (*domain.SaleItem, error) {
	item := &domain.SaleItem{}
	err := r.db.QueryRow(`
		INSERT INTO sale_items (sale_id, product_variant_id, quantity, unit_price, discount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sale_id, product_variant_id, quantity, unit_price, discount, created_at`,
		saleID, productVariantID, quantity, unitPrice, discount).
		Scan(&item.ID, &item.SaleID, &item.ProductVariantID, &item.Quantity,
			&item.UnitPrice, &item.Discount, &item.CreatedAt)
	return item, err
}

func (r *SaleRepository) GetItems(saleID int) ([]domain.SaleItem, error) {
	rows, err := r.db.Query(`
		SELECT si.id, si.sale_id, si.product_variant_id,
		       COALESCE(p.name, ''), COALESCE(sz.name, ''), COALESCE(c.name, ''),
		       si.quantity, si.unit_price, si.discount, si.created_at,
		       COALESCE(img.url, '')
		FROM sale_items si
		JOIN product_variants pv ON pv.id = si.product_variant_id
		JOIN products p           ON p.id  = pv.product_id
		JOIN sizes sz             ON sz.id = pv.size_id
		JOIN colors c             ON c.id  = pv.color_id
		LEFT JOIN LATERAL (
			SELECT url FROM product_images
			WHERE product_id = pv.product_id
			  AND (color_id = pv.color_id OR color_id IS NULL)
			ORDER BY CASE WHEN color_id = pv.color_id THEN 0 ELSE 1 END, sort_order
			LIMIT 1
		) img ON true
		WHERE si.sale_id = $1 ORDER BY si.id`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.SaleItem
	for rows.Next() {
		var it domain.SaleItem
		if err := rows.Scan(&it.ID, &it.SaleID, &it.ProductVariantID,
			&it.ProductName, &it.SizeName, &it.ColorName,
			&it.Quantity, &it.UnitPrice, &it.Discount, &it.CreatedAt,
			&it.ImageURL); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}
