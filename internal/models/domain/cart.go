package domain

import "time"

type Cart struct {
	ID         int        `json:"id"`
	UserID     *int       `json:"user_id,omitempty"`
	Status     string     `json:"status"`
	SharedLink *string    `json:"shared_link,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Items      []CartItem `json:"items,omitempty"`
}

type CartItem struct {
	ID               int       `json:"id"`
	CartID           int       `json:"cart_id"`
	ProductVariantID int       `json:"product_variant_id"`
	ProductID        int       `json:"product_id,omitempty"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	Discount         float64   `json:"discount"`
	CreatedAt        time.Time `json:"created_at"`
	// Joined display fields
	ProductName string   `json:"product_name,omitempty"`
	SizeName    string   `json:"size_name,omitempty"`
	ColorName   string   `json:"color_name,omitempty"`
	ColorHex    string   `json:"color_hex,omitempty"`
	ImageURLs   []string `json:"image_urls,omitempty"`
}
