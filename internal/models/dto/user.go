package dto

// CreateUserRequest is the payload for POST /users.
type CreateUserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Phone    *string `json:"phone,omitempty"`
}

// UpdateUserRequest is the payload for PUT /users/{id}.
type UpdateUserRequest struct {
	Name  string  `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

// UserResponse is the public representation of a user (no password).
type UserResponse struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Phone *string `json:"phone,omitempty"`
}
