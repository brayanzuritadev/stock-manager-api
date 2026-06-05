package domain

import "time"

// Size is a global size entry (e.g. XS, S, M, 38, 40).
type Size struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
