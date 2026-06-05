package dto

//  Product request/response DTOs

// CreateProductRequest is the payload for POST /products.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	SalePrice   float64 `json:"sale_price"`
	CategoryID  *int    `json:"category_id,omitempty"`
}

// UpdateProductRequest is the payload for PUT /products/{id}.
type UpdateProductRequest struct {
	Name        string   `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	SalePrice   *float64 `json:"sale_price,omitempty"`
	CategoryID  *int     `json:"category_id,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

// AddProductImageRequest is the payload for POST /products/{id}/images.
type AddProductImageRequest struct {
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	ColorID   *int   `json:"color_id,omitempty"`
	SortOrder int    `json:"sort_order"`
}

// ProductImageResponse is the public representation of a product image.
type ProductImageResponse struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	ColorID   *int   `json:"color_id,omitempty"`
	URL       string `json:"url"`
	PublicID  string `json:"public_id"`
	SortOrder int    `json:"sort_order"`
}

// ProductVariantResponse is the public representation of a product variant.
type ProductVariantResponse struct {
	ID            int      `json:"id"`
	ProductID     int      `json:"product_id"`
	SizeID        int      `json:"size_id"`
	SizeName      string   `json:"size_name"`
	ColorID       int      `json:"color_id"`
	ColorName     string   `json:"color_name"`
	ColorHex      *string  `json:"color_hex,omitempty"`
	Stock         int      `json:"stock"`
	PriceOverride *float64 `json:"price_override,omitempty"`
}

// CreateProductVariantRequest is the payload for POST /products/{id}/variants.
// Stock is intentionally omitted — initial stock must be registered through inventory movements.
type CreateProductVariantRequest struct {
	SizeID        int      `json:"size_id"`
	ColorID       int      `json:"color_id"`
	PriceOverride *float64 `json:"price_override,omitempty"`
}

// UpdateProductVariantRequest is the payload for PUT /products/{id}/variants/{variantId}.
// Stock is intentionally omitted — stock changes must go through inventory movements.
type UpdateProductVariantRequest struct {
	PriceOverride *float64 `json:"price_override,omitempty"`
}

// ProductResponse is the public representation of a product.
type ProductResponse struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	SalePrice   float64                  `json:"sale_price"`
	CategoryID  *int                     `json:"category_id,omitempty"`
	Active      bool                     `json:"active"`
	Images      []ProductImageResponse   `json:"images,omitempty"`
	Variants    []ProductVariantResponse `json:"variants,omitempty"`
}

//  Size DTOs

type CreateSizeRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type SizeResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

//  Color DTOs

type CreateColorRequest struct {
	Name    string  `json:"name"`
	HexCode *string `json:"hex_code,omitempty"`
}

type ColorResponse struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	HexCode *string `json:"hex_code,omitempty"`
}
