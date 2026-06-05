package domain

import "time"

// Product is the base product (stock lives in ProductVariant).
type Product struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	SalePrice   float64          `json:"sale_price"`
	CategoryID  *int             `json:"category_id,omitempty"`
	Active      bool             `json:"active"`
	Images      []ProductImage   `json:"images,omitempty"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   *time.Time       `json:"deleted_at,omitempty"`
}
