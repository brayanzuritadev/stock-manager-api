package domain

import "time"

// ProductVariant is a unique combination of product + size + color with its own stock.
type ProductVariant struct {
	ID            int      `json:"id"`
	ProductID     int      `json:"product_id"`
	ProductName   string   `json:"product_name,omitempty"`
	SizeID        int      `json:"size_id"`
	ColorID       int      `json:"color_id"`
	Stock         int      `json:"stock"`
	PriceOverride *float64 `json:"price_override,omitempty"`
	// Joined fields (populated when queried with JOINs)
	SizeName  string    `json:"size_name,omitempty"`
	ColorName string    `json:"color_name,omitempty"`
	ColorHex  *string   `json:"color_hex,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
