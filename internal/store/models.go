package store

import "time"

// WordItem represents a forbidden word entry in the database.
type WordItem struct {
	ID        int64     `json:"id"`
	Word      string    `json:"word"`
	CreatedAt time.Time `json:"created_at"`
}

// TitleWordItem represents a forbidden title word entry in the database.
type TitleWordItem struct {
	ID        int64     `json:"id"`
	Word      string    `json:"word"`
	CreatedAt time.Time `json:"created_at"`
}

// WhitelistItem represents an excluded sender entry in the database.
type WhitelistItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
