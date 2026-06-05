package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// ICartRepository defines the persistence contract for carts.
type ICartRepository interface {
	GetAll() ([]domain.Cart, error)
	GetByID(id int) (*domain.Cart, error)
	GetBySharedLink(link string) (*domain.Cart, error)
	Create(userID *int) (*domain.Cart, error)
	UpdateStatus(id int, status string) (*domain.Cart, error)
	SetSharedLink(cartID int, link string) (*domain.Cart, error)
	UpdateNotesByLink(link string, notes string) (*domain.Cart, error)
	SoftDelete(id int) error
	GetItems(cartID int) ([]domain.CartItem, error)
	AddItem(cartID, productVariantID, quantity int, unitPrice, discount float64) (*domain.CartItem, error)
	UpdateItem(cartID, itemID, quantity int, discount float64) (*domain.CartItem, error)
	DeleteItem(cartID, itemID int) error
}

// CartRepository is the PostgreSQL implementation of ICartRepository.
type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) ICartRepository {
	return &CartRepository{db: db}
}

// ──────────────────────────
// Cart CRUD
// ──────────────────────────

func (r *CartRepository) GetAll() ([]domain.Cart, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, status, shared_link, notes, created_at, updated_at
		FROM carts
		WHERE deleted_at IS NULL
		ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carts []domain.Cart
	for rows.Next() {
		var c domain.Cart
		if err := rows.Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		carts = append(carts, c)
	}
	return carts, nil
}

func (r *CartRepository) GetByID(id int) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		SELECT id, user_id, status, shared_link, notes, created_at, updated_at
		FROM carts
		WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CartRepository) GetBySharedLink(link string) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		SELECT id, user_id, status, shared_link, notes, created_at, updated_at
		FROM carts
		WHERE shared_link = $1 AND deleted_at IS NULL`, link).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CartRepository) Create(userID *int) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		INSERT INTO carts (user_id, status)
		VALUES ($1, 'pending')
		RETURNING id, user_id, status, shared_link, notes, created_at, updated_at`,
		userID).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CartRepository) UpdateStatus(id int, status string) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		UPDATE carts
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, status, shared_link, notes, created_at, updated_at`,
		status, id).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CartRepository) SetSharedLink(cartID int, link string) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		UPDATE carts
		SET shared_link = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, status, shared_link, notes, created_at, updated_at`,
		link, cartID).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CartRepository) UpdateNotesByLink(link string, notes string) (*domain.Cart, error) {
	c := &domain.Cart{}
	err := r.db.QueryRow(`
		UPDATE carts
		SET notes = $1, updated_at = NOW()
		WHERE shared_link = $2 AND deleted_at IS NULL AND status = 'pending'
		RETURNING id, user_id, status, shared_link, notes, created_at, updated_at`,
		notes, link).
		Scan(&c.ID, &c.UserID, &c.Status, &c.SharedLink, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CartRepository) SoftDelete(id int) error {
	res, err := r.db.Exec(`
		UPDATE carts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cart not found")
	}
	return nil
}

// ──────────────────────────
// CartItem CRUD
// ──────────────────────────

func (r *CartRepository) GetItems(cartID int) ([]domain.CartItem, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.cart_id, ci.product_variant_id, p.id,
		       COALESCE(p.name, ''), COALESCE(sz.name, ''), COALESCE(c.name, ''), COALESCE(c.hex_code, ''),
		       ci.quantity, ci.unit_price, ci.discount, ci.created_at,
		       COALESCE(
		           (SELECT json_agg(pi.url ORDER BY pi.sort_order) FROM product_images pi WHERE pi.product_id = p.id AND pi.color_id = pv.color_id),
		           (SELECT json_agg(pi.url ORDER BY pi.sort_order) FROM product_images pi WHERE pi.product_id = p.id),
		           '[]'::json
		       ) AS image_urls
		FROM cart_items ci
		JOIN product_variants pv ON pv.id = ci.product_variant_id
		JOIN products p           ON p.id  = pv.product_id
		JOIN sizes sz             ON sz.id = pv.size_id
		JOIN colors c             ON c.id  = pv.color_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem
	for rows.Next() {
		var it domain.CartItem
		var imagesJSON json.RawMessage
		if err := rows.Scan(&it.ID, &it.CartID, &it.ProductVariantID, &it.ProductID,
			&it.ProductName, &it.SizeName, &it.ColorName, &it.ColorHex,
			&it.Quantity, &it.UnitPrice, &it.Discount, &it.CreatedAt, &imagesJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(imagesJSON, &it.ImageURLs)
		items = append(items, it)
	}
	return items, nil
}

func (r *CartRepository) AddItem(cartID, productVariantID, quantity int, unitPrice, discount float64) (*domain.CartItem, error) {
	it := &domain.CartItem{}
	err := r.db.QueryRow(`
		INSERT INTO cart_items (cart_id, product_variant_id, quantity, unit_price, discount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, cart_id, product_variant_id, quantity, unit_price, discount, created_at`,
		cartID, productVariantID, quantity, unitPrice, discount).
		Scan(&it.ID, &it.CartID, &it.ProductVariantID, &it.Quantity,
			&it.UnitPrice, &it.Discount, &it.CreatedAt)
	return it, err
}

func (r *CartRepository) UpdateItem(cartID, itemID, quantity int, discount float64) (*domain.CartItem, error) {
	it := &domain.CartItem{}
	err := r.db.QueryRow(`
		UPDATE cart_items
		SET quantity = $1, discount = $2
		WHERE id = $3 AND cart_id = $4
		RETURNING id, cart_id, product_variant_id, quantity, unit_price, discount, created_at`,
		quantity, discount, itemID, cartID).
		Scan(&it.ID, &it.CartID, &it.ProductVariantID, &it.Quantity,
			&it.UnitPrice, &it.Discount, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return it, err
}

func (r *CartRepository) DeleteItem(cartID, itemID int) error {
	res, err := r.db.Exec(`
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2`, itemID, cartID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}
