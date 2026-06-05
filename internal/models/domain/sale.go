package domain

import "time"

type Sale struct {
	ID            int        `json:"id"`
	UserID        *int       `json:"user_id,omitempty"`
	Total         float64    `json:"total"`
	TotalDiscount float64    `json:"total_discount"`
	GrossTotal    float64    `json:"gross_total"`
	Channel       *string    `json:"channel,omitempty"`
	Status        string     `json:"status"` // pending | paid | cancelled | refunded
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ItemsCount    int        `json:"items_count"`
	Items         []SaleItem `json:"items,omitempty"`
}

type SaleItem struct {
	ID               int       `json:"id"`
	SaleID           int       `json:"sale_id"`
	ProductVariantID int       `json:"product_variant_id"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	Discount         float64   `json:"discount"`
	CreatedAt        time.Time `json:"created_at"`
	// Joined display fields
	ProductName string `json:"product_name,omitempty"`
	SizeName    string `json:"size_name,omitempty"`
	ColorName   string `json:"color_name,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}
