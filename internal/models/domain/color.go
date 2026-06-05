package domain

import "time"

// Color represents a product colour option (e.g. Negro, Rojo, #FF0000).
type Color struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	HexCode   *string    `json:"hex_code,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
