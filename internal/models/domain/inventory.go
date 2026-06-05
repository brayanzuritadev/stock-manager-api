package domain

import "time"

// InventoryMovement records a stock change for a specific product variant.
//
//	Type "IN"  — stock received (purchase, return, adjustment).
//	Type "OUT" — stock consumed (sale, damage, adjustment).
type InventoryMovement struct {
	ID               int       `json:"id"`
	ProductVariantID int       `json:"product_variant_id"`
	ProductID        int       `json:"product_id,omitempty"`
	Type             string    `json:"type"` // "IN" | "OUT"
	Quantity         int       `json:"quantity"`
	UnitCost         *float64  `json:"unit_cost,omitempty"`
	ReferenceType    *string   `json:"reference_type,omitempty"`
	ReferenceID      *int      `json:"reference_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	// Joined display fields
	ProductName string `json:"product_name,omitempty"`
	SizeName    string `json:"size_name,omitempty"`
	ColorName   string `json:"color_name,omitempty"`
	ColorHex    string `json:"color_hex,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}
