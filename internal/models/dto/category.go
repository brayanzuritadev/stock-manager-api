package dto

// CreateCategoryRequest is the payload for POST /categories.
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// UpdateCategoryRequest is the payload for PUT /categories/{id}.
type UpdateCategoryRequest struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CategoryResponse is the public representation of a category.
type CategoryResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}
