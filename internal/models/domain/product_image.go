package domain

import "time"

// ProductImage stores a URL for a product, optionally linked to a specific color.
type ProductImage struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	ColorID   *int      `json:"color_id,omitempty"`
	URL       string    `json:"url"`
	PublicID  string    `json:"public_id"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}
