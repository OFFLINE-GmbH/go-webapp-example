package entity

import (
	"time"
)

// A single quote entity.
type Quote struct {
	ID      int    `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`

	CreatedAt time.Time `json:"created_at" diff:"-"`
	UpdatedAt time.Time `json:"updated_at" diff:"-"`
}

// Primary returns the primary key of this entity.
func (s Quote) Primary() int {
	return s.ID
}

// Type returns a string representation of this entity's type.
func (s Quote) Type() Kind {
	return KindQuote
}
