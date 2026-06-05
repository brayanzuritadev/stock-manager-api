package dto

//  Sale DTOs

type CreateSaleItemRequest struct {
	ProductVariantID int     `json:"product_variant_id"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	Discount         float64 `json:"discount"`
}

type CreateSaleRequest struct {
	UserID        *int                    `json:"user_id,omitempty"`
	Channel       *string                 `json:"channel,omitempty"`
	Status        *string                 `json:"status,omitempty"`
	TotalDiscount float64                 `json:"total_discount"`
	Notes         *string                 `json:"notes,omitempty"`
	Items         []CreateSaleItemRequest `json:"items"`
}

type SaleItemResponse struct {
	ID               int     `json:"id"`
	SaleID           int     `json:"sale_id"`
	ProductVariantID int     `json:"product_variant_id"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	Discount         float64 `json:"discount"`
	ProductName      string  `json:"product_name,omitempty"`
	SizeName         string  `json:"size_name,omitempty"`
	ColorName        string  `json:"color_name,omitempty"`
	ImageURL         string  `json:"image_url,omitempty"`
}

type SaleResponse struct {
	ID            int                `json:"id"`
	UserID        *int               `json:"user_id,omitempty"`
	Total         float64            `json:"total"`
	TotalDiscount float64            `json:"total_discount"`
	Channel       *string            `json:"channel,omitempty"`
	Status        string             `json:"status"`
	CreatedAt     string             `json:"created_at"`
	Items         []SaleItemResponse `json:"items,omitempty"`
}

//  Inventory Movement DTOs

type CreateInventoryMovementRequest struct {
	ProductVariantID int      `json:"product_variant_id"`
	Type             string   `json:"type"` // "IN" | "OUT"
	Quantity         int      `json:"quantity"`
	UnitCost         *float64 `json:"unit_cost,omitempty"`
	ReferenceType    *string  `json:"reference_type,omitempty"`
	ReferenceID      *int     `json:"reference_id,omitempty"`
}

type InventoryMovementResponse struct {
	ID               int      `json:"id"`
	ProductVariantID int      `json:"product_variant_id"`
	Type             string   `json:"type"`
	Quantity         int      `json:"quantity"`
	UnitCost         *float64 `json:"unit_cost,omitempty"`
	ReferenceType    *string  `json:"reference_type,omitempty"`
	ReferenceID      *int     `json:"reference_id,omitempty"`
	CreatedAt        string   `json:"created_at"`
	ProductName      string   `json:"product_name,omitempty"`
	SizeName         string   `json:"size_name,omitempty"`
	ColorName        string   `json:"color_name,omitempty"`
	ImageURL         string   `json:"image_url,omitempty"`
}
