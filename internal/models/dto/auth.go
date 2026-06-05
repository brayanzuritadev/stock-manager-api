package dto

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the payload for POST /auth/register.
type RegisterRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Phone    *string `json:"phone,omitempty"`
}

// AuthResponse contains the JWT token returned after login/register.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
